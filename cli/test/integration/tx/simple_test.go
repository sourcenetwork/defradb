// Copyright 2026 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tx

import (
	"testing"
	"time"

	"github.com/sourcenetwork/defradb/cli/test/action"
	"github.com/sourcenetwork/defradb/cli/test/integration"
)

func TestCreateTxWithRawTTL(t *testing.T) {
	test := &integration.Test{
		Timeout: time.Second * 5, // test wide timeout
		Actions: []action.Action{
			&action.CreateTx{
				TTL: "1", // 1 second
			},
			&action.Wait{
				Duration: time.Second * 2,
			},
			action.WithTxnI(
				&action.AddCollection{
					InlineSDL: `
						type User {
							name: String
						}
					`,
					ExpectError: "transaction not found",
				},
				0,
			),
		},
	}

	test.Execute(t)
}

func TestCreateTxWithFormattedTTL(t *testing.T) {
	test := &integration.Test{
		Timeout: time.Second * 5, // test wide timeout
		Actions: []action.Action{
			&action.CreateTx{
				TTL: "1s", // 1 second
			},
			&action.Wait{
				Duration: time.Second * 2,
			},
			action.WithTxnI(
				&action.AddCollection{
					InlineSDL: `
						type User {
							name: String
						}
					`,
					ExpectError: "transaction not found",
				},
				0,
			),
		},
	}

	test.Execute(t)
}
