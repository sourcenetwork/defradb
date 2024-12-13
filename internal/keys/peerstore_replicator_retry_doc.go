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

type ReplicatorRetryDocIDKey struct {
	PeerID string
	DocID  string
}

var _ Key = (*ReplicatorRetryDocIDKey)(nil)

func NewReplicatorRetryDocIDKey(peerID, docID string) ReplicatorRetryDocIDKey {
	return ReplicatorRetryDocIDKey{
		PeerID: peerID,
		DocID:  docID,
	}
}

// NewReplicatorRetryDocIDKeyFromString creates a new [ReplicatorRetryDocIDKey] from a string.
//
// It expects the input string to be in the format `/rep/retry/doc/[PeerID]/[DocID]`.
func NewReplicatorRetryDocIDKeyFromString(key string) (ReplicatorRetryDocIDKey, error) {
	trimmedKey := strings.TrimPrefix(key, REPLICATOR_RETRY_DOC+"/")
	keyArr := strings.Split(trimmedKey, "/")
	if len(keyArr) != 2 {
		return ReplicatorRetryDocIDKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewReplicatorRetryDocIDKey(keyArr[0], keyArr[1]), nil
}

func (k ReplicatorRetryDocIDKey) ToString() string {
	keyString := REPLICATOR_RETRY_DOC + "/" + k.PeerID
	if k.DocID != "" {
		keyString += "/" + k.DocID
	}
	return keyString
}

func (k ReplicatorRetryDocIDKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k ReplicatorRetryDocIDKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
