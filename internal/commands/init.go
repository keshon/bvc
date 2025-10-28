package commands

import (
	_ "app/internal/commands/add"
	_ "app/internal/commands/analyze"
	_ "app/internal/commands/branch"
	_ "app/internal/commands/branch-name"
	_ "app/internal/commands/checkout"
	_ "app/internal/commands/cherry-pick"
	_ "app/internal/commands/commit"
	_ "app/internal/commands/help"
	_ "app/internal/commands/init"
	_ "app/internal/commands/log"
	_ "app/internal/commands/merge"
	_ "app/internal/commands/reset"
	_ "app/internal/commands/status"
	_ "app/internal/commands/verify"
)

// import all commands to trigger init
