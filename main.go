package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/amscotti/reddit_image_downloader/structs"
	jsoniter "github.com/json-iterator/go"
)

// createHTTPClient returns a configured HTTP client optimized for connection reuse
func createHTTPClient() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		ForceAttemptHTTP2:   true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}

func parseSubreddits(ctx context.Context, httpClient *http.Client, subreddits <-chan string, fileExtToDownload map[string]bool, files chan<- structs.DownloadFile, wg *sync.WaitGroup) {
	for subreddit := range subreddits {
		log.Printf("Parsing %s", subreddit)

		req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://www.reddit.com/r/%s.json", subreddit), nil)
		if err != nil {
			log.Printf("Error creating request for %s: %v", subreddit, err)
			wg.Done()
			continue
		}

		req.Header.Set("User-Agent", "Reddit Image Downloader/1.0")

		resp, err := httpClient.Do(req)
		if err != nil {
			log.Printf("Error fetching %s: %v", subreddit, err)
			wg.Done()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Error response from Reddit for %s: %s", subreddit, resp.Status)
			resp.Body.Close()
			wg.Done()
			continue
		}

		// Parse the JSON response
		var redditResponse struct {
			Data struct {
				Children []struct {
					Data struct {
						URL string `json:"url"`
					} `json:"data"`
				} `json:"children"`
			} `json:"data"`
		}

		json := jsoniter.ConfigCompatibleWithStandardLibrary
		if err := json.NewDecoder(resp.Body).Decode(&redditResponse); err != nil {
			log.Printf("Error decoding JSON for %s: %v", subreddit, err)
			resp.Body.Close()
			wg.Done()
			continue
		}

		// Process each post
		for _, child := range redditResponse.Data.Children {
			formatedURL := html.UnescapeString(child.Data.URL)
			_, filename := filepath.Split(formatedURL)
			if fileExtToDownload[filepath.Ext(filename)] {
				wg.Add(1)
				files <- structs.DownloadFile{Filename: filename, Folder: subreddit, URL: formatedURL}
			}
		}

		resp.Body.Close()
		wg.Done()
	}
}

func downloadFiles(ctx context.Context, httpClient *http.Client, files <-chan structs.DownloadFile, downloadPath string, wg *sync.WaitGroup) {
	for file := range files {
		select {
		case <-ctx.Done():
			wg.Done()
			return
		default:
			if err := file.DownloadFile(ctx, httpClient, downloadPath); err != nil {
				log.Printf("Error downloading %s: %v", file.URL, err)
			}
			wg.Done()
		}
	}
}

func downloadImagesFromSubreddits(ctx context.Context, subreddits []string, downloadPath string, fileExtToDownload map[string]bool) error {
	var wg sync.WaitGroup
	countOfSubreddits := len(subreddits)

	if countOfSubreddits == 0 {
		return fmt.Errorf("no subreddits specified")
	}

	// Ensure download directory exists
	if err := os.MkdirAll(downloadPath, 0777); err != nil {
		return fmt.Errorf("failed to create download directory: %w", err)
	}

	// Create a shared HTTP client for connection reuse
	httpClient := createHTTPClient()

	wg.Add(countOfSubreddits)

	subredditsToRead := make(chan string, countOfSubreddits)
	files := make(chan structs.DownloadFile, 1000)

	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go parseSubreddits(ctx, httpClient, subredditsToRead, fileExtToDownload, files, &wg)
		go downloadFiles(ctx, httpClient, files, downloadPath, &wg)
	}

	for _, subreddit := range subreddits {
		subredditsToRead <- subreddit
	}
	close(subredditsToRead)

	// Set up a channel to signal the completion of the wait group
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// Wait for either completion or context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func main() {
	var configFile string
	var timeout int
	var config structs.Config

	flag.StringVar(&configFile, "c", "config.toml", "Location of configuration file to use")
	flag.IntVar(&timeout, "timeout", 300, "Timeout in seconds for the entire operation")
	flag.Parse()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	log.Print("Starting")

	log.Printf("Reading config file at %s", configFile)
	if err := config.LoadConfig(configFile); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Download path: %s", config.DownloadPath)
	log.Printf("Subreddits to process: %d", len(config.Subreddits))
	log.Printf("File extensions to download: %v", config.FileExt)

	err := downloadImagesFromSubreddits(ctx, config.Subreddits, config.DownloadPath, config.FileExt)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	log.Print("Finish")
}
