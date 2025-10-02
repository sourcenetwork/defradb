// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package many_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestManyToMany_QueryFromSecondary_Succeeds(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddSchema{
				Schema: `
                    type Student {
                        name: String
                        enrollments: [Enrollment] @relation(name: "student_enrollments")
                    }

                    type Course {
                        name: String
                    }

                    type Enrollment {
                        student: Student @relation(name: "student_enrollments")
                        course: Course @relation(name: "course_enrollments")
                    }
                `,
			},

			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},

			testUtils.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "Math"}`,
			},
			testUtils.CreateDoc{
				CollectionID: 1,
				Doc:          `{"name": "Science"}`,
			},

			testUtils.CreateDoc{
				CollectionID: 2, // Enrollment
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 0), // Alice
					"course":  testUtils.NewDocIndex(1, 0), // Math
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2, // Enrollment
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 0), // Alice
					"course":  testUtils.NewDocIndex(1, 1), // Science
				},
			},
			testUtils.CreateDoc{
				CollectionID: 2,
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 1), // Bob
					"course":  testUtils.NewDocIndex(1, 0), // Math
				},
			},

			// Query Alice and access her course names
			testUtils.Request{
				Request: `query {
					Student(filter: {name: {_eq: "Alice"}}) {
						name
						enrollments(order: {course: {name: ASC}}) {
							course { name }
						}
					}
				}`,
				Results: map[string]any{
					"Student": []map[string]any{
						{
							"name": "Alice",
							"enrollments": []map[string]any{
								{"course": map[string]any{"name": "Math"}},
								{"course": map[string]any{"name": "Science"}},
							},
						},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
