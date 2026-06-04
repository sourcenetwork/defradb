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

package index

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestIndex_QueryWithIndexOnOneToManyRelationAndFilter_NoData(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
					"name": "Source Inc."
				}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "LensVM",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
					"name": "Zanzi"
				}`,
			},
			&action.Request{
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
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "Source Inc."
                }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "LensVM",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
                    "name": "Zanzi"
                }`,
			},
			&action.Request{
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
							"name": "LensVM",
						},
						{
							"name": "DefraDB",
						},
						{
							"name": "Zanzi",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestIndex_QueryWithIndexOnOneToManyRelationNotFilter_Data(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
				  type Program {
					name: String
					certificationBodyOrg: Organization
				  }

				  type Organization {
					name: String @index
					programs: [Program]
				  }`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "Source Inc."
                }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "DefraDB",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc: `{
                    "name": "ESA"
                }`,
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":                 "Horizon",
					"certificationBodyOrg": testUtils.NewDocIndex(1, 1),
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
                    "name": "Zanzi"
                }`,
			},
			&action.Request{
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
							"name": "Horizon",
						},
						{
							"name": "Zanzi",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
