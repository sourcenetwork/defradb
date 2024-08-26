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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationCreateFieldKinds_WithDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						time: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			testUtils.Request{
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
			testUtils.SchemaUpdate{
				Schema: `
					type User {
						time: DateTime
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000001Z",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000002Z",
				},
			},
			testUtils.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 1, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 2, time.UTC),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
