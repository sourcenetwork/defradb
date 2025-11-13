// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"testing"

	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestQuerySimpleWithLimit0(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
					"Name": "John",
					"Age": 21
				}`,
			},
			testUtils.CreateDoc{
				Doc: `{
					"Name": "Bob",
					"Age": 32
				}`,
			},
			testUtils.Request{
				Request: `query {
					Users(limit: 0) {
						Name
					}
				}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
						},
						{
							"Name": "John",
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 1) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit2(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Carlo",
						"Age": 55
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Alice",
						"Age": 19
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 2) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimitBiggerThanTotalDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 3) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithOffset0(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(offset: 0) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithOffset1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(offset: 1) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithOffset2(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Carlo",
						"Age": 55
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Alice",
						"Age": 19
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Melynda",
						"Age": 30
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(offset: 2) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithOffsetBiggerThanTotalDocuments(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(offset: 3) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit0AndOffset0(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 0, offset: 0) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
						{
							"Name": "John",
							"Age":  int64(21),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit1AndOffset1(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 1, offset: 1) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Bob",
							"Age":  int64(32),
						},
					},
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestQuerySimpleWithLimit2AndOffset2(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			testUtils.CreateDoc{
				Doc: `{
						"Name": "John",
						"Age": 21
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Bob",
						"Age": 32
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Carlo",
						"Age": 55
					}`,
			},
			testUtils.CreateDoc{
				Doc: `{
						"Name": "Alice",
						"Age": 19
					}`,
			},
			testUtils.Request{
				Request: `query {
						Users(limit: 2, offset: 2) {
							Name
							Age
						}
					}`,
				Results: map[string]any{
					"Users": []map[string]any{
						{
							"Name": "Carlo",
							"Age":  int64(55),
						},
						{
							"Name": "Alice",
							"Age":  int64(19),
						},
					},
				},
				NonOrderedResults: true,
			},
		},
	}

	executeTestCase(t, test)
}
