# BVC (Block Version Control)

BVC is a content-addressable version control system for managing snapshots of files and organizing them into streams (catalogs). It is designed to provide efficient and reliable project backup, versioning, and history tracking.  

BVC supports two workflow modes:  
- **Snapshot-first**: Each snapshot represents a full point-in-time backup of the project. This mode is suited for strict versioning and incremental backups, allowing you to restore or compare any snapshot independently.  
- **Stream-first**: Snapshots are organized into streams, which act like evolving project branches or catalogs. This mode is optimized for linear workflows, where the latest snapshot in a stream represents the current state, and earlier snapshots can be incrementally applied or merged.  

Both modes use content-addressable storage for deduplication and integrity, ensuring that identical file blocks are stored only once while maintaining a complete history of changes.

## Key Concepts

- **Snapshots**: Immutable versions of the project at a point in time. Each snapshot stores file blocks identified by content hashes.
- **Streams**: Ordered collections of snapshots, similar to branches or catalogs, allowing workflows over sequences of snapshots.
- **Blocks**: Individual pieces of file data stored once and referenced by snapshots, enabling deduplication.

## Available Commands

### Repository Initialization

- `init` — Initialize a new repository in the current directory. Choose between `snapshot-first` or `stream-first` modes.

### Snapshots

- `snapshot create` — Create a new snapshot of the project.
- `snapshot list` — List all snapshots.
- `snapshot show <id>` — Show detailed information of a snapshot.
- `snapshot checkout <id>` — Restore project files from a snapshot.
- `snapshot diff <a> <b>` — Compare two snapshots.
- `snapshot merge <a> <b> <new>` — Merge two snapshots into a new one.

### Streams

- `stream create <name>` — Create a new stream.
- `stream add <stream> <snapshotID>` — Add a snapshot to a stream.
- `stream list` — List all streams.
- `stream show <name>` — Show snapshots in a stream.
- `stream checkout <name>` — Checkout the latest snapshot(s) from a stream.
- `stream clone <src> <dst>` — Clone a stream.
- `stream remove <name>` — Remove a stream.

### Maintenance

- `prune [-dry]` — Remove unused blocks not referenced by any snapshot. Use `-dry` to see what would be removed without deleting.

### Repository Status

- `status` — Show changed files since the last snapshot or current stream.

### Synchronization

- `sync pull` — Pull snapshots and blocks from remote storage (not implemented).
- `sync push` — Push snapshots and blocks to remote storage (not implemented).

### Help

- `help` — Show available commands and their descriptions.

## Features

- Parallelized operations for snapshot creation, checkout, and pruning for improved performance.
- Content-addressable storage for block deduplication and integrity.
- Flexible workflow modes: snapshot-first and stream-first.
- Safe conflict handling and merge capabilities for snapshots.

