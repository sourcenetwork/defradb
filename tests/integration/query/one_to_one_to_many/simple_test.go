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

package one_to_one_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryOneToOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query across a one-to-one-to-many chain returns the full linked hierarchy.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"_indicatorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Observation": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToManyFromSecondaryOnOneToMany(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query from the secondary one-to-one side traverses into the one-to-many observations list.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Observable1",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Indicator1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Indicator": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToManyFromSecondaryOnOneToOne(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query from the secondary one-to-one side traverses back through the one-to-one link to indicator.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name": "Observable1",
				},
			},
			&action.AddDoc{
				CollectionID: 0,
				DocMap: map[string]any{
					"name":          "Indicator1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Observation": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryOneToOneToManyFromSecondary(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Query from the secondary indicator side retrieves the observable with its many observations.",
		Actions: []any{
			&action.AddCollection{
				SDL: `
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
			&action.AddDoc{
				CollectionID: 0,
				Doc: `{
						"name":	"Indicator1"
					}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				DocMap: map[string]any{
					"name":         "Observable1",
					"_indicatorID": testUtils.NewDocIndex(0, 0),
				},
			},
			&action.AddDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"name":          "Observation1",
					"_observableID": testUtils.NewDocIndex(1, 0),
				},
			},
			&action.Request{
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
				Results: map[string]any{
					"Indicator": []map[string]any{
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
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
