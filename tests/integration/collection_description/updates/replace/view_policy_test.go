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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// todo: The inverse of this test is not currently possible, make sure it also is tested when
// resolving https://github.com/sourcenetwork/defradb/issues/2983
func TestColDescrUpdateReplaceIsMaterialized_GivenPolicyOnNonMAterializedView_Errors(t *testing.T) {
	test := testUtils.TestCase{
		SupportedViewTypes: immutable.Some([]testUtils.ViewType{
			testUtils.CachelessViewType,
		}),
		Actions: []any{
			testUtils.AddPolicy{
				Identity: immutable.Some(1),
				Policy: `
                    name: test
                    description: a test policy which marks a collection in a database as a resource

                    actor:
                      name: actor

                    resources:
                      userView:
                        permissions:
                          read:
                            expr: owner + reader
                          write:
                            expr: owner

                        relations:
                          owner:
                            types:
                              - actor
                          reader:
                            types:
                              - actor
                          admin:
                            manages:
                              - reader
                            types:
                              - actor
                `,
				ExpectedPolicyID: "7a698a9c5fe74a5854c2e1e8d00c606926c64ad883a157db2f345749e8609fcb",
			},
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						name: String
					}
				`,
			},
			testUtils.CreateView{
				Query: `
					User {
						name
					}
				`,
				SDL: `
					type UserView @policy(
						id: "7a698a9c5fe74a5854c2e1e8d00c606926c64ad883a157db2f345749e8609fcb",
						resource: "userView"
					) @materialized(if: false) {
						name: String
					}
				`,
			},
			testUtils.PatchCollection{
				Patch: `
					[
						{ "op": "replace", "path": "/2/IsMaterialized", "value": true }
					]
				`,
				ExpectedError: "materialized views do not support ACP. Collection: UserView",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
