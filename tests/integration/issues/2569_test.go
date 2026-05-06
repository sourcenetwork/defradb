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

package issues

import (
	"fmt"
	"math"
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/state"
)

// These tests document https://github.com/sourcenetwork/defradb/issues/2569

func TestP2PUpdate_WithPNCounterFloatOverflowIncrement_PreventsQuerying(t *testing.T) {
	test := testUtils.TestCase{
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// This issue only affects the http and the cli clients
				state.HTTPClientType,
				state.CLIClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64),
			},
			&action.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, math.MaxFloat64/10),
			},
			&action.Request{
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// This issue only affects the http and the cli clients
				state.HTTPClientType,
				state.CLIClientType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, -math.MaxFloat64),
			},
			&action.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(-1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, -math.MaxFloat64/10),
			},
			&action.Request{
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex},
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				// This issue only affects the http and the cli clients
				state.HTTPClientType,
				state.CLIClientType,
			},
		),
		SupportedMutationTypes: immutable.Some(
			[]state.MutationType{
				// We limit the test to Collection mutation calls, as the test framework
				// will make a `Get` call before submitting the document, which is where the error
				// will surface (not the update itelf)
				state.CollectionSaveMutationType,
				state.CollectionNamedMutationType,
			},
		),
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: fmt.Sprintf(`{
					"name": "John",
					"points": %g
				}`, math.MaxFloat64),
			},
			&action.UpdateDoc{
				// Overflow the points field, this results in a value of `math.Inf(1)`
				Doc: fmt.Sprintf(`{
					"points": %g
				}`, math.MaxFloat64/10),
			},
			&action.UpdateDoc{
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
