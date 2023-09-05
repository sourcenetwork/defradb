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
	inlineArray "github.com/sourcenetwork/defradb/tests/integration/mutation/inline_array"
)

func TestMutationInlineArrayWithNillableStrings(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple inline array with no filter, nillable strings",
		Request: `mutation {
					update_Users(data: "{\"pageHeaders\": [\"\", \"the previous\", null, \"empty string\", \"blank string\", \"hitchi\"]}") {
						name
						pageHeaders
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"name": "John",
					"pageHeaders": ["", "the previous", "the first", "empty string", null]
				}`,
			},
		},
		Results: []map[string]any{
			{
				"name": "John",
				"pageHeaders": []immutable.Option[string]{
					immutable.Some(""),
					immutable.Some("the previous"),
					immutable.None[string](),
					immutable.Some("empty string"),
					immutable.Some("blank string"),
					immutable.Some("hitchi"),
				},
			},
		},
	}

	inlineArray.ExecuteTestCase(t, test)
}
