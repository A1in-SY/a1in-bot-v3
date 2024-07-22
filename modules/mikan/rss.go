package mikan

import (
	"encoding/xml"
	"io"
	"net/http"
)

type MikanRSSFeed struct {
	XMLName xml.Name        `xml:"rss"`
	Version string          `xml:"version,attr"`
	Channel mikanRSSChannel `xml:"channel"`
}

type mikanRSSChannel struct {
	Title       string         `xml:"title"`
	Link        string         `xml:"link"`
	Description string         `xml:"description"`
	Items       []mikanRSSItem `xml:"item"`
}

type mikanRSSItem struct {
	Guid mikanRSSGuid `xml:"guid"`
	// Mikan 页面 URL
	Link  string `xml:"link"`
	Title string `xml:"title"`
	// 描述
	Description string            `xml:"description"`
	Torrent     mikanRSSTorrent   `xml:"torrent"`
	Enclosure   mikanRSSEnclosure `xml:"enclosure"`
}

type mikanRSSGuid struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

type mikanRSSTorrent struct {
	Link          string `xml:"link"`
	ContentLength int    `xml:"contentLength"`
	PubDate       string `xml:"pubDate"`
}

type mikanRSSEnclosure struct {
	Type   string `xml:"type,attr"`
	Length int    `xml:"length,attr"`
	// 种子 URL
	URL string `xml:"url,attr"`
}

func (mikan *Mikan) getRSSFeed(rssUrl string) (feed *MikanRSSFeed, err error) {
	req, _ := http.NewRequest(http.MethodGet, rssUrl, nil)
	resp, err := mikan.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	feed = &MikanRSSFeed{}
	err = xml.Unmarshal(data, feed)
	if err != nil {
		return nil, err
	}
	return feed, nil
}
