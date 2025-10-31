package command

// Middleware is a function that wraps a command
type Middleware func(Command) Command

// WrappedCommand represents a command wrapped with a middleware
type WrappedCommand struct {
	Command
	Wrap func(ctx *Context) error
}

// Run executes the wrapped command
func (w *WrappedCommand) Run(ctx *Context) error {
	if w.Wrap != nil {
		return w.Wrap(ctx)
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
