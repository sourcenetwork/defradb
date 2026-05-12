// Copyright 2026 Democratized Data Foundation
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
	"context"
	"errors"
	"slices"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/cursor"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
)

type cursorNode struct {
	docMapper

	p    *Planner
	plan planNode

	first        immutable.Option[uint64]
	afterCursor  immutable.Option[string]
	afterPayload *cursor.CursorPayload

	last          immutable.Option[uint64]
	beforeCursor  immutable.Option[string]
	beforePayload *cursor.CursorPayload

	backwardBuffer []core.Doc
	bufferIndex    int

	pastCursor      bool   // true once cursor position has been passed
	collected       uint64 // count of results yielded so far
	indexSeekActive bool

	orderFields []mapper.OrderCondition

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
func (n *cursorNode) Source() planNode { return n.plan }

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

func (n *cursorNode) nextBackward() (bool, error) {
	if n.backwardBuffer != nil {
		n.bufferIndex++
		if n.bufferIndex >= len(n.backwardBuffer) {
			return false, nil
		}
		n.collected++
		return true, nil
	}

	if n.indexSeekActive {
		return n.bufferBackwardPage()
	}

	return n.drainBackwardPage()
}

func (n *cursorNode) bufferBackwardPage() (bool, error) {
	var buf []core.Doc
	limit, hasLimit := 0, false
	if n.last.HasValue() {
		limit = int(n.last.Value())
		hasLimit = true
	}

	for !hasLimit || len(buf) < limit {
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

	if hasLimit {
		hasMore, err := n.plan.Next()
		if err != nil {
			return false, err
		}
		n.hasPreviousPage = hasMore
	}

	if n.beforeCursor.HasValue() {
		hasNext, err := n.beforeCursorBoundarySurvives()
		if err != nil {
			return false, err
		}
		n.hasNextPage = hasNext
	}

	slices.Reverse(buf)
	return n.initBackwardBuffer(buf), nil
}

func (n *cursorNode) drainBackwardPage() (bool, error) {
	var buf []core.Doc
	foundBefore := false
	for {
		hasNext, err := n.plan.Next()
		if err != nil {
			return false, err
		}
		if !hasNext {
			break
		}
		doc := n.plan.Value()
		if n.beforePayload != nil && doc.GetID() == n.beforePayload.DocID {
			foundBefore = true
			break
		}
		buf = append(buf, doc.Clone())
	}

	if n.last.HasValue() && len(buf) > int(n.last.Value()) {
		n.hasPreviousPage = true
		buf = buf[len(buf)-int(n.last.Value()):]
	}

	if n.beforeCursor.HasValue() {
		n.hasNextPage = foundBefore
	}

	return n.initBackwardBuffer(buf), nil
}

func (n *cursorNode) beforeCursorBoundarySurvives() (bool, error) {
	if n.beforePayload == nil {
		return false, nil
	}

	scan := getNode[*scanNode](n.plan)
	if scan == nil || !scan.index.HasValue() {
		return false, nil
	}

	seekKey := scan.buildCursorSeekKey()
	if seekKey == nil {
		return false, nil
	}
	seekKey.Offset = 0
	if scan.index.Value().Unique {
		appendDocIDToNilUniqueIndexKey(n.p.ctx, seekKey, n.beforePayload.DocID)
	}

	txn := datastore.CtxMustGetTxn(n.p.ctx)
	val, err := txn.Datastore().Get(n.p.ctx, seekKey)
	if errors.Is(err, corekv.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(val) > 0 {
		entryDocShortID, err := keys.DecodeDocShortID(val)
		if err != nil {
			return false, err
		}
		beforeDocShortID, found, err := id.GetDocShortID(n.p.ctx, seekKey.CollectionShortID, n.beforePayload.DocID)
		if err != nil || !found {
			return false, err
		}
		if entryDocShortID != beforeDocShortID {
			return false, nil
		}
	}

	return n.beforeCursorDocMatches(scan)
}

func (n *cursorNode) beforeCursorDocMatches(scan *scanNode) (bool, error) {
	docID, err := client.NewDocIDFromString(n.beforePayload.DocID)
	if err != nil {
		return false, nil
	}

	shortID, err := id.GetCollectionShortID(n.p.ctx, scan.col.Version().CollectionID)
	if err != nil {
		return false, err
	}

	docShortID, found, err := id.GetDocShortID(n.p.ctx, shortID, docID.String())
	if err != nil || !found {
		return false, err
	}

	txn := datastore.CtxMustGetTxn(n.p.ctx)
	f := fetcher.NewDocumentFetcher()
	err = f.Init(
		n.p.ctx,
		n.p.identity,
		txn,
		n.p.nodeACP,
		n.p.documentACP,
		immutable.None[client.IndexDescription](),
		scan.col,
		scan.fields,
		scan.filter,
		nil,
		scan.documentMapping,
		scan.showDeleted,
	)
	if err != nil {
		return false, err
	}
	defer func() { _ = f.Close() }()

	err = f.Start(n.p.ctx, keys.DataStoreKey{
		CollectionShortID: shortID,
		DocShortID:        docShortID,
	})
	if err != nil {
		return false, err
	}

	doc, _, err := f.FetchNext(n.p.ctx)
	if err != nil {
		return false, err
	}
	return doc != nil, nil
}

// appendDocIDToNilUniqueIndexKey embeds the document's short ID in the key, matching how
// unique index entries with nil field values are stored (see makeUniqueKeyValueRecord).
func appendDocIDToNilUniqueIndexKey(ctx context.Context, key *keys.IndexDataStoreKey, docID string) {
	for _, field := range key.Fields {
		if field.Value.IsNil() {
			docShortID, found, err := id.GetDocShortID(ctx, key.CollectionShortID, docID)
			if err == nil && found {
				key.DocShortID = docShortID
			}
			return
		}
	}
	}

func (n *cursorNode) initBackwardBuffer(buf []core.Doc) bool {
	n.backwardBuffer = buf
	n.bufferIndex = 0

	if len(buf) == 0 {
		return false
	}

	last := len(buf) - 1
	n.firstDocID = buf[0].GetID()
	n.firstDoc = buf[0].Clone()
	n.lastDocID = buf[last].GetID()
	n.lastDoc = buf[last].Clone()
	n.collected = 1
	return true
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
