package middleware

import (
	"fmt"

	"github.com/keshon/bvc/internal/command"
	"github.com/keshon/bvc/internal/config"
)

// WithBlockIntegrityCheck is a middleware that checks the integrity of the repository blocks
func WithDebugArgsPrint() command.Middleware {
	return func(cmd command.Command) command.Command {
		return &command.WrappedCommand{
			Command: cmd,
			Wrap: func(ctx *command.Context) error {
				if config.IsDev {
					fmt.Printf("Args: %+v\n", ctx.Args)
				}
				return cmd.Run(ctx)
			},
		}
	}
}
