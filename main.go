package main

import (
	"flag"
	"log"
	"os"
	"path"
	"sync"

	"github.com/amscotti/reddit_image_downloader/structs"
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

func genDownloadFileChannel(in <-chan string) <-chan structs.DownloadFile {
	out := make(chan structs.DownloadFile)
	go func() {
		for r := range in {
			var reddit structs.Reddit
			reddit.GetReddits(r, out)
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
		outputPath := path.Join(config.DownloadPath, file.Folder)
		os.Mkdir(outputPath, 0777)
		wg.Add(1)
		f := file
		go func() {
			f.DownloadFile(outputPath)
			wg.Done()
		}()
	}
	wg.Wait()
}
