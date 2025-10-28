package middleware

import (
	"app/internal/cli"
	"app/internal/repo"
	"fmt"
)

// WithBlockIntegrityCheck is a middleware that checks the integrity of the repository blocks
func WithBlockIntegrityCheck() cli.Middleware {
	return func(cmd cli.Command) cli.Command {
		return &cli.WrappedCommand{
			Command: cmd,
			Wrap: func(ctx *cli.Context) error {
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
