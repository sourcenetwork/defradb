// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.

package schema

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
)

func TestDefaultSchemaSDLWritesEquivalentDefinition(t *testing.T) {
	manager, err := NewSchemaManager(false)
	require.NoError(t, err)

	var writtenSDL bytes.Buffer
	require.NoError(t, manager.WriteSDL(&writtenSDL))
	expected := introspectSDL(t, defaultSchemaSDL)
	actual := introspectSDL(t, writtenSDL.String())
	require.Equal(t, len(expected.Types), len(actual.Types))
	for _, gqlType := range expected.Types {
		require.Equal(t, typeSurface(gqlType), typeSurface(actual.TypeByName(gqlType.Name)), gqlType.Name)
	}
}

func TestDefaultSchemaSDLGeneratesIntrospection(t *testing.T) {
	document, report := astparser.ParseGraphqlDocumentString(defaultSchemaSDL)
	require.False(t, report.HasErrors(), report.Error())

	var data introspection.Data
	introspection.NewGenerator().Generate(&document, &report, &data)
	require.False(t, report.HasErrors(), report.Error())
	require.Equal(t, "Query", data.Schema.QueryType.Name)
	require.Equal(t, "Mutation", data.Schema.MutationType.Name)
	require.Equal(t, "Subscription", data.Schema.SubscriptionType.Name)
	require.NotNil(t, data.Schema.TypeByName("Blob"))
	require.NotNil(t, data.Schema.TypeByName("Commit"))
}
