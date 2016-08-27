package structs

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//Config hold data on what Reddits to pull from
type Config struct {
	Reddits      []string
	DownloadPath string `json:"download_folder"`
}

// LoadConfig will load in a config from a path
func (c *Config) LoadConfig(confFile string) {
	rawConf, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(rawConf, &c)
}
