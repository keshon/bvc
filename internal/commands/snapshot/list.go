package snapshot

import (
	"flag"
	"fmt"
	"sort"

	"github.com/keshon/bvc/pkg/command"

	"github.com/keshon/bvc/internal/repo"
	"github.com/keshon/bvc/internal/snapshot"
)

type ListCmd struct{}

func (c *ListCmd) Name() string                   { return "list" }
func (c *ListCmd) Brief() string                  { return "List all snapshots" }
func (c *ListCmd) Help() string                   { return "List all snapshots" }
func (c *ListCmd) Flags(fs *flag.FlagSet)         {}
func (c *ListCmd) SubCommands() []command.Command { return nil }

func (c *ListCmd) Run(ctx *command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}
	if err := r.RequireMode("snapshot-first"); err != nil {
		return err
	}

	listIds, err := r.Snaps.List()
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}

	list := make([]snapshot.Meta, len(listIds))
	for i, id := range listIds {
		snap, err := r.Snaps.Load(id)
		if err != nil {
			return err
		}
		list[i] = *snap
	}

	// sort by created_at
	sort.Slice(list, func(i, j int) bool {
		return list[i].CreatedAt.Before(list[j].CreatedAt)
	})

	for _, s := range list {
		desc := s.Description
		if len(desc) > 30 {
			desc = desc[:30] + "..."
		}
		fmt.Printf("%s\t%s\t%s\t%s\tfiles=%d\n",
			s.ID,
			s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			s.Name,
			desc,
			len(s.Files),
		)
	}

	return nil
}
