# Block Version Control (BVC)

BVC is a content-addressable version control system for managing snapshots of files and organizing them into streams (catalogs). It provides reliable project backup, versioning, and history tracking.

> [!WARNING]
> This proof-of-concept pet project is obviously not intended for production use. 

## The idea

**Rustic blocks backup + Git versioning = BVC Frankenstein**

BVC merges concepts from **Rustic** and **Git**, combining simple snapshot-based workflows with content-addressable versioning.

* **Snapshots (Rustic-inspired)**
  Each snapshot captures the state of the project at a point in time. Files are split into blocks for deduplication, and snapshots store the exact content of these blocks. Snapshots are independent—you can restore any snapshot without worrying about intermediate changes.

* **Streams**
  Streams group snapshots and behave differently depending on the workflow mode:

  * **Snapshot-first mode**: streams act like **tags or collections**. They do not imply a sequence of changes but are named groups of snapshots for easier reference.
  * **Stream-first mode**: streams behave like **Git branches**, representing evolving lines of development. Snapshots in a stream act like commits that can be added, merged, or checkout the latest snapshot to recreate a working state.

* **Blocks (Content-addressable storage)**
  Files are split into blocks identified by their hash, enabling:

  * Deduplication: identical data is stored only once across snapshots and streams.
  * Integrity: block corruption can be detected and recovered if a healthy copy exists.

---

## Basic Commands

### Repository Initialization

* `init` — Initialize a repository in the current directory. Choose between `snapshot-first` or `stream-first` modes.

### Snapshots

* `snapshot create` — Create a new snapshot of the project.
* `snapshot list` — List all snapshots.
* `snapshot show <id>` — Show detailed information for a snapshot.
* `snapshot checkout <id>` — Restore files from a snapshot.
* `snapshot diff <a> <b>` — Compare two snapshots.
* `snapshot merge <a> <b> <new>` — Merge two snapshots into a new one.

### Streams

* `stream create <name>` — Create a new stream.
* `stream add <stream> <snapshotID>` — Add a snapshot to a stream.
* `stream list` — List all streams.
* `stream show <name>` — Show snapshots in a stream.
* `stream checkout <name>` — Checkout the latest snapshot(s) from a stream.
* `stream clone <src> <dst>` — Clone a stream.
* `stream remove <name>` — Remove a stream.

### Maintenance

* `prune [-dry]` — Remove unused blocks not referenced by any snapshot. Use `-dry` to preview what would be removed.

### Repository Status

* `status` — Show changed files since the last snapshot or current stream.

### Integrity Checks

* `check workspace [-repair]` — Verify files in the working directory against the latest snapshot or stream.

  * Without `-repair`: reports missing or corrupted files.
  * With `-repair`: attempts to restore missing files from HEAD snapshots.

* `check blocks [-repair]` — Verify integrity of all blocks referenced by snapshots.

  * Without `-repair`: lists missing or corrupted blocks.
  * With `-repair`: recalculates missing blocks from healthy files in the workspace.

### Synchronization

* `sync pull` — Pull snapshots and blocks from remote storage (**not implemented**).
* `sync push` — Push snapshots and blocks to remote storage (**not implemented**).

### Help

* `help` — Show available commands and descriptions.

---

## License

BVC is licensed under the [MIT License](LICENSE).