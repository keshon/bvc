package cli

import "strings"

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
	Args      []string          // Positional arguments
	Flags     map[string]string // Value flags: -m msg or --message=msg
	BoolFlags map[string]bool   // Boolean flags: --hard or -H
	RawArgs   []string          // Original args
}

// NewContext parses raw arguments
func NewContext(args []string) *Context {
	ctx := &Context{
		Args:      []string{},
		Flags:     make(map[string]string),
		BoolFlags: make(map[string]bool),
		RawArgs:   args,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			if eq := strings.Index(arg, "="); eq >= 0 {
				ctx.Flags[arg[2:eq]] = arg[eq+1:]
			} else {
				ctx.BoolFlags[arg[2:]] = true
			}
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			key := string(arg[1])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				ctx.Flags[key] = args[i+1]
				i++
			} else {
				ctx.BoolFlags[key] = true
			}
		} else {
			ctx.Args = append(ctx.Args, arg)
		}
	}

	return ctx
}
