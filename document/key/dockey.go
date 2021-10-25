// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package key

import (
	// "github.com/google/uuid"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/ipfs/go-cid"
	mbase "github.com/multiformats/go-multibase"
	uuid "github.com/satori/go.uuid"
	"github.com/sourcenetwork/defradb/core"
)

// Key Versions
const (
	V0 = 0x01
)

var validVersions = map[uint16]bool{
	V0: true,
}

var (
	// NamespaceSDNDocKeyV0 reserved domain namespace for Source Data Network (SDN)
	// Design a more appropriate system for future proofing doc key versions, ensuring
	// backwards compatability. RE: CID
	// *At the moment this is an random uuidV4
	NamespaceSDNDocKeyV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// VersionToNamespace is a convenience for mapping between Version number and its UUID Namespace
var VersionToNamespace = map[uint16]uuid.UUID{
	V0: NamespaceSDNDocKeyV0,
}

// DocKey is the root key identifier for documents in DefraDB
type DocKey struct {
	version uint16
	uuid    uuid.UUID
	cid     cid.Cid
	peerID  string
	Key     core.DataStoreKey
}

// Undef can be defined to be a nil like DocKey
var Undef = DocKey{}

// NewDocKeyV0 creates a new doc key identified by the root data CID, peer ID, and
// namespaced by the versionNS
// TODO: Parameterize namespace Version
func NewDocKeyV0(dataCID cid.Cid) DocKey {
	dc := DocKey{
		version: V0,
		uuid:    uuid.NewV5(NamespaceSDNDocKeyV0, dataCID.String()),
		cid:     dataCID,
	}
	dc.Key = core.DataStoreKey{DocKey: dc.String()}
	return dc
}

func NewFromString(key string) (DocKey, error) {
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

	dc := DocKey{
		version: uint16(version),
		uuid:    uuid,
	}
	dc.Key = core.DataStoreKey{DocKey: key}
	return dc, nil
}

// UUID returns the doc key in UUID form
func (key DocKey) UUID() uuid.UUID {
	return key.uuid
}

// UUID returns the doc key in string form
func (key DocKey) String() string {
	buf := make([]byte, 1)
	binary.PutUvarint(buf, uint64(key.version))
	versionStr, _ := mbase.Encode(mbase.Base32, buf)
	return versionStr + "-" + key.uuid.String()
}

// Bytes returns the DocKey in Byte format
func (key DocKey) Bytes() []byte {
	buf := make([]byte, binary.MaxVarintLen16)
	binary.PutUvarint(buf, uint64(key.version))
	return append(buf, key.uuid.Bytes()...)
}

// Verify ensures that the given DocKey is valid as per the DefraDB spec
// to prevent against collions from both honest and dishonest validators
// TODO: Check into better utilizing or dropping context, since we don't recurse
// down
func (key DocKey) Verify(ctx context.Context, data cid.Cid, peerID string) bool {
	parent, ok := ctx.Value("parent").(uuid.UUID)
	// if we have a parent then assume were operating  on a sub key
	// otherwise were the root DocKey
	var comparedUUID uuid.UUID
	if ok {
		comparedUUID = uuid.NewV5(parent, data.String())
	} else {
		comparedUUID = uuid.NewV5(NamespaceSDNDocKeyV0, data.String())
	}

	return comparedUUID.String() == key.uuid.String()
}

// Sub returns a sub DocKey, which is a UUIDv5 generated
// using the parent UUID as the namespace, and the provided
// name
func (key DocKey) Sub(subname string) DocKey {
	subParts := strings.Split(subname, "/")
	return key.subrec(subParts)
}

// recursive Sub call
// prerequisite, subparts needs to be at least 1 element long
func (key DocKey) subrec(subparts []string) DocKey {
	if len(subparts) > 1 {
		subkey := DocKey{
			uuid:   uuid.NewV5(key.uuid, subparts[0]),
			cid:    key.cid,
			peerID: key.peerID,
		}
		return subkey.subrec(subparts[1:])
	}
	// else
	return DocKey{
		uuid:   uuid.NewV5(key.uuid, subparts[0]),
		cid:    key.cid,
		peerID: key.peerID,
	}
}
