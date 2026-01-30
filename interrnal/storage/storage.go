package storage

type BlobStore struct {
	rootDir string
}

func NewBlobStore(rootDir string) (*BlobStore, error ) {

	//create temp dir if not exist

	//retrun error if not exist
	return &BlobStore{
		rootDir: rootDir,
	},  nil
}


// Store writes content to the blob store and returns the SHA-256 hash
func (bs *BlobStore) Store(content []byte) (string, error) {
	return "", nil
}

// Exists checks if a blob exists without reading it
func (bs *BlobStore) Exists(hash string) (bool, error) {
	return true, nil
}


// Read retrieves blob content by hash
func (bs *BlobStore) Read(hash string) ([]byte, error) {

	return nil, nil
}


// blobPath constructs the filesystem path for a blob
func (bs *BlobStore) blobPath(hash string) string {
	return ""
}

func atomicWrite(path string, ){

}