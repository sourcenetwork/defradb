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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderHeightAsc(t *testing.T) {
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
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

func TestQueryCommitsWithDocIDAndOrderCidDesc(t *testing.T) {
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderCidAsc(t *testing.T) {
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
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithDocIDAndOrderAndMultiUpdatesCidAsc(t *testing.T) {
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	24
				}`,
			},
			testUtils.Request{
				Request: `query {
						 _commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreia5jhb6ughpzd2rjszl4qbdd4w5zrdjfoseyrvnmhm2xiyrudvja4",
							"height": int64(2),
						},
						{
							"cid":    "bafyreieira5p74wdicqhelwbjsin7jtnnvvlplngrrcqfapleq2phexqga",
							"height": int64(2),
						},
						{
							"cid":    "bafyreicq5lp5kzbj4pop6prenjfyrlhzm3ihkamhj4if24lxopybmrye5a",
							"height": int64(3),
						},
						{
							"cid":    "bafyreiajxvestc7ya7kb76xyypcmtgwa3zbqugljrbizqya5od24o3bwpm",
							"height": int64(3),
						},
						{
							"cid":    "bafyreigbdxwbmbnuvpjywymoc2bbqxtbu2ofc4dxaxk4o4d2q3wykmzlkm",
							"height": int64(4),
						},
						{
							"cid":    "bafyreifugntzbj44gaf36fi4djxr5jgoo7lbbfbspi3wlgz7zm5hihmzue",
							"height": int64(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
