# BVC – Binary Version Control (Proof of Concept)

BVC is a content-addressed, block-based version control system designed for **large binary projects**.  
It is a **personal pet project** and **not production ready**. Use at your own risk.

---

## Available Commands

### add
`add <file|dir|.>`
Stage files or directories for the next commit

### analyze
`analyze [--sort reuse|unique|size]`
Analyze block reuse across the entire repository (all snapshots and branches)

### branch
`branch [<branch-name>]`
List all branches or create a new one

### checkout
`checkout <branch-name>`
Switch to another branch

### cherry-pick
`cherry-pick <commit-id>`
Apply selected commit to the current branch

### commit
`commit -m "<message>" [--allow-empty]`
Commit staged changes to the current branch

### help
`help [command]`
Show help for commands

### init
`init`
Initialize a new repository

### log
`log [-a|--all]`
Show commit history (current branch by default)

### merge
`merge <branch-name>`
Merge another branch into the current branch

### reset
`reset [<commit-id>] [--soft|--mixed|--hard]`
Reset current branch to a commit or HEAD

### status
`status`
Show uncommitted changes

### verify
`verify [--repair|--auto]`
Verify repository integrity or attempt to repair missing/damaged blocks


