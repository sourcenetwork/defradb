// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package planner

import (
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/cursor"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

// cursorNode implements cursor-based pagination by skipping documents
// until the cursor position is passed, then collecting up to `first` results.
type cursorNode struct {
	docMapper

	p    *Planner
	plan planNode

	// Forward cursor parameters (from mapper.Select)
	first       immutable.Option[uint64]
	afterCursor immutable.Option[string]

	// Pre-decoded forward cursor payload
	afterPayload *cursor.CursorPayload

	// Backward cursor parameters (from mapper.Select)
	last         immutable.Option[uint64]
	beforeCursor immutable.Option[string]

	// Pre-decoded backward cursor payload
	beforePayload *cursor.CursorPayload

	// Backward re-reversal buffer: backward iteration yields results in reverse order,
	// so we buffer them and reverse before yielding to maintain forward result order.
	backwardBuffer []core.Doc
	bufferIndex    int

	// Internal state machine
	pastCursor      bool   // true once cursor position has been passed
	collected       uint64 // count of results yielded so far
	indexSeekActive bool

	orderFields []mapper.OrderCondition

	// _pageInfo state (populated during iteration)
	hasNextPage     bool
	hasPreviousPage bool
	firstDocID      string
	lastDocID       string
	firstDoc        core.Doc
	lastDoc         core.Doc
	pageInfoSelect  *request.PageInfoSelect

	execInfo cursorExecInfo
}

type cursorExecInfo struct {
	// Total number of times cursorNode was executed.
	iterations uint64
}

// Cursor creates a new cursorNode from the parsed select if it is a cursor query.
// The after cursor token is decoded here (once per query) rather than on the
// hot path in Next().
func (p *Planner) Cursor(parsed *mapper.Select) (*cursorNode, error) {
	if !parsed.IsCursor {
		return nil, nil
	}

	var afterPayload *cursor.CursorPayload
	if parsed.CursorAfter.HasValue() {
		payload, err := cursor.Decode(parsed.CursorAfter.Value())
		if err != nil {
			return nil, err
		}
		afterPayload = &payload
	}

	var beforePayload *cursor.CursorPayload
	if parsed.CursorBefore.HasValue() {
		payload, err := cursor.Decode(parsed.CursorBefore.Value())
		if err != nil {
			return nil, err
		}
		beforePayload = &payload
	}

	return &cursorNode{
		p:              p,
		first:          parsed.CursorFirst,
		afterCursor:    parsed.CursorAfter,
		afterPayload:   afterPayload,
		last:           parsed.CursorLast,
		beforeCursor:   parsed.CursorBefore,
		beforePayload:  beforePayload,
		pageInfoSelect: parsed.CursorPageInfo,
		docMapper:      docMapper{parsed.DocumentMapping},
	}, nil
}

func (n *cursorNode) Kind() string {
	return "cursorNode"
}

func (n *cursorNode) Init() error {
	n.pastCursor = false
	n.collected = 0
	n.hasNextPage = false
	n.hasPreviousPage = false
	n.firstDocID = ""
	n.lastDocID = ""
	n.firstDoc = core.Doc{}
	n.lastDoc = core.Doc{}
	n.backwardBuffer = nil
	n.bufferIndex = 0
	return n.plan.Init()
}

func (n *cursorNode) Start() error                      { return n.plan.Start() }
func (n *cursorNode) Prefixes(prefixes []keys.Walkable) { n.plan.Prefixes(prefixes) }
func (n *cursorNode) Close() error                      { return n.plan.Close() }
func (n *cursorNode) Value() core.Doc {
	if n.backwardBuffer != nil {
		return n.backwardBuffer[n.bufferIndex]
	}
	return n.plan.Value()
}
func (n *cursorNode) Source() planNode                  { return n.plan }

func (n *cursorNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.last.HasValue() || n.beforeCursor.HasValue() {
		return n.nextBackward()
	}

	if !n.afterCursor.HasValue() || n.indexSeekActive {
		n.pastCursor = true
	}
	if n.afterCursor.HasValue() {
		n.hasPreviousPage = true
	}

	for {
		// Check if we've already collected enough
		if n.first.HasValue() && n.collected >= n.first.Value() {
			// Probe one more to determine hasNextPage
			hasMore, err := n.plan.Next()
			if err != nil {
				return false, err
			}
			n.hasNextPage = hasMore
			return false, nil
		}

		hasNext, err := n.plan.Next()
		if !hasNext {
			return false, err
		}

		// Skip phase: looking for cursor position
		if !n.pastCursor {
			doc := n.plan.Value()
			if doc.GetID() == n.afterPayload.DocID {
				n.pastCursor = true
				continue // skip the cursor document itself
			}
			continue // keep skipping
		}

		// Collect phase
		n.collected++
		doc := n.plan.Value()
		docID := doc.GetID()
		if n.collected == 1 {
			n.firstDocID = docID
			n.firstDoc = doc.Clone()
		}
		n.lastDocID = docID
		n.lastDoc = doc.Clone()
		return true, nil
	}
}

// nextBackward implements backward pagination by draining the reversed scan into a buffer,
// applying the `last` limit, then yielding results by iterating backward through the buffer
// to restore forward order.
func (n *cursorNode) nextBackward() (bool, error) {
	if n.backwardBuffer != nil {
		n.bufferIndex--
		if n.bufferIndex < 0 {
			return false, nil
		}
		n.collected++
		return true, nil
	}

	var buf []core.Doc
	for {
		hasNext, err := n.plan.Next()
		if err != nil {
			return false, err
		}
		if !hasNext {
			break
		}
		doc := n.plan.Value()
		buf = append(buf, doc.Clone())
	}

	if n.last.HasValue() && len(buf) > int(n.last.Value()) {
		n.hasPreviousPage = true
		buf = buf[len(buf)-int(n.last.Value()):]
	}

	if n.beforeCursor.HasValue() {
		n.hasNextPage = true
	}

	// buf is in reversed scan order; we iterate backward through it
	// to yield results in forward order without an O(n) reverse.
	if len(buf) > 0 {
		last := len(buf) - 1
		n.firstDocID = buf[last].GetID()
		n.firstDoc = buf[last].Clone()
		n.lastDocID = buf[0].GetID()
		n.lastDoc = buf[0].Clone()
	}

	n.backwardBuffer = buf
	n.bufferIndex = len(buf) - 1

	if len(buf) == 0 {
		return false, nil
	}

	n.collected = 1
	return true, nil
}

// buildEnrichedPayload creates a CursorPayload with DocID and index key values from order fields.
func (n *cursorNode) buildEnrichedPayload(docID string, doc core.Doc) cursor.CursorPayload {
	payload := cursor.CursorPayload{DocID: docID}
	if len(n.orderFields) == 0 {
		return payload
	}
	payload.Keys = make(map[string]any, len(n.orderFields))
	for _, oc := range n.orderFields {
		if len(oc.FieldIndexes) == 0 {
			continue
		}
		fieldIdx := oc.FieldIndexes[0]
		name, ok := n.documentMapping.TryToFindNameFromIndex(fieldIdx)
		if !ok {
			continue
		}
		if fieldIdx < len(doc.Fields) {
			payload.Keys[name] = doc.Fields[fieldIdx]
		}
	}
	return payload
}

// PageInfo returns the _pageInfo metadata for this cursor query page.
// Call after iteration is complete (Next() returned false).
func (n *cursorNode) PageInfo() (map[string]any, error) {
	info := map[string]any{}

	sel := n.pageInfoSelect
	if sel == nil {
		return info, nil
	}

	if sel.HasNext {
		info[request.HasNextFieldName] = n.hasNextPage
	}
	if sel.HasPrev {
		info[request.HasPrevFieldName] = n.hasPreviousPage
	}
	if sel.StartCursor {
		if n.firstDocID != "" {
			encoded, err := cursor.Encode(n.buildEnrichedPayload(n.firstDocID, n.firstDoc))
			if err != nil {
				return nil, err
			}
			info["startCursor"] = encoded
		} else {
			info["startCursor"] = nil
		}
	}
	if sel.EndCursor {
		if n.lastDocID != "" {
			encoded, err := cursor.Encode(n.buildEnrichedPayload(n.lastDocID, n.lastDoc))
			if err != nil {
				return nil, err
			}
			info["endCursor"] = encoded
		} else {
			info["endCursor"] = nil
		}
	}

	return info, nil
}

func (n *cursorNode) simpleExplain() (map[string]any, error) {
	m := map[string]any{}
	if n.first.HasValue() {
		m["first"] = n.first.Value()
	} else {
		m["first"] = nil
	}
	if n.afterCursor.HasValue() {
		m["after"] = n.afterCursor.Value()
	} else {
		m["after"] = nil
	}
	if n.last.HasValue() {
		m["last"] = n.last.Value()
	} else {
		m["last"] = nil
	}
	if n.beforeCursor.HasValue() {
		m["before"] = n.beforeCursor.Value()
	} else {
		m["before"] = nil
	}

	payload := n.afterPayload
	if payload == nil {
		payload = n.beforePayload
	}
	if payload != nil {
		cursorInfo := map[string]any{
			"docID": payload.DocID,
		}
		if len(payload.Keys) > 0 {
			cursorInfo["keys"] = payload.Keys
		}
		m["cursorValue"] = cursorInfo
	} else {
		m["cursorValue"] = nil
	}

	return m, nil
}

func (n *cursorNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return map[string]any{
			"iterations": n.execInfo.iterations,
		}, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}
