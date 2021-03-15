package main

import (
	"os"
	"strconv"

	"github.com/JonPulfer/rss_collector/pkg/repository"
	"github.com/JonPulfer/rss_collector/pkg/server"

	"github.com/rs/zerolog/log"
)

func main() {

	var itemRepos repository.FeedItemStore
	var categoryRepos repository.FeedCategoryStore
	var feedRepos repository.FeedSourceStore

	// This memory based repository satisfies all of the object store interfaces
	itemRepos = repository.NewMemoryStore()
	categoryRepos = repository.NewMemoryStore()
	feedRepos = repository.NewMemoryStore()

	if len(os.Getenv("DATABASE_URL")) > 0 {
		dbRepos, err := repository.NewPostgresDB(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Error().Err(err).Msg("failed to connect to db")
			panic(err)
		}
		log.Debug().Msg("connected to database")
		if len(os.Getenv("MIGRATIONS_DIR")) > 0 {
			if err := dbRepos.Migrate(os.Getenv("MIGRATIONS_DIR")); err != nil {
				log.Error().Err(err).Msg("failed to migrate database")
				panic(err)
			}
		}
		itemRepos = dbRepos
		categoryRepos = dbRepos
		feedRepos = dbRepos
	}

	port := uint(8080)
	if len(os.Getenv("PORT")) > 0 {
		envPort, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Error().Err(err).Msg("failed to parse supplied PORT envvar")
			panic(err)
		}
		port = uint(envPort)
		log.Info().Msgf("port configured as %d", port)
	}
	s := server.NewHTTPFeedServer(feedRepos, itemRepos, categoryRepos, &server.Config{Port: port})
	err := s.Start()
	if err != nil {
		log.Error().Err(err).Msg("server detected an error")
		panic(err)
	}
}
