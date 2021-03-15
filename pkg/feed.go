package pkg

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FeedSourcePartial provides just the details of the source feed without the
// collected items.
type FeedSourcePartial struct {
	ID            string    `json:"id"`
	Link          string    `json:"link"`
	FeedURL       string    `json:"feedUrl"`
	Title         string    `json:"title"`
	CategoryIDs   []string  `json:"categoryIDs,omitempty"`
	LastCollected time.Time `json:"lastCollected"`
}

func NewFeedSourcePartial(source FeedSource) FeedSourcePartial {
	return FeedSourcePartial{
		ID:            source.ID,
		Link:          FeedSourceLink(source.ID),
		FeedURL:       source.FeedURL,
		Title:         source.Title,
		CategoryIDs:   source.CategoryIDs,
		LastCollected: source.LastCollected,
	}
}

// FeedSource is the full representation that includes the FeedItems we have
// collected.
type FeedSource struct {
	FeedSourcePartial
	FeedItems []*FeedItem `json:"feedItems"`
}

func FeedSourceLink(id string) string {
	return fmt.Sprintf("/feeds/%s", id)
}

// FeedItem collected from a FeedSource.
type FeedItem struct {
	ID          string            `json:"id,omitempty"`
	SourceID    string            `json:"sourceId,omitempty"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Content     string            `json:"content,omitempty"`
	Link        string            `json:"link,omitempty"`
	Updated     *time.Time        `json:"updated,omitempty"`
	Published   *time.Time        `json:"published,omitempty"`
	Author      string            `json:"author,omitempty"`
	GUID        string            `json:"guid,omitempty"`
	Image       *FeedItemImage    `json:"image,omitempty"`
	Categories  []string          `json:"categories,omitempty"`
	CategoryIDs []string          `json:"categoryIds,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

type FeedItems []*FeedItem

func (f FeedItems) Len() int {
	return len(f)
}

func (f FeedItems) Less(i, j int) bool {
	return f[i].Published.Unix() < f[j].Published.Unix()
}

func (f FeedItems) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type FeedItemImage struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type FeedCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewCategory(name string) FeedCategory {
	id, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	return FeedCategory{
		ID:   id.String(),
		Name: name,
	}
}
