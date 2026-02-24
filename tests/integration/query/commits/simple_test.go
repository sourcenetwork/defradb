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

	"github.com/sourcenetwork/defradb/tests/action"
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Shahzad",
						"age":	28
					}`,
			},
			&action.Request{
				Request: `query {
						_commits {
							cid
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
						},
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
						{
							"cid": "bafyreibeindczotofvlnhqgoeedqrphxxayp4ugsrcutr4kgda3uq3fy2y",
						},
						{
							"cid": "bafyreigbqcwozrscjscuclr3hxehca4skh5n7eqvtfsmbzcql7uz2pyk7e",
						},
						{
							"cid": "bafyreihaypay6hfz3czcdjzasyqthwf6kw3hhpucmsmrdx4et6wwokjw64",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryCommitsWithCollectionVersionIDField(t *testing.T) {
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
			&action.Request{
				Request: `query {
						_commits {
							cid
							collectionVersionId
						}
					}`,
				Results: map[string]any{
					"_commits": []map[string]any{
						{
							"cid":                 "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
							"collectionVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
						{
							"cid":                 "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
							"collectionVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
						},
						{
							"cid":                 "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
							"collectionVersionId": "bafyreicrgjxxcviov5jawe2haq5fbtd4jxt63vsdhqpcyaaahiothj72tu",
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
			&action.AddDoc{
				Doc: `{
						"name":	"John",
						"age":	21
					}`,
			},
			&action.Request{
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
			&action.AddDoc{
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
			&action.Request{
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
			&action.AddDoc{
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
			&action.Request{
				Request: `
					query {
						_commits {
							cid
							delta
							docID
							fieldName
							height
							links {
								fieldName
								cid
							}
							heads {
								cid
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
							"links":     []map[string]any{},
							"heads": []map[string]any{
								{
									"cid": ageCreateCid,
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
							"heads":     []map[string]any{},
							"signature": nil,
						},
						{
							"cid":       gomega.And(nameCreateCid, uniqueCid),
							"delta":     testUtils.CBORValue("John"),
							"docID":     "bae-1084671a-e3fb-5f2e-97a0-eb9d684e9738",
							"fieldName": "name",
							"height":    int64(1),
							"links":     []map[string]any{},
							"heads":     []map[string]any{},
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
									"cid":       ageUpdateCid,
									"fieldName": "age",
								},
							},
							"heads": []map[string]any{
								{
									"cid": createCompositeCid,
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
									"cid":       ageCreateCid,
									"fieldName": "age",
								},
								{
									"cid":       nameCreateCid,
									"fieldName": "name",
								},
							},
							"heads":      []map[string]any{},
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name":	"John",
					"age":	21
				}`,
			},
			&action.Request{
				Request: `query {
					history: _commits {
						cid
					}
				}`,
				Results: map[string]any{
					"history": []map[string]any{
						{
							"cid": "bafyreiajq6jmyblg2b6vupjdapzkaodbt7kkwqp4fijekdvydnyxvr4y7q",
						},
						{
							"cid": "bafyreigonvri5vfdosfgp4qxtq46snjxm7cnjlzizrod2wy3l53jbxiysm",
						},
						{
							"cid": "bafyreiejjfevlp5wrfl5o7bxbdtjj4th36lbdjov5gdkmy5n5jzs6dcmpu",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
