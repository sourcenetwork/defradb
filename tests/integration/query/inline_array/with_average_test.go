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

package inline_array

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArrayWithAverageAndNullArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndEmptyArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndZeroArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [0, 0, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineIntegerArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [-1, 0, 9, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableIntegerArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"testScores": [-1, null, 13, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(testScores: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(4),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndNullArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndEmptyArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": []
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndZeroArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [0, 0, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{

							"name": "John",
							"AVG":  float64(0),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineFloatArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [-0.1, 0, 0.9, 0]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0.2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQueryInlineNillableFloatArrayWithAverageAndPopulatedArray(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"pageRatings": [-0.1, 0, 0.9, 0, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						AVG(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"AVG":  float64(0.2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
