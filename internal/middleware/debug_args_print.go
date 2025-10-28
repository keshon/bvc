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
				fmt.Printf("Flags: %+v\n", ctx.Flags)
				fmt.Printf("BoolFlags: %+v\n", ctx.BoolFlags)
				return cmd.Run(ctx)
			},
		}
	}
}
