package net_pb

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
)

// customGogoType aggregates the interfaces that custom Gogo types need to implement.
// it is only used for type assertions.
type customGogoType interface {
	proto.Marshaler
	proto.Unmarshaler
	json.Marshaler
	json.Unmarshaler
	proto.Sizer
	MarshalTo(data []byte) (n int, err error)
}
