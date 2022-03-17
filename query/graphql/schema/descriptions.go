// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package schema

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	gql "github.com/graphql-go/graphql"
)

var (
	// this is only here as a reference, and not to be used
	// directly. As it will yield incorrect and unexpected
	// results

	// nolint:deadcode,unused,varcheck
	gqlTypeToFieldKindReference = map[gql.Type]client.FieldKind{
		gql.ID:        client.FieldKind_DocKey,
		gql.Boolean:   client.FieldKind_BOOL,
		gql.Int:       client.FieldKind_INT,
		gql.Float:     client.FieldKind_FLOAT,
		gql.DateTime:  client.FieldKind_DATE,
		gql.String:    client.FieldKind_STRING,
		&gql.Object{}: client.FieldKind_FOREIGN_OBJECT,
		&gql.List{}:   client.FieldKind_FOREIGN_OBJECT_ARRAY,
		// More custom ones to come
		// - JSON
		// - ByteArray
		// - Counters
	}

	// This map is fine to use
	defaultCRDTForFieldKind = map[client.FieldKind]client.CType{
		client.FieldKind_DocKey:               client.LWW_REGISTER,
		client.FieldKind_BOOL:                 client.LWW_REGISTER,
		client.FieldKind_BOOL_ARRAY:           client.LWW_REGISTER,
		client.FieldKind_INT:                  client.LWW_REGISTER,
		client.FieldKind_INT_ARRAY:            client.LWW_REGISTER,
		client.FieldKind_FLOAT:                client.LWW_REGISTER,
		client.FieldKind_FLOAT_ARRAY:          client.LWW_REGISTER,
		client.FieldKind_DATE:                 client.LWW_REGISTER,
		client.FieldKind_STRING:               client.LWW_REGISTER,
		client.FieldKind_STRING_ARRAY:         client.LWW_REGISTER,
		client.FieldKind_FOREIGN_OBJECT:       client.NONE_CRDT,
		client.FieldKind_FOREIGN_OBJECT_ARRAY: client.NONE_CRDT,
	}
)

func gqlTypeToFieldKind(t gql.Type) client.FieldKind {
	switch v := t.(type) {
	case *gql.Scalar:
		switch v.Name() {
		case "ID":
			return client.FieldKind_DocKey
		case "Boolean":
			return client.FieldKind_BOOL
		case "Int":
			return client.FieldKind_INT
		case "Float":
			return client.FieldKind_FLOAT
		case "Date":
			return client.FieldKind_DATE
		case "String":
			return client.FieldKind_STRING
		}
	case *gql.Object:
		return client.FieldKind_FOREIGN_OBJECT
	case *gql.List:
		if scalar, isScalar := v.OfType.(*gql.Scalar); isScalar {
			switch scalar.Name() {
			case "Boolean":
				return client.FieldKind_BOOL_ARRAY
			case "Int":
				return client.FieldKind_INT_ARRAY
			case "Float":
				return client.FieldKind_FLOAT_ARRAY
			case "String":
				return client.FieldKind_STRING_ARRAY
			}
		}
		return client.FieldKind_FOREIGN_OBJECT_ARRAY
	}

	return client.FieldKind_None
}

func (g *Generator) CreateDescriptions(types []*gql.Object) ([]client.CollectionDescription, error) {
	// create a indexable cached map
	typeMap := make(map[string]*gql.Object)
	for _, t := range types {
		typeMap[t.Name()] = t
	}

	descs := make([]client.CollectionDescription, len(types))
	// do the real generation
	for i, t := range types {
		desc := client.CollectionDescription{
			Name: t.Name(),
		}

		// add schema
		desc.Schema.Fields = []client.FieldDescription{
			{
				Name: "_key",
				Kind: client.FieldKind_DocKey,
				Typ:  client.NONE_CRDT,
			},
		}
		// and schema fields
		for fname, field := range t.Fields() {
			if _, ok := parser.ReservedFields[fname]; ok {
				continue
			}

			// check if we already have a defined field
			// with the same name.
			// NOTE: This will happen for the virtual ID
			// field associated with a related type, as
			// its defined down below in the IsObject block.
			if _, exists := desc.GetField(fname); exists {
				// lets make sure its an _id field, otherwise
				// we might have an error here
				if strings.HasSuffix(fname, "_id") {
					continue
				} else {
					return nil, fmt.Errorf("Error: found a duplicate field '%s' for type %s", fname, t.Name())
				}
			}

			fd := client.FieldDescription{
				Name: fname,
				Kind: gqlTypeToFieldKind(field.Type),
			}
			fd.Typ = defaultCRDTForFieldKind[fd.Kind]

			if fd.IsObject() {
				schemaName := field.Type.Name()
				fd.Schema = schemaName

				// check if its a one-to-one, one-to-many, many-to-many
				rel := g.manager.Relations.GetRelationByDescription(
					fname, schemaName, t.Name())
				if rel == nil {
					return nil, fmt.Errorf(
						"Field missing associated relation. FieldName: %s, SchemaType: %s, ObjectType: %s",
						fname,
						field.Type.Name(),
						t.Name())
				}
				fd.RelationName = rel.name

				_, fieldRelationType, ok := rel.GetField(schemaName, fname)
				if !ok {
					return nil, errors.New("Relation is missing field")
				}

				fd.RelationType = rel.Kind() | fieldRelationType

				// handle object id field, defined as {{object_name}}_id
				// with type gql.ID
				// If it exists we need to delete and redefine
				// if it doesn't exist we simply define, and make sure we
				// skip later

				if !fd.IsObjectArray() {
					for i, sf := range desc.Schema.Fields {
						if sf.Name == fmt.Sprintf("%s_id", fname) {
							// delete element matching
							desc.Schema.Fields = append(desc.Schema.Fields[:i], desc.Schema.Fields[i+1:]...)
							break
						}
					}

					// create field
					fdRelated := client.FieldDescription{
						Name:         fmt.Sprintf("%s_id", fname),
						Kind:         gqlTypeToFieldKind(gql.ID),
						RelationType: client.Relation_Type_INTERNAL_ID,
					}
					fdRelated.Typ = defaultCRDTForFieldKind[fdRelated.Kind]
					desc.Schema.Fields = append(desc.Schema.Fields, fdRelated)
				}
			}

			desc.Schema.Fields = append(desc.Schema.Fields, fd)
		}

		// sort the fields lexicographically
		sort.Slice(desc.Schema.Fields, func(i, j int) bool {
			// make sure that the _key is always at the beginning
			if desc.Schema.Fields[i].Name == "_key" {
				return true
			} else if desc.Schema.Fields[j].Name == "_key" {
				return false
			}
			return desc.Schema.Fields[i].Name < desc.Schema.Fields[j].Name
		})

		// add default index
		desc.Indexes = []client.IndexDescription{
			{
				ID: uint32(0),
			},
		}

		// @todo: Add additional indexes based on defined
		// relations and directives

		descs[i] = desc
	}

	return descs, nil
}

/*

type book {
	name: String
	rating: Float
	author: author @primary
}

// don't need to worry about IDs and FieldIDs

return client.CollectionDescription{
		Name: "book",
		ID:   uint32(2),
		Schema: client.SchemaDescription{
			ID:       uint32(2),
			FieldIDs: []uint32{1, 2, 3, 4, 5},
			Fields: []client.FieldDescription{
				client.FieldDescription{
					Name: "_key",
					ID:   base.FieldID(1),
					Kind: base.FieldKind_DocKey,
				},
				client.FieldDescription{
					Name: "name",
					ID:   base.FieldID(2),
					Kind: base.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				client.FieldDescription{
					Name: "rating",
					ID:   base.FieldID(3),
					Kind: base.FieldKind_FLOAT,
					Typ:  client.LWW_REGISTER,
				},
				client.FieldDescription{
					Name:   "author",
					ID:     base.FieldID(5),
					Kind:   base.FieldKind_FOREIGN_OBJECT,
					Schema: "author",
					Typ:    client.NONE_CRDT,
					Meta:   base.Relation_Type_ONE | base.Relation_Type_Primary,
				},
				client.FieldDescription{
					Name: "author_id",
					ID:   base.FieldID(6),
					Kind: base.FieldKind_DocKey,
					Typ:  client.LWW_REGISTER,
				},
			},
		},
		Indexes: []client.IndexDescription{
			client.IndexDescription{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}

	return client.CollectionDescription{
		Name: "author",
		ID:   uint32(3),
		Schema: client.SchemaDescription{
			ID:       uint32(3),
			Name:     "author",
			FieldIDs: []uint32{1, 2, 3, 4, 5},
			Fields: []client.FieldDescription{
				client.FieldDescription{
					Name: "_key",
					ID:   base.FieldID(1),
					Kind: base.FieldKind_DocKey,
				},
				client.FieldDescription{
					Name: "name",
					ID:   base.FieldID(2),
					Kind: base.FieldKind_STRING,
					Typ:  client.LWW_REGISTER,
				},
				client.FieldDescription{
					Name: "age",
					ID:   base.FieldID(3),
					Kind: base.FieldKind_INT,
					Typ:  client.LWW_REGISTER,
				},
				client.FieldDescription{
					Name: "verified",
					ID:   base.FieldID(4),
					Kind: base.FieldKind_BOOL,
					Typ:  client.LWW_REGISTER,
				},
				client.FieldDescription{
					Name:   "published",
					ID:     base.FieldID(5),
					Kind:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
					Schema: "book",
					Typ:    client.NONE_CRDT,
					Meta:   base.Relation_Type_ONEMANY,
				},
			},
		},
		Indexes: []client.IndexDescription{
			client.IndexDescription{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}
*/
