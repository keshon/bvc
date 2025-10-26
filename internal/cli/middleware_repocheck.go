package cli

import (
	"app/internal/verify"
	"fmt"
)

// WithRepoCheck ensures the repository is healthy before running the command
func WithRepoCheck() Middleware {
	return func(cmd Command) Command {
		return &wrappedCommand{
			Command: cmd,
			wrap: func(ctx *Context) error {
				fmt.Println("Checking repository integrity...")
				if err := verify.ScanRepositoryBlocks(); err != nil {
					return fmt.Errorf(
						"repository verification failed: %v\nPlease run `bvc repair` before continuing",
						err,
					)
				}
				return cmd.Run(ctx)
			},
		}
	}
}
