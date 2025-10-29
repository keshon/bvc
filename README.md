# BVC – Binary Version Control (Proof of Concept)

BVC is a content-addressed, block-based version control system designed for **large binary projects**.  
It is a **personal pet project** and **not production ready**. Use at your own risk.

---

## Available Commands

### add
`add <file|dir|.>`
Stage changes for commit.
Usage:
  add .              - stage new and modified files
  add -A or --all    - stage all changes, including deletions
  add -u or --update - stage modifications and deletions (no new files)
  add <path>         - stage a specific file or directory

### analyze
`analyze [--full] [--export]`
Analyze block reuse across all branches and commits.
Use --full to print detailed shared block list.
Use --export to save output to .bvcanalyze.

### blocks
`blocks [branch|name]`
Show repository blocks list with optional sort:
  - default: by block hash
  - branch: sort by branch name
  - name: sort by file name

Useful for identifying shared blocks between branches and associated files.

### branch
`branch [<branch-name>]`
Usage:
  branch             - List all branches (current marked with '*')
  branch <name>      - Create a new branch from the current one

### checkout
`checkout <branch-name>`
Switch to another branch.
Restores the branch's fileset and updates HEAD reference.

### cherry-pick
`cherry-pick <commit-id>`
Apply a specific commit to the current branch.
Use 'bvc log all' to find the commit ID you want to apply.

### commit
`commit -m "<message>" [--allow-empty]`
Create a new commit with the staged changes.
Supports -m / --message for commit message.
Supports --allow-empty to commit even if no staged changes exist.

### help
`help [command]`
Display detailed help information for a specific command, or list all commands if none is provided.

### init
`init`
Initialize a new repository in the current directory.
If the directory is not empty, existing content will be marked as pending changes.

### log
`log [-a|--all]`
List commits for the current branch or all branches if -a / --all is specified.

### merge
`merge <branch-name>`
Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.

### reset
`reset [<commit-id>] [--soft|--mixed|--hard]`
Reset the current branch.
Modes:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory
If <commit-id> is omitted, the last commit is used (mixed).

### status
`status`
List uncommitted changes in the current branch.
WARNING: Switching branches with pending changes may cause data loss.

### verify
`verify [--repair|--auto]`
Verify repository blocks and file integrity.

Usage:
  verify           - Scan all blocks and report missing/damaged ones.
  verify --repair  - Attempt to repair any missing or damaged blocks automatically.



