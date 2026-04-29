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

package commits

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDocIDAndUnknownField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", 
							filter: {fieldName: {_eq: "not a field"}}
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndUnknownFieldId(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", 
							filter: {fieldName: {_eq: "999999"}
						}) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							filter: {fieldName: {_eq: "age"}}
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": testUtils.ValidCID()},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndCompositeField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
				Request: `query {
						_commits(
							docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							filter: {fieldName: {_eq: "_C"}}
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{"cid": testUtils.ValidCID()},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
