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

package upsert

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationUpsert_WithDateTimeField_WithUTCNow_ShouldBeEqual(t *testing.T) {
	timestampMatcher := testUtils.NewSameValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						created_at: DateTime
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"created_at": "2011-07-23T01:11:11-05:00"
				}`,
			},
			// Perform mutations to update using UTC_NOW
			&action.Request{
				Request: `mutation {
					john: upsert_Users(
						filter: {name: {_eq: "John"}},
						add: {name: "John", created_at: UTC_NOW},
						update: {created_at: UTC_NOW}
					) {
						created_at
					}
					chris: upsert_Users(
						filter: {name: {_eq: "Chris"}},
						add: {name: "Chris", created_at: UTC_NOW},
						update: {created_at: UTC_NOW}
					) {
						created_at
					}
				}`,
				Results: map[string]any{
					"john":  []map[string]any{{"created_at": timestampMatcher}},
					"chris": []map[string]any{{"created_at": timestampMatcher}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
