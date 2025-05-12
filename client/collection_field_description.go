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
	// Name contains the name of the [SchemaFieldDescription] that this field uses.
	Name string

	// Kind contains the local field kind if this is a local-only field (e.g. the secondary
	// side of a relation).
	//
	// If the field is globaly defined (on the Schema), this will be [None].
	Kind immutable.Option[FieldKind]

	// RelationName contains the name of this relation, if this field is part of a relationship.
	//
	// Otherwise will be [None].
	RelationName immutable.Option[string]

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
	Name         string
	RelationName immutable.Option[string]
	DefaultValue any
	Size         int

	// Properties below this line are unmarshalled using custom logic in [UnmarshalJSON]
	Kind json.RawMessage
}

func (f *CollectionFieldDescription) UnmarshalJSON(bytes []byte) error {
	var descMap collectionFieldDescription
	err := json.Unmarshal(bytes, &descMap)
	if err != nil {
		return err
	}

	f.Name = descMap.Name
	f.DefaultValue = descMap.DefaultValue
	f.RelationName = descMap.RelationName
	f.Size = descMap.Size
	kind, err := parseFieldKind(descMap.Kind)
	if err != nil {
		return err
	}

	if kind != FieldKind_None {
		f.Kind = immutable.Some(kind)
	}

	return nil
}
