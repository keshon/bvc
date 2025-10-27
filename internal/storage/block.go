package storage

import (
	"app/internal/config"
	"app/internal/core"
	"app/internal/util"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/zeebo/xxh3"
)

const (
	minChunkSize = 2 * 1024 * 1024 // 2 MiB
	maxChunkSize = 8 * 1024 * 1024 // 8 MiB
	rollMod      = 4096
)

// StoreBlocks saves file blocks concurrently.
func StoreBlocks(srcPath string, blocks []BlockRef) error {
	return parallel(blocks, WorkerCount(), func(b BlockRef) error {
		return storeBlock(srcPath, b)
	})
}

// storeBlock writes a single block atomically
func storeBlock(srcPath string, block BlockRef) error {
	dstPath := filepath.Join(config.ObjectsDir, block.Hash+".bin")

	if fi, err := os.Stat(dstPath); err == nil && fi.Size() == block.Size {
		return nil
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	data := make([]byte, block.Size)
	if _, err := f.ReadAt(data, block.Offset); err != nil {
		return fmt.Errorf("read block: %w", err)
	}

	dir := filepath.Dir(dstPath)
	tmpFile, err := os.CreateTemp(dir, "tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	n, err := tmpFile.Write(data)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp block: %w", err)
	}
	if n != len(data) {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("short write: %d/%d bytes", n, len(data))
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("sync temp block: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Final integrity check before rename
	if fi, err := os.Stat(tmpPath); err != nil || fi.Size() != block.Size {
		os.Remove(tmpPath)
		return fmt.Errorf("temp file incomplete for %s", block.Hash)
	}

	return os.Rename(tmpPath, dstPath)
}

// readBlock loads a single block by its hash.
func readBlock(hash string) ([]byte, error) {
	path := filepath.Join(config.ObjectsDir, hash+".bin")
	return os.ReadFile(path)
}

// VerifyBlock checks whether a block exists and matches its hash.
func VerifyBlock(hash string) (BlockStatus, error) {
	path := filepath.Join(config.ObjectsDir, hash+".bin")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return BlockMissing, nil
		}
		return BlockMissing, err
	}

	computed := fmt.Sprintf("%x", xxh3.Hash128(data).Bytes())
	if computed == hash {
		return BlockOK, nil
	}
	return BlockDamaged, nil
}

// VerifyBlocks concurrently verifies many blocks and streams results.
func VerifyBlocks(blocks map[string]struct{}, workers int) <-chan BlockCheck {
	out := make(chan BlockCheck, 128)

	go func() {
		defer close(out)
		tasks := make(chan string, len(blocks))
		for hash := range blocks {
			tasks <- hash
		}
		close(tasks)

		var wg sync.WaitGroup
		if workers <= 0 {
			workers = WorkerCount()
		}

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for hash := range tasks {
					status, _ := VerifyBlock(hash)
					out <- BlockCheck{Hash: hash, Status: status}
				}
			}()
		}

		wg.Wait()
	}()

	return out
}

// SplitFileIntoBlocks divides a file into content-defined chunks.
func SplitFileIntoBlocks(srcPath string) ([]BlockRef, error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var (
		blocks []BlockRef
		chunk  []byte
		buf    = make([]byte, maxChunkSize)
		offset int64
		rh     uint32
	)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			for i := 0; i < n; i++ {
				b := buf[i]
				chunk = append(chunk, b)
				rh = (rh<<1 + uint32(b)) & 0xFFFFFFFF
				if shouldSplit(len(chunk), rh) {
					blocks = append(blocks, hashBlock(chunk, offset))
					offset += int64(len(chunk))
					chunk = chunk[:0]
				}
			}
		}
		if err != nil {
			if err == os.ErrClosed || err.Error() == "EOF" {
				break
			}
			if err.Error() == "EOF" {
				break
			}
			if err != nil {
				break
			}
		}
		if n == 0 {
			break
		}
	}

	if len(chunk) > 0 {
		blocks = append(blocks, hashBlock(chunk, offset))
	}
	return blocks, nil
}

// shouldSplit decides when to end a chunk.
func shouldSplit(size int, rh uint32) bool {
	return (size >= minChunkSize && rh%rollMod == 0) || size >= maxChunkSize
}

// hashBlock computes a hash and creates a BlockRef.
func hashBlock(data []byte, offset int64) BlockRef {
	hash := xxh3.Hash128(data).Bytes()
	return BlockRef{
		Hash:   fmt.Sprintf("%x", hash),
		Size:   int64(len(data)),
		Offset: offset,
	}
}

// countAllBlocks returns the total number of unique blocks in the repository.
func CountAllBlocks() (int, error) {
	branches, err := core.Branches()
	if err != nil {
		return 0, err
	}

	blockHashes := map[string]struct{}{}

	for _, branch := range branches {
		commitID, err := core.LastCommitID(branch.Name)
		if err != nil {
			return 0, err
		}
		if commitID == "" {
			continue
		}

		var commit core.Commit
		if err := util.ReadJSON(filepath.Join(config.CommitsDir, commitID+".json"), &commit); err != nil {
			continue
		}

		var fs Fileset
		if err := util.ReadJSON(filepath.Join(config.FilesetsDir, commit.FilesetID+".json"), &fs); err != nil {
			continue
		}

		for _, f := range fs.Files {
			for _, b := range f.Blocks {
				blockHashes[b.Hash] = struct{}{}
			}
		}
	}

	return len(blockHashes), nil
}

// CleanupTmpBlocks removes orphaned or zero-size temp files in the objects dir.
func CleanupTmpBlocks() error {
	entries, err := os.ReadDir(config.ObjectsDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() && len(e.Name()) > 4 && e.Name()[:4] == "tmp-" {
			path := filepath.Join(config.ObjectsDir, e.Name())
			info, err := os.Stat(path)
			if err != nil || info.Size() == 0 {
				_ = os.Remove(path)
			}
		}
	}
	return nil
}
