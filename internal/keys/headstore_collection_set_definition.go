// Copyright 2025 Democratized Data Foundation
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

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

// HeadstoreCollectionSetDefinition is the headstore key for CollectionSet blocks.
type HeadstoreCollectionSetDefinition struct {
	// The ID of the collection with the lexographically smallest name in the collection set
	// at time of creation.
	//
	// This is totally arbitrary and is used to create a deterministic and unique prefix for
	// the set's cids.
	FirstCollectionID string

	// The CID of the collection set.
	Cid cid.Cid
}

var _ HeadstoreKey = (*HeadstoreCollectionSetDefinition)(nil)

func NewHeadstoreCollectionSetDefinitionFromString(key string) (HeadstoreCollectionSetDefinition, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadstoreCollectionSetDefinition{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[3])
	if err != nil {
		return HeadstoreCollectionSetDefinition{}, err
	}

	return HeadstoreCollectionSetDefinition{
		// elements[0] is empty (key has leading '/')
		FirstCollectionID: elements[2],
		Cid:               cid,
	}, nil
}

func (k HeadstoreCollectionSetDefinition) WithCid(c cid.Cid) HeadstoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreCollectionSetDefinition) GetCid() cid.Cid {
	return k.Cid
}

func (k HeadstoreCollectionSetDefinition) ToString() string {
	result := HEADSTORE_COL_SET_DEF

	if k.FirstCollectionID != "" {
		result = result + "/" + k.FirstCollectionID
	}

	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadstoreCollectionSetDefinition) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadstoreCollectionSetDefinition) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadstoreCollectionSetDefinition) PrefixEnd() Walkable {
	newKey := k

	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	if k.FirstCollectionID != "" {
		newKey.FirstCollectionID = string(bytesPrefixEnd([]byte(k.FirstCollectionID)))
		return newKey
	}

	return newKey
}
