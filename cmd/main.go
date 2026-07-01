package main

import (
	"log"
	"path/filepath"

	"github.com/richd0tcom/yadm/internal/config"
	"github.com/richd0tcom/yadm/internal/diff"
	"github.com/richd0tcom/yadm/internal/snapshot"
	"github.com/richd0tcom/yadm/internal/storage"

	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tracker <command>")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(config.DEFAULT_CONFIG_PATH)
	if err != nil {
		log.Fatal(err)
	}

	trackerRoot := cfg.TrackerRoot

	if cfg.TrackerRoot == "" {
		trackerRoot = config.DEFAULT_TRACKER_ROOT
	}

	bs, err := storage.NewBlobStore(filepath.Join(trackerRoot, config.BLOB_FOLDER))
	if err != nil {
		log.Fatal(err)
	}

	//TODO: implement flags and robust argument parsing https://cobra.dev
	command := os.Args[1]

	switch command {
	case "snapshot":
		handleSnapshot(cfg, bs)
	case "list":
		handleList()
	case "restore":
		if len(os.Args) < 3 {
			fmt.Println("Usage: tracker restore <snapshotID>")
			os.Exit(1)
		}

		dryRun := false

		if len(os.Args) > 3 && os.Args[3] == "--dryrun" {
			dryRun = true
		}
		snapID := os.Args[2]
		handleRestore(snapID, bs, dryRun)
	case "diff":
		if len(os.Args) < 4 {
			fmt.Println("Usage: tracker restore <snapshotID-A> <snapshotID-B>")
			os.Exit(1)
		}

        var snapA, snapB string
        snapA = os.Args[2]
        snapB = os.Args[3]

        handleDiff(snapA, snapB)

	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleSnapshot(cfg *config.Config, bs *storage.BlobStore) {

	var paths []string = make([]string, 0, len(cfg.Dotfiles))

	for _, file := range cfg.Dotfiles {

		paths = append(paths, file.Path)
	}

	ss, err := snapshot.Create(paths, bs, config.GetSnaphotDir())
	if err != nil {
		log.Fatal(err)
	}

	for _, snap := range ss.Files {

		if snap.Error != "" {
			fmt.Printf("WARN: %s  ->  %s \n", snap.AbsolutePath, snap.Error)
		} else {
			fmt.Printf("OK: %s  ->  %s \n", snap.AbsolutePath, snap.Hash)
		}
	}
}

func handleList() {

	entries, err := os.ReadDir(config.GetSnaphotDir())
	if err != nil {
		log.Fatal(err)
	}

	if len(entries) == 0 {
		fmt.Println("No snapshots found")
		return
	}

	for _, entry := range entries {
		fmt.Printf(" %s \n", entry.Name())
	}
}

func handleRestore(snapID string, blobstore *storage.BlobStore, dryRun bool) {

	restores, err := snapshot.RestoreSnapshot(snapID, blobstore, dryRun)

	if err != nil {
		log.Fatal(err)
	}

	for _, r := range restores {
		if !r.Success {
			fmt.Printf("WARN: %s  ->  %s \n", r.Path, r.Error)
		} else {
			fmt.Printf("OK[restored]: %s  \n", r.Path)
		}
	}
}

func handleDiff(snapA, snapB string) {

	sA, err := snapshot.ReadSnapshot(snapA)
	if err != nil {
		log.Fatal(err)
	}

	sB, err := snapshot.ReadSnapshot(snapB)
	if err != nil {
		log.Fatal(err)
	}

    changes:= diff.DiffSnapshots(sA, sB)

    diff.PrettyPrint(changes)
}
