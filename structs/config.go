package structs

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//Config hold data on what Reddits to pull from
type Config struct {
	Subreddits   []string
	DownloadPath string          `json:"downloadPath"`
	FileExt      map[string]bool `json:"fileExt"`
}

// LoadConfig will load in a config from a path
func (c *Config) LoadConfig(confFile string) {
	rawConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(rawConf, &c)
}
