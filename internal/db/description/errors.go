// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package description

import "github.com/sourcenetwork/defradb/errors"

const (
	errFailedToCloseCollectionVersionQuery string = "failed to close collection version prefix query"
	errFailedToCloseCollectionQuery        string = "failed to close collection prefix query"
	errSaveCollection                      string = "failed to save collection"
	errGetCollectionByID                   string = "failed to get collection by version ID"
	errGetCollectionByName                 string = "failed to get collection by name"
	errGetCollections                      string = "failed to get collections"
	errGetActiveCollections                string = "failed to get active collections"
	errGetCollectionVersions               string = "failed to get collection versions"
	errDeleteCollection                    string = "failed to delete collection"
	errCheckCollectionExists               string = "failed to check if collection exists"
)

// NewErrFailedToCloseCollectionVersionQuery returns a new error indicating that the query
// to get a collection version failed to close.
func NewErrFailedToCloseCollectionVersionQuery(inner error) error {
	return errors.Wrap(errFailedToCloseCollectionVersionQuery, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed to close.
func NewErrFailedToCloseCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCloseCollectionQuery, inner)
}

// NewErrSaveCollection returns a new error indicating that saving the collection failed.
func NewErrSaveCollection(inner error, collectionID string) error {
	return errors.Wrap(errSaveCollection, inner, errors.NewKV("CollectionID", collectionID))
}

// NewErrGetCollectionByID returns a new error indicating that getting the collection by version ID failed.
func NewErrGetCollectionByID(inner error, versionID string) error {
	return errors.Wrap(errGetCollectionByID, inner, errors.NewKV("VersionID", versionID))
}

// NewErrGetCollectionByName returns a new error indicating that getting the collection by name failed.
func NewErrGetCollectionByName(inner error, name string) error {
	return errors.Wrap(errGetCollectionByName, inner, errors.NewKV("Name", name))
}

// NewErrGetCollections returns a new error indicating that getting all collections failed.
func NewErrGetCollections(inner error) error {
	return errors.Wrap(errGetCollections, inner)
}

// NewErrGetActiveCollections returns a new error indicating that getting active collections failed.
func NewErrGetActiveCollections(inner error) error {
	return errors.Wrap(errGetActiveCollections, inner)
}

// NewErrGetCollectionVersions returns a new error indicating that getting collection versions failed.
func NewErrGetCollectionVersions(inner error, collectionID string) error {
	return errors.Wrap(errGetCollectionVersions, inner, errors.NewKV("CollectionID", collectionID))
}

// NewErrDeleteCollection returns a new error indicating that deleting the collection failed.
func NewErrDeleteCollection(inner error, collectionID string) error {
	return errors.Wrap(errDeleteCollection, inner, errors.NewKV("CollectionID", collectionID))
}

// NewErrCheckCollectionExists returns a new error indicating that checking collection existence failed.
func NewErrCheckCollectionExists(inner error, name string) error {
	return errors.Wrap(errCheckCollectionExists, inner, errors.NewKV("Name", name))
}
