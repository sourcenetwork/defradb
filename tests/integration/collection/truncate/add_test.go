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

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestTruncateCollectionAdd_RemovesDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

// todo - this test shouldn't need to exist, but we dont currently test
// Defra with digital signatures enabled besides a handful of tests that
// explicitly enable it. Remove this test as part of:
// https://github.com/sourcenetwork/defradb/issues/4671
func TestTruncateCollectionAdd_RemovesSignedDocument(t *testing.T) {
	test := testUtils.TestCase{
		EnableSigning: true,
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestTruncateCollectionAdd_RemovesEncryptedDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID:   0,
				IsDocEncrypted: true,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
					Users {
						name
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

func TestTruncateCollectionAdd_RemovesBlocks(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.Request{
				Request: `query {
						_commits (filter: {fieldName: {_eq: "_C"}}) {
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

func TestTruncateCollectionAdd_AddsDocWithSameDocIDAsOriginal(t *testing.T) {
	docID := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
					Users {
						_docID
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": docID,
						},
					},
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				// Assert that there is only one User, and that it has the same docID as the
				// original, truncated, document.
				Request: `query {
					Users {
						_docID
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"_docID": docID,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionAdd_AddsDocWithSameCIDAsOriginal(t *testing.T) {
	compositeCID := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": compositeCID,
						},
					},
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				// Assert that new document composite commit has the same cid as the
				// original, truncated, document.
				Request: `query {
						_commits(filter: {fieldName: {_eq: "_C"}}) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": compositeCID,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestTruncateCollectionAdd_AddsDocWithBlocksAtHeight1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Users {
						name: String
					}
				`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Request{
				// Query the commits api and make sure that the document has been created with
				// blocks at height 1.
				Request: `
					query {
						_commits {
							fieldName
							height
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "name",
							"height":    int64(1),
						},
						{
							"fieldName": "_C",
							"height":    int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
