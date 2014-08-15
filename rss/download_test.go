package rss

import (
	"testing"
	"fmt"
	"net/http"
	"net/http/httptest"
	"gopod/opml"
	"io/ioutil"
	"os"
	"path/filepath"
)

func servers(rssModel string) (*httptest.Server, *httptest.Server) {
	mp3Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "This is a fake podcast")
	}))

	rssServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, fmt.Sprintf(rssModel, mp3Server.URL))
	}))

	return mp3Server, rssServer
}

func download(updateRss func(Rss), outline *opml.OpmlOutline) (rss Rss, downloadDir string, numDownloaded int, err error) {
	rssModel := Rss{
		Channel{
			Title: "Test Podcast",
			Items: []Item{{
				Title: "Podcast Item 1",
				PubDate: "Mon, 11 Aug 2014 21:20:36 +0000"}}}}

	rssModel.Channel.Items[0].Enclosure.Type = "audio/mpeg"
	rssModel.Channel.Items[0].Media.Type = "audio/mpeg"

	updateRss(rssModel)

	mp3Server, rssServer := servers(rssModel.String())
	defer mp3Server.Close()
	defer rssServer.Close()

	downloadDir, err = ioutil.TempDir("", "podcasts")
	if err != nil {
		panic("Unable to create temporary download dir")
	}

	head := opml.OpmlHead{DownloadDir: downloadDir}
	outline.XmlUrl = rssServer.URL
	numEpisodesDownloaded, err := Download(head, outline)

	return rssModel, downloadDir, numEpisodesDownloaded, err
}

func Test_DownloadEnclosureHasUrl(t *testing.T) {
	outline := &opml.OpmlOutline{LastUpdate:"Mon, 10 Aug 2014 21:20:36 +0000"}

	rssModel, downloadDir, downloaded, err := download(func(rssModel Rss) {
			rssModel.Channel.Items[0].Enclosure.Url = "%s"
		}, outline)

	if err != nil {
		t.Fatal(err)
	}

	if downloaded != 1 {
		t.Errorf("Expected 1 episode to be downloaded: %d", downloaded)
	}

	downloadDirFile, err := os.Open(filepath.Join(downloadDir, rssModel.Channel.Title))
	if err != nil {
		t.Fatalf("Unable to open in download directory: %v", err)
	}
	files, err := downloadDirFile.Readdir(-1)
	if err != nil {
		t.Fatalf("Unable to list files in download directory: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Unexpected number of files downloaded: %d", len(files))
	}

	if files[0].Size() == 0 {
		t.Errorf("No data was written to file %q", files[0].Name())
	}

	if filepath.Ext(files[0].Name()) != ".mp3" {
		t.Errorf("Expected file to end with mp3: %q, %q", filepath.Ext(files[0].Name()), files[0].Name())
	}

	items := rssModel.Channel.Items
	if outline.LastUpdate != items[0].PubDate {
		t.Errorf("outline last update was not updated. Expected %q but got %q", outline.LastUpdate, items[0].PubDate)
	}
}

func Test_DownloadMediaHasUrl(t *testing.T) {
	outline := &opml.OpmlOutline{}

	rssModel, downloadDir, downloaded, err := download(func(rssModel Rss) {
			rssModel.Channel.Items[0].Media.Url = "%s"
		}, outline)

	if err != nil {
		t.Fatal(err)
	}

	if downloaded != 1 {
		t.Errorf("Expected 1 episode to be downloaded: %d", downloaded)
	}


	downloadDirFile, err := os.Open(filepath.Join(downloadDir, rssModel.Channel.Title))
	if err != nil {
		t.Fatalf("Unable to open in download directory: %v", err)
	}
	files, err := downloadDirFile.Readdir(-1)

	if len(files) != 1 {
		t.Fatalf("Unexpected number of files downloaded: %d", len(files))
	}

	if files[0].Size() == 0 {
		t.Error("No data was written to file %q", files[0].Name())
	}

	if filepath.Ext(files[0].Name()) != ".mp3" {
		t.Errorf("Expected file to end with mp3: %q, %q", filepath.Ext(files[0].Name()), files[0].Name())
	}
}

func Test_DownloadIsUpToDate(t *testing.T) {
	outline := &opml.OpmlOutline{}

	rssModel, downloadDir, downloaded, err := download(func(rssModel Rss) {
			rssModel.Channel.Items[0].Media.Url = "%s"
			outline.LastUpdate = rssModel.Channel.Items[0].PubDate
		}, outline)

	if err != nil {
		t.Fatal(err)
	}

	if downloaded != 0 {
		t.Errorf("Expected 0 episodes to be downloaded: %d", downloaded)
	}


	_, err = os.Open(filepath.Join(downloadDir, rssModel.Channel.Title))
	if os.IsExist(err) {
		t.Fatalf("Should not have created the podcast dir if no files were downloaded: %v", err)
	}
}

