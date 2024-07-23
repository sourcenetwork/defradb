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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestUpdateWithInvalidFilterType_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test update users with invalid filter type",
		Actions: []any{
			testUtils.UpdateWithFilter{
				CollectionID:  0,
				Filter:        t,
				Updater:       `{"name": "Eric"}`,
				ExpectedError: "invalid filter",
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithEmptyFilter_ReturnsError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test update users with empty filter",
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
		Description: "Test update users with filter and invalid JSON",
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
		Description: "Test update users with filter and invalid updator",
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
		Description: "Test update users with filter and patch updator (not implemented so no change)",
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
				Updater:      `[{"name": "Eric"}, {"name": "Sam"}]`,
			},
			testUtils.Request{
				Request: `query{
					Users {
						name
					}
				}`,
				Results: []map[string]any{
					{"name": "John"},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateWithFilter_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test update users with filter",
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
				Results: []map[string]any{
					{"name": "Eric"},
				},
			},
		},
	}

	executeTestCase(t, test)
}
