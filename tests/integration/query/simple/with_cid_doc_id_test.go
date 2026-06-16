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
	"github.com/sourcenetwork/defradb/tests/multiplier"
)

func TestQuerySimpleWithInvalidCidAndInvalidDocID(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
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
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.Request{
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
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CID0_0_0}}",
							docID: "{{.DocID0_0}}"
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
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CID0_0_0}}",
							docID: "{{.DocID0_0}}"
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
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CID0_0_1}}",
							docID: "{{.DocID0_0}}"
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
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnnn"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CID0_0_1}}",
							docID: "{{.DocID0_0}}"
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
									"cid": "{{.CID0_0_1}}",
								},
								{
									"cid": "{{.CID0_0_0}}",
								},
							},
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuerySimpleWithUpdateAndFirstCidAndDocIDAndSchemaVersion(t *testing.T) {
	test := testUtils.TestCase{
		// hardcoded CIDs would change under encryption
		MultiplierExcludes: []string{multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John"
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"name": "Johnn"
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
							cid: "{{.CID0_0_0}}",
							docID: "{{.DocID0_0}}"
						) {
						name
						_version {
							collectionVersionId
						}
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_version": []map[string]any{
								{
									"collectionVersionId": "{{.CollectionVersionID0}}",
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Int @crdt(type: pncounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": -5
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
						cid: "{{.CID0_0_0}}",
						docID: "{{.DocID0_0}}"
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float @crdt(type: pncounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": -5.3
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
						cid: "{{.CID0_0_0}}",
						docID: "{{.DocID0_0}}"
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Int @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 10
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": 20
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
						cid: "{{.CID0_0_0}}",
						docID: "{{.DocID0_0}}"
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
		// Accumulated CRDT fields (pncounter/pcounter) cannot be indexed.
		// https://github.com/sourcenetwork/defradb/issues/4439
		MultiplierExcludes: []string{multiplier.SecondaryIndex, multiplier.EncryptedDocs},
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
						points: Float @crdt(type: pcounter)
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"points": 10.2
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"points": 20.6
				}`,
			},
			&action.Request{
				Request: `query {
					Users (
						cid: "{{.CID0_0_0}}",
						docID: "{{.DocID0_0}}"
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
