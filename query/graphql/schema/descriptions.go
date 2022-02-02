// Copyright 2020 Source Inc.
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

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/parser"

	gql "github.com/graphql-go/graphql"
)

var (
	// this is only here as a reference, and not to be used
	// directly. As it will yield incorrect and unexpected
	// results

	// nolint:deadcode,unused,varcheck
	gqlTypeToFieldKindReference = map[gql.Type]base.FieldKind{
		gql.ID:        base.FieldKind_DocKey,
		gql.Boolean:   base.FieldKind_BOOL,
		gql.Int:       base.FieldKind_INT,
		gql.Float:     base.FieldKind_FLOAT,
		gql.DateTime:  base.FieldKind_DATE,
		gql.String:    base.FieldKind_STRING,
		&gql.Object{}: base.FieldKind_FOREIGN_OBJECT,
		&gql.List{}:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
		// More custom ones to come
		// - JSON
		// - ByteArray
		// - Counters
	}

	// This map is fine to use
	defaultCRDTForFieldKind = map[base.FieldKind]core.CType{
		base.FieldKind_DocKey:               core.LWW_REGISTER,
		base.FieldKind_BOOL:                 core.LWW_REGISTER,
		base.FieldKind_BOOL_ARRAY:           core.LWW_REGISTER,
		base.FieldKind_INT:                  core.LWW_REGISTER,
		base.FieldKind_INT_ARRAY:            core.LWW_REGISTER,
		base.FieldKind_FLOAT:                core.LWW_REGISTER,
		base.FieldKind_FLOAT_ARRAY:          core.LWW_REGISTER,
		base.FieldKind_DATE:                 core.LWW_REGISTER,
		base.FieldKind_STRING:               core.LWW_REGISTER,
		base.FieldKind_STRING_ARRAY:         core.LWW_REGISTER,
		base.FieldKind_FOREIGN_OBJECT:       core.NONE_CRDT,
		base.FieldKind_FOREIGN_OBJECT_ARRAY: core.NONE_CRDT,
	}
)

func gqlTypeToFieldKind(t gql.Type) base.FieldKind {
	switch v := t.(type) {
	case *gql.Scalar:
		switch v.Name() {
		case "ID":
			return base.FieldKind_DocKey
		case "Boolean":
			return base.FieldKind_BOOL
		case "Int":
			return base.FieldKind_INT
		case "Float":
			return base.FieldKind_FLOAT
		case "Date":
			return base.FieldKind_DATE
		case "String":
			return base.FieldKind_STRING
		}
	case *gql.Object:
		return base.FieldKind_FOREIGN_OBJECT
	case *gql.List:
		if scalar, isScalar := v.OfType.(*gql.Scalar); isScalar {
			switch scalar.Name() {
			case "Boolean":
				return base.FieldKind_BOOL_ARRAY
			case "Int":
				return base.FieldKind_INT_ARRAY
			case "Float":
				return base.FieldKind_FLOAT_ARRAY
			case "String":
				return base.FieldKind_STRING_ARRAY
			}
		}
		return base.FieldKind_FOREIGN_OBJECT_ARRAY
	}

	return base.FieldKind_None
}

func (g *Generator) CreateDescriptions(types []*gql.Object) ([]base.CollectionDescription, error) {
	// create a indexable cached map
	typeMap := make(map[string]*gql.Object)
	for _, t := range types {
		typeMap[t.Name()] = t
	}

	descs := make([]base.CollectionDescription, len(types))
	// do the real generation
	for i, t := range types {
		desc := base.CollectionDescription{
			Name: t.Name(),
		}

		// add schema
		desc.Schema = base.SchemaDescription{
			Name: t.Name(),
			Fields: []base.FieldDescription{
				{
					Name: "_key",
					Kind: base.FieldKind_DocKey,
					Typ:  core.NONE_CRDT,
				},
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

			fd := base.FieldDescription{
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

				fd.Meta = rel.Kind() | fieldRelationType

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
					fdRelated := base.FieldDescription{
						Name: fmt.Sprintf("%s_id", fname),
						Kind: gqlTypeToFieldKind(gql.ID),
						Meta: base.Meta_Relation_INTERNAL_ID,
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
		desc.Indexes = []base.IndexDescription{
			{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
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

// dont need to worry about IDs and FieldIDs

return base.CollectionDescription{
		Name: "book",
		ID:   uint32(2),
		Schema: base.SchemaDescription{
			ID:       uint32(2),
			FieldIDs: []uint32{1, 2, 3, 4, 5},
			Fields: []base.FieldDescription{
				base.FieldDescription{
					Name: "_key",
					ID:   base.FieldID(1),
					Kind: base.FieldKind_DocKey,
				},
				base.FieldDescription{
					Name: "name",
					ID:   base.FieldID(2),
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "rating",
					ID:   base.FieldID(3),
					Kind: base.FieldKind_FLOAT,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name:   "author",
					ID:     base.FieldID(5),
					Kind:   base.FieldKind_FOREIGN_OBJECT,
					Schema: "author",
					Typ:    core.NONE_CRDT,
					Meta:   base.Meta_Relation_ONE | base.Meta_Relation_Primary,
				},
				base.FieldDescription{
					Name: "author_id",
					ID:   base.FieldID(6),
					Kind: base.FieldKind_DocKey,
					Typ:  core.LWW_REGISTER,
				},
			},
		},
		Indexes: []base.IndexDescription{
			base.IndexDescription{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}

	return base.CollectionDescription{
		Name: "author",
		ID:   uint32(3),
		Schema: base.SchemaDescription{
			ID:       uint32(3),
			Name:     "author",
			FieldIDs: []uint32{1, 2, 3, 4, 5},
			Fields: []base.FieldDescription{
				base.FieldDescription{
					Name: "_key",
					ID:   base.FieldID(1),
					Kind: base.FieldKind_DocKey,
				},
				base.FieldDescription{
					Name: "name",
					ID:   base.FieldID(2),
					Kind: base.FieldKind_STRING,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "age",
					ID:   base.FieldID(3),
					Kind: base.FieldKind_INT,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name: "verified",
					ID:   base.FieldID(4),
					Kind: base.FieldKind_BOOL,
					Typ:  core.LWW_REGISTER,
				},
				base.FieldDescription{
					Name:   "published",
					ID:     base.FieldID(5),
					Kind:   base.FieldKind_FOREIGN_OBJECT_ARRAY,
					Schema: "book",
					Typ:    core.NONE_CRDT,
					Meta:   base.Meta_Relation_ONEMANY,
				},
			},
		},
		Indexes: []base.IndexDescription{
			base.IndexDescription{
				Name:    "primary",
				ID:      uint32(0),
				Primary: true,
				Unique:  true,
			},
		},
	}
*/
