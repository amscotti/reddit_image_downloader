# reddit_image_downloader

This is a command line tool that uses a config file to download images from a list of subreddits.

The idea for this project is from a friend, [icmpzero](https://github.com/icmpzero) and was used to help us better understand writing in Go.

## Building

To build, run
```bash
$ go build
```

## Usage

```
Usage of ./reddit_image_downloader:
  -c string
        Location of configuration file to use (default "config.json")
```

## Config 
* subreddits: List of Subreddits to look for images in
* fileExt: What file type to download
* downloadPath: Path to download files into

### Example

```JSON
{
  "subreddits": [
    "Wallpapers",
    "battlestations",
    "Aww",
    "Beerwithaview",
    "OldSchoolCool",
    "TheWayWeWere",
    "itookapicture"
  ],
  "fileExt": {
    ".jpg": true,
    ".png": true,
    ".gif": true
  },
  "downloadPath": "/home/ascotti/Downloads"
}
```