package db

import "errors"

// errors
var (
	ErrDocumentNotFound = errors.New("No document for the given key exists")
)
