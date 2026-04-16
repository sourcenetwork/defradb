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

package replace

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestCollectionVersionUpdatesReplaceField(t *testing.T) {
	test := testUtils.TestCase{
		// The secondary-index multiplier adds @index on all fields. Replacing a field that
		// has a dependent index is not supported.
		// https://github.com/sourcenetwork/defradb/issues/4722
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						email: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/Users/Fields/1", "value": {"Name": "fax", "Kind": 11} }
					]
				`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						email
					}
				}`,
				ExpectedError: "Cannot query field \"email\" on type \"Users\".",
			},
			&action.Request{
				Request: `query {
					Users {
						name
						fax
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
