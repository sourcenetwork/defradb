// Copyright 2024 Democratized Data Foundation
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

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/keys"
)

// setCollectionIDs sets the IDs on a collection description, including field IDs, mutating the input set.
func (db *db) setCollectionIDs(ctx context.Context, newCollections []client.CollectionDefinition) error {
	err := db.setCollectionID(ctx, newCollections)
	if err != nil {
		return err
	}

	return db.setFieldIDs(ctx, newCollections)
}

// setCollectionID sets the IDs directly on a collection description, excluding stuff like field IDs,
// mutating the input set.
func (db *db) setCollectionID(ctx context.Context, newCollections []client.CollectionDefinition) error {
	colSeq, err := db.getSequence(ctx, keys.CollectionIDSequenceKey{})
	if err != nil {
		return err
	}

	for i := range newCollections {
		if len(newCollections[i].Description.Fields) == 0 {
			// This is a schema-only definition, we should not create a collection for it
			continue
		}

		colID, err := colSeq.next(ctx)
		if err != nil {
			return err
		}

		// Unlike schema, collections can be mutated and thus we need to make sure this function
		// does not assign new IDs to existing collections.
		if newCollections[i].Description.ID == 0 {
			newCollections[i].Description.ID = uint32(colID)
		}

		if newCollections[i].Description.RootID == 0 {
			newCollections[i].Description.RootID = uint32(colID)
		}
	}

	return nil
}

// setFieldIDs sets the field IDs hosted on the given collections, mutating the input set.
func (db *db) setFieldIDs(ctx context.Context, definitions []client.CollectionDefinition) error {
	collectionsByName := map[string]client.CollectionDescription{}
	schemasByName := map[string]client.SchemaDescription{}
	for _, def := range definitions {
		if def.Description.Name.HasValue() {
			collectionsByName[def.Description.Name.Value()] = def.Description
		}
		schemasByName[def.Schema.Name] = def.Schema
	}

	for i := range definitions {
		fieldSeq, err := db.getSequence(ctx, keys.NewFieldIDSequenceKey(definitions[i].Description.RootID))
		if err != nil {
			return err
		}

		for j := range definitions[i].Description.Fields {
			var fieldID client.FieldID
			if definitions[i].Description.Fields[j].ID != client.FieldID(0) {
				fieldID = definitions[i].Description.Fields[j].ID
			} else if definitions[i].Description.Fields[j].Name == request.DocIDFieldName {
				// There is no hard technical requirement for this, we just think it looks nicer
				// if the doc id is at the zero index.  It makes it look a little nicer in commit
				// queries too.
				fieldID = 0
			} else {
				nextID, err := fieldSeq.next(ctx)
				if err != nil {
					return err
				}
				fieldID = client.FieldID(nextID)
			}

			if definitions[i].Description.Fields[j].Kind.HasValue() {
				switch kind := definitions[i].Description.Fields[j].Kind.Value().(type) {
				case *client.NamedKind:
					var newKind client.FieldKind
					if kind.Name == definitions[i].Description.Name.Value() {
						newKind = client.NewSelfKind("", kind.IsArray())
					} else if otherCol, ok := collectionsByName[kind.Name]; ok {
						newKind = client.NewCollectionKind(otherCol.RootID, kind.IsArray())
					} else if otherSchema, ok := schemasByName[kind.Name]; ok {
						newKind = client.NewSchemaKind(otherSchema.Root, kind.IsArray())
					} else {
						// Continue, and let the validation stage return user friendly errors
						// if appropriate
						continue
					}

					definitions[i].Description.Fields[j].Kind = immutable.Some(newKind)
				default:
					// no-op
				}
			}

			definitions[i].Description.Fields[j].ID = fieldID
		}
	}

	return nil
}
