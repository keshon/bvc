package snapshot

import "github.com/keshon/bvc/internal/repo/file"

type SnapshotContextInterface interface {
	BuildWorkingTreeFileset() (Fileset, error)
	BuildStagedFileset() (Fileset, error)
	BuildIgnoredFileset() (Fileset, error)
	BuildAllRepositoryFilesets() (tracked Fileset, staged Fileset, ignored Fileset, err error)
	BuildFilesetFromEntries(entries []file.Entry) (Fileset, error)
	WriteAndSave(fs *Fileset) error
	Save(fs Fileset) error
	Load(filesetID string) (Fileset, error)
	List() ([]Fileset, error)
}
