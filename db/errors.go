package db

import "errors"

// errors
var (
	ErrDocumentNotExists = errors.New("No document for the given key exists")
)
