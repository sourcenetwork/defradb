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

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithInvalidCidAndInvalidDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with invalid cid and invalid docID",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "any non-nil string value - this will be ignored",
							docID: "invalid docID"
						) {
						name
					}
				}`,
				ExpectedError: "invalid cid: selected encoding not supported",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// This test is for documentation reasons only. This is not
// desired behaviour (should just return empty).
func TestQuerySimpleWithUnknownCidAndInvalidDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with unknown cid and invalid docID",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafybeid57gpbwi4i6bg7g357vwwyzsmr4bjo22rmhoxrwqvdxlqxcgaqvu",
							docID: "invalid docID"
						) {
						name
					}
				}`,
				ExpectedError: "failed to get block in blockstore: ipld: could not find",
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with cid and docID",
		Actions: []any{
			&action.AddSchema{
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
							cid: "bafyreifhbll6j3m5imwxdkumaumjl5hevppuzcfbofamsbihla6ob2asyi",
							docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e"
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (first) cid and docID",
		Actions: []any{
			&action.AddSchema{
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
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreifhbll6j3m5imwxdkumaumjl5hevppuzcfbofamsbihla6ob2asyi",
							docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e"
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

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndLastCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (last) cid and docID",
		Actions: []any{
			&action.AddSchema{
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
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreihmihcd4hgubxxtvngdesz2pyehhowakg3sov7kfyselmn3yg7eei",
							docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e"
						) {
						name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Johnn",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndMiddleCidAndDocID(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (middle) cid and docID",
		Actions: []any{
			&action.AddSchema{
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
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreihmihcd4hgubxxtvngdesz2pyehhowakg3sov7kfyselmn3yg7eei",
							docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e"
						) {
						name
						_version {
							cid
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "Johnn",
							"_version": []map[string]any{
								{
									"cid": "bafyreihmihcd4hgubxxtvngdesz2pyehhowakg3sov7kfyselmn3yg7eei",
								},
								{
									"cid": "bafyreifhbll6j3m5imwxdkumaumjl5hevppuzcfbofamsbihla6ob2asyi",
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

func TestQuerySimpleWithUpdateAndFirstCidAndDocIDAndSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with (first) cid and docID and yielded schema version",
		Actions: []any{
			&action.AddSchema{
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
			testUtils.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
							cid: "bafyreifhbll6j3m5imwxdkumaumjl5hevppuzcfbofamsbihla6ob2asyi",
							docID: "bae-0623ed7c-0861-5995-a5d7-cce53642a83e"
						) {
						name
						_version {
							schemaVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"schemaVersionId": "bafyreigsld6ten2pppcu2tgkbexqwdndckp6zt2vfjhuuheykqkgpmwk7i",
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

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPNCounterWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pncounter int type",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": -5
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafyreiggwvfcs3z5v7ovvg7bcgaxgbvxcf3lj2glj4d25ikats6z5mkvra",
						docID: "bae-b9db5d1c-0f75-5d98-a65b-d086dbe9adf0"
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": int64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPNCounterWithFloatKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pncounter and float type",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": -5.3
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafyreifnratxa3bv4tyb2d4kof5zbn4rvz6pkzz64uq2gc4273imfzh2h4",
						docID: "bae-60aa594f-20be-5370-b47b-a95a9191f148"
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": 10.2,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPCounterWithIntKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pcounter int type",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafyreihjdxfwfazfnugiw5j65ixknvoixvnyvo5wluepy3czryzvbdwy44",
						docID: "bae-f0c1cc11-ca60-5db2-8e7b-8ff63234e54c"
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": int64(10),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// Note: Only the first CID is reproducible given the added entropy to the Counter CRDT type.
func TestCidAndDocIDQuery_ContainsPCounterWithFloatKind_NoError(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple query with first cid and docID with pcounter and float type",
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
						points: Float @crdt(type: pcounter)
					}
				`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			testUtils.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users (
						cid: "bafyreibrhd7uh6f2fzijqct32f3kwsmpdcopywl2kde7oeehrjiak53v3m",
						docID: "bae-2e998ed4-a819-5e7d-bc08-8dbbdca860a1"
					) {
						name
						points
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":   "John",
							"points": 10.2,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
