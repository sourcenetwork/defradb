// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package field_kinds

import (
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateFieldKinds_WithDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreateFieldKinds_WithDateTimesNanoSecondsAppart(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000001Z",
				},
			},
			&action.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000002Z",
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 2, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 1, time.UTC),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreateFieldKinds_WithDateTime_WithUTCNow(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.Request{
				Request: `mutation {
                    create_User(input: {time: UTC_NOW}) {
						time
                    }
                }`,
				Results: map[string]any{
					"create_User": []map[string]any{
						{
							"time": testUtils.CurrentTimestamp(),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestMutationCreate_WithDateTime_SetsTwoEqualUTCNowValues(t *testing.T) {
	timestampMatcher := testUtils.NewSameValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type User {
						name: String
						created: DateTime
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					bob: create_User(input: { name: "Bob", created: UTC_NOW }) {
						created
					}

					alice: create_User(input: { name: "Alice", created: UTC_NOW }) {
						created
					}
                }`,
				Results: map[string]any{
					"bob":   []map[string]any{{"created": timestampMatcher}},
					"alice": []map[string]any{{"created": timestampMatcher}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
