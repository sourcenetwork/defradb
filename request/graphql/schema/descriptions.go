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
	gql "github.com/graphql-go/graphql"

	"github.com/sourcenetwork/defradb/client"
)

var (
	// this is only here as a reference, and not to be used
	// directly. As it will yield incorrect and unexpected
	// results

	//nolint:unused
	gqlTypeToFieldKindReference = map[gql.Type]client.FieldKind{
		gql.ID:        client.FieldKind_DocKey,
		gql.Boolean:   client.FieldKind_BOOL,
		gql.Int:       client.FieldKind_INT,
		gql.Float:     client.FieldKind_FLOAT,
		gql.DateTime:  client.FieldKind_DATETIME,
		gql.String:    client.FieldKind_STRING,
		&gql.Object{}: client.FieldKind_FOREIGN_OBJECT,
		&gql.List{}:   client.FieldKind_FOREIGN_OBJECT_ARRAY,
		// More custom ones to come
		// - JSON
		// - ByteArray
		// - Counters
	}

	fieldKindToGQLType = map[client.FieldKind]gql.Type{
		client.FieldKind_DocKey:                gql.ID,
		client.FieldKind_BOOL:                  gql.Boolean,
		client.FieldKind_BOOL_ARRAY:            gql.NewList(gql.NewNonNull(gql.Boolean)),
		client.FieldKind_NILLABLE_BOOL_ARRAY:   gql.NewList(gql.Boolean),
		client.FieldKind_INT:                   gql.Int,
		client.FieldKind_INT_ARRAY:             gql.NewList(gql.NewNonNull(gql.Int)),
		client.FieldKind_NILLABLE_INT_ARRAY:    gql.NewList(gql.Int),
		client.FieldKind_FLOAT:                 gql.Float,
		client.FieldKind_FLOAT_ARRAY:           gql.NewList(gql.NewNonNull(gql.Float)),
		client.FieldKind_NILLABLE_FLOAT_ARRAY:  gql.NewList(gql.Float),
		client.FieldKind_DATETIME:              gql.DateTime,
		client.FieldKind_STRING:                gql.String,
		client.FieldKind_STRING_ARRAY:          gql.NewList(gql.NewNonNull(gql.String)),
		client.FieldKind_NILLABLE_STRING_ARRAY: gql.NewList(gql.String),
	}

	// This map is fine to use
	defaultCRDTForFieldKind = map[client.FieldKind]client.CType{
		client.FieldKind_DocKey:                client.LWW_REGISTER,
		client.FieldKind_BOOL:                  client.LWW_REGISTER,
		client.FieldKind_BOOL_ARRAY:            client.LWW_REGISTER,
		client.FieldKind_NILLABLE_BOOL_ARRAY:   client.LWW_REGISTER,
		client.FieldKind_INT:                   client.LWW_REGISTER,
		client.FieldKind_INT_ARRAY:             client.LWW_REGISTER,
		client.FieldKind_NILLABLE_INT_ARRAY:    client.LWW_REGISTER,
		client.FieldKind_FLOAT:                 client.LWW_REGISTER,
		client.FieldKind_FLOAT_ARRAY:           client.LWW_REGISTER,
		client.FieldKind_NILLABLE_FLOAT_ARRAY:  client.LWW_REGISTER,
		client.FieldKind_DATETIME:              client.LWW_REGISTER,
		client.FieldKind_STRING:                client.LWW_REGISTER,
		client.FieldKind_STRING_ARRAY:          client.LWW_REGISTER,
		client.FieldKind_NILLABLE_STRING_ARRAY: client.LWW_REGISTER,
		client.FieldKind_FOREIGN_OBJECT:        client.NONE_CRDT,
		client.FieldKind_FOREIGN_OBJECT_ARRAY:  client.NONE_CRDT,
	}
)
