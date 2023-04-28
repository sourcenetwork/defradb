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

func TestUpdateOneToOneSaveErrorsGivenNonExistantKeyViaSecondarySide(t *testing.T) {
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
				"author_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
		Docs: map[string][]string{
			"Book": {
				`{
					"name": "Painted House"
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Book": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
		ExpectedError: "no document for the given key exists",
	}

	executeTestCase(t, test)
}

// Note: This test should probably not pass, as it contains a
// reference to a document that doesnt exist. It is doubly odd
// given that saving from the secondary side errors as expected
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
			"Author": {
				`{
					"name": "John Grisham"
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Author": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
	}

	executeTestCase(t, test)
}

func TestUpdateOneToOneSaveFromSecondarySide(t *testing.T) {
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
				"author_id": "bae-2edb7fdd-cad7-5ad4-9c7d-6920245a96ed"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
		Docs: map[string][]string{
			"Author": {
				`{
					"name": "John Grisham"
				}`,
			},
			"Book": {
				`{
					"name": "Painted House"
				}`,
			},
		},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Book": []func(c client.Collection) error{
				func(c client.Collection) error {
					return c.Save(context.Background(), doc)
				},
			},
		},
	}

	executeTestCase(t, test)
}
