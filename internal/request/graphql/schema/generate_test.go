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
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	gql "github.com/sourcenetwork/graphql-go"
)

func TestIntrospectionResult(t *testing.T) {
	ctx := context.Background()
	manager, err := NewSchemaManager(true)
	require.NoError(t, err)

	collections := []client.CollectionVersion{
		{
			Name: "User",
			Fields: []client.CollectionFieldDescription{
				{Name: "email", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "ssn", Kind: client.FieldKind_NILLABLE_STRING},
				{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
			},
		},
	}

	_, err = manager.Generator.Generate(ctx, collections)
	require.NoError(t, err)

	request, err := os.ReadFile("introspection.gql")
	require.NoError(t, err)

	schema := manager.Schema()
	params := gql.Params{Schema: *schema, RequestString: string(request)}
	r := gql.Do(params)

	require.Empty(t, r.Errors)
}
