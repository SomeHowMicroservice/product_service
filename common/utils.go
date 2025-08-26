package common

import (
	"time"

	"github.com/gosimple/slug"
)

func GenerateSlug(str string) string {
	return slug.Make(str)
}

func ParseDate(str string) (time.Time, error) {
	parsedDate, err := time.Parse("2006-01-02", str)
	if err != nil {
		return time.Time{}, err
	}

	return parsedDate, nil
}