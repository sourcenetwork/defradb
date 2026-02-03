// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package kms

import (
	"context"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corelog"

	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/encryption"
)

var (
	log = corelog.NewLogger("kms")
)

type ServiceType string

const (
	// PubSubServiceType is the type of KMS that uses PubSub mechanism to exchange keys
	// between peers.
	PubSubServiceType ServiceType = "pubsub"
)

// Service is interface for key management service (KMS)
type Service interface {
	// GetKeys retrieves the encryption blocks containing encryption keys for the given links.
	// Blocks are fetched asynchronously, so the method returns an [encryption.Results] object
	// that can be used to wait for the results.
	GetKeys(ctx context.Context, cids ...cidlink.Link) (*encryption.Results, error)

	// TryGetLocalKey attempts to retrieve an encryption block from local storage.
	// If hint is provided, it verifies that the local key matches the hint before returning.
	// Returns the encryption block if found locally and hint matches (if provided), nil otherwise.
	// This is a fast local-only operation that doesn't make network requests.
	TryGetLocalKey(ctx context.Context, encryptionCID cidlink.Link, hint []byte) (*coreblock.Encryption, error)
}
