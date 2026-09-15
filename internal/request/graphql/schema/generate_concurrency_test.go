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

	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astvalidation"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

func TestGenerator_EmptyCollectionDoesNotError(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	err = manager.Generate(context.Background(), []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: request.DocIDFieldName, Kind: client.FieldKind_DocID},
		},
	}})
	require.NoError(t, err)

	mutation, ok := manager.Definition().NodeByNameStr("Mutation")
	require.True(t, ok)
	require.False(t, manager.Definition().ObjectTypeDefinitionHasField(mutation.Ref, []byte("add_User")))
	require.False(t, manager.Definition().ObjectTypeDefinitionHasField(mutation.Ref, []byte("update_User")))
	require.False(t, manager.Definition().ObjectTypeDefinitionHasField(mutation.Ref, []byte("upsert_User")))
	require.True(t, manager.Definition().ObjectTypeDefinitionHasField(mutation.Ref, []byte("delete_User")))
	require.True(t, manager.Definition().ObjectTypeDefinitionHasField(mutation.Ref, []byte("truncate_User")))
}

func TestGenerator_InvalidSchemaDoesNotReplaceDefinition(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)
	original := manager.Definition()

	err = manager.Generate(context.Background(), []client.CollectionVersion{{Name: "Invalid Name"}})
	require.Error(t, err)
	require.Same(t, original, manager.Definition())
}

func TestGenerator_SchemaIsSafeForConcurrentUse(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	err = manager.Generate(context.Background(), []client.CollectionVersion{{
		Name: "User",
		Fields: []client.CollectionFieldDescription{
			{Name: "name", Kind: client.FieldKind_NILLABLE_STRING},
		},
	}})
	require.NoError(t, err)

	requests := []string{
		`query { User(filter: {name: {_eq: "Jane"}}) { name } }`,
		`mutation { update_User(input: {name: "Jane"}) { name } }`,
	}
	const requestCount = 20
	start := make(chan struct{})
	errors := make(chan error, requestCount)
	var waitGroup sync.WaitGroup
	for i := range requestCount {
		waitGroup.Add(1)
		go func(request string) {
			defer waitGroup.Done()
			<-start

			toolsDocument, report := astparser.ParseGraphqlDocumentString(request)
			if report.HasErrors() {
				errors <- report
				return
			}
			astnormalization.NormalizeOperation(&toolsDocument, manager.Definition(), &report)
			if report.HasErrors() {
				errors <- report
				return
			}
			if astvalidation.DefaultOperationValidator().Validate(
				&toolsDocument,
				manager.Definition(),
				&report,
			) == astvalidation.Invalid {
				errors <- report
			}
		}(requests[i%len(requests)])
	}

	close(start)
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}
