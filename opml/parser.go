package opml

import (
	"io"
	"encoding/xml"
)

func ParseOpml(reader io.Reader) (*Opml, error) {
	decoder := xml.NewDecoder(reader)
	var opml Opml
	if err := decoder.Decode(&opml); err != nil {
		return nil, err
	}

	return &opml, nil
}
