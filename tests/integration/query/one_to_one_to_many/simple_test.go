// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package one_to_one_to_many

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple one to one to many query, from primary direction",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Indicator {
						name: String
						observable: Observable
					}

					type Observable {
						name: String
						indicator: Indicator @primary
						observations: [Observation]
					}

					type Observation {
						name: String
						observable: Observable
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"indicator_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"observable_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query  {
							Observation {
								name
								observable {
									name
									indicator {
										name
									}
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Observation1",
						"observable": map[string]any{
							"name": "Observable1",
							"indicator": map[string]any{
								"name": "Indicator1",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToManyFromSecondaryOnOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple one to one to many query, secondary direction across the one to many",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Indicator {
						name: String
						observable: Observable @primary
					}

					type Observable {
						name: String
						indicator: Indicator
						observations: [Observation]
					}

					type Observation {
						name: String
						observable: Observable
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"indicator_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"observable_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query  {
							Indicator {
								name
								observable {
									name
									observations {
										name
									}
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Indicator1",
						"observable": map[string]any{
							"name": "Observable1",
							"observations": []map[string]any{
								{
									"name": "Observation1",
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

func TestQueryOneToOneToManyFromSecondaryOnOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple one to one to many query, secondary direction across the one to one",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Indicator {
						name: String
						observable: Observable @primary
					}

					type Observable {
						name: String
						indicator: Indicator
						observations: [Observation]
					}

					type Observation {
						name: String
						observable: Observable
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"indicator_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"observable_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query  {
							Observation {
								name
								observable {
									name
									indicator {
										name
									}
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Observation1",
						"observable": map[string]any{
							"name": "Observable1",
							"indicator": map[string]any{
								"name": "Indicator1",
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToManyFromSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple one to one to many query, from secondary direction ",
		Actions: []any{
			testUtils.SchemaUpdate{
				Schema: `
					type Indicator {
						name: String
						observable: Observable
					}

					type Observable {
						name: String
						indicator: Indicator @primary
						observations: [Observation]
					}

					type Observation {
						name: String
						observable: Observable
					}
				`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"indicator_id": testUtils.NewDocIndex(0, 0),
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"observable_id": testUtils.NewDocIndex(1, 0),
				},
			},
			testUtils.Request{
				Request: `query  {
							Indicator {
								name
								observable {
									name
									observations {
										name
									}
								}
							}
						}`,
				Results: []map[string]any{
					{
						"name": "Indicator1",
						"observable": map[string]any{
							"name": "Observable1",
							"observations": []map[string]any{
								{
									"name": "Observation1",
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
