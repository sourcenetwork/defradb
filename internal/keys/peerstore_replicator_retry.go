// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package keys

import (
	"strings"

	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/errors"
)

type ReplicatorRetryIDKey struct {
	PeerID string
}

var _ Key = (*ReplicatorRetryIDKey)(nil)

func NewReplicatorRetryIDKey(peerID string) ReplicatorRetryIDKey {
	return ReplicatorRetryIDKey{
		PeerID: peerID,
	}
}

// NewReplicatorRetryIDKeyFromString creates a new [ReplicatorRetryIDKey] from a string.
//
// It expects the input string to be in the format `/rep/retry/id/[PeerID]`.
func NewReplicatorRetryIDKeyFromString(key string) (ReplicatorRetryIDKey, error) {
	peerID := strings.TrimPrefix(key, REPLICATOR_RETRY_ID+"/")
	if peerID == "" {
		return ReplicatorRetryIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewReplicatorRetryIDKey(peerID), nil
}

func (k ReplicatorRetryIDKey) ToString() string {
	return REPLICATOR_RETRY_ID + "/" + k.PeerID
}

func (k ReplicatorRetryIDKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ReplicatorRetryIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
