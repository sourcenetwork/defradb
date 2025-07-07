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

type P2PDocumentKey struct {
	DocID string
}

var _ Key = (*P2PDocumentKey)(nil)

// New
func NewP2PDocumentKey(docID string) P2PDocumentKey {
	return P2PDocumentKey{DocID: docID}
}

func NewP2PDocumentKeyFromString(key string) (P2PDocumentKey, error) {
	keyArr := strings.Split(key, "/")
	if len(keyArr) != 4 {
		return P2PDocumentKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", key))
	}
	return NewP2PDocumentKey(keyArr[3]), nil
}

func (k P2PDocumentKey) ToString() string {
	result := P2P_DOCUMENT

	if k.DocID != "" {
		result = result + "/" + k.DocID
	}

	return result
}

func (k P2PDocumentKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k P2PDocumentKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
