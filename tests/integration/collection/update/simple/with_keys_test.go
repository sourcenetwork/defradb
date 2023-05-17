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

func TestUpdateWithKeys(t *testing.T) {
	docStr1 := `{
		"name": "John",
		"age": 21
	}`

	doc1, err := client.NewDocFromJSON([]byte(docStr1))
	if err != nil {
		assert.Fail(t, err.Error())
	}

	docStr2 := `{
		"name": "Sam",
		"age": 32
	}`

	doc2, err := client.NewDocFromJSON([]byte(docStr2))
	if err != nil {
		assert.Fail(t, err.Error())
	}

	tests := []testUtils.TestCase{
		{
			Description: "Test update users with keys and invalid JSON",
			Docs: map[string][]string{
				"Users": {
					docStr1,
					docStr2,
				},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `{
							name: "Eric"
						}`)
						return err
					},
				},
			},
			ExpectedError: "cannot parse JSON: cannot parse object",
		}, {
			Description: "Test update users with keys and invalid updator",
			Docs: map[string][]string{
				"Users": {
					docStr1,
					docStr2,
				},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `"name: Eric"`)
						return err
					},
				},
			},
			ExpectedError: "the updater of a document is of invalid type",
		}, {
			Description: "Test update users with keys and patch updator (not implemented so no change)",
			Docs: map[string][]string{
				"Users": {
					docStr1,
					docStr2,
				},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `[
							{
								"name": "Eric"
							}, {
								"name": "Bob"
							}
						]`)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc1.Key(), false)
						if err != nil {
							return err
						}

						name, err := d.Get("name")
						if err != nil {
							return err
						}

						assert.Equal(t, "John", name)

						d2, err := c.Get(ctx, doc2.Key(), false)
						if err != nil {
							return err
						}

						name2, err := d2.Get("name")
						if err != nil {
							return err
						}

						assert.Equal(t, "Sam", name2)

						return nil
					},
				},
			},
		}, {
			Description: "Test update users with keys",
			Docs: map[string][]string{
				"Users": {
					docStr1,
					docStr2,
				},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithKeys(ctx, []client.DocKey{doc1.Key(), doc2.Key()}, `{
							"age": 40
						}`)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc1.Key(), false)
						if err != nil {
							return err
						}

						name, err := d.Get("age")
						if err != nil {
							return err
						}

						assert.Equal(t, uint64(40), name)

						d2, err := c.Get(ctx, doc2.Key(), false)
						if err != nil {
							return err
						}

						name2, err := d2.Get("age")
						if err != nil {
							return err
						}

						assert.Equal(t, uint64(40), name2)

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
