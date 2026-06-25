package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/richd0tcom/yadm/interrnal/storage"
	"github.com/richd0tcom/yadm/pkg/utils"
)

type FileEntry struct {
	AbsolutePath string `json:"absolutePath"`
	RelativePath string `json:"relativePath"`
	Anchor       string `json:"anchor"`
	Size         int64  `json:"size"`
	Hash         string `json:"hash"`
	PermMode     string `json:"permMode"`
	Error        string `json:"error"`
}

type Snapshot struct {
	ID        string      `json:"id"`
	Timestamp int64       `json:"timestamp"`
	Author    string      `json:"author"`
	Label     string      `json:"label"`
	Files     []FileEntry `json:"files"`
}

func resolveAnchor(absPath string) (string, string) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic("could not determine home directory")
	}

	relativePath := strings.TrimPrefix(absPath, homeDir)
	return relativePath, "$HOME"

	//TODO: handle case where expandTilde
	// didnt produce a path starting from the home dir (is this een possible?)
}

func formatMode(mode os.FileMode) string {
	return fmt.Sprintf("%04o", mode)
}

func Create(dotfiles []string, blobstore *storage.BlobStore, snapshotDir string) (Snapshot, error){
	snap := Snapshot{}

	snap.Timestamp = time.Now().UnixMilli()
	snap.ID = strconv.FormatInt(snap.Timestamp, 10)

	currentUser, err := user.Current()
	if err != nil {
		// log.Fatalf("Failed to look up current user: %v", err)
	}

	snap.Author = currentUser.Username

	snap.Files = make([]FileEntry, len(dotfiles))

	for _, df := range dotfiles {
		entry := FileEntry{}

		absPath := utils.ExpandTilde(df)
		entry.AbsolutePath = absPath

		relativePath, anchor := resolveAnchor(absPath)
		entry.RelativePath = relativePath
		entry.Anchor = anchor

		content, err := os.ReadFile(absPath)
		if err != nil {
			entry.Error = err.Error()

			snap.Files = append(snap.Files, entry)
			continue
		}

		stat, err := os.Stat(absPath)
		if err != nil {
			entry.Error = err.Error()

			snap.Files = append(snap.Files, entry)
			continue
		}
		entry.Size = stat.Size()
		entry.PermMode = formatMode(stat.Mode())

		hash, err := blobstore.Store(content)
		if err != nil {
			entry.Error = err.Error()
		} else {
			entry.Hash = hash
		}

		snap.Files = append(snap.Files, entry)
	}

	if err = saveSnapshot(snapshotDir, snap); err!=nil {
		return Snapshot{}, err
	}

	return snap, err
}

func saveSnapshot(saveDir string, snap Snapshot) error {
	path := filepath.Join(saveDir, snap.ID+".json")

	data, _ := json.Marshal(snap)

	return storage.AtomicWrite(path, data)
}
