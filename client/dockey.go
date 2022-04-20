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
	// "github.com/google/uuid"
	"bytes"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
	uuid "github.com/satori/go.uuid"
)

// Key Versions
const (
	v0 = 0x01
)

var validVersions = map[uint16]bool{
	v0: true,
}

var (
	// NamespaceSDNDocKeyV0 reserved domain namespace for Source Data Network (SDN)
	// Design a more appropriate system for future proofing doc key versions, ensuring
	// backwards compatability. RE: CID
	// *At the moment this is an random uuidV4
	namespaceSDNDocKeyV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// versionToNamespace is a convenience for mapping between Version number and its UUID Namespace
// nolint
var versionToNamespace = map[uint16]uuid.UUID{
	v0: namespaceSDNDocKeyV0,
}

// DocKey is the root key identifier for documents in DefraDB
type DocKey struct {
	version uint16
	uuid    uuid.UUID
	cid     cid.Cid
}

// NewDocKeyV0 creates a new doc key identified by the root data CID, peer ID, and
// namespaced by the versionNS
// TODO: Parameterize namespace Version
func NewDocKeyV0(dataCID cid.Cid) DocKey {
	return DocKey{
		version: v0,
		uuid:    uuid.NewV5(namespaceSDNDocKeyV0, dataCID.String()),
		cid:     dataCID,
	}
}

func NewDocKeyFromString(key string) (DocKey, error) {
	parts := strings.SplitN(key, "-", 2)
	if len(parts) != 2 {
		return DocKey{}, errors.New("Malformed DocKey, missing either version or cid")
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
	if _, ok := validVersions[uint16(version)]; !ok {
		return DocKey{}, errors.New("Invalid DocKey version")
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

// UUID returns the doc key in string form
func (key DocKey) String() string {
	buf := make([]byte, 1)
	binary.PutUvarint(buf, uint64(key.version))
	versionStr, _ := mbase.Encode(mbase.Base32, buf)
	return versionStr + "-" + key.uuid.String()
}
