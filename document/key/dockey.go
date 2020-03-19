package key

import (
	// "github.com/google/uuid"
	"context"
	"encoding/binary"
	"strings"

	"github.com/ipfs/go-cid"
	ds "github.com/ipfs/go-datastore"
	uuid "github.com/satori/go.uuid"
)

// Key Versions
const (
	V0 = 0x01
)

var (
	// NamespaceSDNDocKeyV0 reserved domain namespace for Source Data Network (SDN)
	// Design a more appropriate system for future proofing doc key versions, ensuring
	// backwards compatability. RE: CID
	// *At the moment this is an random uuidV4
	NamespaceSDNDocKeyV0 = uuid.Must(uuid.FromString("c94acbfa-dd53-40d0-97f3-29ce16c333fc"))
)

// VersionToNamespace is a convenience for mapping between Version number and its UUID Namespace
var VersionToNamespace = map[uint64]uuid.UUID{
	V0: NamespaceSDNDocKeyV0,
}

// DocKey is the root key identifier for documents in DefraDB
type DocKey struct {
	version uint64
	uuid    uuid.UUID
	cid     cid.Cid
	peerID  string
	ds.Key
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
	dc.Key = ds.NewKey(dc.String())
	return dc
}

// UUID returns the doc key in UUID form
func (key DocKey) UUID() uuid.UUID {
	return key.uuid
}

// UUID returns the doc key in string form
func (key DocKey) String() string {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, key.version)
	return string(buf) + "-" + key.uuid.String()
}

// Bytes returns the DocKey in Byte format
func (key DocKey) Bytes() []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutUvarint(buf, key.version)
	return append(buf, key.uuid.Bytes()...)
}

// Verify ensures that the given DocKey is valid as per the DefraDB spec
// to prevent against collions from both honest and dishonest validators
// TODO: Check into better utilizing or dropping context, since we dont recurse
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
