package rss

import (
	"net/http"
	"bufio"
	"log"
	"fmt"
	"io/ioutil"
	"bytes"
	"gopod/opml"
	"path/filepath"
	"os"
	"strings"
	"time"
)

func contentTypeToExt(podcastItem Item) (ext string, err error) {

	contentType := podcastItem.Enclosure.Type

	if !strings.Contains(contentType, "/") {
		contentType = podcastItem.Media.Type
	}

	if !strings.Contains(contentType, "/") {
		return "", fmt.Errorf("Unable to determine the extension of the file")
	}

	parts := strings.Split(contentType, "/")

	switch parts[1] {
	case "mpeg" : return "mp3", nil
	case "mp3" : return "mp3", nil
	}

	return "", fmt.Errorf("Unable to figure out extension %s", parts[1])
}

func parseTime(dateString string) (date time.Time, err error) {
	formats := []string {time.RFC1123Z, time.RFC1123, time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z}
	for _, format := range formats {
		date, err = time.Parse(format, dateString)
		if err == nil {
			return date, nil
		}
	}

	return date, nil;
}

func needsUpdate(item Item, outline *opml.OpmlOutline) bool {
	pubDate, err := parseTime(item.PubDate)
	if err != nil {
		panic(err.Error())
	}
	lastUpdate, err := parseTime(outline.LastUpdate)
	if err != nil {
		return true
	}

	return lastUpdate.Before(pubDate)
}

func Download(head opml.OpmlHead, outline *opml.OpmlOutline) (numEpisodesDownloaded int, err error) {
	log.Printf("Downloading Rss feed from %q\n", outline.XmlUrl)
	resp, err := http.Get(outline.XmlUrl)
	if err != nil {
		return 0, err
	}

	value := resp.Body
	defer value.Close()

	rssFeedText, err := ioutil.ReadAll(value)
	rssModel, err := ParseRss(bytes.NewReader(rssFeedText))

	if err != nil {
		return 0, err
	}

	podcastItems := rssModel.Channel.Items

	if len(podcastItems) == 0 {
		return 0, nil
	}

	if !needsUpdate(podcastItems[0], outline) {
		log.Printf("Podcast %q is update to date", rssModel.Channel.Title)
		return 0, nil
	}

	podcastDir := filepath.Join(head.DownloadDir, rssModel.Channel.Title)

	err = os.MkdirAll(podcastDir, os.ModeDir)

	if err != nil {
		return 0, fmt.Errorf("Failed to make podcast directory %q", podcastDir)
	}

	downloadCount := 0
	keep := outline.Keep
	if keep == 0 {
		keep = head.DefaultKeep
	}
	if keep == 0 {
		keep = 1
	}

	for i := 0; i < keep; i++ {
		podcastItem := podcastItems[0]
		var postcastUrl string
		if postcastUrl = podcastItem.Enclosure.Url; len(postcastUrl) > 0 {
			postcastUrl = postcastUrl
		} else if postcastUrl = podcastItem.Media.Url; len(postcastUrl) > 0 {
			postcastUrl = postcastUrl
		} else {
			return 0, fmt.Errorf("No url was found for this podcast: %q\n", rssModel.Channel.Title)
		}

		log.Printf("Downloading podcast: %q from url %q\n", podcastItem.Title, postcastUrl)
		resp, err = http.Get(postcastUrl)

		if err != nil {
			log.Printf("An error occurred downloading podcast: '%v'\n\n", err)
			return downloadCount, err
		}

		podcast := resp.Body
		defer podcast.Close()

		ext, err := contentTypeToExt(podcastItem)

		if err != nil {
			return downloadCount, err
		}

		dest, err := os.Create(filepath.Join(podcastDir, podcastItem.Title+"."+ext))
		defer dest.Close()

		if err != nil {
			return downloadCount, fmt.Errorf("Unable to create file for podcast: %v", podcastItem)
		}

		n, err := bufio.NewReader(podcast).WriteTo(dest)
		log.Printf("Downloaded file with size: %d to %s\n", n, dest.Name())

		if err != nil {
			return downloadCount, err
		}

		downloadCount++
	}

	outline.LastUpdate = podcastItems[0].PubDate

	return downloadCount, nil
}

