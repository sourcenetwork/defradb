// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package cursor

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// userCollectionGQLSchema defines a User type with indexed age field for cursor pagination tests.
var userCollectionGQLSchema = `
	type User {
		name: String
		age: Int @index
	}
`

// makeExplainQuery wraps a query with @explain(type: execute) for index verification.
func makeExplainQuery(req string) string {
	ind := strings.Index(req, "query")
	if ind < 0 {
		panic("Invalid query: " + req)
	}
	return "query @explain(type: execute) " + req[ind+5:]
}

// extractUsers extracts user documents from query results, handling both []any and []map[string]any types.
func extractUsers(usersRaw any) []map[string]any {
	switch users := usersRaw.(type) {
	case []any:
		result := make([]map[string]any, 0, len(users))
		for _, item := range users {
			if m, ok := item.(map[string]any); ok {
				result = append(result, m)
			}
		}
		return result
	case []map[string]any:
		return users
	default:
		return nil
	}
}

func appendCursorUsers(docs *[]map[string]any) action.ResultAsserter {
	return testUtils.ResultAsserterFunc(func(t testing.TB, result map[string]any) (bool, string) {
		cursor := requireType[map[string]any](t, result["_cursor"])
		*docs = append(*docs, extractUsers(cursor["User"])...)
		return true, ""
	})
}

func requireType[T any](t testing.TB, value any) T {
	t.Helper()
	result, ok := value.(T)
	require.True(t, ok, "unexpected value type: %T", value)
	return result
}
