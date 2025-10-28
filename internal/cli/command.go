package cli

// Command represents a cli command
type Command interface {
	Name() string
	Description() string
	DetailedDescription() string
	Usage() string
	Run(ctx *Context) error
	Aliases() []string
	Short() string
}

// Context represents a cli context
type Context struct {
	Args []string // Positional arguments
}
