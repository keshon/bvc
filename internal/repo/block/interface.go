package block

type BlockContextInterface interface {
	Read(hash string) ([]byte, error)
	Write(filePath string, blocks []BlockRef) error
	CleanupTemp() error
	Verify(hashes map[string]struct{}, workers int) <-chan BlockCheck
	VerifyBlock(hash string) (BlockStatus, error)
	SplitFile(path string) ([]BlockRef, error)
	BlocksDir() string
}
