package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"bvc/internal/blockstore"
	"bvc/internal/snapshot"
	"bvc/internal/stream"
	"bvc/internal/workfs"
	"bvc/pkg/storage"
	"bvc/pkg/util"
)

const BlockSize = 4 * 1024 * 1024

type Engine struct {
	Root    string
	Blocks  *blockstore.Store
	Snaps   *snapshot.Store
	Streams *stream.Store
	Ignore  workfs.Ignore
	Head    *HeadStore
	Mode    string // "snapshot-first" or "stream-first"
}

func NewEngine(root string, blocks *blockstore.Store, snaps *snapshot.Store, streams *stream.Store, ig workfs.Ignore, headBackend storage.Storage) *Engine {
	// load HEAD
	defaultMode := "snapshot-first"
	head, err := NewHeadStore(headBackend, "HEAD").Load()
	if err == nil && head.Mode != "" {
		defaultMode = head.Mode
	}

	return &Engine{
		Root:    root,
		Blocks:  blocks,
		Snaps:   snaps,
		Streams: streams,
		Ignore:  ig,
		Head:    NewHeadStore(headBackend, "HEAD"),
		Mode:    defaultMode,
	}
}

func (e *Engine) RequireMode(expected string) error {
	if e.Mode != expected {
		return fmt.Errorf("command allowed only in %s mode, current: %s", expected, e.Mode)
	}
	return nil
}

func (e *Engine) WalkFiles(fn func(rel string, f *os.File) error) error {
	return workfs.WalkFiles(e.Root, e.Ignore, fn)
}

// ------------------------ Block operations ------------------------

// HashFileBlocks reads the file and returns list of block hashes.
func (e *Engine) HashFileBlocks(f *os.File) ([]string, error) {
	var hashes []string
	buf := make([]byte, BlockSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			h, err := e.Blocks.PutIfAbsent(bytes.NewReader(buf[:n]))
			if err != nil {
				return nil, err
			}
			hashes = append(hashes, h)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return hashes, nil
}

// ------------------------ Snapshot operations ------------------------

func (e *Engine) CreateSnapshot(name, desc string) (*snapshot.Meta, error) {
	if name == "" {
		name = "snap-" + time.Now().Format("20060102-150405")
	}
	meta := &snapshot.Meta{
		ID:          makeSnapshotID(name),
		Name:        name,
		Description: desc,
		CreatedAt:   time.Now().UTC(),
		Files:       map[string][]string{},
	}

	type fileEntry struct {
		rel string
		p   string
	}
	var files []fileEntry

	err := workfs.WalkFiles(e.Root, e.Ignore, func(rel string, f *os.File) error {
		abs := f.Name()
		rel = util.Normalize(rel)
		files = append(files, fileEntry{rel: rel, p: abs})
		return nil
	})
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex

	err = util.Parallel(files, 8, func(ctx context.Context, fe fileEntry) error {
		f, err := os.Open(fe.p)
		if err != nil {
			return err
		}
		defer f.Close()

		var blocks []string
		buf := make([]byte, BlockSize)

		for {
			n, err := f.Read(buf)
			if n > 0 {
				h, err := e.Blocks.PutIfAbsent(bytes.NewReader(buf[:n]))
				if err != nil {
					return err
				}
				blocks = append(blocks, h)
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}

		mu.Lock()
		meta.Files[fe.rel] = blocks
		mu.Unlock()

		return nil
	})
	if err != nil {
		return nil, err
	}

	return meta, e.Snaps.Save(meta)
}

func (e *Engine) CheckoutSnapshot(id string, clean bool) error {
	meta, err := e.Snaps.Load(id)
	if err != nil {
		return err
	}

	if clean {
		if err := e.CleanupWorkdir(); err != nil {
			return err
		}
	}

	type fileEntry struct {
		rel    string
		blocks []string
	}

	var items []fileEntry
	for rel, blocks := range meta.Files {
		items = append(items, fileEntry{rel: rel, blocks: blocks})
	}

	err = util.Parallel(items, 8, func(ctx context.Context, fe fileEntry) error {
		return workfs.RestoreFile(e.Root, fe.rel, fe.blocks, func(h string, w io.Writer) error {
			return e.Blocks.Get(h, w, true)
		})
	})
	return err
}

func (e *Engine) CheckoutSnapshotPartial(id string) error {
	if e.Mode == "stream-first" {
		return fmt.Errorf("partial snapshot checkout not allowed in stream-first mode")
	}
	return e.CheckoutSnapshot(id, false)
}

func (e *Engine) MergeSnapshots(aID, bID, newName string) (*snapshot.Meta, error) {
	a, err := e.Snaps.Load(aID)
	if err != nil {
		return nil, err
	}
	b, err := e.Snaps.Load(bID)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}

	for rel, aBlocks := range a.Files {
		rel = util.Normalize(rel)
		if bBlocks, ok := b.Files[rel]; ok {
			if equalSlices(aBlocks, bBlocks) {
				result[rel] = aBlocks
			} else {
				result[rel] = aBlocks
				conflictRel := rel + ".bvc.conflict"
				if err := workfs.RestoreFile(e.Root, conflictRel, bBlocks, func(h string, w io.Writer) error {
					return e.Blocks.Get(h, w, true)
				}); err != nil {
					return nil, fmt.Errorf("writing conflict %s: %w", conflictRel, err)
				}
			}
		} else {
			result[rel] = aBlocks
		}
	}

	for rel, bBlocks := range b.Files {
		rel = util.Normalize(rel)
		if _, ok := result[rel]; !ok {
			result[rel] = bBlocks
		}
	}

	meta := &snapshot.Meta{
		ID:          makeSnapshotID(newName),
		Name:        newName,
		Description: fmt.Sprintf("merge %s + %s", aID, bID),
		CreatedAt:   time.Now().UTC(),
		Files:       result,
	}

	if err := e.Snaps.Save(meta); err != nil {
		return nil, err
	}
	return meta, nil
}

// ------------------------ Stream operations ------------------------

func (e *Engine) StreamCreate(name string) error {
	return e.Streams.Create(name)
}

func (e *Engine) StreamAdd(name, snap string) error {
	meta, err := e.Streams.Load(name)
	if err != nil {
		return err
	}
	for _, s := range meta.Snapshots {
		if s == snap {
			return nil
		}
	}
	meta.Snapshots = append(meta.Snapshots, snap)
	return e.Streams.Save(meta)
}

func (e *Engine) StreamSnapshots(name string) ([]string, error) {
	meta, err := e.Streams.Load(name)
	if err != nil {
		return nil, err
	}
	return meta.Snapshots, nil
}

func (e *Engine) StreamList() ([]string, error) {
	return e.Streams.List()
}

func (e *Engine) StreamCheckout(name string) error {
	meta, err := e.Streams.Load(name)
	if err != nil {
		return err
	}
	if len(meta.Snapshots) == 0 {
		if e.Mode == "stream-first" {
			return e.CleanupWorkdir()
		}
		return nil
	}

	if e.Mode == "snapshot-first" {
		for _, s := range meta.Snapshots {
			if err := e.CheckoutSnapshotPartial(s); err != nil {
				return err
			}
		}
		return nil
	}

	latest := meta.Snapshots[len(meta.Snapshots)-1]
	return e.CheckoutSnapshot(latest, true)
}

// StreamClone creates a new stream as exact copy of another.
func (e *Engine) StreamClone(src, dst string) error {
	// load source
	srcMeta, err := e.Streams.Load(src)
	if err != nil {
		return fmt.Errorf("source stream '%s' not found", src)
	}

	// ensure destination does not exist
	if _, err := e.Streams.Load(dst); err == nil {
		return fmt.Errorf("destination stream '%s' already exists", dst)
	}

	// copy
	dstMeta := &stream.Meta{
		Name:      dst,
		Snapshots: append([]string(nil), srcMeta.Snapshots...),
	}

	return e.Streams.Save(dstMeta)
}

// StreamRemove deletes a stream by name.
// If HEAD currently points to this stream, it clears HEAD.
func (e *Engine) StreamRemove(name string) error {
	// ensure exists
	if _, err := e.Streams.Load(name); err != nil {
		return fmt.Errorf("stream '%s' not found", name)
	}

	// remove
	if err := e.Streams.Delete(name); err != nil {
		return err
	}

	// if HEAD points to that stream, clear it
	head, _ := e.Head.Load()
	if head.Mode == "stream-first" && head.Ref == name {
		_ = e.Head.Clear()
	}

	return nil
}

// ------------------------ Cleanup & Prune ------------------------

func (e *Engine) Prune(dry bool) ([]string, error) {
	live := map[string]struct{}{}

	ids, err := e.Snaps.List()
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		m, err := e.Snaps.Load(id)
		if err != nil {
			continue
		}
		for _, blks := range m.Files {
			for _, h := range blks {
				live[h] = struct{}{}
			}
		}
	}

	all, err := e.Blocks.Backend.List("")
	if err != nil {
		return nil, err
	}

	var toDelete []string
	for _, key := range all {
		if _, ok := live[key]; !ok {
			toDelete = append(toDelete, key)
		}
	}

	if dry {
		return toDelete, nil
	}

	err = util.Parallel(toDelete, 8, func(ctx context.Context, key string) error {
		return e.Blocks.Backend.Delete(key)
	})
	return toDelete, err
}

func (e *Engine) CleanupWorkdir() error {
	ig := e.Ignore
	var files, dirs []string

	err := filepath.Walk(e.Root, func(pathname string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relRaw, _ := filepath.Rel(e.Root, pathname)
		rel := util.Normalize(relRaw)
		if rel == workfs.DefaultRepoDir || ig.Match(rel, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() {
			dirs = append(dirs, pathname)
			return nil
		}
		if info.Mode().IsRegular() {
			files = append(files, pathname)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, f := range files {
		_ = os.Remove(f)
	}
	for i := len(dirs) - 1; i >= 0; i-- {
		d := dirs[i]
		if entries, _ := os.ReadDir(d); len(entries) == 0 {
			_ = os.Remove(d)
		}
	}
	return nil
}

// ------------------------ Head ------------------------

func (e *Engine) HeadSetSnapshot(id string) error {
	if _, err := e.Snaps.Load(id); err != nil {
		return err
	}
	return e.Head.Save("snapshot-first", id)
}

func (e *Engine) HeadSetStream(name string) error {
	if _, err := e.Streams.Load(name); err != nil {
		return err
	}
	return e.Head.Save("stream-first", name)
}

func (e *Engine) GetHeadOrArg(args []string, expectedMode string) (string, error) {
	// если есть аргумент, используем его
	if len(args) > 0 {
		return args[0], nil
	}

	// иначе берем из HEAD
	head, err := e.Head.Load()
	if err != nil {
		return "", fmt.Errorf("load head: %w", err)
	}
	if head.Mode != expectedMode {
		return "", fmt.Errorf("HEAD is not pointing to mode %s", expectedMode)
	}
	if head.Ref == "" {
		return "", fmt.Errorf("HEAD is empty")
	}
	return head.Ref, nil
}

func (e *Engine) CheckoutHead() error {
	m, err := e.Head.Load()
	if err != nil {
		return err
	}
	if m.Ref == "" {
		return nil
	}
	switch m.Mode {
	case "snapshot-first":
		return e.CheckoutSnapshot(m.Ref, true)
	case "stream-first":
		return e.StreamCheckout(m.Ref)
	default:
		return fmt.Errorf("invalid head mode")
	}
}

// ------------------------ Helpers ------------------------

func makeSnapshotID(name string) string {
	h := sha256.New()
	h.Write([]byte(name))
	h.Write([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
