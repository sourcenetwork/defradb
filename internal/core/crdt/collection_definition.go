// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package crdt

import (
	"bytes"
	"context"

	"github.com/fxamacker/cbor/v2"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/keys"
)

type CollectionDefinitionDelta struct {
	Priority uint64

	Name           *string
	QuerySelect    []byte
	QueryTransform *string
}

var _ Delta = (*CollectionDefinitionDelta)(nil)

func (d *CollectionDefinitionDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type CollectionDefinitionDelta struct {
		priority  		Int
		name optional String
		querySelect optional Bytes
		queryTransform optional String
	}`)
}

func (d *CollectionDefinitionDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *CollectionDefinitionDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type CollectionDefinition struct {
	headstorePrefix keys.HeadstoreCollectionDefinition
}

var _ ReplicatedData = (*Collection)(nil)

func NewCollectionDefinition(
	name string,
) *CollectionDefinition {
	return &CollectionDefinition{
		// WARNING: This prefix will need to be rebuilt if/when we allow the mutation of collection
		// name.
		headstorePrefix: keys.HeadstoreCollectionDefinition{
			CollectionName: name,
		},
	}
}

func (c *CollectionDefinition) HeadstorePrefix() keys.HeadstoreKey {
	return c.headstorePrefix
}

func (c *CollectionDefinition) Delta(
	new client.CollectionVersion,
	old client.CollectionVersion,
) (*CollectionDefinitionDelta, bool, error) {
	var name *string
	if new.Name != old.Name {
		name = &new.Name
	}

	var queryDelta []byte
	if new.Query.HasValue() {
		newQuery, err := cbor.Marshal(new.Query.Value().Query)
		if err != nil {
			return &CollectionDefinitionDelta{}, false, err
		}

		if old.Query.HasValue() {
			oldQuery, err := cbor.Marshal(old.Query.Value().Query)
			if err != nil {
				return &CollectionDefinitionDelta{}, false, err
			}

			if !bytes.Equal(newQuery, oldQuery) {
				queryDelta = newQuery
			}
		}
	}

	var transformDelta *string
	if new.Query.HasValue() && new.Query.Value().Transform.HasValue() {
		newLensID := new.Query.Value().Transform.Value()

		if old.Query.HasValue() && old.Query.Value().Transform.HasValue() {
			if new.Query.Value().Transform.Value() != old.Query.Value().Transform.Value() {
				transformDelta = &newLensID
			}
		} else {
			transformDelta = &newLensID
		}
	} else if old.Query.HasValue() && old.Query.Value().Transform.HasValue() {
		newLensID := ""
		transformDelta = &newLensID
	}

	if name == nil && queryDelta == nil && transformDelta == nil {
		return &CollectionDefinitionDelta{}, false, nil
	}

	return &CollectionDefinitionDelta{
		Name:           name,
		QuerySelect:    queryDelta,
		QueryTransform: transformDelta,
	}, true, nil
}

func (c *CollectionDefinition) Merge(ctx context.Context, other Delta) error {
	return nil
}
