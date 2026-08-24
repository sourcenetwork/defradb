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

package update

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestUpdateWithInvalidFilterType_ReturnsError(t *testing.T) {
	type invalidFilterType struct{ Number int }
	test := testUtils.TestCase{
		// http and cli clients will pass the serialize filter into json which will result in
		// the payload deserialized into map[string]any. With Go client the filter is passed as is.
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{state.HTTPClientType, state.CLIClientType}),
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        invalidFilterType{Number: 1},
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "collection not found",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithInvalidFilterType_WithGoClient_ReturnsError(t *testing.T) {
	type invalidFilterType struct{ Number int }
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some([]state.ClientType{state.GoClientType}),
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        invalidFilterType{Number: 1},
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "unsupported filter type",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithEmptyFilter_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        "",
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "filter cannot be empty",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithInvalidJSON_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        `{name: {_eq: "John"}}`,
				Updater:       `{name: "Eric"}`,
				ExpectedError: "cannot parse JSON: cannot parse object",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithInvalidUpdater_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        `{name: {_eq: "John"}}`,
				Updater:       `"name: Eric"`,
				ExpectedError: "the updater of a document is of invalid type",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithPatch_DoesNothing(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID:         0,
				Filter:               `{name: {_eq: "John"}}`,
				Updater:              `[{"name": "Eric"}, {"name": "Sam"}]`,
				SkipLocalUpdateEvent: true,
			},
			&action.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithMapFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID: 0,
				Filter: map[string]any{
					"name": map[string]any{"_eq": "John"},
				},
				Updater: `{"name": "Eric"}`,
			},
			&action.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Eric"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithMapFilter_LargeIntegerStraddling2To53_UpdatesOnlyExactMatch(t *testing.T) {
	test := testUtils.TestCase{
		// JS numbers are IEEE-754 doubles, so an integer filter condition above 2^53 has
		// already lost precision before it reaches Go.
		// To do: https://github.com/sourcenetwork/defradb/issues/5176.
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
			},
		),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Fred",
					"age": 9007199254740992
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 9007199254740993
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID: 0,
				Filter: map[string]any{
					"age": map[string]any{"_eq": int64(9007199254740993)},
				},
				Updater: `{"name": "Eric"}`,
			},
			&action.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Fred"},
						{"name": "Eric"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithSomeOptionFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID: 0,
				Filter: immutable.Some(request.Filter{
					Conditions: map[string]any{
						"name": map[string]any{"_eq": "John"},
					},
				}),
				Updater: `{"name": "Eric"}`,
			},
			&action.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Eric"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithNoneOptionFilter_UpdatesAllDocuments(t *testing.T) {
	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{
				state.GoClientType,
				state.HTTPClientType,
				state.CLIClientType,
				state.CClientType,
				state.JSClientType,
			},
		),
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Fred",
					"age": 25
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID:         0,
				Filter:               immutable.None[request.Filter](),
				Updater:              `{"verified": true}`,
				SkipLocalUpdateEvent: true,
			},
			&action.Request{
				Request: `query{
					Users {
						name
						verified
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "John", "verified": true},
						{"name": "Fred", "verified": true},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "John",
					"age": 21
				}`,
			},
			testUtils.UpdateWithFilter{
				CollectionID: 0,
				Filter:       `{name: {_eq: "John"}}`,
				Updater:      `{"name": "Eric"}`,
			},
			&action.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{"name": "Eric"},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
