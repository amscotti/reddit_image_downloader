package structs

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

// DownloadFile holds the details for the file to download
type DownloadFile struct {
	Filename string
	Folder   string
	URL      string
}

func (f *DownloadFile) outFile(directory string) string {
	return path.Join(directory, f.Folder, f.Filename)
}

// DownloadFile is used to download the file to the local system
// Uses streaming to avoid loading the entire file into memory
func (f *DownloadFile) DownloadFile(ctx context.Context, client *http.Client, directory string) error {
	// Check if file already exists
	if _, err := os.Stat(f.outFile(directory)); err == nil {
		// File already exists
		return nil
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(path.Join(directory, f.Folder), 0777); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the output file
	output, err := os.Create(f.outFile(directory))
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer output.Close()

	// Create request with context for cancelation
	req, err := http.NewRequestWithContext(ctx, "GET", f.URL, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	// Add useful headers
	req.Header.Set("User-Agent", "Reddit Image Downloader/1.0")

	// Execute the request
	response, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not download file: %w", err)
	}
	defer response.Body.Close()

	// Check for successful response
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response: %s", response.Status)
	}

	// Stream directly from response body to file (no memory buffering)
	bytesWritten, err := io.Copy(output, response.Body)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	// Log download success with size information
	sizeKB := float64(bytesWritten) / 1024.0
	log.Printf("Downloaded %s file %s (%.2f KB)", f.Folder, f.Filename, sizeKB)
	return nil
}
