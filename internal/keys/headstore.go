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
)

const (
	HEADSTORE_DOC = "/d"
	HEADSTORE_COL = "/c"
)

// HeadstoreKey represents any key that may be stored in the headstore.
type HeadstoreKey interface {
	Walkable

	// GetCid returns the cid that forms part of this key.
	GetCid() cid.Cid

	// WithCid returns a new HeadstoreKey with the same values as the original,
	// apart from the cid which will have been replaced by the given value.
	WithCid(c cid.Cid) HeadstoreKey
}

// NewHeadstoreKey returns the typed representation of the given key string, or
// an [ErrInvalidKey] error if it's type could not be determined.
func NewHeadstoreKey(key string) (HeadstoreKey, error) {
	switch {
	case strings.HasPrefix(key, HEADSTORE_DOC):
		return NewHeadstoreDocKey(key)
	case strings.HasPrefix(key, HEADSTORE_COL):
		return NewHeadstoreColKeyFromString(key)
	default:
		return nil, ErrInvalidKey
	}
}
