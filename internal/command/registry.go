package command

var tree = NewTree()

// RegisterCommand adds a command to the global tree
func RegisterCommand(cmd Command) {
	tree.Register(cmd)
}

// ResolveCommand finds a command from args
func ResolveCommand(args []string) (*Node, []string, error) {
	return tree.Resolve(args)
}

// GetCommand returns a command by name
func GetCommand(name string) (Command, bool) {
	return tree.Get(name)
}

// AllCommands returns all commands registered in the global tree.
func AllCommands() []Command {
	cmds := make([]Command, 0)
	seen := make(map[Command]struct{})

	var walk func(node *Node)
	walk = func(node *Node) {
		if node.Cmd != nil {
			if _, ok := seen[node.Cmd]; !ok {
				cmds = append(cmds, node.Cmd)
				seen[node.Cmd] = struct{}{}
			}
		}
		for _, sub := range node.Subcommands {
			walk(sub)
		}
	}

	walk(tree.root)
	return cmds
}
