package file

type FileContextInterface interface {
	BuildEntry(path string) (Entry, error)
	BuildEntries(paths []string, silent bool) ([]Entry, error)
	Write(e Entry) error
	Exists(path string) bool

	SaveIndexReplace(entries []Entry) error
	SaveIndexMerge(newEntries []Entry) error
	ClearIndex() error
	LoadIndex() ([]Entry, error)

	RestoreFilesToWorkingTree([]Entry, string) error
}
