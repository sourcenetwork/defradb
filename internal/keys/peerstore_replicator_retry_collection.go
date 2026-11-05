// Copyright 2026 Democratized Data Foundation
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

type ReplicatorRetryCollectionIDKey struct {
	PeerID       string
	CollectionID string
}

var _ Key = (*ReplicatorRetryCollectionIDKey)(nil)

func NewReplicatorRetryCollectionIDKey(peerID string, collectionID string) ReplicatorRetryCollectionIDKey {
	return ReplicatorRetryCollectionIDKey{
		PeerID:       peerID,
		CollectionID: collectionID,
	}
}

// NewReplicatorRetryCollectionIDKeyFromString creates a new [ReplicatorRetryCollectionIDKey] from a string.
//
// It expects the input string to be in the format `/rep/retry/col/[PeerID]/[ColID]`.
func NewReplicatorRetryCollectionIDKeyFromString(key string) (ReplicatorRetryCollectionIDKey, error) {
	trimmedKey := strings.TrimPrefix(key, REPLICATOR_RETRY_COL+"/")
	keyArr := strings.Split(trimmedKey, "/")
	if len(keyArr) != 2 {
		return ReplicatorRetryCollectionIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewReplicatorRetryCollectionIDKey(keyArr[0], keyArr[1]), nil
}

func (k ReplicatorRetryCollectionIDKey) ToString() string {
	keyString := REPLICATOR_RETRY_COL + "/" + k.PeerID
	if k.CollectionID != "" {
		keyString += "/" + k.CollectionID
	}
	return keyString
}

func (k ReplicatorRetryCollectionIDKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ReplicatorRetryCollectionIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
