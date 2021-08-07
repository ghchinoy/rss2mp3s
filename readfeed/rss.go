package readfeed

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
)

// RSS is the RSS object
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Length  int
	Channel Channel `xml:"channel"`
}

// Channel is the RSS:channel element
type Channel struct {
	XMLName xml.Name `xml:"channel"`
	Title   string   `xml:"title"`
	Link    url.URL  `xml:"link"`
	Items   []Item   `xml:"item"`
}

// Item is the RSS:item element
type Item struct {
	XMLName     xml.Name  `xml:"item"`
	Title       string    `xml:"title"`
	Link        url.URL   `xml:"link"`
	GUID        string    `xml:"guid"`
	PubDate     string    `xml:"pubDate"`
	Author      string    `xml:"author"`
	Description string    `xml:"description"`
	Enclosure   Enclosure `xml:"enclosure"`
}

// Enclosure is an RSS:item's media element
type Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"url,attr"`
	Length  int      `xml:"length,attr"`
	Type    string   `xml:"type,attr"`
}

func (rss *RSS) Read(url *url.URL) error {
	res, err := http.Get(url.String())
	if err != nil {
		return err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		return err
	}
	rss.Length = len(body)

	return nil
}
