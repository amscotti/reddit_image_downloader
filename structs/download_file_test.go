package structs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOutFile(t *testing.T) {
	file := DownloadFile{
		Filename: "test.jpg",
		Folder:   "testfolder",
		URL:      "http://example.com/test.jpg",
	}

	result := file.outFile("/tmp")
	expected := filepath.Join("/tmp", "testfolder", "test.jpg")

	if result != expected {
		t.Errorf("Expected path %s, got %s", expected, result)
	}
}

func TestDownloadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set up a test server that returns a simple image
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test image content"))
	}))
	defer server.Close()

	// Create the download file struct
	file := DownloadFile{
		Filename: "test.jpg",
		Folder:   "testfolder",
		URL:      server.URL,
	}

	// Create context and client for test
	ctx := context.Background()
	client := &http.Client{}

	// Test downloading the file
	err := file.DownloadFile(ctx, client, tempDir)
	if err != nil {
		t.Fatalf("Failed to download file: %v", err)
	}

	// Check if the file exists
	expectedPath := filepath.Join(tempDir, "testfolder", "test.jpg")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("File was not downloaded to expected path: %s", expectedPath)
	}

	// Check the content of the file
	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "test image content" {
		t.Errorf("File content doesn't match expected content")
	}

	// Test downloading a file that already exists (should return nil)
	err = file.DownloadFile(ctx, client, tempDir)
	if err != nil {
		t.Errorf("Expected nil error when file already exists, got: %v", err)
	}
}
