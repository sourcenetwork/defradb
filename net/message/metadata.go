// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package message

// MetaData is the information that should be part of every [Message]
type MetaData struct {
	Version    string
	MessageID  string
	PeerID     string
	Pubkey     []byte
	Signature  []byte `cbor:",omitempty"`
	ErrMessage string `cbor:",omitempty"`
}

var _ Message = (*MetaData)(nil)

func (m *MetaData) GetVersion() string {
	return m.Version
}

func (m *MetaData) SetVersion() {
	m.Version = messageVersion
}

func (m *MetaData) GetMessageID() string {
	return m.MessageID
}

func (m *MetaData) SetMessageID(id string) {
	m.MessageID = id
}

func (m *MetaData) GetPeerID() string {
	return m.PeerID
}

func (m *MetaData) SetPeerID(id string) {
	m.PeerID = id
}

func (m *MetaData) GetPubkey() []byte {
	return m.Pubkey
}

func (m *MetaData) SetPubkey(key []byte) {
	m.Pubkey = key
}

func (m *MetaData) GetSignature() []byte {
	return m.Signature
}

func (m *MetaData) SetSignature(signature []byte) {
	m.Signature = signature
}

func (m *MetaData) GetErrMessage() string {
	return m.ErrMessage
}

func (m *MetaData) SetErrMessage(err string) {
	m.ErrMessage = err
}
