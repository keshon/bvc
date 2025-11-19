# BVC – Binary Version Control (Proof of Concept)

BVC is a content-addressed, block-based version control system designed for **large binary projects**.  
It is a **personal pet project** and **not production ready**. Use at your own risk.

---

## Available Commands

### bvc add
```
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

### bvc block
```
Manage repository blocks and analysis

Usage:
  bvc block <subcommand> [options]

Available subcommands:
  bvc block list
  bvc block scan
  bvc block repair
  bvc block reuse

```

### bvc branch
```
List all branches or create a new one.

Usage:
  branch           - list all branches (current marked with '*')
  branch <name>    - create a new branch from the current one
```

### bvc checkout
```
Switch to another branch.

Usage:
  checkout <branch-name>
```

### bvc cherry-pick
```
Apply a specific commit to the current branch.

Usage:
  cherry-pick <commit-id>
```

### bvc commit
```
Create a new commit with the staged changes.

Usage:
  commit -m "<message>"               - commit with a given message
  commit -m "<message>" --allow-empty - empty commit with a given message (no staged files exist)
```

### bvc help
```
Display help information for commands.

Usage:
  help          List all commands.
  help <name>   Show detailed help for a specific command.
```

### bvc init
```
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

### bvc list
```

Display repository blocks list

Usage:
  bvc block list [branch|name]

Examples:
  bvc block list               List all blocks sorted by hash
  bvc block list branch        List blocks sorted by branch name
  bvc block list name          List blocks sorted by file name

```

### bvc log
```
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

### bvc merge
```
Perform a three-way merge of the specified branch into the current branch.
Conflicts may need manual resolution.
```

### bvc repair
```
Repair any missing or damaged blocks automatically.

Examples:
  bvc block repair
	

```

### bvc reset
```
Reset current branch.

Options:
  --soft  : move HEAD only
  --mixed : move HEAD and reset index (default)
  --hard  : move HEAD, reset index and working directory

If <commit-id> is omitted, the last commit is used.

Usage:
  bvc reset [<commit-id>] [--soft|--mixed|--hard]

Examples:
  bvc reset
  bvc reset --mixed
  bvc reset --hard

  bvc reset <commit-id>
  bvc reset --soft <commit-id>
  bvc reset --mixed <commit-id>
  bvc reset --hard <commit-id>

```

### bvc reuse
```
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

### bvc reuse
```
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

### bvc scan
```
Scan all repository blocks and report missing or damaged ones.

Usage:
  bvc block scan	

```

### bvc status
```
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


