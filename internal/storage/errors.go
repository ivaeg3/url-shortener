package storage

import "errors"

var (
	ErrNotFound     = errors.New("URL not found")
	ErrDuplicateURL = errors.New("original URL already exists")
)
