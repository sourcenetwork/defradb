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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

func TestGenerateEncryptedQueryField(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionVersion{
		{
			Name: "User",
			EncryptedIndexes: []client.EncryptedIndexDescription{
				{FieldName: "email", Type: client.EncryptedIndexTypeEquality},
				{FieldName: "ssn", Type: client.EncryptedIndexTypeEquality},
			},
			Fields: []client.CollectionFieldDescription{
				{Name: "email", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "ssn", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
			},
		},
	}

	err = manager.Generate(ctx, collections)
	require.NoError(t, err)

	generated := introspectDocument(t, manager.Definition())
	querySurface := typeSurface(generated.TypeByName("Query"))
	assert.Contains(t, querySurface,
		"encrypted_User(filter:UserEncryptedFilterArg,limit:Int,offset:Int):"+request.EncryptedSearchResultName)
	assert.Equal(t, []string{"docIDs():[ID!]!"}, typeSurface(generated.TypeByName(request.EncryptedSearchResultName)))
}

func TestNoEncryptedQueryFieldWithoutIndexes(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionVersion{
		{
			Name: "Product",
			Fields: []client.CollectionFieldDescription{
				{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "price", Kind: client.FieldKind_NILLABLE_FLOAT64},
			},
		},
	}

	err = manager.Generate(ctx, collections)
	require.NoError(t, err)

	generated := introspectDocument(t, manager.Definition())
	assert.NotContains(t, typeMembers(generated.TypeByName("Query")), "encrypted_Product")
}
