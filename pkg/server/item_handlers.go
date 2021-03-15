package server

import (
	"sort"

	"github.com/gofiber/fiber/v2"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"
)

func (h HTTPFeedServer) getItems(c *fiber.Ctx) error {
	sourceID := c.Query("sourceId")
	if len(sourceID) > 0 {
		if err := validateID(sourceID); err != nil {
			return err
		}
	}
	itemOptions := rsscollector.ItemOptions{
		SourceID: sourceID,
	}
	categoryID := c.Query("categoryId")
	if len(categoryID) > 0 {
		if err := validateID(categoryID); err != nil {
			return err
		}
		itemOptions.CategoryIDs = []string{categoryID}
	}

	items, err := h.itemRepos.FetchAllItems(itemOptions)
	if err != nil {
		return err
	}

	sort.Sort(items)

	return c.JSON(items)
}

func (h HTTPFeedServer) getItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	if err := validateID(itemID); err != nil {
		return err
	}

	item, err := h.itemRepos.FetchItemByID(itemID)
	if err != nil {
		return err
	}
	return c.JSON(item)
}

// UpdateItemRequest provides the canonical list of category IDs that should
// be associated with the item.
type UpdateItemRequest struct {
	CategoryIDs []string `json:"categoryIds"`
}

func (u UpdateItemRequest) Validate() error {
	for _, v := range u.CategoryIDs {
		if err := validateID(v); err != nil {
			return err
		}
	}
	return nil
}

func (h HTTPFeedServer) putItem(c *fiber.Ctx) error {
	var updateItemRequest UpdateItemRequest
	if err := c.BodyParser(&updateItemRequest); err != nil {
		return err
	}
	if err := updateItemRequest.Validate(); err != nil {
		return err
	}
	itemID := c.Params("id")
	if err := validateID(itemID); err != nil {
		return err
	}

	item, err := h.itemRepos.FetchItemByID(itemID)
	if err != nil {
		return err
	}
	item.CategoryIDs = updateItemRequest.CategoryIDs
	if err := h.itemRepos.StoreItem(item.SourceID, &item); err != nil {
		return err
	}
	return c.JSON(item)
}

func (h HTTPFeedServer) deleteItem(c *fiber.Ctx) error {
	itemID := c.Params("id")
	if err := validateID(itemID); err != nil {
		return err
	}
	if err := h.itemRepos.DeleteItemByID(itemID); err != nil {
		return err
	}
	return c.JSON(itemID)
}
