// Copyright 2020 Source Inc.
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
	"fmt"

	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
)

// scans an index for records
type scanNode struct {
	p     *Planner
	desc  base.CollectionDescription
	index *base.IndexDescription

	fields []*base.FieldDescription
	doc    map[string]interface{}
	docKey []byte

	// Commenting out because unused code (structcheck) according to linter.
	// // map between fieldID and index in fields
	// fieldIdxMap map[base.FieldID]int
	// isSecondaryIndex bool

	spans   core.Spans
	reverse bool

	// rowIndex int64

	// filter data
	filter *parser.Filter

	scanInitialized bool

	fetcher fetcher.DocumentFetcher
}

func (n *scanNode) Init() error {
	return n.initScan()
}

func (n *scanNode) initCollection(desc base.CollectionDescription) error {
	n.desc = desc
	n.index = &desc.Indexes[0]
	return nil
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	// init the fetcher
	return n.fetcher.Init(&n.desc, n.index, n.fields, n.reverse)
}

func (n *scanNode) initScan() error {
	if len(n.spans) == 0 {
		start := base.MakeIndexPrefixKey(&n.desc, n.index)
		n.spans = append(n.spans, core.NewSpan(start, start.PrefixEnd()))
	}

	fmt.Println("Initializing scan with the following spans:", n.spans)
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
	if !n.scanInitialized {
		if err := n.initScan(); err != nil {
			return false, err
		}
	}

	// keep scanning until we find a doc that passes the filter
	for {
		var err error
		n.docKey, n.doc, err = n.fetcher.FetchNextMap(n.p.ctx)
		if err != nil {
			return false, err
		}
		if n.doc == nil {
			return false, nil
		}

		passed, err := parser.RunFilter(n.doc, n.filter, n.p.evalCtx)
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

// Values returns the most recent result from Next()
func (n *scanNode) Values() map[string]interface{} {
	return n.doc
}

func (n *scanNode) Close() error {
	return n.fetcher.Close()
}

func (n *scanNode) Source() planNode { return nil }

// Merge implements mergeNode
func (n *scanNode) Merge() bool { return true }

func (p *Planner) Scan() *scanNode {
	return &scanNode{p: p}
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
	*scanNode
	numReaders int
	numCalls   int

	lastBool bool
	lastErr  error
}

func (n *multiScanNode) addReader() {
	n.numReaders++
}

func (n *multiScanNode) Source() planNode {
	return n.scanNode
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

/*
multiscan := p.MultiScan(scan)
multiscan.Register(typeJoin1)
multiscan.Register(typeJoin2)
multiscan.Register(commitScan)
*/
