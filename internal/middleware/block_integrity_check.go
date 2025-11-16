package middleware

import (
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/repotools"
)

// WithBlockIntegrityCheck is a middleware that checks the integrity of the repository blocks
func WithBlockIntegrityCheck() command.Middleware {
	return func(cmd command.Command) command.Command {
		return &command.WrappedCommand{
			Command: cmd,
			Wrap: func(ctx *command.Context) error {
				fmt.Println("Checking repository integrity...")
				r, err := repo.NewRepositoryByPath(config.ResolveRepoDir())
				if err != nil {
					return fmt.Errorf("failed to open repository: %w", err)
				}
				if err := repotools.VerifyBlocks(r.Meta, r.Config, true); err != nil {
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
