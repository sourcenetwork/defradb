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

func TestQueryCommitsWithDocIDAndOrderHeightDesc(t *testing.T) {
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
			testUtils.UpdateDoc{
				CollectionID: 0,
				DocID:        0,
				Doc: `{
					"age":	22
				}`,
			},
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
						},
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
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
			&action.AddDoc{
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
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
						},
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
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
			&action.AddDoc{
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
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {cid: DESC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
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
			&action.AddDoc{
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
			&action.Request{
				Request: `query {
						_commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {cid: ASC}) {
							cid
							height
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
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
			&action.AddDoc{
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
			&action.Request{
				Request: `query {
						 _commits(docID: "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738", order: {height: ASC}) {
							 cid
							 height
						 }
					 }`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":    "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"height": int64(1),
						},
						{
							"cid":    "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"height": int64(1),
						},
						{
							"cid":    "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"height": int64(1),
						},
						{
							"cid":    "bafyreihht6jz3vxk3fvr4sp3kqnvuplmva36hivbjtpdum7zydvb2yztwu",
							"height": int64(2),
						},
						{
							"cid":    "bafyreia4x5ju33jenbimdqbtnuqc7pby4lydpa7efyk5iu4nl6urm6ofla",
							"height": int64(2),
						},
						{
							"cid":    "bafyreiayx64xmsfgk2dz6mga2hcgm5ajbwrx2nhiroxyzdk7tfojjrl3fe",
							"height": int64(3),
						},
						{
							"cid":    "bafyreicbj6l6nnv6mlkjfhbc4ij36coaui7bejn7zbtxvhdl23d2w6qm5i",
							"height": int64(3),
						},
						{
							"cid":    "bafyreidw723v77miekvwuoouci6npeb6kkv6hidnplh5ob5sbahwsjduuy",
							"height": int64(4),
						},
						{
							"cid":    "bafyreigkfzwwfmcw2lkss4xecb27a6s3f6uzfaj5d3mqx5wninv6dnnhhq",
							"height": int64(4),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
