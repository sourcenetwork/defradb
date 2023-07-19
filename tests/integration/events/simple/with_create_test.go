// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package simple

import (
	"context"
	"testing"

	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"

	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/events"
)

func TestEventsSimpleWithCreate(t *testing.T) {
	doc1, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "John"
			}`,
		),
	)
	assert.Nil(t, err)
	docKey1 := doc1.Key().String()

	doc2, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "Shahzad"
			}`,
		),
	)
	assert.Nil(t, err)
	docKey2 := doc2.Key().String()

	test := testUtils.TestCase{
		CollectionCalls: map[string][]func(client.Collection){
			"Users": []func(c client.Collection){
				func(c client.Collection) {
					err = c.Save(context.Background(), doc1)
					assert.Nil(t, err)
				},
				func(c client.Collection) {
					err = c.Save(context.Background(), doc2)
					assert.Nil(t, err)
				},
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocKey: immutable.Some(docKey1),
			},
			{
				DocKey: immutable.Some(docKey2),
			},
		},
	}

	executeTestCase(t, test)
}
