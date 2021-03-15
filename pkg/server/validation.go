package server

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

type ValidationError struct {
	Err error  `json:"error"`
	Msg string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %v", v.Msg, v.Err)
}

func validateFeedURL(feedURL string) error {
	u, err := url.Parse(feedURL)
	if err != nil {
		return err
	}
	if len(u.Scheme) == 0 {
		return ValidationError{
			Err: nil,
			Msg: "provided URL has no scheme",
		}
	}
	return nil
}

func validateID(id string) error {
	_, err := uuid.Parse(id)
	if err != nil {
		return ValidationError{
			Err: err,
			Msg: fmt.Sprintf("provided ID is not valid: %s", id),
		}
	}
	return nil
}

func validateString(s string) error {
	if len(strings.Replace(s, " ", "", -1)) == 0 {
		return ValidationError{
			Err: nil,
			Msg: "empty value provided",
		}
	}
	return nil
}
