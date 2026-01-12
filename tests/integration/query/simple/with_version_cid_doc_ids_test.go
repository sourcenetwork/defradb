// Copyright 2025 Democratized Data Foundation
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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithVersionAndCidAndCorrectDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			testUtils.Request{
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
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			testUtils.Request{
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
			&action.AddSchema{
				Schema: `
					type Users {
						Name: String
					}
				`,
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "John",
				},
			},
			testUtils.CreateDoc{
				DocMap: map[string]any{
					"Name": "Chris",
				},
			},
			testUtils.Request{
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
