// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/fulltextindex"
	"github.com/sourcenetwork/defradb/internal/db/id"
)

// collectionFullTextIndex adapts documents and the generic CollectionIndex lifecycle to the
// algorithm/storage module in internal/db/fulltextindex.
type collectionFullTextIndex struct {
	collectionBaseIndex
	fullTextDesc *client.FullTextIndexDescription

	collectionShortID uint32
	shortIDResolved   bool
}

var _ client.CollectionIndex = (*collectionFullTextIndex)(nil)

func newCollectionFullTextIndex(base collectionBaseIndex) (client.CollectionIndex, error) {
	desc, ok := base.desc.GetFullText()
	if !ok || desc == nil {
		return nil, NewErrCorruptedIndexKindDescription(base.desc.Name, base.desc.Kind)
	}
	return &collectionFullTextIndex{collectionBaseIndex: base, fullTextDesc: desc}, nil
}

func (i *collectionFullTextIndex) Save(ctx context.Context, doc *client.Document) error {
	text, err := i.documentText(doc)
	if err != nil {
		return err
	}
	collectionShortID, docShortID, err := i.documentShortIDs(ctx, doc)
	if err != nil {
		return err
	}
	index, err := fulltextindex.Open(
		ctx, collectionShortID, i.desc.ID, i.epoch, *i.fullTextDesc,
	)
	if err != nil {
		return err
	}
	return index.Insert(docShortID, text)
}

func (i *collectionFullTextIndex) Update(
	ctx context.Context,
	oldDoc *client.Document,
	newDoc *client.Document,
) error {
	if !isUpdatingIndexedFields(i, oldDoc, newDoc) {
		return nil
	}
	if err := i.Delete(ctx, oldDoc); err != nil {
		return NewErrUpdateIndex(err, i.desc.Name)
	}
	if err := i.Save(ctx, newDoc); err != nil {
		return NewErrUpdateIndex(err, i.desc.Name)
	}
	return nil
}

func (i *collectionFullTextIndex) Delete(ctx context.Context, doc *client.Document) error {
	text, err := i.documentText(doc)
	if err != nil {
		return err
	}
	collectionShortID, docShortID, err := i.documentShortIDs(ctx, doc)
	if err != nil {
		return err
	}
	index, err := fulltextindex.Open(
		ctx, collectionShortID, i.desc.ID, i.epoch, *i.fullTextDesc,
	)
	if err != nil {
		return err
	}
	found, err := index.Delete(docShortID, text, i.building)
	if err != nil {
		return err
	}
	if !found && !i.building {
		return NewErrCorruptedIndex(i.desc.Name)
	}
	return nil
}

func (i *collectionFullTextIndex) documentText(doc *client.Document) (string, error) {
	values, err := i.getDocFieldValues(doc)
	if err != nil {
		return "", err
	}
	if text, ok := values[0].String(); ok {
		return text, nil
	}
	if text, ok := values[0].NillableString(); ok && text.HasValue() {
		return text.Value(), nil
	}
	return "", nil
}

func (i *collectionFullTextIndex) documentShortIDs(
	ctx context.Context,
	doc *client.Document,
) (uint32, uint64, error) {
	collectionShortID, err := i.resolveCollectionShortID(ctx)
	if err != nil {
		return 0, 0, err
	}
	docShortID, found, err := id.GetDocShortID(ctx, collectionShortID, doc.ID().String())
	if err != nil {
		return 0, 0, err
	}
	if !found {
		return 0, 0, client.ErrDocumentNotFoundOrNotAuthorized
	}
	return collectionShortID, docShortID, nil
}

func (i *collectionFullTextIndex) resolveCollectionShortID(ctx context.Context) (uint32, error) {
	if i.shortIDResolved {
		return i.collectionShortID, nil
	}
	shortID, err := id.GetCollectionShortID(ctx, i.collection.Version().CollectionID)
	if err != nil {
		return 0, err
	}
	i.collectionShortID = shortID
	i.shortIDResolved = true
	return shortID, nil
}
