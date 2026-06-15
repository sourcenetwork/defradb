// Copyright 2026 Democratized Data Foundation
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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

var dateTimeArraySchema = (`
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
            &action.AddSchema{
                Schema: dateTimeArraySchema,
            },
        },
        test.Actions...,
    )

    testUtils.ExecuteTestCase(t, fullTest)
}

func TestQueryInlineArrayWithDateTime_Null(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.CreateDoc{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
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
			&action.CreateDoc{
				Doc: `{
					"name": "Mixed Event",
					"optionalTimes": ["2024-03-15T10:00:00Z", null, "2024-03-16T10:00:00Z"]
				}`,
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				Doc: `{
					"name": "Morning Event",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-15T10:00:00Z"]
				}`,
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				Doc: `{
					"name": "Waitlisted",
					"scheduledTimes": ["2024-01-15T09:00:00Z", "2024-01-16T09:00:00Z"]
				}`,
			},
			&action.CreateDoc{
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
			&action.Request{ // _ge matches both docs since all their times are >= the target
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
			&action.CreateDoc{
				Doc: `{
					"name": "Future Event",
					"scheduledTimes": ["2024-02-01T00:00:00Z"]
				}`,
			},
			&action.CreateDoc{
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
			&action.AddSchema{
				Schema: `
					type Events {
						name: String
						scheduledTimes: [DateTime!] @index
					}
				`,
			},
			&action.CreateDoc{
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
			&action.CreateDoc{
				Doc: `{
					"name": "Rescheduled Event",
					"scheduledTimes": ["2024-01-01T00:00:00Z"]
				}`,
			},
			testUtils.UpdateDoc{
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
        Description: "Error on malformed DateTime string in array",
        Actions: []any{
            &action.CreateDoc{
                Doc: `{
                    "name": "Bad Date",
                    "scheduledTimes": ["not-a-date"]
                }`,
                ExpectedError: "cannot parse \"not-a-date\"",
            },
        },
    }

    executeTestCaseDateTime(t, test)
}

// TestQueryInlineArrayWithDateTime_UTC_NOW tests using UTC_NOW inside a DateTime array
// as requested in review by @ChrisBQu.
func TestQueryInlineArrayWithDateTime_UTC_NOW(t *testing.T) {
    test := testUtils.TestCase{
        Description: "DateTime array with UTC_NOW value",
        Actions: []any{
            &action.AddSchema{
                Schema: `type Events { name: String; times: [DateTime!] }`,
            },
            &action.CreateDoc{
                Doc: `{
                    "name": "Now Event",
                    "times": ["UTC_NOW"]
                }`,
            },
            &action.Request{
                Request: `query { Events { name times } }`,
                Results: map[string]any{
                    "Events": []map[string]any{
                        {
                            "name": "Now Event",
                            "times": []any{testUtils.CurrentTimestamp()},
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
        Description: "Non-nillable DateTime array ([DateTime!])",
        Actions: []any{
            &action.AddSchema{
                Schema: `type Events { name: String; times: [DateTime!] }`,
            },
            &action.CreateDoc{
                Doc: `{
                    "name": "Conference",
                    "times": ["2025-01-01T10:00:00Z", "2025-01-02T14:00:00Z"]
                }`,
            },
            &action.Request{
                Request: `query {
                    Events {
                        name
                        times
                    }
                }`,
                Results: map[string]any{
                    "Events": []map[string]any{
                        {
                            "name": "Conference",
                            "times": []time.Time{
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
