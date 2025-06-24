// Copyright 2025 Democratized Data Foundation
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
	"context"
	"testing"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
)

func TestGenerateEncryptedQueryField(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager()
	require.NoError(t, err)
	gen := manager.Generator

	collections := []client.CollectionDefinition{
		{
			Version: client.CollectionVersion{
				Name: "User",
				EncryptedIndexes: []client.EncryptedIndexDescription{
					{
						FieldName: "email",
						Type:      client.EncryptedIndexTypeEquality,
					},
					{
						FieldName: "ssn",
						Type:      client.EncryptedIndexTypeEquality,
					},
				},
			},
			Schema: client.SchemaDescription{
				Name: "User",
				Fields: []client.SchemaFieldDescription{
					{
						Name: "email",
						Kind: client.FieldKind_NILLABLE_STRING,
					},
					{
						Name: "ssn",
						Kind: client.FieldKind_NILLABLE_STRING,
					},
					{
						Name: "name",
						Kind: client.FieldKind_NILLABLE_STRING,
					},
				},
			},
		},
	}

	_, err = gen.Generate(ctx, collections)
	require.NoError(t, err)

	err = manager.ResolveTypes()
	require.NoError(t, err)

	queryType := manager.schema.QueryType()
	require.NotNil(t, queryType)

	fields := queryType.Fields()
	encryptedField, ok := fields["User_encrypted"]
	require.True(t, ok, "User_encrypted field should exist")
	assert.Equal(t, "Query encrypted fields for User", encryptedField.Description)

	var filterArg *gql.Argument
	var hasLimit, hasOffset bool

	for _, arg := range encryptedField.Args {
		switch arg.Name() {
		case "filter":
			filterArg = arg
		case "limit":
			hasLimit = true
		case "offset":
			hasOffset = true
		}
	}

	require.NotNil(t, filterArg, "filter arg should exist")
	assert.True(t, hasLimit, "limit arg should exist")
	assert.True(t, hasOffset, "offset arg should exist")

	filterType := filterArg.Type.(*gql.InputObject)
	assert.Equal(t, "UserEncryptedFilterArg", filterType.Name())

	filterFields := filterType.Fields()
	assert.Len(t, filterFields, 2)
	_, hasEmail := filterFields["email"]
	assert.True(t, hasEmail, "email field should exist in filter")
	_, hasSSN := filterFields["ssn"]
	assert.True(t, hasSSN, "ssn field should exist in filter")
	_, hasName := filterFields["name"]
	assert.False(t, hasName, "name field should not exist in filter")
}

func TestNoEncryptedQueryFieldWithoutIndexes(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager()
	require.NoError(t, err)
	gen := manager.Generator

	collections := []client.CollectionDefinition{
		{
			Version: client.CollectionVersion{
				Name: "Product",
			},
			Schema: client.SchemaDescription{
				Name: "Product",
				Fields: []client.SchemaFieldDescription{
					{
						Name: "name",
						Kind: client.FieldKind_NILLABLE_STRING,
					},
					{
						Name: "price",
						Kind: client.FieldKind_NILLABLE_FLOAT64,
					},
				},
			},
		},
	}

	_, err = gen.Generate(ctx, collections)
	require.NoError(t, err)

	queryType := manager.schema.QueryType()
	require.NotNil(t, queryType)

	_, ok := queryType.Fields()["Product_encrypted"]
	assert.False(t, ok, "Product_encrypted field should not exist")
}
