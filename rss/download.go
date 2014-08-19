package rss

import (
	"bufio"
	"bytes"
	"fmt"
	"gopod/opml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"regexp"
)

type ContentType string

const (
	audio ContentType = "audio"
	video             = "video"
	unknown           = "unknown"
)

var pathCleanUpRegexp = regexp.MustCompile(`[^a-zA-Z0-9_\-&^!+=\)\(\[\].]`)

func cleanPath(path string) string {
	cleaned := pathCleanUpRegexp.ReplaceAll([]byte(path), []byte("+"))
	return string(cleaned)
}
func contentTypeToExt(podcastItem Item) (ext string, ctype ContentType, err error) {

	contentType := podcastItem.Enclosure.Type

	if !strings.Contains(contentType, "/") {
		contentType = podcastItem.Media.Type
	}

	if !strings.Contains(contentType, "/") {
		return "", unknown, fmt.Errorf("Unable to determine the extension of the file")
	}

	parts := strings.Split(contentType, "/")

	switch parts[1] {
	case "mpeg":
		fallthrough
	case "mp3":
		return "mp3", audio, nil
	case "mp4":
		return "mp4", video, nil
	}

	return "", unknown, fmt.Errorf("Unable to figure out extension %s", parts[1])
}

func parseTime(dateString string) (date time.Time, err error) {
	formats := []string{time.RFC1123Z, time.RFC1123, time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z}
	for _, format := range formats {
		date, err = time.Parse(format, dateString)
		if err == nil {
			return date, nil
		}
	}

	oneYearAgo, _ := time.ParseDuration("-8760h")
	return time.Now().Add(oneYearAgo), nil
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

func createPodcastDir(head opml.OpmlHead, rssModel *Rss, podcastItem Item) (podcastDir, ext string, err error) {

	ext, contentType, err := contentTypeToExt(podcastItem)

	if err != nil {
		return "", "", err
	}

	podcastDir = filepath.Join(head.DownloadDir, string(contentType), cleanPath(rssModel.Channel.Title))

	err = os.MkdirAll(podcastDir, os.ModeDir)

	if err != nil {
		return "", "", fmt.Errorf("Failed to make podcast directory %q", podcastDir)
	}


	return podcastDir, ext, nil
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

	outline.Title = rssModel.Channel.Title
	outline.DirectoryName = cleanPath(rssModel.Channel.Title)

	podcastItems := rssModel.Channel.Items

	if len(podcastItems) == 0 {
		return 0, nil
	}

	if !needsUpdate(podcastItems[0], outline) {
		log.Printf("Podcast %q is update to date", rssModel.Channel.Title)
		return 0, nil
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
		podcastItem := podcastItems[i]
		var postcastUrl string
		if postcastUrl = podcastItem.Enclosure.Url; len(postcastUrl) > 0 {
			postcastUrl = postcastUrl
		} else if postcastUrl = podcastItem.Media.Url; len(postcastUrl) > 0 {
			postcastUrl = postcastUrl
		} else {
			return 0, fmt.Errorf("No url was found for this podcast: %q\n", rssModel.Channel.Title)
		}

		podcastDir, ext, err := createPodcastDir(head, rssModel, podcastItem)
		if err != nil {
			return downloadCount, err
		}

		podcastFile := filepath.Join(podcastDir, cleanPath(podcastItem.Title)+"."+ext)

		fi, err := os.Stat(podcastFile)

		if os.IsExist(err) && fi.Size() > 0 {
			log.Printf("Podcast has been previously downloaded, Skipping download of %s\n", podcastItem.Title)
		} else {

			dest, err := os.Create(podcastFile)
			defer dest.Close()

			if err != nil {
				return downloadCount, fmt.Errorf("Unable to create file for podcast: %v", podcastItem)
			}



			log.Printf("Downloading podcast: %q from url %q\n", podcastItem.Title, postcastUrl)
			resp, err = http.Get(postcastUrl)

			if err != nil {
				log.Printf("An error occurred downloading podcast: '%v'\n\n", err)
				return downloadCount, err
			}

			podcast := resp.Body
			defer podcast.Close()

			n, err := bufio.NewReader(podcast).WriteTo(dest)
			if err != nil {
				return downloadCount, fmt.Errorf("Failed to copy %v to %v due to %v", podcastItem.Title, dest, err)
			}

			log.Printf("Downloaded file with size: %d to %s\n", n, dest.Name())
		}
		downloadCount++
	}

	outline.LastUpdate = podcastItems[0].PubDate

	return downloadCount, nil
}
