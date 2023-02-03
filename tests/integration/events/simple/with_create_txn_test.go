// Copyright 2023 Democratized Data Foundation
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

func TestEventsSimpleWithCreateWithTxnDiscarded(t *testing.T) {
	test := testUtils.TestCase{
		DatabaseCalls: []func(context.Context, client.DB){
			func(ctx context.Context, d client.DB) {
				r := d.ExecRequest(
					ctx,
					`mutation {
						create_users(data: "{\"Name\": \"John\"}") {
							_key
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
				r := d.ExecTransactionalRequest(
					ctx,
					`mutation {
						create_users(data: "{\"Name\": \"Shahzad\"}") {
							_key
						}
					}`,
					txn,
				)
				for _, err := range r.GQL.Errors {
					assert.Nil(t, err)
				}
				txn.Discard(ctx)
			},
		},
		ExpectedUpdates: []testUtils.ExpectedUpdate{
			{
				DocKey: immutable.Some("bae-43deba43-f2bc-59f4-9056-fef661b22832"),
			},
			// No event should be recieved for Shahzad, as the transaction was discarded.
		},
	}

	executeTestCase(t, test)
}
