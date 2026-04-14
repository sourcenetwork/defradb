// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package id

import "github.com/sourcenetwork/defradb/errors"

const (
	errGetShortCollectionID    string = "failed to get short collection ID"
	errParseShortCollectionID  string = "failed to parse short collection ID"
	errCheckShortCollectionID  string = "failed to check short collection ID"
	errGetCollectionIDSequence string = "failed to get collection ID sequence"
	errNextCollectionIDSeq     string = "failed to get next collection ID sequence value"
	errStoreShortCollectionID  string = "failed to store short collection ID"
	errDeleteShortCollectionID string = "failed to delete short collection ID"
	errGetShortFieldIDs        string = "failed to get short field IDs"
	errParseShortFieldID       string = "failed to parse short field ID"
	errCheckShortFieldID       string = "failed to check short field ID"
	errGetFieldIDSequence      string = "failed to get field ID sequence"
	errNextFieldIDSeq          string = "failed to get next field ID sequence value"
	errStoreShortFieldID       string = "failed to store short field ID"
	errDeleteShortFieldID      string = "failed to delete short field ID"
)

func NewErrGetShortCollectionID(inner error, collectionID string) error {
	return errors.Wrap(errGetShortCollectionID, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrParseShortCollectionID(inner error, collectionID string) error {
	return errors.Wrap(errParseShortCollectionID, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrCheckShortCollectionID(inner error, collectionID string) error {
	return errors.Wrap(errCheckShortCollectionID, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrGetCollectionIDSequence(inner error, collectionID string) error {
	return errors.Wrap(errGetCollectionIDSequence, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrNextCollectionIDSeq(inner error, collectionID string) error {
	return errors.Wrap(errNextCollectionIDSeq, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrStoreShortCollectionID(inner error, collectionID string) error {
	return errors.Wrap(errStoreShortCollectionID, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrDeleteShortCollectionID(inner error, collectionID string) error {
	return errors.Wrap(errDeleteShortCollectionID, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrGetShortFieldIDs(inner error, collectionShortID uint32, fieldID string) error {
	return errors.Wrap(errGetShortFieldIDs, inner,
		errors.NewKV("CollectionShortID", collectionShortID),
		errors.NewKV("FieldID", fieldID))
}

func NewErrParseShortFieldID(inner error, collectionShortID uint32) error {
	return errors.Wrap(errParseShortFieldID, inner, errors.NewKV("CollectionShortID", collectionShortID))
}

func NewErrCheckShortFieldID(inner error, collectionShortID uint32, fieldID string) error {
	return errors.Wrap(errCheckShortFieldID, inner,
		errors.NewKV("CollectionShortID", collectionShortID),
		errors.NewKV("FieldID", fieldID))
}

func NewErrGetFieldIDSequence(inner error, collectionShortID uint32) error {
	return errors.Wrap(errGetFieldIDSequence, inner, errors.NewKV("CollectionShortID", collectionShortID))
}

func NewErrNextFieldIDSeq(inner error, collectionShortID uint32) error {
	return errors.Wrap(errNextFieldIDSeq, inner, errors.NewKV("CollectionShortID", collectionShortID))
}

func NewErrStoreShortFieldID(inner error, collectionShortID uint32, fieldID string) error {
	return errors.Wrap(errStoreShortFieldID, inner,
		errors.NewKV("CollectionShortID", collectionShortID),
		errors.NewKV("FieldID", fieldID))
}

func NewErrDeleteShortFieldID(inner error, collectionShortID uint32, fieldID string) error {
	return errors.Wrap(errDeleteShortFieldID, inner,
		errors.NewKV("CollectionShortID", collectionShortID),
		errors.NewKV("FieldID", fieldID))
}
