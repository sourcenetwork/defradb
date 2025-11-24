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
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"

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

	request, err := os.ReadFile("introspection_query.gql")
	require.NoError(t, err)

	schema := manager.Schema()
	params := gql.Params{Schema: *schema, RequestString: string(request)}
	r := gql.Do(params)

	require.Empty(t, r.Errors)

	buf, err := json.Marshal(r.Data)
	require.NoError(t, err)

	var converter introspection.JsonConverter
	_, err = converter.GraphQLDocument(bytes.NewReader(buf))
	require.NoError(t, err)
}
