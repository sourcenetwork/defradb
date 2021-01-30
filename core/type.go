package core

// CType indicates CRDT type
// @todo: Migrate core/crdt.Type and merkle/crdt.Type to unifiied /core.CRDTType
type CType byte

const (
	//no lint
	NONE_CRDT = CType(iota) // reserved none type
	LWW_REGISTER
	OBJECT
	COMPOSITE
)

var (
	ByteToType = map[byte]CType{
		byte(0): NONE_CRDT,
		byte(1): LWW_REGISTER,
		byte(2): OBJECT,
		byte(3): COMPOSITE,
	}
)
