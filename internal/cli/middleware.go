package cli

type Middleware func(Command) Command

type wrappedCommand struct {
	Command
	wrap func(ctx *Context) error
}

func (w *wrappedCommand) Run(ctx *Context) error {
	if w.wrap != nil {
		return w.wrap(ctx)
	}
	return w.Command.Run(ctx)
}

// ApplyMiddlewares wraps a command with any number of middlewares
func ApplyMiddlewares(cmd Command, mws ...Middleware) Command {
	for _, mw := range mws {
		cmd = mw(cmd)
	}
	return cmd
}
