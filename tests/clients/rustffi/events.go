// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

//go:build rust_ffi

package rustffi

/*
#include <stdlib.h>
#include "defra.h"
*/
import "C"

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
	"unsafe"

	"github.com/ipfs/go-cid"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/event"
)

// bridgedEvents are the events the test harness waits on.
var bridgedEvents = []event.Name{
	event.UpdateName,
	event.MergeCompleteName,
	event.ReplicatorCompletedName,
	event.SEArtifactReceivedName,
	event.TopicPeerEventName,
	event.ActionExecutionName,
}

// pollInterval is how long the bridge waits after a poll that found nothing.
const pollInterval = 5 * time.Millisecond

// eventBridge republishes the Rust node's events on a Go event bus.
type eventBridge struct {
	bus  event.Bus
	id   string
	stop chan struct{}
	done chan struct{}
}

// rustEvent is an event as a poll of defra_event_subscribe yields it.
type rustEvent struct {
	Name            string          `json:"name"`
	DocID           string          `json:"doc_id"`
	Cid             string          `json:"cid"`
	CollectionID    string          `json:"collection_id"`
	Block           string          `json:"block"`
	IsRetry         bool            `json:"is_retry"`
	IsRelay         bool            `json:"is_relay"`
	ByPeer          string          `json:"by_peer"`
	PeerID          string          `json:"peer_id"`
	Topic           string          `json:"topic"`
	EventType       string          `json:"event_type"`
	ActionExecution json.RawMessage `json:"action_execution"`
}

// newEventBridge subscribes to w's events, as its node identity.
func newEventBridge(w *Wrapper) (*eventBridge, error) {
	var args cstrings
	defer args.free()

	names := make([]string, len(bridgedEvents))
	for i, name := range bridgedEvents {
		names[i] = string(name)
	}
	owner, err := identityHandle(immutable.Some(w.nodeIdentity))
	if err != nil {
		return nil, err
	}
	defer freeIdentity(owner)

	res := take(C.defra_event_subscribe(w.node, args.str(strings.Join(names, ",")), owner))
	if err := res.err(); err != nil {
		return nil, err
	}

	b := &eventBridge{
		bus:  event.NewChannelBus(0, 100),
		id:   res.Value,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go b.run()
	return b, nil
}

// run polls the subscription until it ends or the bridge closes.
func (b *eventBridge) run() {
	defer close(b.done)
	cID := C.CString(b.id)
	defer C.free(unsafe.Pointer(cID))
	for {
		select {
		case <-b.stop:
			return
		default:
		}
		res := take(C.defra_subscription_poll(cID))
		switch res.Status {
		case 0:
			if msg, ok := goMessage(res.Value); ok {
				b.bus.Publish(msg)
			}
		case subscriptionPending:
			time.Sleep(pollInterval)
		default:
			return
		}
	}
}

// close ends the subscription and the bus.
func (b *eventBridge) close() {
	close(b.stop)
	<-b.done
	cID := C.CString(b.id)
	_ = take(C.defra_subscription_close(cID))
	C.free(unsafe.Pointer(cID))
	b.bus.Close()
}

// goMessage is the Go event a Rust event stands for.
func goMessage(raw string) (event.Message, bool) {
	var e rustEvent
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		return event.Message{}, false
	}
	switch event.Name(e.Name) {
	case event.UpdateName:
		blockCid, err := cid.Decode(e.Cid)
		if err != nil {
			return event.Message{}, false
		}
		block, err := base64.StdEncoding.DecodeString(e.Block)
		if err != nil {
			return event.Message{}, false
		}
		return event.NewMessage(event.UpdateName, event.Update{
			DocID:        e.DocID,
			Cid:          blockCid,
			CollectionID: e.CollectionID,
			Block:        block,
			IsRetry:      e.IsRetry,
			IsRelay:      e.IsRelay,
		}), true

	case event.MergeCompleteName:
		blockCid, err := cid.Decode(e.Cid)
		if err != nil {
			return event.Message{}, false
		}
		return event.NewMessage(event.MergeCompleteName, event.MergeComplete{
			Merge: event.Merge{
				DocID:        e.DocID,
				ByPeer:       e.ByPeer,
				Cid:          blockCid,
				CollectionID: e.CollectionID,
			},
		}), true

	case event.ReplicatorCompletedName:
		return event.NewMessage(event.ReplicatorCompletedName, nil), true

	case event.SEArtifactReceivedName:
		return event.NewMessage(event.SEArtifactReceivedName, event.SEArtifactReceived{
			DocID: e.DocID,
		}), true

	case event.TopicPeerEventName:
		return event.NewMessage(event.TopicPeerEventName, event.TopicPeerEvent{
			PeerID:    e.PeerID,
			Topic:     e.Topic,
			EventType: e.EventType,
		}), true

	case event.ActionExecutionName:
		var execution event.ActionExecution
		if err := json.Unmarshal(e.ActionExecution, &execution); err != nil {
			return event.Message{}, false
		}
		return event.NewMessage(event.ActionExecutionName, execution), true

	default:
		return event.Message{}, false
	}
}
