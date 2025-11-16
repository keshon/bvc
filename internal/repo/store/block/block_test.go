package block_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"app/internal/fs"
	"app/internal/repo/store/block"
)

// Helper to create BlockContext with in-memory FS.
func newTestBC(t *testing.T) (*block.BlockContext, string) {
	t.Helper()
	tmpDir := t.TempDir()
	fs := fs.NewMemoryFS()
	err := fs.MkdirAll(filepath.Join(tmpDir, "blocks"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	blockCtx := block.NewBlockContext(filepath.Join(tmpDir, "blocks"), fs)

	return blockCtx, tmpDir
}

func TestWriteAndRead(t *testing.T) {
	bc, _ := newTestBC(t)

	data := []byte("hello-world-1234567890")
	src := filepath.Join(bc.BlocksDir, "src.bin")

	// Write to in-memory FS
	if err := bc.FS.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	refs, err := bc.SplitFile(src)
	if err != nil {
		t.Fatal(err)
	}

	// Write using same path inside FS
	if err := bc.Write(src, refs); err != nil {
		t.Fatal(err)
	}

	out, err := bc.Read(refs[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out, data) {
		t.Fatalf("read data mismatch")
	}
}

func TestVerifyBlock_OK(t *testing.T) {
	bc, _ := newTestBC(t)

	data := []byte("abcdef1234567890")
	src := filepath.Join(bc.BlocksDir, "src.bin")

	if err := bc.FS.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	refs, err := bc.SplitFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if err := bc.Write(src, refs); err != nil {
		t.Fatal(err)
	}

	status, err := bc.VerifyBlock(refs[0].Hash)
	if err != nil {
		t.Fatal(err)
	}
	if status != block.OK {
		t.Fatalf("expected OK, got %v", status)
	}
}

func TestVerifyBlock_Missing(t *testing.T) {
	bc, _ := newTestBC(t)

	status, err := bc.VerifyBlock("deadbeef")
	if err != nil {
		t.Fatal(err)
	}
	if status != block.Missing {
		t.Fatalf("expected Missing, got %v", status)
	}
}

func TestVerifyBlock_Damaged(t *testing.T) {
	bc, _ := newTestBC(t)

	// write garbaged block with wrong hash to objects dir
	wrongPath := "wronghash.bin"
	err := bc.FS.WriteFile(filepath.Join(bc.BlocksDir, wrongPath), []byte("XXX"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	status, err := bc.VerifyBlock("wronghash")
	if status != block.Damaged {
		t.Fatalf("expected Damaged, got %v", status)
	}

	// err may be non-nil; allowed.
	_ = err
}

func TestCleanupTemp(t *testing.T) {
	bc, _ := newTestBC(t)

	// good file (should stay)
	err := bc.FS.WriteFile(filepath.Join(bc.BlocksDir, "good.bin"), []byte("123"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	// temp files (should be removed)
	err = bc.FS.WriteFile(filepath.Join(bc.BlocksDir, "tmp-abc"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	err = bc.FS.WriteFile(filepath.Join(bc.BlocksDir, ".tmp-xyz"), []byte{}, 0o644)
	if err != nil {
		t.Fatal(err)
	}

	if err := bc.CleanupTemp(); err != nil {
		t.Fatal(err)
	}

	entries, _ := bc.FS.ReadDir(bc.BlocksDir)
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}

	if names["tmp-abc"] {
		t.Fatalf("tmp-abc should be removed")
	}
	if names[".tmp-xyz"] {
		t.Fatalf(".tmp-xyz should be removed")
	}
	if !names["good.bin"] {
		t.Fatalf("good.bin should remain")
	}
}

func TestSplitFile(t *testing.T) {
	bc, _ := newTestBC(t)

	data := bytes.Repeat([]byte("A"), 5*1024*1024) // 5 MiB
	src := filepath.Join(bc.BlocksDir, "big.bin")

	// write source file into the fake filesystem
	if err := bc.FS.WriteFile(src, data, 0o644); err != nil {
		t.Fatal(err)
	}

	blocks, err := bc.SplitFile(src)
	if err != nil {
		t.Fatal(err)
	}

	if len(blocks) == 0 {
		t.Fatalf("expected multiple blocks, got 0")
	}

	var sum int64
	for _, b := range blocks {
		sum += b.Size
	}

	if sum != int64(len(data)) {
		t.Fatalf("sum of block sizes mismatch: %d vs %d", sum, len(data))
	}
}
