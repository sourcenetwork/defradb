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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var dateTimeArrayCollection = (`
	type Events {
		name: String
		scheduledTimes: [DateTime!]
		optionalTimes: [DateTime]
	}
`)

// executeTestCaseDateTime adds the DateTime array schema then runs the test
// while preserving all other TestCase fields (Description, etc.).
func executeTestCaseDateTime(t *testing.T, test testUtils.TestCase) {
	fullTest := test // copy struct
	fullTest.Actions = append(
		[]any{
			&action.AddCollection{
				SDL: dateTimeArrayCollection,
			},
		},
		test.Actions...,
	)

	testUtils.ExecuteTestCase(t, fullTest)
}

func TestQueryInlineArrayWithDateTime_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Birthday Party",
					"scheduledTimes": null
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						scheduledTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name":           "Birthday Party",
							"scheduledTimes": nil,
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_EmptyList(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Meeting",
					"scheduledTimes": []
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						scheduledTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name":           "Meeting",
							"scheduledTimes": []time.Time{},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_NotEmpty(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Conference",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-15T14:00:00Z", "2024-01-16T10:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						scheduledTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Conference",
							"scheduledTimes": []time.Time{
								testUtils.MustParseTime("2024-01-15T09:00:00Z"),
								testUtils.MustParseTime("2024-01-15T14:00:00Z"),
								testUtils.MustParseTime("2024-01-16T10:00:00Z"),
							},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithNillableDateTime(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Workshop",
					"optionalTimes": ["2024-02-20T10:00:00Z", null, "2024-02-21T15:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						optionalTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Workshop",
							"optionalTimes": []immutable.Option[time.Time]{
								immutable.Some(testUtils.MustParseTime("2024-02-20T10:00:00Z")),
								immutable.None[time.Time](),
								immutable.Some(testUtils.MustParseTime("2024-02-21T15:00:00Z")),
							},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithNillableDateTime_AllNull(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Pending",
					"optionalTimes": [null, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						optionalTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Pending",
							"optionalTimes": []immutable.Option[time.Time]{
								immutable.None[time.Time](),
								immutable.None[time.Time](),
							},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

// TestQueryInlineArrayWithNillableDateTime_FilterAny tests filtering on [DateTime] arrays
// which contain nullable elements. This exercises the immutable.Option[time.Time] handling
// in the connor filter engine.
func TestQueryInlineArrayWithNillableDateTime_FilterAny(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Mixed Event",
					"optionalTimes": ["2024-03-15T10:00:00Z", null, "2024-03-16T10:00:00Z"]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Empty Event",
					"optionalTimes": [null, null]
				}`,
			},
			&action.Request{
				Request: `query {
					Events(filter: {optionalTimes: {_any: {_eq: "2024-03-15T10:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Mixed Event",
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_FilterAny(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Morning Event",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-15T10:00:00Z"]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Evening Event",
					"scheduledTimes": ["2024-01-15T18:00:00Z", "2024-01-15T19:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events(filter: {scheduledTimes: {_any: {_eq: "2024-01-15T09:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Morning Event",
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_FilterAll(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Waitlisted",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-16T09:00:00Z"]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Confirmed",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-15T10:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events(filter: {scheduledTimes: {_all: {_eq: "2024-01-15T09:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{},
				},
			},
			&action.Request{ // _ge matches both docs since all their scheduledTimes are >= the target
				Request: `query {
					Events(filter: {scheduledTimes: {_all: {_geq: "2024-01-15T09:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Confirmed",
						},
						{
							"name": "Waitlisted",
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_FilterNone(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Future Event",
					"scheduledTimes": ["2024-02-01T00:00:00Z"]
				}`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Past Event",
					"scheduledTimes": ["2023-01-01T00:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events(filter: {scheduledTimes: {_none: {_lt: "2024-01-01T00:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Future Event",
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_Index(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
					type Events {
						name: String
						scheduledTimes: [DateTime!] @index
					}
				`,
			},
			&action.AddDoc{
				Doc: `{
					"name": "Indexed Event",
					"scheduledTimes": ["2024-06-01T12:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events(filter: {scheduledTimes: {_any: {_eq: "2024-06-01T12:00:00Z"}}}) {
						name
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Indexed Event",
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}

func TestQueryInlineArrayWithDateTime_Update(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
					"name": "Rescheduled Event",
					"scheduledTimes": ["2024-01-01T00:00:00Z"]
				}`,
			},
			&action.UpdateDoc{
				Doc: `{
					"scheduledTimes": ["2024-02-01T00:00:00Z", "2024-03-01T00:00:00Z"]
				}`,
			},
			&action.Request{
				Request: `query {
					Events {
						name
						scheduledTimes
					}
				}`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Rescheduled Event",
							"scheduledTimes": []time.Time{
								testUtils.MustParseTime("2024-02-01T00:00:00Z"),
								testUtils.MustParseTime("2024-03-01T00:00:00Z"),
							},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_ErrorMalformed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
                    "name": "Bad Date",
                    "scheduledTimes": ["not-a-date"]
                }`,
				// Error wording differs between mutation paths (collection vs GQL),
				// but both surface the offending value verbatim.
				ExpectedError: "not-a-date",
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

func TestQueryInlineArrayWithDateTime_UTC_NOW(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
                    "name": "Now Event",
                    "scheduledTimes": ["UTC_NOW"]
                }`,
			},
			&action.Request{
				Request: `query { Events { name scheduledTimes } }`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name":           "Now Event",
							"scheduledTimes": testUtils.NewArrayMatcher(testUtils.CurrentTimestamp()),
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}

// TestQueryInlineArrayWithDateTime_NonNillable ensures non-nillable [DateTime!] arrays work
// for consistency with recent changes in the codebase.
func TestQueryInlineArrayWithDateTime_NonNillable(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddDoc{
				Doc: `{
                    "name": "Conference",
                    "scheduledTimes": ["2025-01-01T10:00:00Z", "2025-01-02T14:00:00Z"]
                }`,
			},
			&action.Request{
				Request: `query {
                    Events {
                        name
                        scheduledTimes
                    }
                }`,
				Results: map[string]any{
					"Events": []map[string]any{
						{
							"name": "Conference",
							"scheduledTimes": []time.Time{
								testUtils.MustParseTime("2025-01-01T10:00:00Z"),
								testUtils.MustParseTime("2025-01-02T14:00:00Z"),
							},
						},
					},
				},
			},
		},
	}

	executeTestCaseDateTime(t, test)
}
