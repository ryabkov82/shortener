package storage

import "errors"

var (
	ErrURLNotFound    = errors.New("url not found")
	ErrURLExists      = errors.New("url exists")
	ErrShortURLExists = errors.New("ShortURL already exists")
	ErrURLDeleted     = errors.New("URL has been deleted")
)
