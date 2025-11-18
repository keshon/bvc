# BVC – Binary Version Control (Proof of Concept)

BVC is a content-addressed, block-based version control system designed for **large binary projects**.  
It is a **personal pet project** and **not production ready**. Use at your own risk.

---

## Available Commands

### add
```
add <file|dir|.> [options]
Stage changes for commit.

Options:
  -a, --all             Stage all changes, including deletions (-A)
	  --update          Stage modifications and deletions only (-u)

Usage:
  bvc add <file|dir|.> [options]

Examples:
  bvc add .
  bvc add 'main.go'
  bvc add dir/

```

### block
```
block <subcommand> [options]
Manage repository blocks and analysis

Usage:
  bvc block <subcommand> [options]

Available subcommands:
  bvc block list
  bvc block scan
  bvc block repair
  bvc block reuse

```

### branch
```
branch [options] [<branch-name>]
List all branches or create a new one.

Usage:
  branch           - list all branches (current marked with '*')
  branch <name>    - create a new branch from the current one
```

### checkout
```
checkout <branch-name>
Switch to another branch.

Usage:
  checkout <branch-name>
```

### cherry-pick
```
cherry-pick <commit-id>
Apply a specific commit to the current branch.

Usage:
  cherry-pick <commit-id>
```

### commit
```
commit -m "<message>" [--allow-empty]
Create a new commit with the staged changes.

Usage:
  commit -m "<message>"               - commit with a given message
  commit -m "<message>" --allow-empty - empty commit with a given message (no staged files exist)
```

### help
```
help [command]
Display help information for commands.

Usage:
  help          List all commands.
  help <name>   Show detailed help for a specific command.
```

### init
```
init [options]
Initialize a new repository in the current directory.

Options:
  -q, --quiet                 Suppress normal output.
      --separate-bvc-dir=<d>  Store repository data in a separate directory.
  -b, --initial-branch=<name> Use a custom initial branch name (default: main).
  
Usage:
  bvc init [options]

Examples:
  bvc init
  bvc init -q
  bvc init --separate-bvc-dir=~/.bvc
  bvc init --initial-branch=master

```

### list
```
block list [branch|name]

Display repository blocks list

Usage:
  bvc block list [branch|name]

Examples:
  bvc block list               List all blocks sorted by hash
  bvc block list branch        List blocks sorted by branch name
  bvc block list name          List blocks sorted by file name

```

### log
```
log [options] [branch]
Show commit logs.

Options:
  -a, --all             Show commits from all branches.
      --oneline         Show each commit as a single line (ID + message).
  -n <count>            Limit to the last N commits.
      --since <date>    Show commits after the given date (YYYY-MM-DD).
      --until <date>    Show commits before the given date (YYYY-MM-DD).

Usage:
  bvc log [options]

Examples:
  bvc log
  bvc log -a
  bvc log --oneline -n 10
  bvc log main
```

### merge
```
merge <branch-name>
Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.
```

### repair
```
block repair
Repair any missing or damaged blocks automatically.

Examples:
  bvc block repair
	

```

### reset
```
reset [<commit-id>] [--soft|--mixed|--hard]
Reset current branch.

Modes:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory

If <commit-id> is omitted, the last commit is used.
```

### reuse
```
block reuse [--full] [--export]
Analyze block reuse across branches
Options:
  -f, --full            Print detailed shared block list
  -e, --export          Save output to file

Usage:
  bvc block reuse [options]

Examples:
  bvc block reuse
  bvc block reuse --full
  bvc block reuse --export

```

### reuse
```
block reuse [--full] [--export]
Analyze block reuse across branches
Options:
  -f, --full            Print detailed shared block list
  -e, --export          Save output to file

Usage:
  bvc block reuse [options]

Examples:
  bvc block reuse
  bvc block reuse --full
  bvc block reuse --export

```

### scan
```
block scan
Scan all repository blocks and report missing or damaged ones.

Usage:
  bvc block scan	

```

### status
```
status [options]
Show the working tree status.

Options:
  -s, --short                    Show short summary (XY path)
      --porcelain                Machine-readable short output
  -b, --branch                   Show branch info
  -u, --untracked-files=<mode>   Show untracked files: no, normal, all (default: normal)
      --ignored                  Show ignored files
  -q, --quiet                    Suppress normal output

Usage:
  bvc status [options]

Examples:
  bvc status
  bvc status -s
  bvc status --branch

```


