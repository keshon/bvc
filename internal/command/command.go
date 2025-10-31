package command

// Command represents a cli command
type Command interface {
	Name() string
	Short() string
	Aliases() []string
	Usage() string
	Brief() string
	Help() string
	Run(ctx *Context) error
}

// Context represents a cli context
type Context struct {
	Args []string
}
