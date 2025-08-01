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
	"github.com/sourcenetwork/defradb/client/request"
)

func TestGenerateEncryptedQueryField(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionDefinition{
		{
			Version: client.CollectionVersion{
				Name: "User",
				EncryptedIndexes: []client.EncryptedIndexDescription{
					{FieldName: "email", Type: client.EncryptedIndexTypeEquality},
					{FieldName: "ssn", Type: client.EncryptedIndexTypeEquality},
				},
			},
			Schema: client.SchemaDescription{
				Name: "User",
				Fields: []client.SchemaFieldDescription{
					{Name: "email", Kind: client.FieldKind_NILLABLE_STRING},
					{Name: "ssn", Kind: client.FieldKind_NILLABLE_STRING},
					{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
				},
			},
		},
	}

	_, err = manager.Generator.Generate(ctx, collections)
	require.NoError(t, err)
	require.NoError(t, manager.ResolveTypes())

	queryType := manager.schema.QueryType()
	require.NotNil(t, queryType)

	encryptedField, ok := queryType.Fields()["User_encrypted"]
	require.True(t, ok, "User_encrypted field should exist")

	hasFilter, hasLimit, hasOffset := false, false, false
	for _, arg := range encryptedField.Args {
		switch arg.Name() {
		case "filter":
			hasFilter = true
		case "limit":
			hasLimit = true
		case "offset":
			hasOffset = true
		}
	}
	assert.True(t, hasFilter, "should have filter argument")
	assert.True(t, hasLimit, "should have limit argument")
	assert.True(t, hasOffset, "should have offset argument")

	returnType := encryptedField.Type
	if nonNull, ok := returnType.(*gql.NonNull); ok {
		returnType = nonNull.OfType
	}
	assert.Equal(t, request.EncryptedSearchResultName, returnType.Name())

	resultObj := returnType.(*gql.Object)
	docIDsField, ok := resultObj.Fields()["docIDs"]
	assert.True(t, ok, "EncryptedSearchResult should have docIDs field")

	docIDsType := docIDsField.Type
	if nonNull, ok := docIDsType.(*gql.NonNull); ok {
		docIDsType = nonNull.OfType
	}
	list, ok := docIDsType.(*gql.List)
	assert.True(t, ok, "docIDs should be a list")
	if ok {
		elementType := list.OfType
		if nonNull, ok := elementType.(*gql.NonNull); ok {
			elementType = nonNull.OfType
		}
		assert.Equal(t, "String", elementType.Name())
	}
}

func TestNoEncryptedQueryFieldWithoutIndexes(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionDefinition{
		{
			Version: client.CollectionVersion{
				Name: "Product",
			},
			Schema: client.SchemaDescription{
				Name: "Product",
				Fields: []client.SchemaFieldDescription{
					{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
					{Name: "price", Kind: client.FieldKind_NILLABLE_FLOAT64},
				},
			},
		},
	}

	_, err = manager.Generator.Generate(ctx, collections)
	require.NoError(t, err)

	queryType := manager.schema.QueryType()
	require.NotNil(t, queryType)

	_, ok := queryType.Fields()["Product_encrypted"]
	assert.False(t, ok, "Product_encrypted field should not exist without encrypted indexes")
}
