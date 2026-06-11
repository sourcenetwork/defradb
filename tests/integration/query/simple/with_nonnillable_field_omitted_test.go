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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test documents a known gap: a non-nillable field that is omitted at
// document creation time returns nil at query time, violating the GQL non-null
// contract for that field.
func TestQuerySimple_WithNonNillableStringFieldOmitted_ReturnsNil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `type Users { name: String! }`,
			},
			&action.AddDoc{
				Doc: `{}`,
			},
			&action.Request{
				Request: `query { Users { name } }`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": nil},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}
