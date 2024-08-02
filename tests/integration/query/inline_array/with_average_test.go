// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package inline_array

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQueryInlineIntegerArrayWithAverageAndNullArray(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Simple inline array with no filter, average of nil integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of empty integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": []
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of zero integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [0, 0, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of populated integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [-1, 0, 9, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteIntegers: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(2),
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
		Description: "Simple inline array with no filter, average of populated nillable integer array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"testScores": [-1, null, 13, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(testScores: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(4),
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
		Description: "Simple inline array with no filter, average of nil float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": null
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of empty float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": []
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of zero float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [0, 0, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{

							"name": "John",
							"_avg": float64(0),
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
		Description: "Simple inline array with no filter, average of populated float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"favouriteFloats": [-0.1, 0, 0.9, 0]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(favouriteFloats: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0.2),
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
		Description: "Simple inline array with no filter, average of populated nillable float array",
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"name": "John",
					"pageRatings": [-0.1, 0, 0.9, 0, null]
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users {
						name
						_avg(pageRatings: {})
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name": "John",
							"_avg": float64(0.2),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}
