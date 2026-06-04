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
	"math"
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"

	"github.com/sourcenetwork/immutable"
)

func TestQuerySimple_WithMinOnUndefinedObject_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					MIN
				}`,
				ExpectedError: "aggregate must be provided with a property to aggregate",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinOnUndefinedField_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					MIN(Users: {})
				}`,
				ExpectedError: "Argument \"Users\" has invalid value {}.\nIn field \"field\": Expected \"UsersNumericFieldsArg!\", found null.",
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					MIN(Users: {field: Age})
				}`,
				Results: map[string]any{
					"MIN": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMin_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 30
				}`,
			},
			&action.Request{
				Request: `query {
					MIN(Users: {field: Age})
				}`,
				Results: map[string]any{
					"MIN": int64(21),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithMinAndMaxValueInt_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{
			// JavaScript does not support 64 bit int
			state.GoClientType,
			state.CLIClientType,
			state.HTTPClientType,
		}),
		SupportedMutationTypes: immutable.Some([]state.MutationType{
			// GraphQL does not support 64 bit int
			state.CollectionSaveMutationType,
			state.CollectionNamedMutationType,
		}),
		Actions: []any{
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
					"Age":  int64(math.MaxInt64),
				},
			},
			&action.Request{
				Request: `query {
					MAX(Users: {field: Age})
				}`,
				Results: map[string]any{
					"MAX": int64(math.MaxInt64),
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimple_WithAliasedMinOnEmptyCollection_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.Request{
				Request: `query {
					minimum: MIN(Users: {field: Age})
				}`,
				Results: map[string]any{
					"minimum": nil,
				},
			},
		},
	}

	executeTestCase(t, test)
}
