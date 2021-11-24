package structs

import (
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
func (f *DownloadFile) DownloadFile(directory string) bool {
	if _, err := os.Stat(f.outFile(directory)); err == nil {
		return true
	}
	os.Mkdir(path.Join(directory, f.Folder), 0777)

	output, err := os.Create(f.outFile(directory))
	if err != nil {
		log.Fatal("Could not create output file ", err)
		return false
	}
	defer output.Close()

	response, err := http.Get(f.URL)
	if err != nil {
		log.Fatal("Could not download file ", err)
		return false
	}
	defer response.Body.Close()

	_, err = io.Copy(output, response.Body)
	if err != nil {
		log.Fatal("Error writing file ", err)
		return false
	}
	log.Printf("Downloaded %s file %s", f.Folder, f.Filename)

	return true
}
