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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQueryCommitsWithFieldAndCID(t *testing.T) {
	test := testUtils.TestCase{
		// Result CIDs are hardcoded because template placeholders are not
		// resolved inside Request.Results.
		// See https://github.com/sourcenetwork/defradb/issues/4745.
		MultiplierExcludes: []string{multiplier.SignedDocs, multiplier.EncryptedDocs},
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
						_commits (

							filter: {fieldName: {_eq: "age"}},
							cid: "{{.FieldCID0_0_age_0}}"
						) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": testUtils.ValidCID(),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithWrongFieldAndCID_ReturnEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs},
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
						_commits (

							filter: {fieldName: {_eq: "name"}},
							cid: "{{.FieldCID0_0_age_0}}"
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

func TestQueryCommits_WithInvalidFieldAndCID_ReturnEmptyList(t *testing.T) {
	test := testUtils.TestCase{
		MultiplierExcludes: []string{multiplier.SignedDocs},
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
						_commits (

							filter: {fieldName: {_eq: "NOT_A_FIELD"}},
							cid: "{{.FieldCID0_0_age_0}}"
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
