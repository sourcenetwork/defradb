// Copyright 2026 Democratized Data Foundation
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
	"encoding/base64"
	"strings"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"

	"github.com/sourcenetwork/defradb/errors"
)

// BlockSignatureKey indexes signature blocks by the canonical block CID and signer identity.
type BlockSignatureKey struct {
	BlockCID string
	Identity string
}

var _ Key = (*BlockSignatureKey)(nil)

func NewBlockSignatureKey(blockCID cid.Cid, identity string) BlockSignatureKey {
	return BlockSignatureKey{
		BlockCID: blockCID.String(),
		Identity: identity,
	}
}

func NewBlockSignaturePrefix(blockCID cid.Cid) BlockSignatureKey {
	return BlockSignatureKey{
		BlockCID: blockCID.String(),
	}
}

func NewBlockSignatureKeyFromString(keyString string) (BlockSignatureKey, error) {
	keyString = strings.TrimPrefix(keyString, BLOCK_SIGNATURE+"/")
	elements := strings.Split(keyString, "/")
	if len(elements) != 2 {
		return BlockSignatureKey{}, errors.WithStack(ErrInvalidKey, errors.NewKV("Key", keyString))
	}

	identity, err := base64.RawURLEncoding.DecodeString(elements[1])
	if err != nil {
		return BlockSignatureKey{}, err
	}

	return BlockSignatureKey{
		BlockCID: elements[0],
		Identity: string(identity),
	}, nil
}

func (k BlockSignatureKey) ToString() string {
	result := BLOCK_SIGNATURE

	if k.BlockCID != "" {
		result = result + "/" + k.BlockCID
	}

	if k.Identity != "" {
		result = result + "/" + base64.RawURLEncoding.EncodeToString([]byte(k.Identity))
	}

	return result
}

func (k BlockSignatureKey) Bytes() []byte {
	return []byte(k.ToString())
}

func (k BlockSignatureKey) ToDS() ds.Key {
	return ds.NewKey(k.ToString())
}
