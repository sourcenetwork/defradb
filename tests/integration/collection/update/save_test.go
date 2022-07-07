// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package update

import (
	"context"
	"testing"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration"
	"github.com/stretchr/testify/assert"
)

func TestSaveErrorsGivenUnknownField(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
					"Name": "John",
					"Age": 21
				}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	err = doc.SetWithJSON(
		[]byte(
			`{
				"FieldDoesNotExist": 21
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.QueryTestCase{
		Description: "Simple query with no filter",
		Query: `query {
					users {
						_key
						Name
						Age
					}
				}`,
		Docs: map[int][]string{
			0: {
				(`{
				"Name": "John",
				"Age": 21
			}`)},
		},
		UpdateFuncs: map[int]func(client.Collection) error{
			0: func(c client.Collection) error {
				return c.Save(context.Background(), doc)
			},
		},
		ExpectedError: "The given field does not exist",
	}

	executeTestCase(t, test)
}
