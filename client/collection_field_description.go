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
	"fmt"

	"github.com/sourcenetwork/immutable"
)

// FieldID is a unique identifier for a field in a schema.
type FieldID uint32

// CollectionFieldDescription describes the local components of a field on a collection.
type CollectionFieldDescription struct {
	// Name contains the name of the [SchemaFieldDescription] that this field uses.
	Name string

	// ID contains the local, internal ID of this field.
	ID FieldID

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

	// Embedding contains the configuration for generating embedding vectors.
	//
	// This is only usable with array fields.
	//
	// When configured, embeddings may call 3rd party APIs inline with document mutations.
	// This may cause increase latency in the completion of the mutation requests.
	// This is necessary to ensure that the generated docID is representative of the
	// content of the document.
	Embedding *EmbeddingDescription
}

func (f FieldID) String() string {
	return fmt.Sprint(uint32(f))
}

// collectionFieldDescription is a private type used to facilitate the unmarshalling
// of json to a [CollectionFieldDescription].
type collectionFieldDescription struct {
	Name         string
	ID           FieldID
	RelationName immutable.Option[string]
	DefaultValue any
	Size         int
	Embedding    *EmbeddingDescription

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
	f.ID = descMap.ID
	f.DefaultValue = descMap.DefaultValue
	f.RelationName = descMap.RelationName
	f.Size = descMap.Size
	f.Embedding = descMap.Embedding
	kind, err := parseFieldKind(descMap.Kind)
	if err != nil {
		return err
	}

	if kind != FieldKind_None {
		f.Kind = immutable.Some(kind)
	}

	return nil
}

// EmbeddingDescription hold the relevant information to generate embeddings.
//
// Embeddings are vectors generated based on one or multiple fields, optionaly added to a template.
type EmbeddingDescription struct {
	// Fields are the fields in the parent schema that will be used as the basis of the
	// vector generation.
	Fields []string
	// Model is the LLM of the provider to use for generating the embeddings.
	Model string
	// Provider is the API provider to use for generating the embeddings.
	Provider string
	// (Optional) Template is the local path of the template to use with the
	// field values to form the content to send to the model.
	Template string
	// URL is the url enpoint of the provider's API.
	URL string
}
