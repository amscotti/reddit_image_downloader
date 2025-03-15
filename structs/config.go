package structs

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds data on what subreddits to pull from
type Config struct {
	Subreddits   []string        `toml:"subreddits"`
	DownloadPath string          `toml:"downloadPath"`
	FileExt      map[string]bool `toml:"fileExt"`
}

// LoadConfig will load a config from a path
func (c *Config) LoadConfig(confFile string) error {
	_, err := toml.DecodeFile(confFile, c)
	if err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// Validate checks if the config is valid
func (c *Config) Validate() error {
	if len(c.Subreddits) == 0 {
		return fmt.Errorf("no subreddits specified")
	}

	if c.DownloadPath == "" {
		return fmt.Errorf("download path is required")
	}

	if len(c.FileExt) == 0 {
		return fmt.Errorf("no file extensions specified")
	}

	// Check if download path exists or can be created
	info, err := os.Stat(c.DownloadPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(c.DownloadPath, 0777); err != nil {
				return fmt.Errorf("download path does not exist and cannot be created: %w", err)
			}
		} else {
			return fmt.Errorf("error accessing download path: %w", err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("download path is not a directory")
	}

	return nil
}
