package rss

import (
	"bytes"
	"encoding/xml"
	"io"
)

type Rss struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title         string `xml:"title"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate"`
	Items         []Item `xml:"item"`
}

type Item struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	PubDate     string    `xml:"pubDate"`
	Category    string    `xml:"category"`
	Guid        string    `xml:"guid"`
	Enclosure   Enclosure `xml:"enclosure"`
	Media       Media     `xml:"http://search.yahoo.com/mrss/ content"`
}

type Media struct {
	Url  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}
type Enclosure struct {
	Url    string `xml:"url,attr"`
	Type   string `xml:"type,attr"`
	Length string `xml:"length,attr"`
}

func (rss *Rss) Write(writer io.Writer) (int, error) {
	return write(rss, writer)
}
func (rss Rss) String() string {
	return toString(&rss)
}

func (rss *Channel) Write(writer io.Writer) (int, error) {
	return write(rss, writer)
}
func (rss Channel) String() string {
	return toString(&rss)
}

func (rss *Item) Write(writer io.Writer) (int, error) {
	return write(rss, writer)
}
func (rss Item) String() string {
	return toString(&rss)
}

func (rss *Media) Write(writer io.Writer) (int, error) {
	return write(rss, writer)
}
func (rss Media) String() string {
	return toString(&rss)
}

type writeable interface {
	Write(io.Writer) (int, error)
}

func toString(el writeable) string {
	var buffer = bytes.Buffer{}
	if _, err := el.Write(&buffer); err != nil {
		panic(err)
	}

	return buffer.String()
}

func write(el interface{}, writer io.Writer) (int, error) {
	bytes, err := xml.MarshalIndent(el, "", "  ")
	if err != nil {
		return -1, err
	}

	return writer.Write(bytes)
}
