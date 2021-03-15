package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateID(t *testing.T) {
	testCases := []struct {
		Name          string
		TestID        string
		ErrorExpected bool
	}{
		{
			"Valid UUIDv4",
			"525c540e-a051-44d3-b31e-8ff882365c7f",
			false,
		},
		{
			"Invalid ID",
			"bad",
			true,
		},
		{
			"Empty ID",
			"  ",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateID(tc.TestID)
			if tc.ErrorExpected {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateFeedURL(t *testing.T) {
	testCases := []struct {
		Name          string
		TestFeedURL   string
		ErrorExpected bool
	}{
		{
			"Valid URL",
			"http://feeds.bbci.co.uk/news/technology/rss.xml",
			false,
		},
		{
			"Invalid URL",
			"feeds.bbci.co.uk",
			true,
		},
		{
			"Empty URL",
			" ",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateFeedURL(tc.TestFeedURL)
			if tc.ErrorExpected {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	testCases := []struct {
		Name          string
		Input         string
		ErrorExpected bool
	}{
		{
			"Valid string",
			"Good stuff",
			false,
		},
		{
			"Empty string",
			"",
			true,
		},
		{
			"Just Spaces",
			"   ",
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateString(tc.Input)
			if tc.ErrorExpected {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
