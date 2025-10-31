package middleware

import (
	"app/internal/command"
	"app/internal/repo"
	"fmt"
)

// WithBlockIntegrityCheck is a middleware that checks the integrity of the repository blocks
func WithBlockIntegrityCheck() command.Middleware {
	return func(cmd command.Command) command.Command {
		return &command.WrappedCommand{
			Command: cmd,
			Wrap: func(ctx *command.Context) error {
				fmt.Println("Checking repository integrity...")
				if err := repo.VerifyBlocks(false); err != nil {
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
