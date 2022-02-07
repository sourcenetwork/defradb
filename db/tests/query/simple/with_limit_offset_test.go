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

	testUtils "github.com/sourcenetwork/defradb/db/tests"
)

func TestQuerySimpleWithLimit(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple query with basic limit",
			Query: `query {
						users(limit: 1) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
			},
		},
		{
			Description: "Simple query with basic limit, more rows",
			Query: `query {
						users(limit: 2) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "Bob",
					"Age":  uint64(32),
				},
				{
					"Name": "Alice",
					"Age":  uint64(19),
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}

func TestQuerySimpleWithLimitAndOffset(t *testing.T) {
	tests := []testUtils.QueryTestCase{
		{
			Description: "Simple query with basic limit & offset",
			Query: `query {
						users(limit: 1, offset: 1) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
			},
		},
		{
			Description: "Simple query with basic limit & offset, more rows",
			Query: `query {
						users(limit: 2, offset: 2) {
							Name
							Age
						}
					}`,
			Docs: map[int][]string{
				0: {
					(`{
					"Name": "John",
					"Age": 21
				}`),
					(`{
					"Name": "Bob",
					"Age": 32
				}`),
					(`{
					"Name": "Carlo",
					"Age": 55
				}`),
					(`{
					"Name": "Alice",
					"Age": 19
				}`)},
			},
			Results: []map[string]interface{}{
				{
					"Name": "John",
					"Age":  uint64(21),
				},
				{
					"Name": "Carlo",
					"Age":  uint64(55),
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
