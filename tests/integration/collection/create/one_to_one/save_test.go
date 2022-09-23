// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package create

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/collection"
)

func TestCollectionCreateSaveErrorsGivenRelationSetOnWrongSide(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "Painted House",
				"author_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
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

func TestCollectionCreateSaveCreatesDoc(t *testing.T) {
	doc, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "John",
				"published_id": "bae-fd541c25-229e-5280-b44b-e5c2af3e374d"
			}`,
		),
	)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	test := testUtils.TestCase{
		CollectionCalls: map[string][]func(client.Collection) error{
			"author": []func(c client.Collection) error{
				func(c client.Collection) error {
					err := c.Save(context.Background(), doc)
					if err != nil {
						return err
					}

					d, err := c.Get(context.Background(), doc.Key())
					if err != nil {
						return err
					}

					name, err := d.Get("name")
					if err != nil {
						return err
					}

					assert.Equal(t, "John", name)

					return nil
				},
			},
		},
	}

	executeTestCase(t, test)
}
