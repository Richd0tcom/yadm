package storage

import (
	"crypto/sha256"
	"fmt"
	"io"

	"encoding/hex"
	"os"
	"path/filepath"
)

type BlobStore struct {
	rootDir string
}

func NewBlobStore(rootDir string) (*BlobStore, error) {

	//create blob dir if not exist
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		err := os.MkdirAll(rootDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	//retrun error if not exist
	return &BlobStore{
		rootDir: rootDir,
	}, nil
}

// Store writes content to the blob store and returns the SHA-256 hash
func (bs *BlobStore) Store(content []byte) (string, error) {

	hashBytes := sha256.Sum256(content)
	hashString := hex.EncodeToString(hashBytes[:])

	exist, err := bs.Exists(hashString)
	if err != nil {
		return "", err
	}
	if exist {
		//verify if valid
		return hashString, nil
	}
	blobPath := bs.blobPath(hashString)

	err = atomicWrite(blobPath, content)
	if err != nil {
		return "", err
	}

	return hashString, nil
}

// Exists checks if a blob exists without reading it
func (bs *BlobStore) Exists(hash string) (bool, error) {

	_, err := os.Stat(bs.blobPath(hash))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Read retrieves blob content by hash
func (bs *BlobStore) Read(hash string) ([]byte, error) {

	path:= bs.blobPath(hash)

	file,  err:= os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err:= io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	computedHash := computeHash(content)
	if computedHash != hash {
		//TODO: format proper error
		return nil, fmt.Errorf("hash mismatch: expected %s, got %s", hash, computedHash)
	}

	return content, nil
}

// blobPath constructs the filesystem path for a blob
func (bs *BlobStore) blobPath(hash string) string {
	prefix := hash[:2]
	return filepath.Join(bs.rootDir, prefix, hash)
}

func atomicWrite(path string, content []byte) error {

	dir := filepath.Dir(path)

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer func() {

		if tmpFile != nil {
			tmpFile.Close()
			os.Remove(tmpPath)
		}

	}()

	if _, err := tmpFile.Write(content); err != nil {
		return err
	}

	if err := tmpFile.Sync(); err != nil {
		return err
	}

	if err := tmpFile.Close(); err != nil {
		return err
	}

	tmpFile = nil

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}


	return nil
}

func computeHash(content []byte) string {
	hashBytes := sha256.Sum256(content)
	return hex.EncodeToString(hashBytes[:])
}
