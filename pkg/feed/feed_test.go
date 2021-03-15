package feed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollection(t *testing.T) {
	source, err := NewSource("http://feeds.bbci.co.uk/news/uk/rss.xml")
	require.Nil(t, err)

	require.Nil(t, source.Collect())
	require.NotNil(t, source.Feed)
	assert.Equal(t, "BBC News - UK", source.Feed.Title)
	assert.NotEmpty(t, source.Feed.Items)
}
