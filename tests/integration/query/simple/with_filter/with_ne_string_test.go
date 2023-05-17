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

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithStringNotEqualsFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne string filter",
		Request: `query {
					Users(filter: {Name: {_ne: "John"}}) {
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Age": uint64(32),
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithStringNotEqualsNilFilterBlock(t *testing.T) {
	test := testUtils.RequestTestCase{
		Description: "Simple query with ne string nil filter",
		Request: `query {
					Users(filter: {Name: {_ne: null}}) {
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				`{
					"Name": "John",
					"Age": 21
				}`,
				`{
					"Name": "Bob",
					"Age": 32
				}`,
				`{
					"Age": 36
				}`,
			},
		},
		Results: []map[string]any{
			{
				"Age": uint64(32),
			},
			{
				"Age": uint64(21),
			},
		},
	}

	executeTestCase(t, test)
}
