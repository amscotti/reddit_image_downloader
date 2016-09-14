package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	// My personal preference about package layout:
	// I'd rename 'structs' to something liked 'redditdl' (for lack of a better name)
	//
	// The main package (this file) could go under cmd/redditdl/main.go
	//
	// cmd/redditdl/main.go would import: "github.com/amscotti/redditdl", and then have
	// only a main() method which calls the 'redditdl' public API:
	//
	// See https://github.com/scottfrazer/maple for an example of this kind of layout
	//
	// redditdl's public API could be something like:
	//
	// redditdl.DownloadImages(subreddits []string) []string
	//
	// where the return value is a slice of paths to the downloaded images
	// the parallelism could be handled behind this public API, and maybe with
	// tuning flags:
	//
	// redditdl.DownloadImages(subreddits []string, maxConcurrent int) []string
	"github.com/amscotti/reddit_image_downloader/structs"
	"github.com/buger/jsonparser"
)

func genRedditChannel(reddits []string) <-chan string {
	out := make(chan string)
	go func() {
		for _, r := range reddits {
			out <- r
		}
		close(out)
	}()
	return out
}

// Personally, I'd have the parameter to this be a []string and get rid of genRedditChannel()
func genDownloadFileChannel(in <-chan string) <-chan structs.DownloadFile {
	fileExtToDownload := map[string]bool{".jpg": true, ".png": true, ".gif": true}
	out := make(chan structs.DownloadFile)
	go func() {
		for r := range in {
			client := &http.Client{}

			req, err := http.NewRequest("GET", fmt.Sprintf("http://www.reddit.com/r/%s.json", r), nil)
			if err != nil {
				log.Fatal(err)
			}

			req.Header.Set("User-Agent", "GoLang Img Downloadeder/0.1")

			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			// Not that it really matters in this case because of the log.Fatal
			defer resp.Body.Close()

			data, _ := ioutil.ReadAll(resp.Body)
			jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				url, _, _, _ := jsonparser.Get(value, "data", "url")
				formatedURL := html.UnescapeString(string(url))
				_, filename := filepath.Split(formatedURL)
				if fileExtToDownload[filepath.Ext(filename)] {
					out <- structs.DownloadFile{Filename: filename, Folder: r, URL: formatedURL}
				}
			}, "data", "children")
		}
		close(out)
	}()
	return out
}

func main() {
	log.Print("Starting")
	var configFile string
	flag.StringVar(&configFile, "c", "config.json", "Location of configuration file to use")
	flag.Parse()
	log.Printf("Reading Config file at %s", configFile)

	var config structs.Config
	config.LoadConfig(configFile)
	log.Printf("Download path %s", config.DownloadPath)

	var wg sync.WaitGroup
	for file := range genDownloadFileChannel(genRedditChannel(config.Reddits)) {
		wg.Add(1)
		go func(f structs.DownloadFile) {
			f.DownloadFile(config.DownloadPath)
			wg.Done()
		}(file)
	}
	wg.Wait()
	log.Print("Done")
}
