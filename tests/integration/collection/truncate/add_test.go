// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package truncate

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestCollectionTruncateAdd_RemovesDocument(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.CreateDoc{
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

func TestCollectionTruncateAdd_RemovesBlocks(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.CreateDoc{
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

func TestCollectionTruncateAdd_CreatesDocWithSameDocIDAsOriginal(t *testing.T) {
	docID := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
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

func TestCollectionTruncateAdd_CreatesDocWithSameCIDAsOriginal(t *testing.T) {
	compositeCID := testUtils.NewSameValue()

	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
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

func TestCollectionTruncateAdd_CreatesDocWithBlocksAtHeight1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
					type Users {
						name: String
					}
				`,
			},
			&action.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name": "John",
				},
			},
			&action.Truncate{
				CollectionIndex: 0,
			},
			&action.CreateDoc{
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
