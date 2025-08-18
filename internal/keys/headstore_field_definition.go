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

type HeadstoreFieldDefinition struct {
	CollectionName string
	FieldName      string
	Cid            cid.Cid
}

var _ HeadstoreKey = (*HeadstoreFieldDefinition)(nil)

func NewHeadstoreFieldDefinitionFromString(key string) (HeadstoreFieldDefinition, error) {
	elements := strings.Split(key, "/")
	if len(elements) != 5 {
		return HeadstoreFieldDefinition{}, ErrInvalidKey
	}

	cid, err := cid.Decode(elements[4])
	if err != nil {
		return HeadstoreFieldDefinition{}, err
	}

	return HeadstoreFieldDefinition{
		// elements[0] is empty (key has leading '/')
		CollectionName: elements[2],
		FieldName:      elements[3],
		Cid:            cid,
	}, nil
}

func (k HeadstoreFieldDefinition) WithCid(c cid.Cid) HeadstoreKey {
	newKey := k
	newKey.Cid = c
	return newKey
}

func (k HeadstoreFieldDefinition) GetCid() cid.Cid {
	return k.Cid
}

func (k HeadstoreFieldDefinition) ToString() string {
	result := HEADSTORE_FIELD_DEF

	if k.CollectionName != "" {
		result = result + "/" + k.CollectionName
	}

	if k.FieldName != "" {
		result = result + "/" + k.FieldName
	}

	if k.Cid.Defined() {
		result = result + "/" + k.Cid.String()
	}

	return result
}

func (k HeadstoreFieldDefinition) Bytes() []byte {
	return []byte(k.ToString())
}

func (k HeadstoreFieldDefinition) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}

func (k HeadstoreFieldDefinition) PrefixEnd() Walkable {
	newKey := k

	if k.Cid.Defined() {
		newKey.Cid = cid.MustParse(bytesPrefixEnd(k.Cid.Bytes()))
		return newKey
	}

	if k.FieldName != "" {
		newKey.FieldName = string(bytesPrefixEnd([]byte(k.FieldName)))
		return newKey
	}

	if k.CollectionName != "" {
		newKey.CollectionName = string(bytesPrefixEnd([]byte(k.CollectionName)))
		return newKey
	}

	return newKey
}
