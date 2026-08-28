package client

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) GetRssFeed() (*t.Rss, error) {
	url := fmt.Sprintf("%s/rss", c.BaseUrl)
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("error fetching RSS feed: %w", err)
		log.Println(err)
		return nil, err
	}
	defer resp.Body.Close()

	var rss t.Rss
	err = xml.NewDecoder(resp.Body).Decode(&rss)
	if err != nil {
		err = fmt.Errorf("error decoding RSS feed XML: %w", err)
		return nil, err
	}

	return &rss, nil
}
