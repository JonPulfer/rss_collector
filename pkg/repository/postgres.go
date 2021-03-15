package repository

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"

	rsscollector "github.com/JonPulfer/rss_collector/pkg"
)

const ConnectionRetryLimit = 20
const ConnectionRetryDelay = 3 * time.Second

type PostgresDB struct {
	conn *sql.DB
}

func (p PostgresDB) StoreCategory(category *rsscollector.FeedCategory) error {
	insertSql := `
insert into categories (id, category_name) values($1, $2)
on conflict do nothing;`

	if len(category.ID) == 0 {
		u, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		category.ID = u.String()
		_, err = p.conn.Exec(insertSql, category.ID, category.Name)
		return err
	}

	updateSql := `update categories set category_name = $2 where id = $1;`
	_, err := p.conn.Exec(updateSql, category.ID, category.Name)
	return err
}

func (p PostgresDB) FetchCategoryByID(id string) (rsscollector.FeedCategory, error) {
	selectSql := `select category_name from categories where id = $1;`
	var result rsscollector.FeedCategory
	rows, err := p.conn.Query(selectSql, id)
	if err != nil {
		return rsscollector.FeedCategory{}, err
	}
	if rows.Err() != nil {
		return rsscollector.FeedCategory{}, rows.Err()
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return rsscollector.FeedCategory{}, err
		}
		result = rsscollector.FeedCategory{
			ID:   id,
			Name: name,
		}
	}

	if len(result.Name) > 0 {
		return result, nil
	}

	return rsscollector.FeedCategory{}, fmt.Errorf("no category found with id: %s", id)
}

func (p PostgresDB) FetchCategoryByName(name string) (rsscollector.FeedCategory, error) {
	selectSql := `select id from categories where category_name = $1;`
	var result rsscollector.FeedCategory
	rows, err := p.conn.Query(selectSql, name)
	if err != nil {
		return rsscollector.FeedCategory{}, err
	}
	if rows.Err() != nil {
		return rsscollector.FeedCategory{}, rows.Err()
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return rsscollector.FeedCategory{}, err
		}
		result = rsscollector.FeedCategory{
			ID:   id,
			Name: name,
		}
	}

	if len(result.ID) > 0 {
		return result, nil
	}

	return rsscollector.FeedCategory{}, fmt.Errorf("no category found with name: %s", name)
}

func (p PostgresDB) FetchCategoriesForIDs(ids []string) ([]rsscollector.FeedCategory, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	selectSql := `select id, category_name from categories where id = ANY($1);`
	results := make([]rsscollector.FeedCategory, 0)
	rows, err := p.conn.Query(selectSql, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		results = append(results, rsscollector.FeedCategory{
			ID:   id,
			Name: name,
		})
	}

	if len(results) > 0 {
		return results, nil
	}

	return nil, fmt.Errorf("no categories found with ids: %v", ids)
}

func (p PostgresDB) FetchAllCategories() ([]rsscollector.FeedCategory, error) {
	selectSql := `select id, category_name from categories;`
	rows, err := p.conn.Query(selectSql)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	results := make([]rsscollector.FeedCategory, 0)
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		results = append(results, rsscollector.FeedCategory{
			ID:   id,
			Name: name,
		})
	}
	return results, nil
}

func (p PostgresDB) DeleteCategoryByID(id string) error {
	deleteItemLinksSql := `delete from item_categories where category_id = $1;`
	_, err := p.conn.Exec(deleteItemLinksSql, id)
	if err != nil {
		return err
	}
	deleteSourceLinksSql := `delete from feed_categories where category_id = $1;`
	_, err = p.conn.Exec(deleteSourceLinksSql, id)
	if err != nil {
		return err
	}
	deleteCategorySql := `delete from categories where id = $1;`
	_, err = p.conn.Exec(deleteCategorySql, id)
	if err != nil {
		return err
	}
	return nil
}

func (p PostgresDB) StoreItem(sourceID string, item *rsscollector.FeedItem) error {

	var buf bytes.Buffer

	if len(item.ID) == 0 {
		insertSql := `insert into items (id, source_id, item_data) 
values($1, $2, $3) on conflict do nothing;`
		u, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		item.ID = u.String()
		if err := json.NewEncoder(&buf).Encode(&item); err != nil {
			return err
		}

		_, err = p.conn.Exec(insertSql, item.ID, sourceID, buf.String())
		if err != nil {
			return err
		}
		return p.updateCategoriesForItem(item)
	}

	if err := p.updateCategoriesForItem(item); err != nil {
		return err
	}
	if err := json.NewEncoder(&buf).Encode(&item); err != nil {
		return err
	}
	updateSql := `update items set (source_id, item_data) = ($2, $3) where id = $1;`
	_, err := p.conn.Exec(updateSql, item.ID, sourceID, buf.String())
	if err != nil {
		return err
	}
	return nil
}

func (p PostgresDB) updateCategoriesForItem(item *rsscollector.FeedItem) error {
	if len(item.CategoryIDs) > 0 {
		storedCategoryIDs := make([]string, 0)
		selectItemCategoriesSql := `select category_id from item_categories where item_id = $1`
		rows, err := p.conn.Query(selectItemCategoriesSql, item.ID)
		if err != nil {
			return err
		}
		if rows.Err() != nil {
			return rows.Err()
		}

		for rows.Next() {
			var storedCategoryID string
			if err := rows.Scan(&storedCategoryID); err != nil {
				return err
			}
			storedCategoryIDs = append(storedCategoryIDs, storedCategoryID)
		}
		rows.Close()

		categoryIDsToAdd := linkCategoryIDs(item.CategoryIDs, storedCategoryIDs)
		if len(categoryIDsToAdd) > 0 {
			for _, categoryID := range categoryIDsToAdd {
				relateCategorySql := `
insert into item_categories (item_id, category_id) values ($1, $2) on conflict do nothing;`
				_, err := p.conn.Exec(relateCategorySql, item.ID, categoryID)
				if err != nil {
					return err
				}
			}
		}

		categoryIDsToDelete := unlinkCategoryIDs(item.CategoryIDs, storedCategoryIDs)
		if len(categoryIDsToDelete) > 0 {
			for _, deleteCategoryID := range categoryIDsToDelete {
				deleteSql := `delete from item_categories where item_id = $1 and category_id = $2;`
				_, err := p.conn.Exec(deleteSql, item.ID, deleteCategoryID)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func linkCategoryIDs(newList, oldList []string) []string {
	results := make([]string, 0)
	for _, item := range newList {
		var itemFound bool
		for _, potentialMatch := range oldList {
			if item == potentialMatch {
				itemFound = true
			}
		}
		if !itemFound {
			results = append(results, item)
		}
	}
	return results
}

func unlinkCategoryIDs(newList, oldList []string) []string {
	results := make([]string, 0)
	for _, oldItem := range oldList {
		var stillRequired bool
		for _, potentialMatch := range newList {
			if oldItem == potentialMatch {
				stillRequired = true
			}
		}
		if !stillRequired {
			results = append(results, oldItem)
		}
	}
	return results
}

func (p PostgresDB) StoreItems(sourceID string, items []*rsscollector.FeedItem) error {
	for _, item := range items {
		if err := p.StoreItem(sourceID, item); err != nil {
			return err
		}
	}
	return nil
}

func (p PostgresDB) FetchItemByID(id string) (rsscollector.FeedItem, error) {
	selectSql := `select source_id, item_data from items where id = $1;`
	rows, err := p.conn.Query(selectSql, id)
	if err != nil {
		return rsscollector.FeedItem{}, err
	}
	if rows.Err() != nil {
		return rsscollector.FeedItem{}, rows.Err()
	}
	defer rows.Close()

	var result rsscollector.FeedItem
	for rows.Next() {
		var sourceID, data string
		if err := rows.Scan(&sourceID, &data); err != nil {
			return rsscollector.FeedItem{}, err
		}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return rsscollector.FeedItem{}, err
		}
		result.SourceID = sourceID
	}
	if len(result.ID) > 0 {
		return result, nil
	}
	return rsscollector.FeedItem{}, fmt.Errorf("no item found with id: %s", id)
}

func (p PostgresDB) FetchAllItems(options rsscollector.ItemOptions) (rsscollector.FeedItems, error) {
	var rows *sql.Rows
	var err error

	if len(options.CategoryIDs) > 0 {
		itemIDs := make([]string, 0)
		for _, categoryID := range options.CategoryIDs {
			selectSql := `select item_id from item_categories where category_id = $1;`
			rows, err := p.conn.Query(selectSql, categoryID)
			if err != nil {
				return rsscollector.FeedItems{}, err
			}
			if rows.Err() != nil {
				return rsscollector.FeedItems{}, rows.Err()
			}

			for rows.Next() {
				var itemID string
				if err := rows.Scan(&itemID); err != nil {
					return rsscollector.FeedItems{}, err
				}
				itemIDs = append(itemIDs, itemID)
			}
			rows.Close()
		}
		if len(options.SourceID) > 0 {
			selectSql := `select id, source_id, item_data from items where source_id = $1 and id = ANY($2);`
			rows, err = p.conn.Query(selectSql, options.SourceID, pq.Array(itemIDs))
			if err != nil {
				return rsscollector.FeedItems{}, err
			}
			if rows.Err() != nil {
				return rsscollector.FeedItems{}, rows.Err()
			}
		} else {
			selectSql := `select id, source_id, item_data from items where id = ANY($1);`
			rows, err = p.conn.Query(selectSql, pq.Array(itemIDs))
			if err != nil {
				return rsscollector.FeedItems{}, err
			}
			if rows.Err() != nil {
				return rsscollector.FeedItems{}, rows.Err()
			}

		}
	} else if len(options.SourceID) > 0 {
		selectSql := `select id, source_id, item_data from items where source_id = $1;`
		rows, err = p.conn.Query(selectSql, options.SourceID)
		if err != nil {
			return rsscollector.FeedItems{}, err
		}
		if rows.Err() != nil {
			return rsscollector.FeedItems{}, rows.Err()
		}
	} else {
		selectSql := `select id, source_id, item_data from items;`
		rows, err = p.conn.Query(selectSql)
		if err != nil {
			return rsscollector.FeedItems{}, err
		}
		if rows.Err() != nil {
			return rsscollector.FeedItems{}, rows.Err()
		}
	}
	defer rows.Close()

	results := make(rsscollector.FeedItems, 0)
	for rows.Next() {
		var id, sourceID, data string
		if err := rows.Scan(&id, &sourceID, &data); err != nil {
			return rsscollector.FeedItems{}, err
		}
		var feedItem rsscollector.FeedItem
		if err := json.Unmarshal([]byte(data), &feedItem); err != nil {
			return rsscollector.FeedItems{}, err
		}
		feedItem.SourceID = sourceID
		results = append(results, &feedItem)
	}

	if len(results) > 0 {
		return results, nil
	}

	return rsscollector.FeedItems{}, fmt.Errorf("no items found for ItemOptions: %v", options)
}

func (p PostgresDB) DeleteItemByID(id string) error {
	deleteItemCategoriesSql := `delete from item_categories where item_id = $1;`
	_, err := p.conn.Exec(deleteItemCategoriesSql, id)
	if err != nil {
		return err
	}
	deleteSql := `delete from items where id = $1;`
	_, err = p.conn.Exec(deleteSql, id)
	return err
}

func (p PostgresDB) StoreSource(source *rsscollector.FeedSource) error {

	if len(source.ID) == 0 {
		insertSql := `insert into feeds (id, feed_url, feed_data) values($1, $2, $3);`
		u, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		source.ID = u.String()
		source.Link = rsscollector.FeedSourceLink(source.ID)
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(&source); err != nil {
			return err
		}
		if len(source.CategoryIDs) > 0 {
			for _, categoryID := range source.CategoryIDs {
				linkCategorySql := `insert into feed_categories (feed_id, category_id) values ($1, $2);`
				_, err := p.conn.Exec(linkCategorySql, source.ID, categoryID)
				if err != nil {
					return err
				}
			}
		}
		_, err = p.conn.Exec(insertSql, source.ID, source.FeedURL, buf.String())
		if err != nil {
			return err
		}
		return nil
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&source); err != nil {
		return err
	}

	updateSql := `update feeds set (feed_url, feed_data) = ($2, $3) where id = $1;`
	_, err := p.conn.Exec(updateSql, source.ID, source.FeedURL, buf.String())
	if err != nil {
		return err
	}

	if len(source.CategoryIDs) > 0 {
		selectCategoriesSql := `select category_id from feed_categories where feed_id = $1;`
		rows, err := p.conn.Query(selectCategoriesSql, source.ID)
		if err != nil {
			return err
		}
		if rows.Err() != nil {
			return rows.Err()
		}
		storedCategoryIDs := make([]string, 0)
		for rows.Next() {
			var categoryID string
			if err := rows.Scan(&categoryID); err != nil {
				return err
			}
			storedCategoryIDs = append(storedCategoryIDs, categoryID)
		}
		rows.Close()

		feedCategoriesToAdd := linkCategoryIDs(source.CategoryIDs, storedCategoryIDs)
		if len(feedCategoriesToAdd) > 0 {
			for _, categoryToAdd := range feedCategoriesToAdd {
				linkFeedCategoriesSql := `insert into feed_categories (feed_id, category_id) values ($1, $2);`
				_, err := p.conn.Exec(linkFeedCategoriesSql, source.ID, categoryToAdd)
				if err != nil {
					return err
				}
			}
		}

		feedCategoriesToRemove := unlinkCategoryIDs(source.CategoryIDs, storedCategoryIDs)
		if len(feedCategoriesToRemove) > 0 {
			for _, categoryToRemove := range feedCategoriesToRemove {
				unlinkFeedCategorySql := `delete from feed_categories where feed_id = $1 and category_id = $2;`
				_, err := p.conn.Exec(unlinkFeedCategorySql, source.ID, categoryToRemove)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (p PostgresDB) FetchSource(feedID string) (rsscollector.FeedSource, error) {
	selectSql := `select feed_data from feeds where id = $1`
	rows, err := p.conn.Query(selectSql, feedID)
	if err != nil {
		return rsscollector.FeedSource{}, err
	}
	if rows.Err() != nil {
		return rsscollector.FeedSource{}, rows.Err()
	}
	defer rows.Close()

	var result rsscollector.FeedSource
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return rsscollector.FeedSource{}, err
		}
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return rsscollector.FeedSource{}, err
		}
	}
	if len(result.ID) > 0 {
		return result, nil
	}

	return rsscollector.FeedSource{}, fmt.Errorf("no feed source found for id: %s", feedID)
}

func (p PostgresDB) FetchAllSources() ([]rsscollector.FeedSourcePartial, error) {
	selectSql := `select feed_data from feeds;`
	rows, err := p.conn.Query(selectSql)
	if err != nil {
		return []rsscollector.FeedSourcePartial{}, err
	}
	if rows.Err() != nil {
		return []rsscollector.FeedSourcePartial{}, rows.Err()
	}
	defer rows.Close()

	results := make([]rsscollector.FeedSourcePartial, 0)
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return []rsscollector.FeedSourcePartial{}, err
		}
		var feed rsscollector.FeedSourcePartial
		if err := json.Unmarshal([]byte(data), &feed); err != nil {
			return []rsscollector.FeedSourcePartial{}, err
		}
		results = append(results, feed)
	}

	if len(results) > 0 {
		return results, nil
	}

	return []rsscollector.FeedSourcePartial{}, fmt.Errorf("no feeds found")
}

func (p PostgresDB) DeleteSourceByID(id string) error {
	selectSourceItemIDs := `select id from items where source_id = $1;`
	rows, err := p.conn.Query(selectSourceItemIDs, id)
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err != nil {
			return err
		}
		if err := p.DeleteItemByID(itemID); err != nil {
			return err
		}
	}
	rows.Close()

	deleteSourceCategoriesSql := `delete from feed_categories where feed_id = $1;`
	_, err = p.conn.Exec(deleteSourceCategoriesSql, id)
	if err != nil {
		return err
	}

	deleteSourceSql := `delete from feeds where id = $1;`
	_, err = p.conn.Exec(deleteSourceSql, id)
	return err
}

func NewPostgresDB(connectionString string) (*PostgresDB, error) {
	conn, err := retryConnection(connectionString)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{conn: conn}, nil
}

func retryConnection(connectionString string) (*sql.DB, error) {

	for i := 0; i < ConnectionRetryLimit; i++ {
		time.Sleep(ConnectionRetryDelay)
		conn, err := sql.Open("postgres", connectionString)
		if err == nil {
			return conn, nil
		}
		log.Debug().Msg("retrying database connection")
	}
	return nil, fmt.Errorf("failed to connect to database after %d retries",
		ConnectionRetryLimit)
}

func (p PostgresDB) Migrate(migrationsDirectory string) error {
	driver, err := postgres.WithInstance(p.conn, &postgres.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file:///%s", migrationsDirectory),
		"postgres", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil {
		if err != migrate.ErrNoChange {
			return err
		}
	}
	return nil
}
