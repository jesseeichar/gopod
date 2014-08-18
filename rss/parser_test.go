package rss

import (
	"bytes"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"
)

func TestParseRss(t *testing.T) {
	rssBytes, err := ioutil.ReadFile("example.rss.xml")
	if err != nil {
		t.Fatal(err)
	}
	rss, err := ParseRss(bytes.NewReader(rssBytes))
	if err != nil {
		t.Fatal(err)
	}

	if rss.Channel.Description != "On your side." {
		t.Errorf("Wrong channel description: : \n%q\n%q\n", "On your side", rss.Channel.Description)
	}

	if rss.Channel.Title != "Daily Tech News Show" {
		t.Errorf("Wrong channel title: \n%q\n%q\n", "Daily Tech News Show", rss.Channel.Title)
	}

	if rss.Channel.LastBuildDate != "Mon, 11 Aug 2014 21:51:56 +0000" {
		t.Errorf("Wrong channel LastBuildDate: \n%q\n%q\n", "Mon, 11 Aug 2014 21:51:56 +0000", rss.Channel.LastBuildDate)
	}

	if len(rss.Channel.Items) != 25 {
		t.Fatalf("Not enough items in channel, expected 25 but got %v", len(rss.Channel.Items))
	}

	checkItem(t, 0, rss.Channel.Items, Item{
		Title:       "DTNS 2297 – Antitrust Prime",
		Link:        "http://feedproxy.google.com/~r/DailyTechNewsShow/~3/CeBqn9GHRKk/",
		Description: "Nate Lanxon is on the show today to chat about the Hachette-Amazon spat, as well as a little on Broadwell chips and the $300 million 60 Tb/s cable Google wants to lay. MP3 Multiple versions (ogg, video etc.) from Archive.org. Please SUBSCRIBE HERE. A special thanks to all our Patreon supporters&#8211;without you, none of this [&#8230;]",
		PubDate:     "Mon, 11 Aug 2014 21:20:36 +0000",
		Category:    "Episode",
		Guid:        "http://www.dailytechnewsshow.com/?p=1844",
		Enclosure: Enclosure{
			Url:  "http://archive.org/download/DTNS20140811/DTNS20140811.mp3",
			Type: "audio/mpeg"},
		Media: Media{
			Url:  "http://archive.org/download/DTNS20140811/DTNS20140811.mp3",
			Type: "audio/mpeg"}})
}

func checkItem(t *testing.T, index int, items []Item, expected Item) {
	actual := items[index]

	if actual != expected {
		if actual.Title != expected.Title {
			t.Errorf("Item[%d] does not have the correct title: \n%q\n%q", index, expected.Title, actual.Title)
		}
		if actual.Category != expected.Category {
			t.Errorf("Item[%d] does not have the correct category: \n%q\n%q", index, expected.Category, actual.Category)
		}
		if actual.Guid != expected.Guid {
			t.Errorf("Item[%d] does not have the correct Guid: \n%q\n%q", index, expected.Guid, actual.Guid)
		}
		if actual.Link != expected.Link {
			t.Errorf("Item[%d] does not have the correct Link: \n%q\n%q", index, expected.Link, actual.Link)
		}
		if actual.PubDate != expected.PubDate {
			t.Errorf("Item[%d] does not have the correct PubDate: \n%q\n%q", index, expected.PubDate, actual.PubDate)
		}
		if actual.Media.Url != expected.Media.Url {
			t.Errorf("Item[%d] does not have the correct Media.Url: \n%q\n%q", index, expected.Media.Url, actual.Media.Url)
		}
		if actual.Media.Type != expected.Media.Type {
			t.Errorf("Item[%d] does not have the correct Media.Type: \n%q\n%q", index, expected.Media.Type, actual.Media.Type)
		}
		if actual.Enclosure.Url != expected.Enclosure.Url {
			t.Errorf("Item[%d] does not have the correct Enclosure.Url: \n%q\n%q", index, expected.Enclosure.Url, actual.Enclosure.Url)
		}
		if actual.Enclosure.Type != expected.Enclosure.Type {
			t.Errorf("Item[%d] does not have the correct Enclosure.Type: \n%q\n%q", index, expected.Enclosure.Type, actual.Enclosure.Type)
		}
		whiteSpaceMatcher := regexp.MustCompile(`\s+`)
		actualDesc := string(whiteSpaceMatcher.ReplaceAll([]byte(actual.Description), []byte("")))
		expectedDesc := string(whiteSpaceMatcher.ReplaceAll([]byte(expected.Description), []byte("")))
		if actualDesc != expectedDesc {
			t.Errorf("Item[%d] does not have the correct Description: \n%q\n%q", index, expectedDesc, actualDesc)
		}
	}
}

func Test_WriteReadCycle(t *testing.T) {
	original := Rss{
		Channel: Channel{
			Title:         "ChannelTitle",
			Description:   "ChannelDescription",
			LastBuildDate: "Mon, 01 Aug 2014 21:20:36 +0000",
			Items: []Item{
				Item{
					Title:       "ItemTitle",
					Link:        "http://item.link",
					Description: "ItemDescription",
					PubDate:     "Mon, 11 Aug 2014 21:20:36 +0000",
					Category:    "tech",
					Guid:        "guid"}}}}

	parsed, err := ParseRss(strings.NewReader(original.String()))

	if err != nil {
		t.Fatalf("Error occurred during parsing of rss: %q", err.Error())
	}

	if original.Channel.Title != parsed.Channel.Title {
		t.Errorf("Titles don't match:\n%q\n%q", original.Channel.Title, parsed.Channel.Title)
	}

	if original.Channel.Description != parsed.Channel.Description {
		t.Errorf("Descriptions don't match:\n%q\n%q", original.Channel.Description, parsed.Channel.Description)
	}

	if original.Channel.LastBuildDate != parsed.Channel.LastBuildDate {
		t.Errorf("LastBuildDates don't match:\n%q\n%q", original.Channel.LastBuildDate, parsed.Channel.LastBuildDate)
	}

	for i, item := range original.Channel.Items {
		checkItem(t, i, parsed.Channel.Items, item)
	}
}
