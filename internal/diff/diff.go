package diff

import (
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/richd0tcom/yadm/internal/snapshot"
)

type DiffStatus string

const (
	Added     DiffStatus = "added"
	Removed   DiffStatus = "removed"
	Modified  DiffStatus = "modified"
	Unchanged DiffStatus = "unchanged"
)

type FileDiff struct {
	Path    string
	Status  DiffStatus
	OldHash string
	NewHash string
}

func DiffSnapshots(snapA, snapB snapshot.Snapshot) []FileDiff {
	results := make([]FileDiff, 0)

	mapA := make(map[string]snapshot.FileEntry)

	for _, file := range snapA.Files {
		mapA[file.AbsolutePath] = file
	}

	mapB := make(map[string]snapshot.FileEntry)

	for _, file := range snapB.Files {
		mapB[file.AbsolutePath] = file
	}

	for path, fileA := range mapA {
		if _, ok := mapB[path]; !ok {
			results = append(results, FileDiff{
				Path:    path,
				Status:  Removed,
				OldHash: fileA.Hash,
			})
		} else if fileA.Hash != mapB[path].Hash {
			results = append(results, FileDiff{
				Path:    path,
				Status:  Modified,
				OldHash: fileA.Hash,
				NewHash: mapB[path].Hash,
			})
		} else {
			results = append(results, FileDiff{
				Path:    path,
				Status:  Unchanged,
				OldHash: fileA.Hash,
				NewHash: mapB[path].Hash,
			})
		}
	}

	for path := range mapB {
		if _, ok := mapA[path]; !ok {
			results = append(results, FileDiff{
				Path:    path,
				Status:  Added,
				NewHash: mapB[path].Hash,
			})
		}
	}

	return results
}


func PrettyPrint(diffs []FileDiff) {
    if len(diffs) == 0 {
        fmt.Println("No differences found.")
        return
    }

    // Sort by path for stable output. do we need this?
    sort.Slice(diffs, func(i, j int) bool {
        return diffs[i].Path < diffs[j].Path
    })

    // Define color styles
    added := color.New(color.FgGreen, color.Bold)
    removed := color.New(color.FgRed, color.Bold)
    modified := color.New(color.FgYellow, color.Bold)
    unchanged := color.New(color.FgHiBlack) // dim gray

    // print a summary header
    fmt.Printf("Summary: %d files changed\n\n", len(diffs))

    for _, d := range diffs {
        switch d.Status {
        case Added:
            added.Printf("[+] Added    ")
            fmt.Printf("%s  (new: %s)\n", d.Path, d.NewHash)
        case Removed:
            removed.Printf("[-] Removed  ")
            fmt.Printf("%s  (old: %s)\n", d.Path, d.OldHash)
        case Modified:
            modified.Printf("[*] Modified ")
            fmt.Printf("%s  (old: %s → new: %s)\n", d.Path, d.OldHash, d.NewHash)
        case Unchanged:
            unchanged.Printf("[ ] Unchanged")
            fmt.Printf(" %s  (%s)\n", d.Path, d.OldHash)
        }
    }
}


/*
TODO:
Group by status – print added, removed, modified, unchanged sections for better readability.

Use emojis – e.g., ✅, ❌, 🔄, ➖ for visual cues.

Suppress unchanged files – we may only care about changes; add a flag to hide Unchanged.

Colorize hash values – we can also color the hashes themselves or use a shorter representation.

Wrap paths – if paths are long, we might want to shorten them or align columns.

*/
