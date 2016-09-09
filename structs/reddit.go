package structs

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"path/filepath"
)

type Reddit struct {
	Data struct {
		Children []struct {
			Data Item
		}
	}
}

type Item struct {
	Title    string
	URL      string
	Comments int `json:"num_comments"`
}

func (i *Item) GetFilename() string {
	_, filename := filepath.Split(i.URL)
	return filename
}

func (i *Item) GetFilenameExt() string {
	return filepath.Ext(i.GetFilename())
}

func (i *Item) GetDownloadFile(folder string) DownloadFile {
	return DownloadFile{Filename: i.GetFilename(), Folder: folder, URL: html.UnescapeString(i.URL)}
}

func (r *Reddit) GetReddits(subreddit string, out chan DownloadFile) {
	fileExtToDownload := map[string]bool{".jpg": true, ".png": true, ".gif": true}
	getURL := fmt.Sprintf("http://www.reddit.com/r/%s.json", subreddit)
	client := &http.Client{}

	req, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("User-Agent", "GoLang Img Downloadeder/0.1")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = json.NewDecoder(resp.Body).Decode(r)
	if err != nil {
		log.Print(err)
	}

	rPosts := make([]Item, len(r.Data.Children))
	for i, child := range r.Data.Children {
		rPosts[i] = child.Data
	}

	for _, item := range rPosts {
		if fileExtToDownload[item.GetFilenameExt()] {
			out <- item.GetDownloadFile(subreddit)
		}
	}
}
