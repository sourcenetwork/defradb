// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"github.com/sourcenetwork/lens/host-go/config/model"
)

// LensConfig represents the configuration of a Lens migration in Defra.
type LensConfig struct {
	// SourceCollectionVersionID is the ID of the collection version from which to migrate
	// from.
	//
	// The source and destination versions must be next to each other in the history.
	SourceCollectionVersionID string

	// DestinationCollectionVersionID is the ID of the collection version from which to migrate
	// to.
	//
	// The source and destination versions must be next to each other in the history.
	DestinationCollectionVersionID string

	// The configuration of the Lens module.
	//
	// For now, the wasm module must remain at the location specified as long as the
	// migration is active.
	model.Lens
}

// TxnSource represents an object capable of constructing the transactions that
// implicit-transaction registries need internally.
type TxnSource interface {
	NewTxn(bool) (Txn, error)
}
