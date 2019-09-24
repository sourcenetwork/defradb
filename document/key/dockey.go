package key

import (
	// "github.com/google/uuid"
	"github.com/ipfs/go-cid"
	uuid "github.com/satori/go.uuid"
)

var (
	// Design a more appropriate system for
	// future proofing doc key versions, ensuring
	// backwards compatability. RE: CID
	// At the moment this is an random uuidV4
	NamespaceSDNDocKeyV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// DocKey is the root key identifier for documents in DefraDB
type DocKey struct {
	uuid   uuid.UUID
	cid    cid.Cid
	peerID string
}

// Creates a new doc key identified by the root data CID, peer ID, and
// namespaced by the versionNS
func NewDocKey(versionNS uuid.UUID, dataCID cid.Cid, peerID string) DocKey {
	return DocKey{
		uuid:   uuid.NewV5(versionNS, dataCID.String()),
		cid:    dataCID,
		peerID: peerID,
	}
}

// UUID returns the doc key in UUID form
func (key DocKey) UUID() uuid.UUID {
	return key.uuid
}

// UUID returns the doc key in string form
func (key DocKey) String() string {
	return key.uuid.String()
}

// Sub returns a sub DocKey, which is a UUIDv5 generated
// using the parent UUID as the namespace, and the provided
// name
func (key DocKey) Sub(subname string) DocKey {
	return DocKey{
		uuid:   uuid.NewV5(key.uuid, subname),
		cid:    key.cid,
		peerID: key.peerID,
	}
}
