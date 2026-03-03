// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

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
