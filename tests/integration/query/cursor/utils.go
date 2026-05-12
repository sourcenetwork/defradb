// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cursor

import (
	"strings"
	"testing"

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

func executeTestCase(t *testing.T, test testUtils.TestCase) {
	testUtils.ExecuteTestCase(
		t,
		testUtils.TestCase{
			SupportedMutationTypes: test.SupportedMutationTypes,
			SupportedClientTypes:   test.SupportedClientTypes,
			Actions: append(
				[]any{
					&action.AddCollection{
						SDL: userCollectionGQLSchema,
					},
				},
				test.Actions...,
			),
		},
	)
}

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
		cursor := result["_cursor"].(map[string]any)
		*docs = append(*docs, extractUsers(cursor["User"])...)
		return true, ""
	})
}
