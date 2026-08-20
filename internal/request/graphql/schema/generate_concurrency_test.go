// Copyright 2026 Democratized Data Foundation
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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	gql "github.com/sourcenetwork/graphql-go"
	gqlp "github.com/sourcenetwork/graphql-go/language/parser"
	"github.com/sourcenetwork/graphql-go/language/source"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

func TestGenerator_EmptyCollectionDoesNotError(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	_, err = manager.Generator.Generate(context.Background(), []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: request.DocIDFieldName, Kind: client.FieldKind_DocID},
		},
	}})
	require.NoError(t, err)
}

func TestGenerator_MutationInputIsSafeForConcurrentUse(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	_, err = manager.Generator.Generate(context.Background(), []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
		},
	}})
	require.NoError(t, err)

	const requestCount = 20
	start := make(chan struct{})
	errors := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for range requestCount {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			<-start

			document, err := gqlp.Parse(gqlp.ParseParams{Source: source.NewSource(&source.Source{
				Body: []byte(`mutation { update_User(input: {name: "Jane"}) { name } }`),
			})})
			if err != nil {
				errors <- err
				return
			}
			result := gql.ValidateDocument(manager.Schema(), document, nil)
			if !result.IsValid {
				errors <- result.Errors[0]
			}
		}()
	}

	close(start)
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}
