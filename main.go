package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

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
	for _, items := range rss.Channel.Items {
		log.Printf("%s", items.Title)
		log.Printf("\t%v", items.Enclosure)
	}
}
