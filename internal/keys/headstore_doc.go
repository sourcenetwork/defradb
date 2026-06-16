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
	"bytes"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
)

type HeadstoreDocKey struct {
	CollectionShortID uint32
	DocShortID        uint32
	FieldID           string //can be 'C'
	Cid               cid.Cid
}

var _ HeadstoreKey = (*HeadstoreDocKey)(nil)

// Creates a new HeadstoreDocKey from a string as best as it can,
// splitting the input using '/' as a field deliminator.  It assumes
// that the input string is in the following format:
//
// /d/[CollectionShortID]/[DocShortID]/[FieldId]/[Cid]
//
// Any properties before the above are ignored
func NewHeadstoreDocKey(key string) (HeadstoreDocKey, error) {
	elements := bytes.Split([]byte(key), []byte{'/'})
	if len(elements) != 6 {
		return HeadstoreDocKey{}, ErrInvalidKey
	}

	collectionShortID, err := DecodeCollectionShortID(elements[2])
	if err != nil {
		return HeadstoreDocKey{}, err
	}

	docShortID, err := DecodeDocShortID(elements[3])
	if err != nil {
		return HeadstoreDocKey{}, err
	}

	cid, err := cid.Decode(string(elements[5]))
	if err != nil {
		return HeadstoreDocKey{}, err
	}

	return HeadstoreDocKey{
		// elements[0] is empty (key has leading '/')
		CollectionShortID: collectionShortID,
		DocShortID:        docShortID,
		FieldID:           string(elements[4]),
		Cid:               cid,
	}, nil
}

// WithCollectionShortID returns a copy of the key scoped to the given collection.
func (k HeadstoreDocKey) WithCollectionShortID(collectionShortID uint32) HeadstoreDocKey {
	newKey := k
	newKey.CollectionShortID = collectionShortID
	return newKey
}

func (k HeadstoreDocKey) WithDocShortID(docShortID uint32) HeadstoreDocKey {
	newKey := k
	newKey.DocShortID = docShortID
	return newKey
}

func (k HeadstoreDocKey) WithCid(c cid.Cid) HeadstoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreDocKey) GetCid() cid.Cid {
	return k.Cid
}

func (k HeadstoreDocKey) WithFieldID(fieldID string) HeadstoreDocKey {
	newKey := k
	newKey.FieldID = fieldID
	return newKey
}

func (k HeadstoreDocKey) ToString() string {
	return string(k.Bytes())
}

func (k HeadstoreDocKey) Bytes() []byte {
	result := []byte(HEADSTORE_DOC)

	if k.CollectionShortID != 0 {
		result = append(result, '/')
		result = append(result, EncodeCollectionShortID(k.CollectionShortID)...)
	}
	if k.DocShortID != 0 {
		result = append(result, '/')
		result = append(result, EncodeDocShortID(k.DocShortID)...)
	}
	if k.FieldID != "" {
		result = append(result, '/')
		result = append(result, []byte(k.FieldID)...)
	}
	if k.Cid.Defined() {
		result = append(result, '/')
		result = append(result, []byte(k.Cid.String())...)
	}

	return result
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
	if k.DocShortID != 0 {
		newKey.DocShortID++
		return newKey
	}
	if k.CollectionShortID != 0 {
		newKey.CollectionShortID++
		return newKey
	}
	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	return newKey
}
