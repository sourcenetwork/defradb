// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package replace

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// todo: The inverse of this test is not currently possible, make sure it also is tested when
// resolving https://github.com/sourcenetwork/defradb/issues/2983
func TestColVersionUpdateReplaceIsMaterialized_GivenPolicyOnNonMAterializedView_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.CachelessViewType,
		}),
		Actions: []any{
			testUtils.AddDACPolicy{
				Identity: testUtils.ClientIdentity(1),
				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource
                    resources:
                    - name: userView
                      permissions:
                      - name: read
                        expr: reader
                      - name: update
                      - name: delete
                      relations:
                      - name: reader
                        types:
                        - actor
                      - name: admin
                        manages:
                        - reader
                        types:
                        - actor
                `,
			},
			&action.AddCollection{
				SDL: `
					type User {
						name: String
					}
				`,
			},
			&action.AddView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @policy(
						id: "62cff38630eb2732c5f5e763ab31478a4bac7077ed66c9ad0c061c86a5b498c9",
						resource: "userView"
					) @materialized(if: false) {
						name: String
					}
				`,
			},
			&action.PatchCollection{
				Patch: `
					[
						{
							"op": "replace",
							"path": "/UserView/IsMaterialized",
							"value": true
						}
					]
				`,
				ExpectedError: "materialized views do not support ACP. Collection: UserView",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
