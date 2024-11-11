// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/tests/change_detector"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithInvalidCid(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with cid",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (cid: "any non-nil string value - this will be ignored") {
						Name
					}
				}`,
				ExpectedError: "invalid cid: selected encoding not supported",
			},
		},
	}

	executeTestCase(t, test)
}

// This test documents a bug:
// https://github.com/sourcenetwork/defradb/issues/3214
func TestQuerySimpleWithCid(t *testing.T) {
	if change_detector.Enabled {
		t.Skipf("Change detector does not support requiring panics")
	}

	test := testUtils.TestCase{
		SupportedClientTypes: immutable.Some(
			[]testUtils.ClientType{
				// The CLI/Http clients don't panic in this context
				testUtils.GoClientType,
			},
		),
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreib7afkd5hepl45wdtwwpai433bhnbd3ps5m2rv3masctda7b6mmxe"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
						},
					},
				},
			},
		},
	}

	require.Panics(t, func() {
		testUtils.ExecuteTestCase(t, test)
	})
}
