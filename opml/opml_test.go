package opml

import (
	"bytes"
	"testing"
)

func TestParseOpml(t *testing.T) {
	opmlData := `
<?xml version='1.0' encoding='UTF-8' standalone='yes' ?>
<opml version="2.0">
    <head>
        <dateCreated>Sun Aug 10 11:03:04 CEST 2014</dateCreated>
    </head>
    <body>
        <outline xmlUrl="http://downloads.bbc.co.uk/podcasts/worldservice/globalnews/rss.xml" />
        <outline xmlUrl="http://feeds.feedburner.com/DailyTechNewsShow" />
    </body>
</opml>`

	model, err := ParseOpml(bytes.NewReader([]byte(opmlData)))
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(model.Body.Outline) != 2 {
		t.Fatalf("Incorrect number of Outline elements: \n%v", model.Body.Outline)
	}

	if model.Head.DateCreated != "Sun Aug 10 11:03:04 CEST 2014" {
		t.Fatal("Wrong date created: " + model.Head.DateCreated)
	}

	if model.Body.Outline[0].XmlUrl != "http://downloads.bbc.co.uk/podcasts/worldservice/globalnews/rss.xml" {
		t.Fatal("Outline[0] has wrong url: " + model.Body.Outline[0].XmlUrl)
	}

	if model.Body.Outline[1].XmlUrl != "http://feeds.feedburner.com/DailyTechNewsShow" {
		t.Fatal("Outline[1] has wrong url: " + model.Body.Outline[0].XmlUrl)
	}
}

func TestWriteOpml(t *testing.T) {
	model := New()
	model.Head.DateCreated = "Today"
	model.Body.Outline = make([]OpmlOutline, 2)
	model.Body.Outline[0] = OpmlOutline{"http://url0"}
	model.Body.Outline[1] = OpmlOutline{"http://url1"}

	buffer := &bytes.Buffer{}
	if _, err := model.Write(buffer); err != nil {
		t.Fatal(err)
	}
	if parsedModel, err := ParseOpml(buffer); err != nil {
		t.Fatal(err)
	} else {
		equal(&model, &parsedModel, t)
	}
}

func equal(o1, o2 *Opml, t *testing.T) {
	if o1.Head != o2.Head {
		t.Fatalf("Opml Head elements differ: \n%v\n%v\n", o1.Head, o2.Head)
	}

	if o1.Version != o2.Version {
		t.Fatalf("Opml Version differ: \n%v\n%v\n", o1.Head, o2.Head)
	}

	if len(o1.Body.Outline) != len(o2.Body.Outline) {
		t.Fatalf("Opml Outline lengths differ: \n%v\n%v\n", o1.Body.Outline, o2.Body.Outline)
	}

	for i, bo1 := range o1.Body.Outline {
		bo2 := o2.Body.Outline[i]

		if bo1 != bo2 {
			t.Errorf("Outline %d do not match: \n%v\n%v", bo1, bo2)
		}
	}

}
