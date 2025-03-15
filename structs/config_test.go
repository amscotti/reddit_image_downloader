package structs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary test config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.toml")

	configData := `
# Reddit Image Downloader configuration
subreddits = ["test1", "test2"]
downloadPath = "/tmp/test"

[fileExt]
".jpg" = true
".png" = true
`

	if err := os.WriteFile(configPath, []byte(configData), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Test loading the config
	var config Config
	err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config was loaded correctly
	if len(config.Subreddits) != 2 {
		t.Errorf("Expected 2 subreddits, got %d", len(config.Subreddits))
	}

	if config.Subreddits[0] != "test1" || config.Subreddits[1] != "test2" {
		t.Errorf("Subreddits not loaded correctly, got %v", config.Subreddits)
	}

	if config.DownloadPath != "/tmp/test" {
		t.Errorf("Expected download path /tmp/test, got %s", config.DownloadPath)
	}

	if len(config.FileExt) != 2 {
		t.Errorf("Expected 2 file extensions, got %d", len(config.FileExt))
	}

	if !config.FileExt[".jpg"] || !config.FileExt[".png"] {
		t.Errorf("File extensions not loaded correctly, got %v", config.FileExt)
	}
}

func TestLoadConfigInvalidPath(t *testing.T) {
	var config Config
	err := config.LoadConfig("non_existent_file.toml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestLoadConfigInvalidTOML(t *testing.T) {
	// Create a temporary test config file with invalid TOML
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid_config.toml")

	invalidTOML := `
# Invalid TOML file
subreddits = ["test1", "test2"
downloadPath = "/tmp/test"

[fileExt]
".jpg" = true
".png" = true
` // Missing closing bracket

	if err := os.WriteFile(configPath, []byte(invalidTOML), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	var config Config
	err := config.LoadConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid TOML, got nil")
	}
}
