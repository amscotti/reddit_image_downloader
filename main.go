package main

import (
	"flag"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/amscotti/reddit_image_downloader/structs"
	"github.com/buger/jsonparser"
)

func parseSubreddits(subreddits <-chan string, fileExtToDownload map[string]bool, files chan<- structs.DownloadFile, wg *sync.WaitGroup) {
	client := &http.Client{}
	for subreddit := range subreddits {
		log.Printf("Parsing %s", subreddit)

		req, err := http.NewRequest("GET", fmt.Sprintf("http://www.reddit.com/r/%s.json", subreddit), nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("User-Agent", "GoLang Img Downloadeder/0.1")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		data, _ := ioutil.ReadAll(resp.Body)
		jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			url, _, _, _ := jsonparser.Get(value, "data", "url")
			formatedURL := html.UnescapeString(string(url))
			_, filename := filepath.Split(formatedURL)
			if fileExtToDownload[filepath.Ext(filename)] {
				wg.Add(1)
				files <- structs.DownloadFile{Filename: filename, Folder: subreddit, URL: formatedURL}
			}
		}, "data", "children")

		resp.Body.Close()
		wg.Done()
	}
}

func downloadFiles(files <-chan structs.DownloadFile, downloadPath string, wg *sync.WaitGroup) {
	for file := range files {
		file.DownloadFile(downloadPath)
		wg.Done()
	}
}

func main() {
	var configFile string

	flag.StringVar(&configFile, "c", "config.json", "Location of configuration file to use")
	flag.Parse()

	log.Print("Starting")

	log.Printf("Reading Config file at %s", configFile)

	var config structs.Config
	config.LoadConfig(configFile)
	log.Printf("Download path %s", config.DownloadPath)

	var wg sync.WaitGroup

	subreddits := make(chan string, 100)
	files := make(chan structs.DownloadFile, 1000)

	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < runtime.NumCPU(); i++ {
		go parseSubreddits(subreddits, config.FileExt, files, &wg)
		go downloadFiles(files, config.DownloadPath, &wg)
	}

	for _, subreddit := range config.Subreddits {
		wg.Add(1)
		subreddits <- subreddit
	}

	wg.Wait()

	log.Print("Finish")
}
