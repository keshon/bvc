# BVC – Binary Version Control (Proof of Concept)

BVC is a content-addressed, block-based version control system designed for **large binary projects**.  
It is a **personal pet project** and **not production ready**. Use at your own risk.

---

## Available Commands

### back
`back <commit-id>`
Revert to a specific commit

### block
`block [branch|name]`
Display repository blocks overview

### commit
`commit "<message>"`
Commit current changes

### drop
`drop`
Discard pending changes

### goto
`goto <branch-name>`
Switch to another branch

### help
`help <command-name>`
Show help for commands

### init
`init`
Initialize a new repository

### list
`list`
List all branches

### log
`log [all]`
Show commit history (use 'all' for all branches)

### merge
`merge <branch-name>`
Merge another branch into current

### new
`new <branch-name>`
Create a new branch

### pending
`pending`
Show uncommitted changes

### pick
`pick <commit-id>`
Apply selected commit to current branch

### repair
`repair`
Repair missing or damaged repository blocks

### scan
`scan`
Verify repository blocks and file integrity


