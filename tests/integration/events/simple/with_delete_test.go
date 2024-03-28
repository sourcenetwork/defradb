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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/events"
)

func TestEventsSimpleWithDelete(t *testing.T) {
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

	test := testUtils.TestCase{
		CollectionCalls: map[string][]func(client.Collection){
			"Users": []func(c client.Collection){
				func(c client.Collection) {
					err = c.Save(context.Background(), acpIdentity.NoIdentity, doc1)
					assert.Nil(t, err)
				},
				func(c client.Collection) {
					wasDeleted, err := c.Delete(context.Background(), acpIdentity.NoIdentity, doc1.ID())
					assert.Nil(t, err)
					assert.True(t, wasDeleted)
				},
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocID: immutable.Some(docID1),
			},
			{
				DocID: immutable.Some(docID1),
			},
		},
	}

	executeTestCase(t, test)
}
