package block

import (
	"app/internal/fsio"
	"app/internal/util"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

// Write stores all blocks for a given file.
func (bc *BlockContext) Write(filePath string, blocks []BlockRef) error {
	if err := fsio.MkdirAll(bc.Root, 0o755); err != nil {
		return fmt.Errorf("create objects dir: %w", err)
	}
	workers := util.WorkerCount()
	return util.Parallel(blocks, workers, func(b BlockRef) error {
		return bc.writeBlockAtomic(filePath, b)
	})
}

// Read retrieves a block by its hash.
func (bc *BlockContext) Read(hash string) ([]byte, error) {
	path := filepath.Join(bc.Root, hash+".bin")
	data, err := fsio.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read block %q: %w", hash, err)
	}
	return data, nil
}

func (bc *BlockContext) writeBlockAtomic(filePath string, block BlockRef) error {
	dst := filepath.Join(bc.Root, block.Hash+".bin")

	if fi, err := fsio.StatFile(dst); err == nil && fi.Size() == block.Size {
		return nil // already exists
	}

	if err := fsio.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure dir for %q: %w", dst, err)
	}

	src, err := fsio.Open(filePath)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", filePath, err)
	}
	defer src.Close()

	tmp, err := fsio.CreateTempFile(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %q: %w", filepath.Dir(dst), err)
	}
	tmpPath := tmp.Name()
	defer fsio.Remove(tmpPath)

	if _, err := src.Seek(block.Offset, io.SeekStart); err != nil {
		tmp.Close()
		return fmt.Errorf("seek to offset %d in %q: %w", block.Offset, filePath, err)
	}
	if _, err := io.CopyN(tmp, src, block.Size); err != nil && !errors.Is(err, io.EOF) {
		tmp.Close()
		return fmt.Errorf("copy block %q: %w", block.Hash, err)
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync temp file %q: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file %q: %w", tmpPath, err)
	}

	if err := fsio.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename temp file %q to %q: %w", tmpPath, dst, err)
	}

	return nil
}
