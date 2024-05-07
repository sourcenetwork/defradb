// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package issues

import (
	"fmt"
	"math"
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

// These tests document https://github.com/sourcenetwork/defradb/issues/2569

func TestP2PUpdate_WithPNCounterFloatOverflowIncrement_PreventsQuerying(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// This issue only affects the http and the cli clients
				testUtils.HTTPClientType,
				testUtils.CLIClientType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, math.MaxFloat64/10),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				ExpectedError: "unexpected end of JSON input",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PUpdate_WithPNCounterFloatOverflowDecrement_PreventsQuerying(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// This issue only affects the http and the cli clients
				testUtils.HTTPClientType,
				testUtils.CLIClientType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, -math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(-1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, -math.MaxFloat64/10),
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						points
					}
				}`,
				ExpectedError: "unexpected end of JSON input",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestP2PUpdate_WithPNCounterFloatOverflow_PreventsCollectionGet(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// This issue only affects the http and the cli clients
				testUtils.HTTPClientType,
				testUtils.CLIClientType,
			},
		),
		SupportedMutationTypes: immutable.Some(
			[]testUtils.MutationType{
				// We limit the test to Collection mutation calls, as the test framework
				// will make a `Get` call before submitting the document, which is where the error
				// will surface (not the update itelf)
				testUtils.CollectionSaveMutationType,
				testUtils.CollectionNamedMutationType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: "pncounter")
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64),
			},
			testUtils.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, math.MaxFloat64/10),
			},
			testUtils.UpdateDoc{
				// Try and update the document again, the value used does not matter.
				Doc: `{
					"points": 1
				}`,
				// WARNING: This error is just an artifact of our test harness, what actually happens
				// is the test harness calls `collection.Get`, which returns an empty string and no error.
				ExpectedError: "cannot parse JSON: cannot parse empty string",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
