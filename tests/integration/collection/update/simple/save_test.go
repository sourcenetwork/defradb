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

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/collection"
)

func TestUpdateSaveErrorsGivenUnknownField(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "John",
				"age": 21
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

	test := testUtils.TestCase{
		Docs: map[string][]string{
			"Users": {
				`{
					"name": "John",
					"age": 21
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Users": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
		ExpectedError: "The given field does not exist",
	}

	executeTestCase(t, test)
}
