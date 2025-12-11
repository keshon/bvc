package status

import (
	"bvc/command"
	"bvc/internal/repo"
	"bvc/util"
	"flag"
	"fmt"
	"os"
	"sort"
)

type StatusCmd struct{}

func (c *StatusCmd) Name() string                   { return "status" }
func (c *StatusCmd) Help() string                   { return "Show changed files since last snapshot or current stream" }
func (c *StatusCmd) Flags(fs *flag.FlagSet)         {}
func (c *StatusCmd) SubCommands() []command.Command { return nil }

func (c *StatusCmd) Run(ctx command.Context) error {
	r, err := repo.OpenRepo(".")
	if err != nil {
		return err
	}

	var snapFiles map[string][]string

	head, err := r.Head.Load()
	if err != nil || head.Mode == "" {
		// HEAD doesn't exist or is empty - consider empty snapshot
		snapFiles = map[string][]string{}
	} else {
		switch head.Mode {
		case "snapshot-first":
			if head.Ref == "" {
				snapFiles = map[string][]string{}
			} else {
				meta, err := r.Snaps.Load(head.Ref)
				if err != nil {
					return fmt.Errorf("load snapshot: %w", err)
				}
				snapFiles = meta.Files
			}
		case "stream-first":
			if head.Ref == "" {
				snapFiles = map[string][]string{}
			} else {
				st, err := r.Streams.Load(head.Ref)
				if err != nil {
					return fmt.Errorf("load stream: %w", err)
				}
				if len(st.Snapshots) == 0 {
					snapFiles = map[string][]string{}
				} else {
					latestSnapID := st.Snapshots[len(st.Snapshots)-1]
					meta, err := r.Snaps.Load(latestSnapID)
					if err != nil {
						return fmt.Errorf("load snapshot from stream: %w", err)
					}
					snapFiles = meta.Files
				}
			}
		default:
			// unknown mode - consider empty snapshot
			snapFiles = map[string][]string{}
		}
	}

	// collect current disk files
	currFiles := map[string][]string{}
	if err := r.WalkFiles(func(rel string, f *os.File) error {
		hashes, err := r.HashFileBlocks(f)
		if err != nil {
			return err
		}
		currFiles[rel] = hashes
		return nil
	}); err != nil {
		return err
	}

	var changes []string

	// detect NEW and MODIFIED
	for rel, curr := range currFiles {
		old, ok := snapFiles[rel]
		if !ok {
			changes = append(changes, "NEW: "+rel)
			continue
		}
		if !util.EqualStringSlices(curr, old) {
			changes = append(changes, "MODIFIED: "+rel)
		}
	}

	// detect DELETED
	for rel := range snapFiles {
		if _, ok := currFiles[rel]; !ok {
			changes = append(changes, "DELETED: "+rel)
		}
	}

	if len(changes) == 0 {
		fmt.Println("No changes since last snapshot")
		return nil
	}

	// sort for stable output
	sort.Strings(changes)
	for _, line := range changes {
		fmt.Println(line)
	}
	return nil
}
