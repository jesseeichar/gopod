package opml

import (
	"io"
	"encoding/xml"
	"time"
	"bytes"
)

type OpmlHead struct {
	DateCreated string `xml:"dateCreated"`
}

type OpmlOutline struct {
	XmlUrl string `xml:"xmlUrl,attr"`
}

type OpmlBody struct {
	Outline []OpmlOutline `xml:"outline"`
}
type Opml struct {
	Head OpmlHead `xml:"head"`
	Body OpmlBody  `xml:"body"`
	Version string  `xml:"version,attr"`
}

func New() Opml {
	model := Opml{}
	model.Head.DateCreated = time.Now().Format(time.UnixDate)
	model.Version = "2.0"
	return model
}

func (opml *Opml) Write(writer io.Writer) (int, error) {
	bytes, err := xml.MarshalIndent(opml, "", "  ")
	if err != nil {
		return -1, err
	}

	return writer.Write(bytes)
}

func (opml Opml) String() string {
	var buffer = bytes.Buffer{}
	if _, err := opml.Write(&buffer); err != nil {
		panic(err)
	}

	return buffer.String()

}
