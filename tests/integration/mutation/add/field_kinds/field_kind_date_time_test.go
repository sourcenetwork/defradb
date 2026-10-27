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

package field_kinds

import (
	"testing"
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestMutationAddFieldKinds_WithDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithDateTimesNanoSecondsAppart(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000Z",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000001Z",
				},
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": "2017-07-23T03:46:56.000000002Z",
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 2, time.UTC),
						},
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 1, time.UTC),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithDateTime_WithUTCNow(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.Request{
				Request: `mutation {
                    add_User(input: {time: UTC_NOW}) {
						time
                    }
                }`,
				Results: map[string]any{
					"add_User": []map[string]any{
						{
							"time": testUtils.CurrentTimestamp(),
						},
					},
				},
			},
		},
	}
	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableDateTime_Nil(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						time: DateTime
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": nil,
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": nil,
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						time: DateTime!
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
				},
			},
			&action.Request{
				Request: `query {
					User {
						time
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"time": time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableDateTimeArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []time.Time{
						time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []immutable.Option[time.Time]{
								immutable.Some(time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC)),
								immutable.Some(time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNillableDateTimeArray_WithNilElement(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []any{
						time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						nil,
						time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []immutable.Option[time.Time]{
								immutable.Some(time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC)),
								immutable.None[time.Time](),
								immutable.Some(time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableDateTimeArray_FromTypedSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []time.Time{
						time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []time.Time{
								time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
								time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAddFieldKinds_WithNonNillableDateTimeArray_FromAnySlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []any{
						time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
						time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []time.Time{
								time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
								time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestMutationAdd_WithDateTime_SetsTwoEqualUTCNowValues(t *testing.T) {
	timestampMatcher := testUtils.NewSameValue()
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						name: String
						created: DateTime
					}
				`,
			},
			&action.Request{
				Request: `mutation {
					bob: add_User(input: { name: "Bob", created: UTC_NOW }) {
						created
					}

					alice: add_User(input: { name: "Alice", created: UTC_NOW }) {
						created
					}
                }`,
				Results: map[string]any{
					"bob":   []map[string]any{{"created": timestampMatcher}},
					"alice": []map[string]any{{"created": timestampMatcher}},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestMutationAddFieldKinds_WithNillableDateTimeArray_FromStringSlice verifies
// that a []string of RFC3339 timestamps is parsed element-by-element rather than
// silently dropped, matching the []any behaviour.
func TestMutationAddFieldKinds_WithNillableDateTimeArray_FromStringSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []string{
						"2017-07-23T03:46:56Z",
						"2018-08-24T04:47:57Z",
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []immutable.Option[time.Time]{
								immutable.Some(time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC)),
								immutable.Some(time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC)),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

// TestMutationAddFieldKinds_WithNonNillableDateTimeArray_FromStringSlice verifies
// that a []string of RFC3339 timestamps is parsed rather than silently dropped for
// a non-nillable array field.
func TestMutationAddFieldKinds_WithNonNillableDateTimeArray_FromStringSlice(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type User {
						times: [DateTime!]
					}
				`,
			},
			&action.AddDoc{
				DocMap: map[string]any{
					"times": []string{
						"2017-07-23T03:46:56Z",
						"2018-08-24T04:47:57Z",
					},
				},
			},
			&action.Request{
				Request: `query {
					User {
						times
					}
				}`,
				Results: map[string]any{
					"User": []map[string]any{
						{
							"times": []time.Time{
								time.Date(2017, time.July, 23, 3, 46, 56, 0, time.UTC),
								time.Date(2018, time.August, 24, 4, 47, 57, 0, time.UTC),
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
