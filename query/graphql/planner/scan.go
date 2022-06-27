// Copyright 2022 Democratized Data Foundation
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
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
)

// scans an index for records
type scanNode struct {
	documentIterator
	docMapper

	p    *Planner
	desc client.CollectionDescription

	fields []*client.FieldDescription
	docKey []byte

	spans   core.Spans
	reverse bool

	filter *mapper.Filter

	scanInitialized bool

	fetcher fetcher.Fetcher
}

func (n *scanNode) Kind() string {
	return "scanNode"
}

func (n *scanNode) Init() error {
	// init the fetcher
	if err := n.fetcher.Init(&n.desc, n.fields, n.reverse); err != nil {
		return err
	}
	return n.initScan()
}

func (n *scanNode) initCollection(desc client.CollectionDescription) error {
	n.desc = desc
	return nil
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	return nil // no op
}

func (n *scanNode) initScan() error {
	if len(n.spans) == 0 {
		start := base.MakeCollectionKey(n.desc)
		n.spans = append(n.spans, core.NewSpan(start, start.PrefixEnd()))
	}

	err := n.fetcher.Start(n.p.ctx, n.p.txn, n.spans)
	if err != nil {
		return err
	}

	n.scanInitialized = true
	return nil
}

// Next gets the next result.
// Returns true, if there is a result,
// and false otherwise.
func (n *scanNode) Next() (bool, error) {
	// keep scanning until we find a doc that passes the filter
	for {
		var err error
		n.docKey, n.currentValue, err = n.fetcher.FetchNextDoc(n.p.ctx, n.documentMapping)
		if err != nil {
			return false, err
		}

		if len(n.currentValue.Fields) == 0 {
			return false, nil
		}

		passed, err := mapper.RunFilter(n.currentValue, n.filter)
		if err != nil {
			return false, err
		}
		if passed {
			return true, nil
		}
	}
}

func (n *scanNode) Spans(spans core.Spans) {
	n.spans = spans
}

func (n *scanNode) Close() error {
	return n.fetcher.Close()
}

func (n *scanNode) Source() planNode { return nil }

// explainSpans explains the spans attribute.
func (n *scanNode) explainSpans() []map[string]interface{} {
	spansExplainer := []map[string]interface{}{}
	for _, span := range n.spans {
		spanExplainer := map[string]interface{}{
			"start": span.Start().ToString(),
			"end":   span.End().ToString(),
		}

		spansExplainer = append(spansExplainer, spanExplainer)
	}

	return spansExplainer
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *scanNode) Explain() (map[string]interface{}, error) {
	explainerMap := map[string]interface{}{}

	// Add the filter attribute if it exists.
	if n.filter == nil || n.filter.ExternalConditions == nil {
		explainerMap[filterLabel] = nil
	} else {
		explainerMap[filterLabel] = n.filter.ExternalConditions
	}

	// Add the collection attributes.
	explainerMap[collectionNameLabel] = n.desc.Name
	explainerMap[collectionIDLabel] = n.desc.IDString()

	// Add the spans attribute.
	explainerMap[spansLabel] = n.explainSpans()

	return explainerMap, nil
}

// Merge implements mergeNode
func (n *scanNode) Merge() bool { return true }

func (p *Planner) Scan(parsed *mapper.Select) *scanNode {
	var f fetcher.Fetcher
	if parsed.Cid != "" {
		f = new(fetcher.VersionedFetcher)
	} else {
		f = new(fetcher.DocumentFetcher)
	}
	return &scanNode{
		p:         p,
		fetcher:   f,
		docMapper: docMapper{&parsed.DocumentMapping},
	}
}

// multiScanNode is a buffered scanNode that has
// multiple readers. Each reader is unaware of the
// others, so we need a system, that will correctly
// manage *when* to increment through the scanNode
// plan.
//
// If we have two readers on our multiScanNode, then
// we call Next() on the underlying scanNode only
// once every 2 Next() calls on the multiScan
type multiScanNode struct {
	docMapper

	scanNode   *scanNode
	numReaders int
	numCalls   int

	lastBool bool
	lastErr  error
}

func (n *multiScanNode) Init() error {
	return n.scanNode.Init()
}

func (n *multiScanNode) Start() error {
	return n.scanNode.Start()
}

// Next only calls Next() on the underlying
// scanNode every numReaders.
func (n *multiScanNode) Next() (bool, error) {
	if n.numCalls == 0 {
		n.lastBool, n.lastErr = n.scanNode.Next()
	}
	n.numCalls++

	// if the number of calls equals the numbers of readers
	// reset the counter, so our next call actually executes the Next()
	if n.numCalls == n.numReaders {
		n.numCalls = 0
	}

	return n.lastBool, n.lastErr
}

func (n *multiScanNode) Value() core.Doc {
	return n.scanNode.documentIterator.Value()
}

func (n *multiScanNode) Spans(spans core.Spans) {
	n.scanNode.Spans(spans)
}

func (n *multiScanNode) Source() planNode {
	return n.scanNode
}

func (n *multiScanNode) Kind() string {
	return "multiScanNode"
}

func (n *multiScanNode) Close() error {
	return n.scanNode.Close()
}

func (n *multiScanNode) addReader() {
	n.numReaders++
}
