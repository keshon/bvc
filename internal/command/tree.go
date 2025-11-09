package command

import (
	"errors"
)

// Node represents a node in the command tree.
type Node struct {
	Cmd         Command
	Subcommands map[string]*Node
}

// CommandTree manages all commands and subcommands.
type CommandTree struct {
	root *Node
}

// NewTree creates a new empty command tree.
func NewTree() *CommandTree {
	return &CommandTree{
		root: &Node{Subcommands: make(map[string]*Node)},
	}
}

// Register inserts a command and all its subcommands recursively.
func (t *CommandTree) Register(cmd Command) {
	t.insert(t.root, cmd)
}

// Get returns a command by name.
func (t *CommandTree) Get(name string) (Command, bool) {
	node, ok := t.root.Subcommands[name]
	if !ok {
		return nil, false
	}
	return node.Cmd, true
}

func (t *CommandTree) insert(node *Node, cmd Command) {
	names := append([]string{cmd.Name()}, cmd.Aliases()...)
	for _, n := range names {
		if node.Subcommands == nil {
			node.Subcommands = make(map[string]*Node)
		}
		sub := &Node{Cmd: cmd, Subcommands: make(map[string]*Node)}
		node.Subcommands[n] = sub
		// Recursively add subcommands
		for _, subcmd := range cmd.Subcommands() {
			t.insert(sub, subcmd)
		}
	}
}

// Resolve walks down the command tree following args.
func (t *CommandTree) Resolve(args []string) (*Node, []string, error) {
	node := t.root
	for len(args) > 0 {
		next, ok := node.Subcommands[args[0]]
		if !ok {
			break
		}
		node = next
		args = args[1:]
	}
	if node.Cmd == nil {
		return nil, nil, errors.New("unknown command")
	}
	return node, args, nil
}
