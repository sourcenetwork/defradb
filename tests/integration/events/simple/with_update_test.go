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
				"Name": "John"
			}`,
		),
	)
	assert.Nil(t, err)
	doc2, err := client.NewDocFromJSON(
		[]byte(
			`{
				"Name": "Shahzad"
			}`,
		),
	)
	assert.Nil(t, err)

	test := testUtils.TestCase{
		CollectionCalls: map[string][]func(client.Collection){
			"users": []func(c client.Collection){
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
					doc1.Set("Name", "Johnnnnn")
					err = c.Save(context.Background(), doc1)
					assert.Nil(t, err)
				},
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocKey: immutable.Some("bae-43deba43-f2bc-59f4-9056-fef661b22832"),
				Cid:    immutable.Some("bafybeihm4gewwhzsyqs7tuazdsbusp6pm3oinhnj2ae6funcwshwpxuxri"),
			},
			{
				DocKey: immutable.Some("bae-359dfe55-77a4-59de-a687-fc5049dc61fe"),
			},
			{
				DocKey: immutable.Some("bae-43deba43-f2bc-59f4-9056-fef661b22832"),
				Cid:    immutable.Some("bafybeicxo5sqjzi2mjputtbzwnkliickvowwfigyzziibtwlj7xe3ca7ly"),
			},
		},
	}

	executeTestCase(t, test)
}
