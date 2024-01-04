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

// DocID versions.
const (
	DocIDV0 = 0x01
)

// ValidDocIDVersions is a map of DocID versions and their current validity.
var ValidDocIDVersions = map[uint16]bool{
	DocIDV0: true,
}

var (
	// SDNNamespaceV0 is a reserved domain namespace for Source Data Network (SDN).
	SDNNamespaceV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// DocID is the root identifier for documents in DefraDB.
type DocID struct {
	version uint16
	uuid    uuid.UUID
	cid     cid.Cid
}

// NewDocIDV0 creates a new DocID identified by the root data CID, peerID, and namespaced by the versionNS.
func NewDocIDV0(dataCID cid.Cid) DocID {
	return DocID{
		version: DocIDV0,
		uuid:    uuid.NewV5(SDNNamespaceV0, dataCID.String()),
		cid:     dataCID,
	}
}

// NewDocIDFromString creates a new DocID from a string.
func NewDocIDFromString(docID string) (DocID, error) {
	parts := strings.SplitN(docID, "-", 2)
	if len(parts) != 2 {
		return DocID{}, ErrMalformedDocID
	}
	versionStr := parts[0]
	_, data, err := mbase.Decode(versionStr)
	if err != nil {
		return DocID{}, err
	}
	buf := bytes.NewBuffer(data)
	version, err := binary.ReadUvarint(buf)
	if err != nil {
		return DocID{}, err
	}
	if _, ok := ValidDocIDVersions[uint16(version)]; !ok {
		return DocID{}, ErrInvalidDocIDVersion
	}

	uuid, err := uuid.FromString(parts[1])
	if err != nil {
		return DocID{}, err
	}

	return DocID{
		version: uint16(version),
		uuid:    uuid,
	}, nil
}

// UUID returns the underlying document identifier in UUID form.
func (docID DocID) UUID() uuid.UUID {
	return docID.uuid
}

// String returns the underlying document identifier in string form.
func (docID DocID) String() string {
	buf := make([]byte, 1)
	binary.PutUvarint(buf, uint64(docID.version))
	versionStr, _ := mbase.Encode(mbase.Base32, buf)
	return versionStr + "-" + docID.uuid.String()
}

// Bytes returns the underlying document identifier in Byte format.
func (docID DocID) Bytes() []byte {
	buf := make([]byte, binary.MaxVarintLen16)
	binary.PutUvarint(buf, uint64(docID.version))
	return append(buf, docID.uuid.Bytes()...)
}
