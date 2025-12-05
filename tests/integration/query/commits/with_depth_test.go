// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package commits

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommitsWithDepth1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
						},
						{
							"cid": "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
						},
						{
							"cid": "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// "Age" field head
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth2WithUpdate(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	23
				}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(depth: 2) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							// Composite head
							"cid":    "bafyreicq5lp5kzbj4pop6prenjfyrlhzm3ihkamhj4if24lxopybmrye5a",
							"height": int64(3),
						},
						{
							// Composite head -1
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							// "Name" field head (unchanged from create)
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							// "Age" field head
							"cid":    "bafyreiajxvestc7ya7kb76xyypcmtgwa3zbqugljrbizqya5od24o3bwpm",
							"height": int64(3),
						},
						{
							// "Age" field head -1
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDepth1AndMultipleDocs(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Fred",
						"age":	25
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(depth: 1) {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
						},
						{
							"cid": "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
						},
						{
							"cid": "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
						},
						{
							"cid": "bafyreigccsr2vxscb3izhibriqqcvvrhkcd3nn3at4wisqnpt2xqj5b3vi",
						},
						{
							"cid": "bafyreih6sikzsu2dpjsnxi7la3slidjx66lkjrl3gesa3uv4su3wcz3jbu",
						},
						{
							"cid": "bafyreiaees4qbril4il7zrp5vdgkwp27gafu5v5if5yv3afue5drfkj5ci",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithFilterFieldNameAndDepth_ReturnsCommitsAtAllHeights(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.UpdateDoc{
				Doc: `{"age": 22}`,
			},
			testUtils.UpdateDoc{
				Doc: `{"age": 23}`,
			},
			testUtils.Request{
				Request: `query {
						_commits(filter: {fieldName: {_eq: "age"}}, depth: 2) {
							fieldName
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
							"height":    int64(3),
						},
						{
							"fieldName": "age",
							"height":    int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
