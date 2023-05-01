// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package client

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/gofrs/uuid/v5"
	"github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
)

// DocKey versions.
const (
	DocKeyV0 = 0x01
)

// ValidDocKeyVersions is a map of DocKey versions and their current validity.
var ValidDocKeyVersions = map[uint16]bool{
	DocKeyV0: true,
}

var (
	// SDNNamespaceV0 is a reserved domain namespace for Source Data Network (SDN).
	SDNNamespaceV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// DocKey is the root key identifier for documents in DefraDB.
type DocKey struct {
	version uint16
	uuid    uuid.UUID
	cid     cid.Cid
}

// NewDocKeyV0 creates a new dockey identified by the root data CID,peerID, and namespaced by the versionNS.
func NewDocKeyV0(dataCID cid.Cid) DocKey {
	return DocKey{
		version: DocKeyV0,
		uuid:    uuid.NewV5(SDNNamespaceV0, dataCID.String()),
		cid:     dataCID,
	}
}

// NewDocKeyFromString creates a new DocKey from a string.
func NewDocKeyFromString(key string) (DocKey, error) {
	parts := strings.SplitN(key, "-", 2)
	if len(parts) != 2 {
		return DocKey{}, ErrMalformedDocKey
	}
	versionStr := parts[0]
	_, data, err := mbase.Decode(versionStr)
	if err != nil {
		return DocKey{}, err
	}
	buf := bytes.NewBuffer(data)
	version, err := binary.ReadUvarint(buf)
	if err != nil {
		return DocKey{}, err
	}
	if _, ok := ValidDocKeyVersions[uint16(version)]; !ok {
		return DocKey{}, ErrInvalidDocKeyVersion
	}

	uuid, err := uuid.FromString(parts[1])
	if err != nil {
		return DocKey{}, err
	}

	return DocKey{
		version: uint16(version),
		uuid:    uuid,
	}, nil
}

// UUID returns the doc key in UUID form.
func (key DocKey) UUID() uuid.UUID {
	return key.uuid
}

// String returns the doc key in string form.
func (key DocKey) String() string {
	buf := make([]byte, 1)
	binary.PutUvarint(buf, uint64(key.version))
	versionStr, _ := mbase.Encode(mbase.Base32, buf)
	return versionStr + "-" + key.uuid.String()
}

// Bytes returns the DocKey in Byte format.
func (key DocKey) Bytes() []byte {
	buf := make([]byte, binary.MaxVarintLen16)
	binary.PutUvarint(buf, uint64(key.version))
	return append(buf, key.uuid.Bytes()...)
}
