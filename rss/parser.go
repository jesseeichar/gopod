package rss

import (
	"encoding/xml"
	"io"
)

func ParseRss(reader io.Reader) (*Rss, error) {
	decoder := xml.NewDecoder(reader)
	var rss Rss
	if err := decoder.Decode(&rss); err != nil {
		return nil, err
	}

	return &rss, nil
}
