package repository

import (
	rsscollector "github.com/JonPulfer/rss_collector/pkg"
)

type FeedSourceStore interface {
	StoreSource(source *rsscollector.FeedSource) error
	FetchSource(feedID string) (rsscollector.FeedSource, error)
	FetchAllSources() ([]rsscollector.FeedSourcePartial, error)
	DeleteSourceByID(feedID string) error
}

type FeedItemStore interface {
	StoreItem(sourceID string, item *rsscollector.FeedItem) error
	StoreItems(sourceID string, items []*rsscollector.FeedItem) error
	FetchItemByID(id string) (rsscollector.FeedItem, error)
	FetchAllItems(options rsscollector.ItemOptions) (rsscollector.FeedItems, error)
	DeleteItemByID(id string) error
}

type FeedCategoryStore interface {
	StoreCategory(category *rsscollector.FeedCategory) error
	FetchAllCategories() ([]rsscollector.FeedCategory, error)
	FetchCategoryByID(id string) (rsscollector.FeedCategory, error)
	FetchCategoryByName(name string) (rsscollector.FeedCategory, error)
	FetchCategoriesForIDs(ids []string) ([]rsscollector.FeedCategory, error)
	DeleteCategoryByID(id string) error
}
