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

type HeadstoreDocKey struct {
	DocID   string
	FieldID string //can be 'C'
	Cid     cid.Cid
}

var _ Walkable = (*HeadstoreDocKey)(nil)

// Creates a new HeadstoreDocKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /d/[DocID]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadstoreDocKey(key string) (HeadstoreDocKey, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 5 {
		return HeadstoreDocKey{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[4])
	if err != nil {
		return HeadstoreDocKey{}, err
	}

	return HeadstoreDocKey{
		// elements[0] is empty (key has leading '/')
		DocID:   elements[2],
		FieldID: elements[3],
		Cid:     cid,
	}, nil
}

func (k HeadstoreDocKey) WithDocID(docID string) HeadstoreDocKey {
	newKey := k
	newKey.DocID = docID
	return newKey
}

func (k HeadstoreDocKey) WithCid(c cid.Cid) HeadstoreDocKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreDocKey) WithFieldID(fieldID string) HeadstoreDocKey {
	newKey := k
	newKey.FieldID = fieldID
	return newKey
}

func (k HeadstoreDocKey) ToString() string {
	result := HEADSTORE_DOC

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

func (k HeadstoreDocKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadstoreDocKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadstoreDocKey) PrefixEnd() Walkable {
	newKey := k

	if k.FieldID != "" {
		newKey.FieldID = string(bytesPrefixEnd([]byte(k.FieldID)))
		return newKey
	}
	if k.DocID != "" {
		newKey.DocID = string(bytesPrefixEnd([]byte(k.DocID)))
		return newKey
	}
	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	return newKey
}
