package server

import (
	"github.com/gofiber/fiber/v2"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"
)

func (h HTTPFeedServer) getCategories(c *fiber.Ctx) error {
	categories, err := h.categoryRepos.FetchAllCategories()
	if err != nil {
		return err
	}
	return c.JSON(categories)
}

type CreateCategoryRequest struct {
	Name string `json:"categoryName"`
}

func (c CreateCategoryRequest) Validate() error {
	return validateString(c.Name)
}

func (h HTTPFeedServer) postCategories(c *fiber.Ctx) error {
	var createCategoryRequest CreateCategoryRequest
	if err := c.BodyParser(&createCategoryRequest); err != nil {
		return err
	}
	if err := createCategoryRequest.Validate(); err != nil {
		return err
	}

	category := rsscollector.FeedCategory{
		Name: createCategoryRequest.Name,
	}

	if err := h.categoryRepos.StoreCategory(&category); err != nil {
		return err
	}

	return c.JSON(category)
}

func (h HTTPFeedServer) getCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	if err := validateID(categoryID); err != nil {
		return err
	}

	category, err := h.categoryRepos.FetchCategoryByID(categoryID)
	if err != nil {
		return err
	}

	return c.JSON(category)
}

type UpdateCategoryRequest struct {
	Name string `json:"categoryName"`
}

func (u UpdateCategoryRequest) Validate() error {
	return validateString(u.Name)
}

func (h HTTPFeedServer) putCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	if err := validateID(categoryID); err != nil {
		return err
	}
	var updateCategoryRequest UpdateCategoryRequest
	if err := c.BodyParser(&updateCategoryRequest); err != nil {
		return err
	}
	if err := updateCategoryRequest.Validate(); err != nil {
		return err
	}

	category, err := h.categoryRepos.FetchCategoryByID(categoryID)
	if err != nil {
		return err
	}

	category.Name = updateCategoryRequest.Name

	if err := h.categoryRepos.StoreCategory(&category); err != nil {
		return err
	}

	return c.JSON(category)
}

func (h HTTPFeedServer) deleteCategory(c *fiber.Ctx) error {
	categoryID := c.Params("id")
	if err := validateID(categoryID); err != nil {
		return err
	}
	if err := h.categoryRepos.DeleteCategoryByID(categoryID); err != nil {
		return err
	}

	return c.JSON(categoryID)
}
