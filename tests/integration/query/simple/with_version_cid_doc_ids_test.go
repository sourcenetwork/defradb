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

package simple

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithVersionAndCidAndCorrectDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			&action.Request{
				Request: `query {
					Users(cid: "bafyreic2xpowsfqw5vh42kjlyykrewjd77rsofsdfuz4slgvaeviv7hbbq", docID: ["bae-97a6033e-d2b5-564d-828f-d5edd9d4d536"]) {
						Name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Chris",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithVersionAndCidAndCorrectAndIncorrectDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			&action.Request{
				Request: `query {
					Users(
						cid: "bafyreic2xpowsfqw5vh42kjlyykrewjd77rsofsdfuz4slgvaeviv7hbbq",
						docID: ["bae-97a6033e-d2b5-564d-828f-d5edd9d4d536", "bae-fda35cb5-cd39-5d52-80b8-b324f2d7a8b0"]
					) {
						Name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Chris",
							"_version": []map[string]any{
								{
									"fieldName": "_C",
								},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithVersionAndCidAndIncorrectDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						Name: String
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			&action.Request{
				Request: `query {
					Users(
						cid: "bafyreic2xpowsfqw5vh42kjlyykrewjd77rsofsdfuz4slgvaeviv7hbbq",
						docID: ["bae-fda35cb5-cd39-5d52-80b8-b324f2d7a8b0"]
					) {
						Name
						_version {
							fieldName
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
