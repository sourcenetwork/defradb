// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package clients

import (
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/event"
)

// Client implements the DB interface along with a few other methods
// required for testing.
type Client interface {
	client.TxnStore
	Close()
	MaxTxnRetries() int
	Events() event.Bus
}
