// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package peer_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// This test documents https://github.com/sourcenetwork/defradb/issues/2566
func TestP2PUpdate_WithPNCounterSimultaneousOverflowIncrement_DoesNotReachConsitency(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				// Create John on all nodes
				Doc: fmt.Sprintf(`{
					"Name": "John",
					"Age": %g
				}`, math.MaxFloat64/10),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: fmt.Sprintf(`{
					"Age": %g
				}`, math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: fmt.Sprintf(`{
					"Age": %g
				}`, -math.MaxFloat64),
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						// Node 0 overflows before subtraction, and because subtracting from infinity
						// results in infinity the value remains infinate
						"Age": math.Inf(1),
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						// Node 1 subtracts before adding, meaning no overflow is achieved and the value
						// remains finate
						"Age": float64(1.7976931348623155e+307),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test documents https://github.com/sourcenetwork/defradb/issues/2566
func TestP2PUpdate_WithPNCounterSimultaneousOverflowDecrement_DoesNotReachConsitency(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.RandomNetworkingConfig(),
			testUtils.RandomNetworkingConfig(),
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						Name: String
						Age: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				// Create John on all nodes
				Doc: fmt.Sprintf(`{
					"Name": "John",
					"Age": %g
				}`, -math.MaxFloat64/10),
			},
			testUtils.ConnectPeers{
				SourceNodeID: 0,
				TargetNodeID: 1,
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(1),
				Doc: fmt.Sprintf(`{
					"Age": %g
				}`, math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				NodeID: immutable.Some(0),
				Doc: fmt.Sprintf(`{
					"Age": %g
				}`, -math.MaxFloat64),
			},
			testUtils.WaitForSync{},
			testUtils.Request{
				NodeID: immutable.Some(0),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						// Node 0 overflows before addition, and because adding to infinity
						// results in infinity the value remains infinate
						"Age": math.Inf(-1),
					},
				},
			},
			testUtils.Request{
				NodeID: immutable.Some(1),
				Request: `query {
					Users {
						Age
					}
				}`,
				Results: []map[string]any{
					{
						// Node 1 adds before subtracting, meaning no overflow is achieved and the value
						// remains finate
						"Age": float64(-1.7976931348623155e+307),
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
