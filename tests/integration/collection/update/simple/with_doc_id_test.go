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

func TestUpdateWithDocID(t *testing.T) {
	docStr := `{
		"name": "John",
		"age": 21
	}`

	doc, err := client.NewDocFromJSON([]byte(docStr), colDefMap["Users"].Schema)
	if err != nil {
		assert.Fail(t, err.Error())
	}

	tests := []testUtils.TestCase{
		{
			Description: "Test update users with docID and invalid JSON",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithDocID(
							ctx,
							doc.ID(),
							`{name: "Eric"}`,
						)
						return err
					},
				},
			},
			ExpectedError: "cannot parse JSON: cannot parse object",
		}, {
			Description: "Test update users with docID and invalid updator",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithDocID(ctx, doc.ID(), `"name: Eric"`)
						return err
					},
				},
			},
			ExpectedError: "the updater of a document is of invalid type",
		}, {
			Description: "Test update users with docID and patch updator (not implemented so no change)",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithDocID(
							ctx,

							doc.ID(),
							`[
								{
									"name": "Eric"
								}, {
									"name": "Sam"
								}
							]`,
						)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc.ID(), false)
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
		}, {
			Description: "Test update users with docID",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithDocID(
							ctx,

							doc.ID(),
							`{"name": "Eric"}`,
						)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc.ID(), false)
						if err != nil {
							return err
						}

						name, err := d.Get("name")
						if err != nil {
							return err
						}

						assert.Equal(t, "Eric", name)

						return nil
					},
				},
			},
		},
	}

	for _, test := range tests {
		executeTestCase(t, test)
	}
}
