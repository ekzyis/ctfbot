package types

import (
	"encoding/xml"
	"time"
)

type RssItem struct {
	Guid        string    `xml:"guid"`
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     RssDate   `xml:"pubDate"`
	Author      RssAuthor `xml:"author"`
}

type RssChannel struct {
	Title         string    `xml:"title"`
	Description   string    `xml:"description"`
	Link          string    `xml:"link"`
	Items         []RssItem `xml:"item"`
	LastBuildDate RssDate   `xml:"lastBuildDate"`
}

type Rss struct {
	Channel RssChannel `xml:"channel"`
}

type RssDate struct {
	time.Time
}

type RssAuthor struct {
	Name string `xml:"name"`
}

func (c *RssDate) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	dateFormat := "Mon, 02 Jan 2006 15:04:05 GMT"
	d.DecodeElement(&v, &start)
	parse, err := time.Parse(dateFormat, v)
	if err != nil {
		return err
	}
	*c = RssDate{parse}
	return nil
}
