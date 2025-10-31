package command

// Command registry
var registry = map[string]Command{}

// RegisterCommand registers a command
func RegisterCommand(cmd Command) {
	names := append([]string{cmd.Name()}, cmd.Aliases()...)
	for _, n := range names {
		registry[n] = cmd
	}
	if short := cmd.Short(); short != "" {
		registry[short] = cmd
	}
}

// GetCommand returns a registered command
func GetCommand(name string) (Command, bool) {
	cmd, ok := registry[name]
	return cmd, ok
}

// AllCommands returns all registered commands
func AllCommands() []Command {
	list := make([]Command, 0, len(registry))
	seen := map[Command]bool{}
	for _, cmd := range registry {
		if !seen[cmd] {
			list = append(list, cmd)
			seen[cmd] = true
		}
	}
	return list
}
