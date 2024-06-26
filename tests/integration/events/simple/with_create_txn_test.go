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
	"github.com/sourcenetwork/defradb/internal/db"
	testUtils "github.com/sourcenetwork/defradb/tests/integration/events"
)

func TestEventsSimpleWithCreateWithTxnDiscarded(t *testing.T) {
	test := testUtils.TestCase{
		DatabaseCalls: []func(context.Context, client.DB){
			func(ctx context.Context, d client.DB) {
				r := d.ExecRequest(
					ctx,
					`mutation {
						create_Users(input: {name: "John"}) {
							_docID
						}
					}`,
				)
				for _, err := range r.GQL.Errors {
					assert.Nil(t, err)
				}
			},
			func(ctx context.Context, d client.DB) {
				txn, err := d.NewTxn(ctx, false)
				assert.Nil(t, err)

				ctx = db.SetContextTxn(ctx, txn)
				r := d.ExecRequest(
					ctx,
					`mutation {
						create_Users(input: {name: "Shahzad"}) {
							_docID
						}
					}`,
				)
				for _, err := range r.GQL.Errors {
					assert.Nil(t, err)
				}
				txn.Discard(ctx)
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocID: immutable.Some("bae-6845cfdf-cb0f-56a3-be3a-b5a67be5fbdc"),
			},
			// No event should be received for Shahzad, as the transaction was discarded.
		},
	}

	executeTestCase(t, test)
}
