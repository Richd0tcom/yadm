package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/BurntSushi/toml"
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
		err := createDefaultConfig(path)

		if err != nil {
			return &Config{}, err
		}

		fmt.Printf("Created default config at: %s\n", fullPath)
		fmt.Println("Please edit it to specify your dotfiles.")

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
			path[:2],
			homeDir,
		)
	}

	//TODO: Handle ~username

	return path
}


