package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/BurntSushi/toml"
)

const (
	DEFAULT_CONFIG_PATH         = "~/.dotfile-tracker/config.toml"
	DEFAULT_TRACKER_ROOT = "~/.dotfile-tracker/"
	SNAPSHOT_FOLDER      = "snapshots"
	BLOB_FOLDER          = "blobs"
)

var (
	snapshotDir string

)

//go:embed sample-config.toml
var configBytes []byte

type Config struct {
	TrackerRoot string    `toml:"tracker_root"`
	Dotfiles    []Dotfile `toml:"dotfiles"`
}

type Dotfile struct {
	Path string `toml:"path"`
}

func LoadConfig(path string) (*Config, error) {
	
	fullPath:= expandTilde(path)


	if _, err := os.Stat(fullPath); os.IsNotExist(err) {

		//create new file
		err := createDefaultConfig(fullPath)

		if err != nil {
			return &Config{}, err
		}

		fmt.Printf("Created default config at: %s\n", fullPath)
		fmt.Println("Please edit it to specify your dotfiles.")

		//create default snapshot folder
		snapshotDir := filepath.Join(expandTilde(DEFAULT_TRACKER_ROOT), SNAPSHOT_FOLDER)
		if err := os.MkdirAll(snapshotDir, 0755); err != nil {
			return &Config{}, err
		}

		//create default blob folder
		blobDir := filepath.Join(expandTilde(DEFAULT_TRACKER_ROOT), BLOB_FOLDER)
		if err := os.MkdirAll(blobDir, 0755); err != nil {
			return &Config{}, err
		}

		//TODO: uninstall should remove all the created directories

	}

	var config Config

	_, err := toml.DecodeFile(fullPath, &config)
	if err != nil {
		return &Config{}, err
	}

	// Expand tildes in all paths
	if err := config.expandPaths(); err != nil {
		return nil, fmt.Errorf("failed to expand paths: %w", err)
	}

	snapshotDir = filepath.Join(config.TrackerRoot, SNAPSHOT_FOLDER)

	return &config, nil

}

// expandPaths expands tildes in TrackerRoot and all Dotfile paths
func (c *Config) expandPaths() error {
	c.TrackerRoot = expandTilde(c.TrackerRoot)

	for i := range c.Dotfiles {
		c.Dotfiles[i].Path = expandTilde(c.Dotfiles[i].Path)
	}

	return nil
}

func createDefaultConfig(path string) error {

	dir:= filepath.Dir(path)

	if err:= os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(configBytes), 0644)
}

func expandTilde(path string) string {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic("could not determine home directory")
	}

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(
			homeDir,
			path[2:],
		)
	}

	//TODO: Handle ~username

	return path
}


func GetSnaphotDir() string {
	return snapshotDir
}

