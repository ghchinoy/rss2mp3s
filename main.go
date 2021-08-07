package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/ghchinoy/rss2mp3s/readfeed"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("please provide an RSS URL")
		fmt.Println("ex. rss2mp3s https://www.robotread.me/podcast/index.xml")
		os.Exit(1)
	}
	feedurl := os.Args[1]
	if feedurl == "" {
		feedurl = "https://www.robotread.me/podcast/index.xml"
	}
	u, err := url.Parse(feedurl)
	if err != nil {
		log.Printf("couldn't read the url: %v", err)
		os.Exit(1)
	}
	var rss readfeed.RSS
	err = rss.Read(u)
	if err != nil {
		log.Printf("couldn't parse url into rss feed: %v", err)
		os.Exit(1)
	}
	log.Printf("ok, rss (%d): '%s'", rss.Length, rss.Channel.Title)
	var waitGroup sync.WaitGroup
	log.Printf("Found %d items", len(rss.Channel.Items))
	for _, items := range rss.Channel.Items {
		//log.Printf("beginning download of %s", items.Title)
		waitGroup.Add(1)
		go downloadEnclosure(items.Title, items.Enclosure.URL, &waitGroup)
	}
	waitGroup.Wait()
}

// downloadEnclosure downloads the target enclosure URL to a local file
func downloadEnclosure(title, enclosureURL string, wg *sync.WaitGroup) {
	defer wg.Done()

	// create a name for the file from the URL
	parts := strings.Split(enclosureURL, "/")
	filename := parts[len(parts)-1:][0]
	var writer io.WriteCloser
	writer, err := os.Create(filename)
	if err != nil {
		log.Printf("unable to create output file '%s': %v", filename, err)
		return
	}
	defer writer.Close()

	r, err := http.Get(enclosureURL)
	if err != nil {
		fmt.Println(err)
		return
	}
	io.Copy(writer, r.Body)
	if err = r.Body.Close(); err != nil {
		fmt.Println(err)
	}
	log.Printf("Downloaded '%s' as %s", title, filename)
}
