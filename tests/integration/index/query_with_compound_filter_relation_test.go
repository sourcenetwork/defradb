// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package index

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndex_QueryWithIndexOnOneToManyRelationAndFilter_NoData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_and: [
								{ certificationBodyOrg: { name: { _eq: "Test" } } }
							]
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationOrFilter_NoData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_or: [
								{ certificationBodyOrg: { name: { _eq: "Test" } } }
							]
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationNotFilter_NoData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_not: {
								certificationBodyOrg: { name: { _eq: "Test" } }
							}
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationAndFilter_Data(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Source Inc."
				}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "LensVM",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zanzi"
				}`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_and: [
								{ certificationBodyOrg: { name: { _eq: "Source Inc." } } }
							]
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{
						{
							"name": "LensVM",
						},
						{
							"name": "DefraDB",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationOrFilter_Data(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "Source Inc."
                }`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "LensVM",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
                    "name": "Zanzi"
                }`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_or: [
								{ certificationBodyOrg: { name: { _eq: "Source Inc." } } },
								{ name: { _eq: "Zanzi" } }
							]
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{
						{
							"name": "Zanzi",
						},
						{
							"name": "LensVM",
						},
						{
							"name": "DefraDB",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationNotFilter_Data(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "Source Inc."
                }`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
                    "name": "Zanzi"
                }`,
			},
			testUtils.Request{
				Request: `query {
					Program(
						filter: {
							_not: {
								certificationBodyOrg: { name: { _eq: "Source Inc." } }
							}
						}
					) {
						name
					}
				}`,
				Results: map[string]any{
					"Program": []map[string]any{
						{
							"name": "Zanzi",
						},
						{
							"name": "Horizon",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
