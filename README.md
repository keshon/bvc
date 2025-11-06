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
`analyze [--detail] [--export]`
Analyze block reuse across all branches and commits.

Usage:
  analyze --detail - print detailed shared block list
  analyze --export - save output to ${config.RepoDir}-analyze

### blocks
`blocks [branch|name]`
Show repository blocks list with optional sort mode.

Usage:
  blocks        - show all blocks
  blocks branch - sort by branch name
  blocks name 	- sort by file name

Useful for identifying shared blocks between branches and associated files.

### branch
`branch [<branch-name>]`
List all branches or create a new one.

Usage:
  branch        - list all branches (current marked with '*')
  branch <name> - create a new branch from the current one

### checkout
`checkout <branch-name>`
Switch to another branch.

Usage:
  checkout <branch-name>

### cherry-pick
`cherry-pick <commit-id>`
Apply a specific commit to the current branch.

Usage:
  cherry-pick <commit-id>

### commit
`commit -m "<message>" [--allow-empty]`
Create a new commit with the staged changes.

Usage:
  commit -m "<message>"               - commit with a given message
  commit -m "<message>" --allow-empty - empty commit with a given message (no staged files exist)
  
 

### help
`help [command]`
Display detailed help information for a specific command, or list all commands

Usage:
  help - list all commands
  help [command] - show help for a specific command

### init
`init [options]`
Initialize a new repository in the current directory.

Options:
  -q, --quiet                 Suppress normal output.
      --bare                  Create a bare repository.
      --object-format=<algo>  Hash algorithm: xxh3-128 or sha256 (default xxh3-128).
      --separate-bvc-dir=<d>  Store repository data in a separate directory.
  -b, --initial-branch=<name> Use a custom initial branch name (default: main).
  
Usage:
  bvc init [options]

Examples:
  bvc init
  bvc init -q
  bvc init --bare
  bvc init --separate-bvc-dir=~/.bvc
  bvc init --initial-branch=master


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



