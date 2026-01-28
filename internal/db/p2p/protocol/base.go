// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package protocol

import (
	"sync"

	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

var log = corelog.NewLogger("p2p")

// baseProto contains the minimum fields that protocols should contain.
type baseProto struct {
	host          client.Host
	mu            sync.Mutex
	responseChans map[string]chan message.Message
}

func newBaseProto(h client.Host) *baseProto {
	return &baseProto{
		host:          h,
		responseChans: make(map[string]chan message.Message),
	}
}

func (proto *baseProto) Host() client.Host {
	return proto.host
}

func (proto *baseProto) SetResponseChan(messageID string, message chan message.Message) {
	proto.mu.Lock()
	defer proto.mu.Unlock()
	proto.responseChans[messageID] = message
}

func (proto *baseProto) GetResponseChan(messageID string) (chan message.Message, bool) {
	proto.mu.Lock()
	defer proto.mu.Unlock()
	m, ok := proto.responseChans[messageID]
	return m, ok
}

func (proto *baseProto) DeleteResponseChan(messageID string) {
	proto.mu.Lock()
	defer proto.mu.Unlock()
	delete(proto.responseChans, messageID)
}
