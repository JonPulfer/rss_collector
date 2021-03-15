package server

import (
	"sort"

	"github.com/gofiber/fiber/v2"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"
	"github.com/JonPulfer/rss_collector/pkg/feed"
)

func (h HTTPFeedServer) getFeeds(c *fiber.Ctx) error {
	feeds, err := h.feedRepos.FetchAllSources()
	if err != nil {
		return err
	}

	return c.JSON(feeds)
}

func (h HTTPFeedServer) getFeed(c *fiber.Ctx) error {
	feedID := c.Params("id")
	if err := validateID(feedID); err != nil {
		return err
	}
	sourceFeed, err := h.feedRepos.FetchSource(feedID)
	if err != nil {
		return err
	}
	feedItems, err := h.itemRepos.FetchAllItems(
		rsscollector.ItemOptions{
			SourceID: sourceFeed.ID,
		})
	if err != nil {
		return err
	}
	sort.Sort(feedItems)
	sourceFeed.FeedItems = feedItems

	return c.JSON(sourceFeed)
}

// CreateFeedRequest to add a new feed source to track.
type CreateFeedRequest struct {
	FeedURL string `json:"feedUrl"`
}

func (c CreateFeedRequest) Validate() error {
	return validateFeedURL(c.FeedURL)
}

// CreateFeedResponse provides the identifying information for the newly created
// feed.
type CreateFeedResponse struct {
	ID   string `json:"id"`
	Link string `json:"link"`
}

func (h HTTPFeedServer) postFeeds(c *fiber.Ctx) error {
	var feedRequest CreateFeedRequest
	if err := c.BodyParser(&feedRequest); err != nil {
		return err
	}
	if err := feedRequest.Validate(); err != nil {
		return err
	}
	source, err := feed.NewSource(feedRequest.FeedURL)
	if err != nil {
		return err
	}
	if err := source.Collect(); err != nil {
		return err
	}

	feedSource := source.FeedSource()
	if err := h.feedRepos.StoreSource(&feedSource); err != nil {
		return err
	}

	if err := h.itemRepos.StoreItems(feedSource.ID, source.Items()); err != nil {
		return err
	}

	resp := CreateFeedResponse{
		ID:   feedSource.ID,
		Link: feedSource.Link,
	}

	return c.JSON(resp)
}

// UpdateFeedRequest with new and additional information.
type UpdateFeedRequest struct {
	FeedURL     string   `json:"feedURL"`
	CategoryIDs []string `json:"categoryIDs"`
}

func (u UpdateFeedRequest) Validate() error {
	if len(u.FeedURL) > 0 {
		if err := validateFeedURL(u.FeedURL); err != nil {
			return err
		}
	}
	for _, v := range u.CategoryIDs {
		if err := validateID(v); err != nil {
			return err
		}
	}
	return nil
}

// UpdateFeedResponse returns the updated feed partial.
type UpdateFeedResponse struct {
	Feed rsscollector.FeedSourcePartial `json:"feed"`
}

func (h HTTPFeedServer) putFeed(c *fiber.Ctx) error {
	var updateRequest UpdateFeedRequest
	if err := c.BodyParser(&updateRequest); err != nil {
		return err
	}
	if err := updateRequest.Validate(); err != nil {
		return err
	}
	feedID := c.Params("id")
	if err := validateID(feedID); err != nil {
		return err
	}

	feedSource, err := h.feedRepos.FetchSource(feedID)
	if err != nil {
		return err
	}

	if len(updateRequest.FeedURL) > 0 {
		if feedSource.FeedURL != updateRequest.FeedURL {
			if err := validateFeedURL(updateRequest.FeedURL); err != nil {
				return err
			}
			feedSource.FeedURL = updateRequest.FeedURL
		}
	}

	if len(updateRequest.CategoryIDs) > 0 {
		feedSource.CategoryIDs = updateRequest.CategoryIDs
	}

	if err := h.feedRepos.StoreSource(&feedSource); err != nil {
		return err
	}

	resp := UpdateFeedResponse{Feed: rsscollector.NewFeedSourcePartial(feedSource)}

	return c.JSON(resp)
}

func (h HTTPFeedServer) deleteFeed(c *fiber.Ctx) error {
	feedID := c.Params("id")
	if err := validateID(feedID); err != nil {
		return err
	}
	if err := h.feedRepos.DeleteSourceByID(feedID); err != nil {
		return err
	}

	return c.JSON(feedID)
}
