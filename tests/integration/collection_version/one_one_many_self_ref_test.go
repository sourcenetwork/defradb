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

package collection_version

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test covers a bug that was found as part of https://github.com/sourcenetwork/defradb/issues/4710
// The bug has been fixed, but the test remains as coverage of this case is important.
func TestCollectionVersionWith_OneOne_OneMany_SelfRef(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Dev_RC_Domain2 {
						routes: [Dev_RC_RedirectRoute2]
						firstRoute: Dev_RC_RedirectRoute2 @primary @relation(name: "domain_first_route")
					}

					type Dev_RC_RedirectRoute2 {
						firstForDomain: Dev_RC_Domain2 @relation(name: "domain_first_route")

						domain: Dev_RC_Domain2
						after: Dev_RC_RedirectRoute2
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
