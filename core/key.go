package core

import (
	"strconv"

	ds "github.com/ipfs/go-datastore"
	"github.com/pkg/errors"
)

var (
	// KeyMin is a minimum key value which sorts before all other keys.
	KeyMin = NewKey("")
	// KeyMax is a maximum key value which sorts after all other keys.
	KeyMax = NewKeyFromBytes([]byte{0xff, 0xff})
)

type Key struct {
	ds.Key
}

// NewKey creates a new Key from a string
func NewKey(s string) Key {
	return Key{ds.NewKey(s)}
}

// NewKeyFromBytes creates a new Key from a byte array
func NewKeyFromBytes(b []byte) Key {
	return Key{ds.NewKey(string(b))}
}

// PrefixEnd determines the end key given key as a prefix, that is the
// key that sorts precisely behind all keys starting with prefix: "1"
// is added to the final byte and the carry propagated. The special
// cases of nil and KeyMin always returns KeyMax.
func (k Key) PrefixEnd() Key {
	if len(k.Bytes()) == 0 {
		return Key(KeyMax)
	}
	return NewKeyFromBytes(bytesPrefixEnd(k.Bytes()))
}

// FieldID extracts the Field Identifier from the Key.
// In a Primary index, the last key path is the FieldID.
// This may be different in Secondary Indexes.
// An error is returned if it can't correct convert the
// field to a uint32.
func (k Key) FieldID() (uint32, error) {
	// fmt.Println(k.String())
	fieldIDStr := k.Type()
	fieldID, err := strconv.Atoi(fieldIDStr)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to get FieldID of Key")
	}
	return uint32(fieldID), nil
}

func bytesPrefixEnd(b []byte) []byte {
	end := make([]byte, len(b))
	copy(end, b)
	for i := len(end) - 1; i >= 0; i-- {
		end[i] = end[i] + 1
		if end[i] != 0 {
			return end[:i+1]
		}
	}
	// This statement will only be reached if the key is already a
	// maximal byte string (i.e. already \xff...).
	return b
}
