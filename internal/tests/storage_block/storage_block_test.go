package block_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"app/internal/storage/block"
)

func makeTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bvc-block-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	return dir
}

// --- CleanupTemp tests --- //
func TestCleanupTemp(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	// create temp files
	tmpFiles := []string{"tmp-a", ".tmp-b", "keep.txt"}
	for _, name := range tmpFiles {
		f, _ := os.Create(filepath.Join(dir, name))
		f.Close()
	}

	bm := &block.BlockManager{Root: dir}
	if err := bm.CleanupTemp(); err != nil {
		t.Fatalf("CleanupTemp failed: %v", err)
	}

	files, _ := os.ReadDir(dir)
	for _, f := range files {
		if f.Name() == "tmp-a" || f.Name() == ".tmp-b" {
			t.Errorf("expected temp file %q to be removed", f.Name())
		}
		if f.Name() == "keep.txt" {
			continue
		}
	}
}

// --- Write & Read tests --- //
func TestWriteAndReadBlock(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	sourceFile := filepath.Join(dir, "source.dat")
	data := []byte("hello world block test")
	if err := os.WriteFile(sourceFile, data, 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	bm := &block.BlockManager{Root: dir}

	ref := block.BlockRef{
		Hash:   "myhash",
		Size:   int64(len(data)),
		Offset: 0,
	}

	if err := bm.Write(sourceFile, []block.BlockRef{ref}); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	readData, err := bm.Read("myhash")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if !bytes.Equal(readData, data) {
		t.Errorf("read data mismatch: got %q", string(readData))
	}
}

// --- VerifyBlock tests --- //
func TestVerifyBlock(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	bm := &block.BlockManager{Root: dir}
	content := []byte("verify me")
	filePath := filepath.Join(dir, "verify.dat")
	os.WriteFile(filePath, content, 0o644)

	ref := block.HashBlock(content, 0)
	blockFile := filepath.Join(dir, ref.Hash+".bin")
	os.WriteFile(blockFile, content, 0o644)

	status, err := bm.VerifyBlock(ref.Hash)
	if err != nil {
		t.Fatalf("VerifyBlock failed: %v", err)
	}
	if status != block.OK {
		t.Errorf("expected OK, got %v", status)
	}

	// wrong hash
	status, _ = bm.VerifyBlock("nonexistent")
	if status != block.Missing {
		t.Errorf("expected Missing, got %v", status)
	}

	// corrupted block
	os.WriteFile(blockFile, []byte("corrupt"), 0o644)
	status, _ = bm.VerifyBlock(ref.Hash)
	if status != block.Damaged {
		t.Errorf("expected Damaged, got %v", status)
	}
}

// --- Verify channel test --- //
func TestVerifyChannel(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	bm := &block.BlockManager{Root: dir}
	data := []byte("abc123")
	ref := block.HashBlock(data, 0)
	os.WriteFile(filepath.Join(dir, ref.Hash+".bin"), data, 0o644)

	hashes := map[string]struct{}{ref.Hash: {}}
	ch := bm.Verify(hashes, 2)

	for bc := range ch {
		if bc.Hash != ref.Hash || bc.Status != block.OK {
			t.Errorf("unexpected BlockCheck: %+v", bc)
		}
	}
}

// --- shouldSplitBlock and hashBlock --- //
func TestShouldSplitAndHashBlock(t *testing.T) {
	// shouldSplitBlock triggers
	if !block.ShouldSplitBlock(2*1024*1024, 0) {
		t.Errorf("expected split at minChunkSize")
	}
	if !block.ShouldSplitBlock(10*1024*1024, 0) {
		t.Errorf("expected split at maxChunkSize")
	}

	data := []byte("hash me")
	ref := block.HashBlock(data, 123)
	if ref.Size != int64(len(data)) || ref.Offset != 123 {
		t.Errorf("unexpected BlockRef: %+v", ref)
	}
}

// --- SplitFile --- //
func TestSplitFileSmall(t *testing.T) {
	dir := makeTempDir(t)
	defer os.RemoveAll(dir)

	filePath := filepath.Join(dir, "small.dat")
	content := []byte("small file for split")
	os.WriteFile(filePath, content, 0o644)

	bm := &block.BlockManager{Root: dir}
	blocks, err := bm.SplitFile(filePath)
	if err != nil {
		t.Fatalf("SplitFile failed: %v", err)
	}
	if len(blocks) == 0 {
		t.Errorf("expected at least 1 block, got 0")
	}

	total := int64(0)
	for _, b := range blocks {
		total += b.Size
	}
	if total != int64(len(content)) {
		t.Errorf("total block size mismatch: got %d, expected %d", total, len(content))
	}
}
