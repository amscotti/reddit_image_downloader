package main

import (
	"flag"
	"log"
	"os"
	"path"

	"github.com/amscotti/reddit_image_downloader/structs"
)

func main() {
	log.Print("Starting")
	var configFile string
	flag.StringVar(&configFile, "c", "config.json", "Location of configuration file to use")
	flag.Parse()
	log.Printf("Reading Config file at %s", configFile)

	var config structs.Config
	config.LoadConfig(configFile)
	log.Printf("Download path %s", config.DownloadPath)

	for _, r := range config.Reddits {
		log.Printf("Will be downloaded %s", r)
		outputPath := path.Join(config.DownloadPath, r)
		os.Mkdir(outputPath, 0777)

		var reddit structs.Reddit

		fileToDownload := reddit.GetReddits(r)

		for _, file := range fileToDownload {
			file.DownloadFile(outputPath)
		}
		log.Printf("Done with %s", r)
	}
}
