package snapshot

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/richd0tcom/yadm/internal/config"
	"github.com/richd0tcom/yadm/internal/storage"
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

func modeToString(mode os.FileMode) string {
	return fmt.Sprintf("%04o", mode)
}

func stringToMode(s string) (os.FileMode, error) {
	mode, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, err
	}
	return os.FileMode(mode), nil
}

func Create(dotfiles []string, blobstore *storage.BlobStore, snapshotDir string) (Snapshot, error) {
	snap := Snapshot{}

	snap.Timestamp = time.Now().UnixMilli()
	snap.ID = strconv.FormatInt(snap.Timestamp, 10)

	currentUser, err := user.Current()
	if err != nil {
		// log.Fatalf("Failed to look up current user: %v", err)
	}

	snap.Author = currentUser.Username

	snap.Files = make([]FileEntry, 0, len(dotfiles))

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
		entry.PermMode = modeToString(stat.Mode())

		hash, err := blobstore.Store(content)
		if err != nil {
			entry.Error = err.Error()
		} else {
			entry.Hash = hash
		}

		snap.Files = append(snap.Files, entry)
	}

	if err = saveSnapshot(snapshotDir, snap); err != nil {
		return Snapshot{}, err
	}

	return snap, err
}

func saveSnapshot(saveDir string, snap Snapshot) error {
	path := filepath.Join(saveDir, snap.ID+".json")

	data, _ := json.Marshal(snap)

	return storage.AtomicWrite(path, data)
}

type Restore struct {
	Path    string
	Success bool
	Error   string
}

func RestoreSnapshot(snapID string, blobstore *storage.BlobStore, dryrun bool) ([]Restore, error) {

	var snap Snapshot
	snapPath := filepath.Join(config.GetSnaphotDir(), snapID+".json")

	data, err := os.ReadFile(snapPath)

	if err != nil {

	}
	err = json.Unmarshal(data, &snap)

	if err != nil {

	}

	results := make([]Restore, 0, len(snap.Files))

	//TODO:  before restoring, we should probably snapshot the current state first.

	for _, file := range snap.Files {
		if file.Error != "" {
			results = append(results, Restore{
				Path:    file.AbsolutePath,
				Success: false,
				Error:   "skipped: errors present at snapshot time",
			})
			continue
		}
		if dryrun {
			results = append(results, Restore{
				Path:    file.AbsolutePath,
				Success: true,
				Error:   "dry-run: would restore",
			})
			continue
		}

		content, err := blobstore.Read(file.Hash)

		if err != nil {
			results = append(results, Restore{
				Path:    file.AbsolutePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		err = storage.AtomicWrite(file.AbsolutePath, content)

		if err != nil {
			results = append(results, Restore{
				Path:    file.AbsolutePath,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		mode, err := stringToMode(file.PermMode)

		if err != nil {
			log.Default().Printf("unable to parse file permission mode: %s ", err.Error())
		}
		err = os.Chmod(file.AbsolutePath, mode)

		if err != nil {
			log.Default().Printf("unable to restore file permission mode: %s ", err.Error())

		}
		results = append(results, Restore{
			Path:    file.AbsolutePath,
			Success: false,
			Error:   "",
		})

	}

	return results, nil
}


func ReadSnapshot(snapID string) (Snapshot,  error) {
	var snap Snapshot
	snapPath := filepath.Join(config.GetSnaphotDir(), snapID+".json")

	data, err := os.ReadFile(snapPath)

	if err != nil {

		if os.IsNotExist(err) {
			return Snapshot{}, fmt.Errorf("Snapshot not found, nothing to diff")
		}
		return Snapshot{}, fmt.Errorf("Error reading snapshot: %s", err)
	}

	err = json.Unmarshal(data, &snap)

	if err != nil {
		return Snapshot{}, fmt.Errorf("Error unmarshaling snapshot: %s", err)
	}

	return snap , nil
}