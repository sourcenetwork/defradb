// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package test_acp_dac_p2p

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestACP_P2PUpdatePrivateDocumentsOnDifferentNodes_SourceHubACP(t *testing.T) {
	test := testUtils.TestCase{

		SupportedDocumentACPTypes: immutable.Some(
			[]testUtils.DocumentACPType{
				testUtils.SourceHubDocumentACPType,
			},
		),

		Actions: []any{
			testUtils.RandomNetworkingConfig(),

			testUtils.RandomNetworkingConfig(),

			testUtils.AddDACPolicy{

				Identity: testUtils.ClientIdentity(1),

				Policy: `
                    name: Test Policy

                    description: A Policy

                    actor:
                      name: actor

                    resources:
                      users:
                        permissions:
                          read:
                            expr: owner + reader + updater + deleter

                          update:
                            expr: owner + updater

                          delete:
                            expr: owner + deleter

                          nothing:
                            expr: dummy

                        relations:
                          owner:
                            types:
                              - actor

                          reader:
                            types:
                              - actor

                          updater:
                            types:
                              - actor

                          deleter:
                            types:
                              - actor

                          admin:
                            manages:
                              - reader
                            types:
                              - actor

                          dummy:
                            types:
                              - actor
                `,
			},

			&action.AddSchema{
				Schema: `
						type Users @policy(
							id: "{{.Policy0}}",
							resource: "users"
						) {
							name: String
							age: Int
						}
					`,
			},

			testUtils.ConfigureReplicator{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad",
				},
			},

			testUtils.CreateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocMap: map[string]any{
					"name": "Shahzad Lone",
				},
			},

			testUtils.WaitForSync{},

			testUtils.UpdateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(0),

				CollectionID: 0,

				DocID: 0,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},

			testUtils.UpdateDoc{
				Identity: testUtils.ClientIdentity(1),

				NodeID: immutable.Some(1),

				CollectionID: 0,

				DocID: 1,

				Doc: `
					{
						"name": "ShahzadLone"
					}
				`,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
