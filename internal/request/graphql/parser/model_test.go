// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astnormalization"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astparser"

	"github.com/sourcenetwork/defradb/client/request"
)

func TestParseRequestFromDocument(t *testing.T) {
	definition, definitionReport := astparser.ParseGraphqlDocumentString(`
		type Query { User(limit: Int, showDeleted: Boolean = true): [User] }
		type User { name: String }
	`)
	require.False(t, definitionReport.HasErrors(), definitionReport.Error())
	astnormalization.NormalizeDefinition(&definition, &definitionReport)
	require.False(t, definitionReport.HasErrors(), definitionReport.Error())
	document, report := astparser.ParseGraphqlDocumentString(`
		query Selected($limit: Int!, $show: Boolean) @explain(type: execute) {
			users: User(limit: $limit, showDeleted: $show) { name }
		}
	`)
	require.False(t, report.HasErrors(), report.Error())
	document.Input.Variables = []byte(`{"limit":3}`)

	parsed, errs := ParseRequest(&definition, &document)
	require.Empty(t, errs)
	require.Len(t, parsed.Queries, 1)
	require.Len(t, parsed.Queries[0].Selections, 1)
	selection, ok := parsed.Queries[0].Selections[0].(*request.Select)
	require.True(t, ok)
	require.Equal(t, "User", selection.Name)
	require.Equal(t, "users", selection.Alias.Value())
	require.Equal(t, uint64(3), selection.Limit.Value())
	require.True(t, selection.ShowDeleted)
	require.Equal(t, request.ExecuteExplain, parsed.Queries[0].Directives.ExplainType.Value())
}

func TestRepairExponentSeparators(t *testing.T) {
	require.JSONEq(
		t,
		`{"values":[1e-11,2E+12],"label":"e,-11"}`,
		string(repairExponentSeparators([]byte(`{"values":[1e,-11,2E,+12],"label":"e,-11"}`))),
	)
}
