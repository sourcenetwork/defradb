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
	"strconv"
	"strings"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

// HeadstoreColKey are used to store the current collection head in the headstore.
type HeadstoreColKey struct {
	// CollectionRoot is the root of the collection that this head refers to.
	//
	// Including it in the key allows easier identification of a given collection's
	// head.
	CollectionRoot uint32

	// Cid is the cid of this head block.
	Cid cid.Cid
}

var _ HeadstoreKey = (*HeadstoreColKey)(nil)

func NewHeadstoreColKey(colRoot uint32) HeadstoreColKey {
	return HeadstoreColKey{
		CollectionRoot: colRoot,
	}
}

func NewHeadstoreColKeyFromString(key string) (HeadstoreColKey, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadstoreColKey{}, ErrInvalidKey
	}

	root, err := strconv.Atoi(elements[2])
	if err != nil {
		return HeadstoreColKey{}, err
	}

	cid, err := cid.Decode(elements[3])
	if err != nil {
		return HeadstoreColKey{}, err
	}

	return HeadstoreColKey{
		// elements[0] is empty (key has leading '/')
		CollectionRoot: uint32(root),
		Cid:            cid,
	}, nil
}

func (k HeadstoreColKey) WithCid(c cid.Cid) HeadstoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreColKey) GetCid() cid.Cid {
	return k.Cid
}

func (k HeadstoreColKey) ToString() string {
	result := HEADSTORE_COL

	if k.CollectionRoot != 0 {
		result = result + "/" + strconv.Itoa(int(k.CollectionRoot))
	}
	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadstoreColKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadstoreColKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadstoreColKey) PrefixEnd() Walkable {
	newKey := k

	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	if k.CollectionRoot != 0 {
		newKey.CollectionRoot = k.CollectionRoot + 1
		return newKey
	}

	return newKey
}
