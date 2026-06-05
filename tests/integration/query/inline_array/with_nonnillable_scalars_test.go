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

func TestQueryInlineArrayWithNonNillableBooleans_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						likedIndexes
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":         "John",
							"likedIndexes": nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableBooleans_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"likedIndexes": [true, false, true]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						likedIndexes
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":         "John",
							"likedIndexes": []bool{true, false, true},
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_Null(t *testing.T) {
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
						favouriteIntegers
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteIntegers": nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableInts_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteIntegers": [1, 2, 3]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						favouriteIntegers
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteIntegers": []int64{1, 2, 3},
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloat64s": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						favouriteFloat64s
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteFloat64s": nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat64s_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloat64s": [1.1, 2.2, 3.3]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						favouriteFloat64s
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteFloat64s": []float64{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloat32s": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						favouriteFloat32s
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteFloat32s": nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableFloat32s_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"favouriteFloat32s": [1.1, 2.2, 3.3]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						favouriteFloat32s
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":              "John",
							"favouriteFloat32s": []float32{1.1, 2.2, 3.3},
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": null
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":             "John",
							"preferredStrings": nil,
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}

func TestQueryInlineArrayWithNonNillableStrings_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "John",
					"preferredStrings": ["one", "two", "three"]
				}`,
			},
			&action.Request{
				Request: `query {
					Users {
						name
						preferredStrings
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"name":             "John",
							"preferredStrings": []string{"one", "two", "three"},
						},
					},
				},
			},
		},
	}
	executeTestCase(t, test)
}
