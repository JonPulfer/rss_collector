package repository

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"
)

type MemoryFeedStore struct {
	feeds            map[string]rsscollector.FeedSource
	items            map[string]rsscollector.FeedItems
	itemsByID        map[string]rsscollector.FeedItem
	categoriesByID   map[string]string
	categoriesByName map[string]string
	sync.RWMutex
}

func NewMemoryStore() *MemoryFeedStore {
	return &MemoryFeedStore{
		feeds:            make(map[string]rsscollector.FeedSource),
		items:            make(map[string]rsscollector.FeedItems),
		itemsByID:        make(map[string]rsscollector.FeedItem),
		categoriesByID:   make(map[string]string),
		categoriesByName: make(map[string]string),
		RWMutex:          sync.RWMutex{},
	}
}

func (m *MemoryFeedStore) StoreSource(source *rsscollector.FeedSource) error {
	defer m.Unlock()
	m.Lock()
	if len(source.ID) == 0 {
		u, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		source.ID = u.String()
	}
	m.feeds[source.ID] = *source
	return nil
}

func (m *MemoryFeedStore) FetchSource(feedID string) (rsscollector.FeedSource, error) {
	defer m.RUnlock()
	m.RLock()
	if f, ok := m.feeds[feedID]; ok {
		return f, nil
	}
	return rsscollector.FeedSource{}, fmt.Errorf("no feed found for feedID: %s", feedID)
}

func (m *MemoryFeedStore) FetchAllSources() ([]rsscollector.FeedSourcePartial, error) {
	defer m.RUnlock()
	m.RLock()
	feeds := make([]rsscollector.FeedSourcePartial, 0)
	for _, v := range m.feeds {
		feeds = append(feeds, rsscollector.NewFeedSourcePartial(v))
	}
	return feeds, nil
}

func (m *MemoryFeedStore) DeleteSourceByID(id string) error {
	defer m.Unlock()
	m.Lock()
	delete(m.items, id)
	for itemID, item := range m.itemsByID {
		if item.SourceID == id {
			delete(m.itemsByID, itemID)
		}
	}
	delete(m.feeds, id)
	return nil
}

func (m *MemoryFeedStore) StoreItem(sourceID string, item *rsscollector.FeedItem) error {
	defer m.Unlock()
	m.Lock()
	if len(item.ID) > 0 {
		m.itemsByID[item.ID] = *item
		for idx := range m.items[sourceID] {
			if m.items[sourceID][idx].ID == item.ID {
				m.items[sourceID][idx] = item
			}
		}
		return nil
	}
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	item.ID = id.String()
	if _, ok := m.itemsByID[sourceID]; ok {
		m.items[sourceID] = append(m.items[sourceID], item)
		m.itemsByID[item.ID] = *item
		return nil
	}
	m.items[sourceID] = []*rsscollector.FeedItem{item}
	m.itemsByID[item.ID] = *item
	return nil
}

func (m *MemoryFeedStore) FetchItemByID(id string) (rsscollector.FeedItem, error) {
	defer m.RUnlock()
	m.RLock()
	if item, ok := m.itemsByID[id]; ok {
		return item, nil
	}
	return rsscollector.FeedItem{}, fmt.Errorf("no feed item found with id: %s", id)
}

func (m *MemoryFeedStore) StoreItems(sourceID string, items []*rsscollector.FeedItem) error {
	defer m.Unlock()
	m.Lock()

	for idx := range items {
		id, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		items[idx].ID = id.String()
		items[idx].SourceID = sourceID
		m.itemsByID[id.String()] = *items[idx]
	}
	m.items[sourceID] = items
	return nil
}

func (m *MemoryFeedStore) FetchAllItems(options rsscollector.ItemOptions) (rsscollector.FeedItems, error) {
	defer m.RUnlock()
	m.RLock()

	results := make(rsscollector.FeedItems, 0)
	for sourceID, storedItems := range m.items {
		if len(options.SourceID) > 0 {
			if sourceID != options.SourceID {
				continue
			}
		}
		items := make(rsscollector.FeedItems, 0)
		for _, storedItem := range storedItems {
			if len(options.CategoryIDs) > 0 {
				var hasRequestedCategories bool
				for _, categoryID := range options.CategoryIDs {
					for _, itemCategoryID := range storedItem.CategoryIDs {
						if itemCategoryID == categoryID {
							hasRequestedCategories = true
						}
					}
				}
				if hasRequestedCategories {
					items = append(items, storedItem)
				}
			} else {
				items = append(items, storedItem)
			}
		}
		results = append(results, items...)
	}

	if len(results) > 0 {
		sort.Sort(results)
		return results, nil
	}

	return nil, fmt.Errorf("no items found for ItemOptions: %v", options)
}

func (m *MemoryFeedStore) DeleteItemByID(id string) error {
	defer m.Unlock()
	m.Lock()
	if _, ok := m.itemsByID[id]; ok {
		delete(m.itemsByID, id)
	}
	for sourceID, sourceItems := range m.items {
		for idx, sourceItem := range sourceItems {
			if sourceItem.ID == id {
				m.items[sourceID][idx] = nil
			}
		}
	}
	return nil
}

func (m *MemoryFeedStore) StoreCategory(category *rsscollector.FeedCategory) error {
	defer m.Unlock()
	m.Lock()
	m.categoriesByID[category.ID] = category.Name
	m.categoriesByName[category.Name] = category.ID
	return nil
}

func (m *MemoryFeedStore) FetchCategoryByID(id string) (rsscollector.FeedCategory, error) {
	defer m.RUnlock()
	m.RLock()
	if _, ok := m.categoriesByID[id]; !ok {
		return rsscollector.FeedCategory{}, fmt.Errorf("no category found with id: %s", id)
	}
	category := rsscollector.FeedCategory{
		ID:   id,
		Name: m.categoriesByID[id],
	}
	return category, nil
}

func (m *MemoryFeedStore) FetchCategoryByName(name string) (rsscollector.FeedCategory, error) {
	defer m.RUnlock()
	m.RLock()
	if id, ok := m.categoriesByName[name]; ok {
		return rsscollector.FeedCategory{
			ID:   id,
			Name: name,
		}, nil
	}
	return rsscollector.FeedCategory{}, fmt.Errorf("no category found with name: %s", name)
}

func (m *MemoryFeedStore) DeleteCategoryByID(id string) error {
	defer m.Unlock()
	m.Lock()
	if categoryName, ok := m.categoriesByID[id]; ok {
		for _, v := range m.itemsByID {
			for idx, itemCategory := range v.CategoryIDs {
				if itemCategory == id {
					v.CategoryIDs[idx] = ""
				}
			}
		}
		for _, sourceItems := range m.items {
			for _, sourceItem := range sourceItems {
				for idx, itemCategory := range sourceItem.CategoryIDs {
					if itemCategory == id {
						sourceItem.CategoryIDs[idx] = ""
					}
				}
			}
		}
		if _, ok := m.categoriesByName[categoryName]; ok {
			delete(m.categoriesByName, categoryName)
		}
		delete(m.categoriesByID, id)
	}
	return nil
}

func (m *MemoryFeedStore) FetchCategoriesForIDs(ids []string) ([]rsscollector.FeedCategory, error) {
	categories := make([]rsscollector.FeedCategory, 0)
	for _, v := range ids {
		category, err := m.FetchCategoryByID(v)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (m *MemoryFeedStore) FetchAllCategories() ([]rsscollector.FeedCategory, error) {
	defer m.RUnlock()
	m.RLock()
	categories := make([]rsscollector.FeedCategory, 0)
	for k, v := range m.categoriesByID {
		categories = append(categories, rsscollector.FeedCategory{
			ID:   k,
			Name: v,
		})
	}
	return categories, nil
}
