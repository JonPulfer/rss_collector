package feed

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"

	"github.com/mmcdole/gofeed"
)

type Source struct {
	ID            string
	FeedURL       string
	address       *url.URL
	feedParser    *gofeed.Parser
	LastCollected time.Time
	Feed          *gofeed.Feed
}

func NewSource(feedURL string) (*Source, error) {
	if !validFeedURL(feedURL) {
		return nil, fmt.Errorf("invalid feedURL: %s", feedURL)
	}

	address, err := url.Parse(feedURL)
	if err != nil {
		return nil, err
	}

	return &Source{
		FeedURL:       address.String(),
		address:       address,
		feedParser:    gofeed.NewParser(),
		LastCollected: time.Time{},
	}, nil
}

func (s Source) String() string {
	return s.ID
}

func (s *Source) Collect() error {
	defer func() {
		s.LastCollected = time.Now()
	}()

	collected, err := s.feedParser.ParseURL(s.address.String())
	if err != nil {
		return err
	}
	s.Feed = collected
	return nil
}

func (s *Source) Items() rsscollector.FeedItems {
	return itemsFromItems(s.Feed.Items)
}

func (s *Source) FeedSource() rsscollector.FeedSource {
	return rsscollector.FeedSource{
		FeedSourcePartial: rsscollector.FeedSourcePartial{
			ID:            s.ID,
			Link:          rsscollector.FeedSourceLink(s.ID),
			FeedURL:       s.FeedURL,
			Title:         s.Feed.Title,
			LastCollected: s.LastCollected,
		},
		FeedItems: s.Items(),
	}
}

func validFeedURL(feedURL string) bool {
	switch {
	case len(strings.Replace(feedURL, " ", "", -1)) == 0:
		return false
	case !strings.HasPrefix(feedURL, "http"):
		return false
	default:
		return true
	}
}

func itemsFromItems(items []*gofeed.Item) rsscollector.FeedItems {
	results := make([]*rsscollector.FeedItem, 0)
	for _, v := range items {
		if v != nil {
			results = append(results, itemFromItem(*v))
		}
	}
	return results
}

func itemFromItem(item gofeed.Item) *rsscollector.FeedItem {
	var author string
	if item.Author != nil {
		author = item.Author.Name
	}

	return &rsscollector.FeedItem{
		Title:       item.Title,
		Description: item.Description,
		Content:     item.Content,
		Link:        item.Link,
		Updated:     item.UpdatedParsed,
		Published:   item.PublishedParsed,
		Author:      author,
		GUID:        item.GUID,
		Image:       ItemImageFromImage(item.Image),
		Categories:  item.Categories,
		Custom:      item.Custom,
	}
}

func ItemImageFromImage(image *gofeed.Image) *rsscollector.FeedItemImage {
	if image == nil {
		return nil
	}
	return &rsscollector.FeedItemImage{
		URL:   image.URL,
		Title: image.Title,
	}
}
