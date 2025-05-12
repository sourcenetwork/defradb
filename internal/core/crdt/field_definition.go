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
	"context"
	"strconv"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// FieldDefinitionDelta contains the properties changed between two different field versions.
//
// It is not unique to a collection, and the same block may be referenced by multiple
// collections.
type FieldDefinitionDelta struct {
	Priority uint64

	Name         *string
	Crdt         *client.CType
	ScalarKind   *client.ScalarKind
	CollectionID *string
	RelativeID   *int
	IsArray      *bool
}

var _ core.Delta = (*FieldDefinitionDelta)(nil)

func (delta *FieldDefinitionDelta) IPLDSchemaBytes() []byte {
	return []byte(`
	type FieldDefinitionDelta struct {
		priority  		Int
		name optional String
		crdt optional Int
		scalarKind optional Int
		collectionID optional String
		relativeID optional Int
		isArray optional Bool
	}`)
}

func (d *FieldDefinitionDelta) GetPriority() uint64 {
	return d.Priority
}

func (d *FieldDefinitionDelta) SetPriority(priority uint64) {
	d.Priority = priority
}

type FieldDefinition struct {
	headstorePrefix keys.HeadstoreFieldDefinition
}

var _ core.ReplicatedData = (*Collection)(nil)

func NewFieldDefinition(
	collectionName string,
	fieldName string,
) *FieldDefinition {
	return &FieldDefinition{
		// WARNING: This prefix will need to be rebuilt if/when we allow the mutation of collection
		// and/or field names.
		//
		// Whilst the field blocks are not collection specific, the heads are - a patch may update
		// a field definition on one collection but not the other - in which case the headstore
		// should differ.
		headstorePrefix: keys.HeadstoreFieldDefinition{
			CollectionName: collectionName,
			FieldName:      fieldName,
		},
	}
}

func (m *FieldDefinition) HeadstorePrefix() keys.HeadstoreKey {
	return m.headstorePrefix
}

func (m *FieldDefinition) Delta(
	new client.CollectionFieldDescription,
	old client.CollectionFieldDescription,
) (*FieldDefinitionDelta, bool, error) {
	if new.FieldID != "" {
		// This function is currently taking advantage of us not yet having implemented field-mutations
		// parts of this code will need to change.
		return nil, false, nil
	}

	if new.RelationName.HasValue() && !new.IsPrimary {
		// secondary fields are local-only and do not get saved in the blockstore
		return nil, false, nil
	}

	var scalarKind client.ScalarKind
	var relatedCollectionID string
	var relativeID int
	switch k := new.Kind.(type) {
	case client.ScalarKind:
		scalarKind = k
	case client.ScalarArrayKind:
		scalarKind = k.SubKind()
	case *client.CollectionKind:
		relatedCollectionID = k.CollectionID
	case *client.SelfKind:
		var err error
		relativeID, err = strconv.Atoi(k.RelativeID)
		if err != nil {
			return nil, false, nil
		}
	}
	isArray := new.Kind.IsArray()

	return &FieldDefinitionDelta{
		Name:         &new.Name,
		Crdt:         &new.Typ,
		ScalarKind:   &scalarKind,
		CollectionID: &relatedCollectionID,
		RelativeID:   &relativeID,
		IsArray:      &isArray,
	}, true, nil
}

func (c *FieldDefinition) Merge(ctx context.Context, other core.Delta) error {
	// WARNING: This is okay for now, as we dont (yet) support the merging of divergant heads,
	// (this is not *really* a CRDT) however, if we do want to support that at somepoint, this function
	// will need to be implemented.
	return nil
}
