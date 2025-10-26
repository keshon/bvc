package storage

// BuildFileset scans all files and creates a snapshot of blocks.
func BuildFileset() (Fileset, error) {
	filePaths, err := listFiles()
	if err != nil {
		return Fileset{}, err
	}

	fileEntries, err := buildFileEntries(filePaths)
	if err != nil {
		return Fileset{}, err
	}

	return Fileset{
		ID:    HashFileset(fileEntries),
		Files: fileEntries,
	}, nil
}

// StoreFileset persists all file blocks in the fileset.
func StoreFileset(fs Fileset) error {
	return parallel(fs.Files, WorkerCount(), StoreFileEntry)
}
