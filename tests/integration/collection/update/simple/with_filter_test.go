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

func TestUpdateWithInvalidFilterType(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test update users with invalid filter type",
		Docs:        map[string][]string{},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Users": []func(c client.Collection) error{
				func(c client.Collection) error {
					ctx := context.Background()
					// test with an invalid filter type
					_, err := c.UpdateWithFilter(ctx, t, `{
						"name": "Eric"
					}`)
					return err
				},
			},
		},
		ExpectedError: "invalid filter",
	}

	executeTestCase(t, test)
}

func TestUpdateWithEmptyFilter(t *testing.T) {
	test := testUtils.TestCase{
		Description: "Test update users with empty filter",
		Docs:        map[string][]string{},
		CollectionCalls: map[string][]func(client.Collection) error{
			"Users": []func(c client.Collection) error{
				func(c client.Collection) error {
					ctx := context.Background()
					// test with an empty filter
					_, err := c.UpdateWithFilter(ctx, "", `{
						"name": "Eric"
					}`)
					return err
				},
			},
		},
		ExpectedError: "invalid filter",
	}

	executeTestCase(t, test)
}

func TestUpdateWithFilter(t *testing.T) {
	docStr := `{
		"name": "John",
		"age": 21
	}`

	doc, err := client.NewDocFromJSON([]byte(docStr))
	if err != nil {
		assert.Fail(t, err.Error())
	}

	filter := `{name: {_eq: "John"}}`

	tests := []testUtils.TestCase{
		{
			Description: "Test update users with filter and invalid JSON",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithFilter(ctx, filter, `{
							name: "Eric"
						}`)
						return err
					},
				},
			},
			ExpectedError: "cannot parse JSON: cannot parse object",
		}, {
			Description: "Test update users with filter and invalid updator",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithFilter(ctx, filter, `"name: Eric"`)
						return err
					},
				},
			},
			ExpectedError: "the updater of a document is of invalid type",
		}, {
			Description: "Test update users with filter and patch updator (not implemented so no change)",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithFilter(ctx, filter, `[
							{
								"name": "Eric"
							}, {
								"name": "Sam"
							}
						]`)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc.Key(), false)
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
			Description: "Test update users with filter",
			Docs: map[string][]string{
				"Users": {docStr},
			},
			CollectionCalls: map[string][]func(client.Collection) error{
				"Users": []func(c client.Collection) error{
					func(c client.Collection) error {
						ctx := context.Background()
						_, err := c.UpdateWithFilter(ctx, filter, `{
							"name": "Eric"
						}`)
						if err != nil {
							return err
						}

						d, err := c.Get(ctx, doc.Key(), false)
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
