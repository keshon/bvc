# internal/store

This package implements the **core store layer** for BVC.  
It abstracts and manages everything under `.bvc/`, including blocks, files, snapshots, and branches.

---

## Overview

```

store/
├── manager.go       # High-level orchestrator (glues blocks, files, snapshots)
├── block/           # Handles low-level deduplicated chunks (.bvc/objects)
├── file/            # Maps working files to blocks and manages index
└── snapshot/        # Handles immutable filesets (commits / restore)

````

---

## 1. `Manager`
**File:** `store/manager.go`  
Entry point for store ops.

- `InitAt(root)` → Creates `.bvc` subdirs and returns a ready `Manager`
- `NewManager(root)` → Attaches block, file, and snapshot subsystems  
  (no mkdir, used for opening existing repo)

Each subsystem is accessible via:
```go
m.Blocks      // *block.BlockManager
m.Files       // *file.FileManager
m.Snapshots   // *snapshot.SnapshotManager
```

---

## 2. `block` — Content Storage

**Path:** `store/block/`

Handles all chunked binary store in `.bvc/objects`.

* **SplitFile(path)** → splits file into deduplicated chunks (xxh3)
* **Write(filePath, blocks)** → saves chunks atomically to disk
* **Read(hash)** → loads block bytes
* **VerifyBlock(hash)** → checks block integrity
* **CleanupTemp()** → removes temp/orphaned files

---

## 3. `file` — Working Tree Abstraction

**Path:** `store/file/`

Manages mapping between real files and content blocks.

* **CreateEntry(path)** → splits one file into `[]BlockRef`
* **CreateEntries(paths)** → bulk version (parallel)
* **Write(entry)** → writes its blocks via BlockManager
* **Restore(entries)** → rebuilds working files from stored blocks
* **StageFiles(entries)** → writes index (`index.json`)
* **GetIndexFiles()** → loads staged entries

---

## 4. `snapshot` — Filesets & Commits

**Path:** `store/snapshot/`

Immutable snapshots of tracked files.

* **Create(entries)** → builds `Fileset` from staged entries
* **Save(fs)** / **Load(id)** → persist or load `.json` snapshot
* **List()** → list all snapshots in `.bvc/filesets`
* **WriteAndSave(fs)** → store file data + save snapshot atomically
* **CreateCurrent()** → snapshot current working tree directly

Each `Fileset`:

```go
type Fileset struct {
    ID    string       // xxh3 hash of all block hashes
    Files []file.Entry // file→block mapping
}
```

---

## Quick Flow

```go
// 1. Init repo
m, _ := store.InitAt(".bvc")

// 2. Create entries from working files
entries, _ := m.Files.CreateAllEntries()

// 3. Build snapshot
fs, _ := m.Snapshots.Create(entries)

// 4. Save snapshot
_ = m.Snapshots.Save(fs)

// 5. Restore snapshot
_ = m.Files.Restore(fs.Files, "HEAD")
```


---

**In short:**
`block` stores data → `file` maps files → `snapshot` versions sets → `manager` glues all.

```
