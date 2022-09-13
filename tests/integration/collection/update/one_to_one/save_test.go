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
	testUtils "github.com/sourcenetwork/defradb/tests/integration/collection"
	"github.com/stretchr/testify/assert"
)

func TestUpdateOneToOneSaveErrorsGivenWrongSideOfRelation(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "Painted House"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	err = doc.SetWithJSON(
		[]byte(
			`{
				"author_id": "ValueDoesntMatter"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
		Docs: map[string][]string{
			"book": {
				`{
					"name": "Painted House"
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"book": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
		ExpectedError: "The given field does not exist",
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist.
func TestUpdateOneToOneSavesGivenNewRelationValue(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "John Grisham"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	err = doc.SetWithJSON(
		[]byte(
			`{
				"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
		Docs: map[string][]string{
			"author": {
				`{
					"name": "John Grisham"
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"author": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
	}

	executeTestCase(t, test)
}
