package main

import (
	"log"
	"path/filepath"

	"github.com/richd0tcom/yadm/interrnal/config"
	"github.com/richd0tcom/yadm/interrnal/snapshot"
	"github.com/richd0tcom/yadm/interrnal/storage"

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

	snapshotDir := filepath.Join(trackerRoot, config.SNAPSHOT_FOLDER)

	//TODO: implement flags and robust argument parsing https://cobra.dev
	command := os.Args[1]

	switch command {
	case "snapshot":
		handleSnapshot(cfg, bs, snapshotDir)
	case "list":
		handleList(snapshotDir)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func handleSnapshot(cfg *config.Config, bs *storage.BlobStore, snapshotDir string) {

	var paths []string = make([]string, 0, len(cfg.Dotfiles))
    
	for _, file := range cfg.Dotfiles {

		paths = append(paths, file.Path)
	}

	ss, err := snapshot.Create(paths, bs, snapshotDir)
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

func handleList(snapshotDir string) {

    entries, err := os.ReadDir(snapshotDir)
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
