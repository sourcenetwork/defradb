// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package net_pb

import (
	"encoding/json"

	"github.com/gogo/protobuf/proto"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-varint"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
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

// LibP2P object custom protobuf types

// ProtoPeerID is a custom type used by gogo to serde raw peer IDs into the peer.ID type, and back.
type ProtoPeerID struct {
	peer.ID
}

var _ customGogoType = (*ProtoPeerID)(nil)

func (id ProtoPeerID) Marshal() ([]byte, error) {
	return []byte(id.ID), nil
}

func (id ProtoPeerID) MarshalTo(data []byte) (n int, err error) {
	return copy(data, id.ID), nil
}

func (id ProtoPeerID) MarshalJSON() ([]byte, error) {
	m, _ := id.Marshal()
	return json.Marshal(m)
}

func (id *ProtoPeerID) Unmarshal(data []byte) (err error) {
	id.ID = peer.ID(string(data))
	return nil
}

func (id *ProtoPeerID) UnmarshalJSON(data []byte) error {
	var v []byte
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	return id.Unmarshal(v)
}

func (id ProtoPeerID) Size() int {
	return len([]byte(id.ID))
}

// ProtoAddr is a custom type used by gogo to serde raw multiaddresses into
// the ma.Multiaddr type, and back.
type ProtoAddr struct {
	ma.Multiaddr
}

var _ customGogoType = (*ProtoAddr)(nil)

func (a ProtoAddr) Marshal() ([]byte, error) {
	return a.Bytes(), nil
}

func (a ProtoAddr) MarshalTo(data []byte) (n int, err error) {
	return copy(data, a.Bytes()), nil
}

func (a ProtoAddr) MarshalJSON() ([]byte, error) {
	m, _ := a.Marshal()
	return json.Marshal(m)
}

func (a *ProtoAddr) Unmarshal(data []byte) (err error) {
	a.Multiaddr, err = ma.NewMultiaddrBytes(data)
	return err
}

func (a *ProtoAddr) UnmarshalJSON(data []byte) error {
	v := new([]byte)
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return a.Unmarshal(*v)
}

func (a ProtoAddr) Size() int {
	return len(a.Bytes())
}

// ProtoCid is a custom type used by gogo to serde raw CIDs into the cid.CID type, and back.
type ProtoCid struct {
	cid.Cid
}

var _ customGogoType = (*ProtoCid)(nil)

func (c ProtoCid) Marshal() ([]byte, error) {
	return c.Bytes(), nil
}

func (c ProtoCid) MarshalTo(data []byte) (n int, err error) {
	return copy(data, c.Bytes()), nil
}

func (c ProtoCid) MarshalJSON() ([]byte, error) {
	m, _ := c.Marshal()
	return json.Marshal(m)
}

func (c *ProtoCid) Unmarshal(data []byte) (err error) {
	c.Cid, err = cid.Cast(data)
	if errors.Is(err, varint.ErrUnderflow) {
		c.Cid = cid.Undef
		return nil
	}
	return err
}

func (c *ProtoCid) UnmarshalJSON(data []byte) error {
	v := new([]byte)
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return c.Unmarshal(*v)
}

func (c ProtoCid) Size() int {
	return len(c.Bytes())
}

// ProtoCid is a custom type used by gogo to serde raw CIDs into the cid.CID type, and back.
type ProtoDocKey struct {
	client.DocKey
}

var _ customGogoType = (*ProtoDocKey)(nil)

func (c ProtoDocKey) Marshal() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c ProtoDocKey) MarshalTo(data []byte) (n int, err error) {
	return copy(data, []byte(c.String())), nil
}

func (c ProtoDocKey) MarshalJSON() ([]byte, error) {
	m, _ := c.Marshal()
	return json.Marshal(m)
}

func (c *ProtoDocKey) Unmarshal(data []byte) (err error) {
	c.DocKey, err = client.NewDocKeyFromString(string(data))
	return err
}

func (c *ProtoDocKey) UnmarshalJSON(data []byte) error {
	v := new([]byte)
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}
	return c.Unmarshal(*v)
}

func (c ProtoDocKey) Size() int {
	return len([]byte(c.String()))
}
