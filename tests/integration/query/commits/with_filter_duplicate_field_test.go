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

package commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// A filter that repeats the same input field inside a nested input object used to be
// accepted silently, with the later value overwriting the earlier one. It must now be
// rejected, as the GraphQL specification requires input object fields to be unique.
func TestQueryCommits_WithDuplicateFieldNameInNestedFilter_Errors(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(filter: {_or: {fieldName: {_eq: "name"}, fieldName: {_eq: "age"}}}) {
							fieldName
						}
					}`,
				ExpectedError: "there can be only one input field with a given name",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
