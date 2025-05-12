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

type HeadstoreCollectionDefinition struct {
	CollectionName string
	Cid            cid.Cid
}

var _ HeadstoreKey = (*HeadstoreCollectionDefinition)(nil)

func NewHeadstoreCollectionDefinitionFromString(key string) (HeadstoreCollectionDefinition, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadstoreCollectionDefinition{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[3])
	if err != nil {
		return HeadstoreCollectionDefinition{}, err
	}

	return HeadstoreCollectionDefinition{
		// elements[0] is empty (key has leading '/')
		CollectionName: elements[2],
		Cid:            cid,
	}, nil
}

func (k HeadstoreCollectionDefinition) WithCid(c cid.Cid) HeadstoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreCollectionDefinition) GetCid() cid.Cid {
	return k.Cid
}

func (k HeadstoreCollectionDefinition) ToString() string {
	result := HEADSTORE_COL_DEF

	if k.CollectionName != "" {
		result = result + "/" + k.CollectionName
	}

	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadstoreCollectionDefinition) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadstoreCollectionDefinition) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadstoreCollectionDefinition) PrefixEnd() Walkable {
	newKey := k

	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	if k.CollectionName != "" {
		newKey.CollectionName = string(bytesPrefixEnd([]byte(k.CollectionName)))
		return newKey
	}

	return newKey
}
