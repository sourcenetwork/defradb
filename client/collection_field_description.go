// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"encoding/json"

	"github.com/sourcenetwork/immutable"
)

// CollectionFieldDescription describes the local components of a field on a collection.
type CollectionFieldDescription struct {
	// The immutable ID of this field.
	//
	// Only fields persisted in the DAG will have a value - virtual fields such as secondary
	// relation fields will not have a FieldID.
	FieldID string

	// The name of this field, it will be visible throughout the system and is
	// the most common way of referencing this field.
	//
	// Must contain a valid value.
	Name string

	// The data type that this field holds.
	//
	// Must contain a valid value.
	Kind FieldKind

	// The CRDT Type of this field. If no type has been provided it will default to [LWW_REGISTER].
	Typ CType

	// RelationName contains the name of this relation, if this field is part of a relationship.
	//
	// Otherwise will be [None].
	RelationName immutable.Option[string]

	// IsPrimary indicates whether this side of the relation hosts the id of the foriegn object or not.
	//
	// If this is a relation's field, and this value is false, this field will not actually hold a value
	// in the documentDAG.
	IsPrimary bool

	// DefaultValue contains the default value for this field.
	//
	// This value has no effect on views.
	DefaultValue any

	// Size is a constraint that can be applied to fields that are arrays.
	//
	// Mutations on fields with a size constraint will fail if the size of the array
	// does not match the constraint.
	Size int
}

// collectionFieldDescription is a private type used to facilitate the unmarshalling
// of json to a [CollectionFieldDescription].
type collectionFieldDescription struct {
	FieldID      string
	Name         string
	RelationName immutable.Option[string]
	DefaultValue any
	Size         int
	Typ          CType
	IsPrimary    bool

	// Properties below this line are unmarshalled using custom logic in [UnmarshalJSON]
	Kind json.RawMessage
}

func (f *CollectionFieldDescription) UnmarshalJSON(bytes []byte) error {
	var descMap collectionFieldDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	f.FieldID = descMap.FieldID
	f.Name = descMap.Name
	f.DefaultValue = descMap.DefaultValue
	f.RelationName = descMap.RelationName
	f.Size = descMap.Size
	f.Typ = descMap.Typ
	f.IsPrimary = descMap.IsPrimary
	kind, err := parseFieldKind(descMap.Kind)
	if err != nil {
		return err
	}

	f.Kind = kind

	return nil
}
