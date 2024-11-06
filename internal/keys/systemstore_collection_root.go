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
	"fmt"
	"strconv"
	"strings"

	ds "github.com/ipfs/go-datastore"
)

// CollectionRootKey points to nil, but the keys/prefix can be used
// to get collections that are of a given RootID.
//
// It is stored in the format `/collection/root/[RootID]/[CollectionID]`.
type CollectionRootKey struct {
	RootID       uint32
	CollectionID uint32
}

var _ Key = (*CollectionRootKey)(nil)

func NewCollectionRootKey(rootID uint32, collectionID uint32) CollectionRootKey {
	return CollectionRootKey{
		RootID:       rootID,
		CollectionID: collectionID,
	}
}

// NewCollectionRootKeyFromString creates a new [CollectionRootKey].
//
// It expects the key to be in the format `/collection/root/[RootID]/[CollectionID]`.
func NewCollectionRootKeyFromString(key string) (CollectionRootKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 5 || keyArr[1] != COLLECTION || keyArr[2] != "root" {
		return CollectionRootKey{}, ErrInvalidKey
	}
	rootID, err := strconv.Atoi(keyArr[3])
	if err != nil {
		return CollectionRootKey{}, err
	}

	collectionID, err := strconv.Atoi(keyArr[4])
	if err != nil {
		return CollectionRootKey{}, err
	}

	return CollectionRootKey{
		RootID:       uint32(rootID),
		CollectionID: uint32(collectionID),
	}, nil
}

func (k CollectionRootKey) ToString() string {
	result := COLLECTION_ROOT

	if k.RootID != 0 {
		result = fmt.Sprintf("%s/%s", result, strconv.Itoa(int(k.RootID)))
	}

	if k.CollectionID != 0 {
		result = fmt.Sprintf("%s/%s", result, strconv.Itoa(int(k.CollectionID)))
	}

	return result
}

func (k CollectionRootKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k CollectionRootKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
