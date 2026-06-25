// Copyright 2026 Democratized Data Foundation
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
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/gofrs/uuid/v5"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/p2p/message"
)

const (
	blockSyncVersion       = "0.0.1"
	blockSyncRequestProto  = "/defradb/blocksync_req/" + blockSyncVersion
	blockSyncResponseProto = "/defradb/blocksync_car/" + blockSyncVersion

	// maxResponseHeaderSize caps the length-prefixed response header. The CAR body that follows
	// the header is streamed and intentionally uncapped.
	maxResponseHeaderSize = 1 << 20 // 1 MiB
)

var (
	// ErrResponseHeaderTooLarge is returned when a block sync response header exceeds the cap.
	ErrResponseHeaderTooLarge = errors.New("block sync response header too large")
)

// BlockSyncRequest asks a peer for the blocks needed to merge Root.
//
// When Full is false the responder may send only the diff: the blocks reachable from Root that are
// not already reachable from HaveHeads. When Full is true the responder sends the whole tree.
type BlockSyncRequest struct {
	message.MetaData
	// DocID is the document the request relates to (empty for collection-level commits).
	DocID string
	// CollectionID is the collection version id the request relates to.
	CollectionID string
	// Root is the head block CID (cid.Bytes()) the requester wants to merge.
	Root []byte
	// HaveHeads are the CIDs the requester already has, used as the diff cut-off.
	HaveHeads [][]byte
	// Full requests the whole tree, ignoring HaveHeads.
	Full bool
}

// blockSyncResponseHeader prefixes the raw CAR stream returned for a request.
//
// It is not signed: responses are correlated by MessageID and the responding peer id, and block
// integrity is guaranteed by content addressing (and, for signed docs, by signature verification
// at merge time), so a forged response cannot inject valid blocks.
type blockSyncResponseHeader struct {
	// MessageID matches the originating request's MessageID.
	MessageID string
	// EncCIDs lists the CIDs in the CAR that belong in the encryption store rather than the
	// block store.
	EncCIDs [][]byte
	// ErrMessage carries an error from the responder, if any.
	ErrMessage string
}

// RequestHandler produces the response for an incoming request on the serving node.
//
// It returns the CIDs that belong in the encryption store and a function that writes the CAR
// payload to the provided writer. Returning a nil writeCAR with a nil error means "nothing to
// send" (e.g. the requester has no access or this node lacks the blocks).
type RequestHandler func(
	ctx context.Context,
	req BlockSyncRequest,
) (encCIDs [][]byte, writeCAR func(io.Writer) error, err error)

// IngestFunc consumes the CAR payload of a response. It is called on the requesting node within
// the response stream handler, so it may stream-read an arbitrarily large CAR.
type IngestFunc func(encCIDs [][]byte, car io.Reader) error

type pendingResponse struct {
	peerID string
	ingest IngestFunc
	result chan error
}

// pendingSet is a concurrency-safe map of in-flight requests keyed by message id.
type pendingSet struct {
	mu sync.Mutex
	m  map[string]pendingResponse
}

func newPendingSet() *pendingSet {
	return &pendingSet{m: make(map[string]pendingResponse)}
}

func (p *pendingSet) set(id string, resp pendingResponse) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[id] = resp
}

func (p *pendingSet) get(id string) (pendingResponse, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	resp, ok := p.m[id]
	return resp, ok
}

func (p *pendingSet) delete(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, id)
}

// BlockSyncProtocol implements a request / streamed-CAR-response transport over libp2p streams.
//
// The requester sends a small signed request and registers a pending entry; the responder streams
// a CAR back on a second stream, which the requester correlates by MessageID and ingests in place.
type BlockSyncProtocol struct {
	*baseProto
	handler RequestHandler
	pending *pendingSet
}

// NewBlockSyncProtocol returns a new [BlockSyncProtocol] and registers its stream handlers.
//
// handler may be nil on nodes that only request blocks (it is required to serve them).
func NewBlockSyncProtocol(h client.Host, handler RequestHandler) *BlockSyncProtocol {
	proto := &BlockSyncProtocol{
		baseProto: newBaseProto(h),
		handler:   handler,
		pending:   newPendingSet(),
	}
	h.SetStreamHandler(blockSyncRequestProto, proto.onRequest)
	h.SetStreamHandler(blockSyncResponseProto, proto.onResponse)
	return proto
}

// RequestBlocks sends a block sync request to peerID and ingests the streamed CAR response.
//
// ingest is invoked with the response's encryption-store CIDs and a reader over the CAR body. The
// caller should set an appropriate context timeout.
func (p *BlockSyncProtocol) RequestBlocks(
	ctx context.Context,
	peerID string,
	req *BlockSyncRequest,
	ingest IngestFunc,
) error {
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	req.SetMessageID(id.String())

	result := make(chan error, 1)
	p.pending.set(id.String(), pendingResponse{peerID: peerID, ingest: ingest, result: result})
	defer p.pending.delete(id.String())

	if err := message.SendAndForget(ctx, p, req, peerID, blockSyncRequestProto); err != nil {
		return err
	}

	select {
	case err := <-result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// onRequest handles an incoming block sync request on the serving node.
func (p *BlockSyncProtocol) onRequest(stream io.Reader, peerID string) {
	ctx := context.Background()
	req := BlockSyncRequest{}
	if err := message.Receive(stream, peerID, p, &req); err != nil {
		return
	}

	header := blockSyncResponseHeader{MessageID: req.GetMessageID()}
	var writeCAR func(io.Writer) error
	if p.handler != nil {
		encCIDs, w, err := p.handler(ctx, req)
		if err != nil {
			header.ErrMessage = err.Error()
		} else {
			header.EncCIDs = encCIDs
			writeCAR = w
		}
	}

	var buf bytes.Buffer
	if err := writeResponseHeader(&buf, header); err != nil {
		return
	}
	if header.ErrMessage == "" && writeCAR != nil {
		if err := writeCAR(&buf); err != nil {
			return
		}
	}
	_ = p.host.Send(ctx, buf.Bytes(), peerID, blockSyncResponseProto)
}

// onResponse handles an incoming CAR response on the requesting node. It runs in its own stream
// goroutine, so it reads and ingests the (potentially large) CAR in place.
func (p *BlockSyncProtocol) onResponse(stream io.Reader, peerID string) {
	r := bufio.NewReader(stream)
	header, err := readResponseHeader(r)
	if err != nil {
		return
	}

	pend, ok := p.pending.get(header.MessageID)
	if !ok || pend.peerID != peerID {
		return
	}

	if header.ErrMessage != "" {
		pend.result <- errors.New(header.ErrMessage)
		return
	}
	pend.result <- pend.ingest(header.EncCIDs, r)
}

func writeResponseHeader(w io.Writer, header blockSyncResponseHeader) error {
	data, err := cbor.Marshal(header)
	if err != nil {
		return err
	}
	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(data)))
	if _, err := w.Write(lenBuf[:n]); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func readResponseHeader(r *bufio.Reader) (blockSyncResponseHeader, error) {
	var header blockSyncResponseHeader
	length, err := binary.ReadUvarint(r)
	if err != nil {
		return header, err
	}
	if length > maxResponseHeaderSize {
		return header, ErrResponseHeaderTooLarge
	}
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return header, err
	}
	err = cbor.Unmarshal(data, &header)
	return header, err
}
