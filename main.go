package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ghchinoy/rss2mp3s/readfeed"
)

func main() {
	max := flag.Int("max", 0, "Max number of episodes to download. Default is all episodes.")
	retries := flag.Int("retries", 0, "Maximum number of retries for a failed download. Default is 0.")
	parallel := flag.Int("parallel", 1, "Set the download parallelism. Default is 1.")
	uri := flag.String("rss", "", "rss feed URL")
	flag.Parse()

	if *parallel < 1 {
		log.Fatalf("Invalid parallelism value: %d", *parallel)
	}

	feedUrl, err := url.Parse(*uri)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return
	}

	var rss readfeed.RSS
	err = rss.Read(feedUrl)
	if err != nil {
		log.Printf("couldn't parse url into rss feed: %v", err)
		os.Exit(1)
	}

	log.Printf("ok, rss (%d): '%s'", rss.Length, rss.Channel.Title)

	limitChan := make(chan bool, *parallel)
	var waitGroup sync.WaitGroup

	if *max > 0 {
		log.Printf("Found %d items but downloading %d", len(rss.Channel.Items), *max)
	} else {
		log.Printf("Downloading all %d episodes.", len(rss.Channel.Items))
	}

	for cnt, items := range rss.Channel.Items {
		if *max > 0 && cnt >= *max {
			break
		}
		waitGroup.Add(1)
		limitChan <- true

		go func(title, enclosureURL string) {
			downloadEnclosure(title, enclosureURL, *retries, 0)
			<-limitChan
			waitGroup.Done()
		}(items.Title, items.Enclosure.URL)
	}

	waitGroup.Wait()
}

// downloadEnclosure downloads the target enclosure URL to a local file
func downloadEnclosure(title, enclosureURL string, retry int, attempt int) {
	title = strings.TrimSpace(title)

	if attempt > retry {
		log.Fatalf("Too many retries for title %s. Aborting!", title)
		os.Exit(1)
	}

	// create a filename from the track title and suffix
	parts := strings.Split(enclosureURL, "/")
	extFilename := parts[len(parts)-1:][0]
	filename := title + filepath.Ext(extFilename)

	var writer io.WriteCloser
	writer, err := os.Create(filename)
	if err != nil {
		if attempt < retry {
			log.Printf("Error downloading title %s: %v. Retrying...", title, err)
			downloadEnclosure(title, enclosureURL, retry, attempt+1)
		} else {
			panic(err)
		}
		return
	}
	defer writer.Close()

	r, err := http.Get(enclosureURL)
	if err != nil || r.StatusCode != http.StatusOK {
		if attempt < retry {
			log.Printf("Error downloading title %s: %v. Retrying...", title, err)
			downloadEnclosure(title, enclosureURL, retry, attempt+1)
		} else {
			panic(err)
		}
		return
	}

	_, err = io.Copy(writer, r.Body)
	if errClose := r.Body.Close(); err != nil || errClose != nil {
		if attempt < retry {
			log.Printf("Error downloading title %s: %v. Retrying...", title, err)
			downloadEnclosure(title, enclosureURL, retry, attempt+1)
		} else {
			panic(err)
		}
		return
	}

	log.Printf("Downloaded '%s' as %s", title, filename)
}
