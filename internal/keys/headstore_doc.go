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

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

type HeadStoreKey struct {
	DocID   string
	FieldID string //can be 'C'
	Cid     cid.Cid
}

var _ Key = (*HeadStoreKey)(nil)

// Creates a new HeadStoreKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /[DocID]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadStoreKey(key string) (HeadStoreKey, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 4 {
		return HeadStoreKey{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[3])
	if err != nil {
		return HeadStoreKey{}, err
	}

	return HeadStoreKey{
		// elements[0] is empty (key has leading '/')
		DocID:   elements[1],
		FieldID: elements[2],
		Cid:     cid,
	}, nil
}

func (k HeadStoreKey) WithDocID(docID string) HeadStoreKey {
	newKey := k
	newKey.DocID = docID
	return newKey
}

func (k HeadStoreKey) WithCid(c cid.Cid) HeadStoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadStoreKey) WithFieldID(fieldID string) HeadStoreKey {
	newKey := k
	newKey.FieldID = fieldID
	return newKey
}

func (k HeadStoreKey) ToString() string {
	var result string

	if k.DocID != "" {
		result = result + "/" + k.DocID
	}
	if k.FieldID != "" {
		result = result + "/" + k.FieldID
	}
	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadStoreKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadStoreKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
