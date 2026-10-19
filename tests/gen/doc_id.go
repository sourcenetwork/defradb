// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package gen

import (
	"context"
	"maps"

	"github.com/sourcenetwork/defradb/client"
	ccid "github.com/sourcenetwork/defradb/internal/core/cid"
)

// GenerateDocIDFromMap returns a stable seed DocID for relation fixtures.
func GenerateDocIDFromMap(
	ctx context.Context,
	data map[string]any,
	collection client.CollectionVersion,
) (client.DocID, error) {
	doc, err := client.NewDocFromMap(ctx, maps.Clone(data), collection)
	if err != nil {
		return client.DocID{}, err
	}
	bytes, err := doc.Bytes()
	if err != nil {
		return client.DocID{}, err
	}
	bytes = append(bytes, []byte(collection.CollectionID)...)
	cid, err := ccid.NewSHA256CidV1(bytes)
	if err != nil {
		return client.DocID{}, err
	}
	return client.NewDocIDV0(cid), nil
}

func NewDocWithGeneratedID(
	ctx context.Context,
	data map[string]any,
	collection client.CollectionVersion,
) (*client.Document, string, error) {
	docID, err := GenerateDocIDFromMap(ctx, data, collection)
	if err != nil {
		return nil, "", err
	}
	doc, err := client.NewDocFromMap(ctx, maps.Clone(data), collection)
	if err != nil {
		return nil, "", err
	}
	return doc, docID.String(), nil
}
