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

func TestEventsSimpleWithUpdate(t *testing.T) {
	doc1, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "John"
			}`,
		),
		colDefMap["Users"].Schema,
	)
	assert.Nil(t, err)
	docID1 := doc1.ID().String()

	doc2, err := client.NewDocFromJSON(
		[]byte(
			`{
				"name": "Shahzad"
			}`,
		),
		colDefMap["Users"].Schema,
	)
	assert.Nil(t, err)
	docID2 := doc2.ID().String()

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
				func(c client.Collection) {
					// Update John
					doc1.Set("name", "Johnnnnn")
					err = c.Save(context.Background(), doc1)
					assert.Nil(t, err)
				},
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocID: immutable.Some(docID1),
				Cid:   immutable.Some("bafybeiffwlvaz742mg2ejv7v2qd62couktzrejfkzzedzw3djmxpvt5lvi"),
			},
			{
				DocID: immutable.Some(docID2),
			},
			{
				DocID: immutable.Some(docID1),
				Cid:   immutable.Some("bafybeidzdbok2dutsjobk62jxgoqttk2omwazvh4ycvrz27ffg7dp4joo4"),
			},
		},
	}

	executeTestCase(t, test)
}
