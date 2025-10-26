package cli

var registry = map[string]Command{}

// RegisterCommand registers a command
func RegisterCommand(cmd Command) {
	registry[cmd.Name()] = cmd
}

// GetCommand returns a command by name
func GetCommand(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}

// AllCommands returns a list of all registered commands
func AllCommands() []Command {
	list := make([]Command, 0, len(registry))
	for _, cmd := range registry {
		list = append(list, cmd)
	}
	return list
}
