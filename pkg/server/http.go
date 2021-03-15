package server

import (
	"fmt"

	"github.com/JonPulfer/rss_collector/pkg/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Config struct {
	Port uint
}

type HTTPFeedServer struct {
	feedRepos     repository.FeedSourceStore
	itemRepos     repository.FeedItemStore
	categoryRepos repository.FeedCategoryStore
	config        *Config
}

func NewHTTPFeedServer(
	feedRepos repository.FeedSourceStore,
	itemRepos repository.FeedItemStore,
	categoryRepos repository.FeedCategoryStore,
	config *Config) *HTTPFeedServer {
	return &HTTPFeedServer{
		feedRepos:     feedRepos,
		itemRepos:     itemRepos,
		categoryRepos: categoryRepos,
		config:        config,
	}
}

func (h HTTPFeedServer) Start() error {

	app := fiber.New()
	app.Use(logger.New())

	// Cache using default expiration of 1 minute but with key generation that
	// includes the query args.
	app.Use(cache.New(cache.Config{
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.Request().URI().String()
		},
	}))

	// Feeds.
	app.Get("/feeds/", h.getFeeds)
	app.Post("/feeds/", h.postFeeds)
	app.Get("/feeds/:id", h.getFeed)
	app.Put("/feeds/:id", h.putFeed)
	app.Delete("/feeds/:id", h.deleteFeed)

	// Items.
	app.Get("/items/", h.getItems)
	app.Get("/items/:id", h.getItem)
	app.Put("/items/:id", h.putItem)
	app.Delete("/items/:id", h.deleteItem)

	// Categories.
	app.Get("/categories/", h.getCategories)
	app.Post("/categories/", h.postCategories)
	app.Get("/categories/:id", h.getCategory)
	app.Put("/categories/:id", h.putCategory)
	app.Delete("/categories/:id", h.deleteCategory)

	return app.Listen(fmt.Sprintf("0.0.0.0:%d", h.config.Port))
}
