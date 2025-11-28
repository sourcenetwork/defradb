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

	"github.com/onsi/gomega"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryCommits(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	nameCid := testUtils.NewSameValue()
	ageCid := testUtils.NewSameValue()
	headCid := testUtils.NewSameValue()

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
						_commits {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": gomega.And(nameCid, uniqueCid),
						},
						{
							"cid": gomega.And(ageCid, uniqueCid),
						},
						{
							"cid": gomega.And(headCid, uniqueCid),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsMultipleDocs(t *testing.T) {
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
						"name":	"Shahzad",
						"age":	28
					}`,
			},
			testUtils.Request{
				Request: `query {
						_commits {
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
							"cid": "bafyreid5ve64mkobcop4bhx6e5pzcyfiysxbvut2hbr7r7udrrgn3tsute",
						},
						{
							"cid": "bafyreihgwlmva5z7odxvnb6dxomrji7gnbomtcoqpnl7qmcl5zg5wdy3mi",
						},
						{
							"cid": "bafyreicefeu4dm3hk5qw2oqwu4x3dpogw7ffxy7fbpy5ggsr7l7ozopvhm",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithSchemaVersionIDField(t *testing.T) {
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
						_commits {
							cid
							schemaVersionId
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":             "bafyreihakk5jjukb4fw7klfejdmniwhuscnckcjo677p3mtcxrdpiahuea",
							"schemaVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
						{
							"cid":             "bafyreihx4lnknvruc6vonsg3dvb3nnlsycwzbbkeulcutnzgidkzfvea64",
							"schemaVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
						{
							"cid":             "bafyreihpq4duzngkledmxkxx3jevlp2q4aimhmbjygpv5chmgbf6u2fsqm",
							"schemaVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameField(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			updateUserCollectionSchema(),
			testUtils.CreateDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						_commits {
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
						},
						{
							"fieldName": "name",
						},
						{
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithFieldNameFieldAndUpdate(t *testing.T) {
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
				Doc: `{
					"age":	22
				}`,
			},
			testUtils.Request{
				Request: `
					query {
						_commits {
							fieldName
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"fieldName": "age",
						},
						{
							"fieldName": "age",
						},
						{
							"fieldName": "name",
						},
						{
							"fieldName": "_C",
						},
						{
							"fieldName": "_C",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQuery_CommitsWithAllFieldsWithUpdate_NoError(t *testing.T) {
	uniqueCid := testUtils.NewUniqueValue()

	ageUpdateCid := testUtils.NewSameValue()
	ageCreateCid := testUtils.NewSameValue()
	nameCreateCid := testUtils.NewSameValue()
	updateCompositeCid := testUtils.NewSameValue()
	createCompositeCid := testUtils.NewSameValue()

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
				Doc: `{
						"age":	22
					}`,
			},
			testUtils.Request{
				Request: `
					query {
						_commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								cid
								name
							}
							signature {
								type
							}
						}
					}
				`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":       gomega.And(ageUpdateCid, uniqueCid),
							"delta":     testUtils.CBORValue(22),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "age",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":  ageCreateCid,
									"name": "_head",
								},
							},
							"signature": nil,
						},
						{
							"cid":       gomega.And(ageCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue(21),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "age",
							"height":    int64(1),
							"links":     []map[string]any{},
							"signature": nil,
						},
						{
							"cid":       gomega.And(nameCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue("John"),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
							"signature": nil,
						},
						{
							"cid":       gomega.And(updateCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "_C",
							"height":    int64(2),
							"links": []map[string]any{
								{
									"cid":  createCompositeCid,
									"name": "_head",
								},
								{
									"cid":  ageUpdateCid,
									"name": "age",
								},
							},
							"signature": nil,
						},
						{
							"cid":       gomega.And(createCompositeCid, uniqueCid),
							"delta":     nil,
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "_C",
							"height":    int64(1),
							"links": []map[string]any{
								{
									"cid":  ageCreateCid,
									"name": "age",
								},
								{
									"cid":  nameCreateCid,
									"name": "name",
								},
							},
							"_signature": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommits_WithAlias_Succeeds(t *testing.T) {
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
					history: _commits {
						cid
					}
				}`,
				Results: map[string]any{
					"history": []map[string]any{
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
