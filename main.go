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

	"github.com/amscotti/reddit_image_downloader/structs"
	"github.com/buger/jsonparser"
)

func genRedditChannel(subreddits []string) <-chan string {
	out := make(chan string)
	go func() {
		for _, r := range subreddits {
			out <- r
		}
		close(out)
	}()
	return out
}

func genDownloadFileChannel(in <-chan string, fileExtToDownload map[string]bool) <-chan structs.DownloadFile {
	out := make(chan structs.DownloadFile)
	go func() {
		client := &http.Client{}
		for r := range in {
			req, err := http.NewRequest("GET", fmt.Sprintf("http://www.reddit.com/r/%s.json", r), nil)
			if err != nil {
				log.Fatal(err)
			}

			req.Header.Set("User-Agent", "GoLang Img Downloadeder/0.1")

			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}
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
	var configFile string

	flag.StringVar(&configFile, "c", "config.json", "Location of configuration file to use")
	flag.Parse()

	log.Print("Starting")

	log.Printf("Reading Config file at %s", configFile)

	var config structs.Config
	config.LoadConfig(configFile)
	log.Printf("Download path %s", config.DownloadPath)

	var wg sync.WaitGroup
	for file := range genDownloadFileChannel(genRedditChannel(config.Subreddits), config.FileExt) {
		wg.Add(1)
		f := file
		go func() {
			f.DownloadFile(config.DownloadPath)
			wg.Done()
		}()
	}
	wg.Wait()

	log.Print("Done")
}
