// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"testing"

	"github.com/sourcenetwork/immutable"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/sourcenetwork/defradb/tests/state"
)

func TestUpdateWithInvalidFilterType_ReturnsError(t *testing.T) {
	type invalidFilterType struct{ Number int }
	test := testUtils.TestCase{
		// http and cli clients will pass the serialize filter into json which will result in
		// the payload deserialized into map[string]any. With Go client the filter is passed as is.
		SupportedClientTypes: immutable.Some(
			[]state.ClientType{testUtils.HTTPClientType, testUtils.CLIClientType}),
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        invalidFilterType{Number: 1},
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "key not found",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithInvalidFilterType_WithGoClient_ReturnsError(t *testing.T) {
	type invalidFilterType struct{ Number int }
	test := testUtils.TestCase{
		Description:          "Test update users with invalid filter type (go client)",
		SupportedClientTypes: immutable.Some([]state.ClientType{testUtils.GoClientType}),
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        invalidFilterType{Number: 1},
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "invalid filter",
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
				ExpectedError: "invalid filter",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithInvalidJSON_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.CreateDoc{
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
			testUtils.Request{
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

func TestUpdateWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
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
			testUtils.Request{
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
