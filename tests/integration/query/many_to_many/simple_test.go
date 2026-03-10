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

package many_to_many

import (
	"testing"

	"github.com/sourcenetwork/defradb/tests/action"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
)

func TestManyToMany_QueryFromJoinCollection_ShouldSucceed(t *testing.T) {
	test := testUtils.TestCase{
		Actions: []any{
			&action.AddCollection{
				SDL: `
                    type Student {
                        name: String
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

			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Alice"}`,
			},
			&action.AddDoc{
				CollectionID: 0,
				Doc:          `{"name": "Bob"}`,
			},

			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "Math"}`,
			},
			&action.AddDoc{
				CollectionID: 1,
				Doc:          `{"name": "Science"}`,
			},

			&action.AddDoc{
				CollectionID: 2, // Enrollment
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 0), // Alice
					"course":  testUtils.NewDocIndex(1, 0), // Math
				},
			},
			&action.AddDoc{
				CollectionID: 2, // Enrollment
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 0), // Alice
					"course":  testUtils.NewDocIndex(1, 1), // Science
				},
			},
			&action.AddDoc{
				CollectionID: 2, // Enrollment
				DocMap: map[string]any{
					"student": testUtils.NewDocIndex(0, 1), // Bob
					"course":  testUtils.NewDocIndex(1, 0), // Math
				},
			},

			// Query course-to-students direction
			&action.Request{
				Request: `query {
					Enrollment(
						filter: {course: {name: {_eq: "Math"}}}
						order: {student: {name: ASC}}
					) {
						student { name }
					}
				}`,
				Results: map[string]any{
					"Enrollment": []map[string]any{
						{"student": map[string]any{"name": "Alice"}},
						{"student": map[string]any{"name": "Bob"}},
					},
				},
			},

			// Query student-to-courses direction
			&action.Request{
				Request: `query {
					Enrollment(
						filter: {student: {name: {_eq: "Alice"}}}
						order: {course: {name: ASC}}
					) {
						course { name }
					}
				}`,
				Results: map[string]any{
					"Enrollment": []map[string]any{
						{"course": map[string]any{"name": "Math"}},
						{"course": map[string]any{"name": "Science"}},
					},
				},
			},
		},
	}

	testUtils.ExecuteTestCase(t, test)
}
