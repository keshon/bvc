package middleware

import (
	"app/internal/cli"
	"fmt"
)

// WithBlockIntegrityCheck is a middleware that checks the integrity of the repository blocks
func WithDebugArgsPrint() cli.Middleware {
	return func(cmd cli.Command) cli.Command {
		return &cli.WrappedCommand{
			Command: cmd,
			Wrap: func(ctx *cli.Context) error {
				fmt.Printf("Args: %+v\n", ctx.Args)
				return cmd.Run(ctx)
			},
		}
	}
}
