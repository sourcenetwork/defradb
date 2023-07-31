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
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/db/fetcher"
	"github.com/sourcenetwork/defradb/lens"
	"github.com/sourcenetwork/defradb/planner/mapper"
	"github.com/sourcenetwork/defradb/request/graphql/parser"
)

// scanExecInfo contains information about the execution of a scan.
type scanExecInfo struct {
	// Total number of times scan was issued.
	iterations uint64

	// Information about fetches.
	fetches fetcher.ExecInfo
}

// scans an index for records
type scanNode struct {
	documentIterator
	docMapper

	p    *Planner
	desc client.CollectionDescription

	fields []client.FieldDescription
	docKey []byte

	showDeleted bool

	spans   core.Spans
	reverse bool

	filter *mapper.Filter
	slct   *mapper.Select

	fetcher fetcher.Fetcher

	execInfo scanExecInfo
}

func (n *scanNode) Kind() string {
	return "scanNode"
}

func (n *scanNode) Init() error {
	// init the fetcher
	if err := n.fetcher.Init(
		n.p.ctx,
		n.p.txn,
		&n.desc,
		n.fields,
		n.filter,
		n.slct.DocumentMapping,
		n.reverse,
		n.showDeleted,
	); err != nil {
		return err
	}
	return n.initScan()
}

func (n *scanNode) initCollection(desc client.CollectionDescription) error {
	n.desc = desc
	return n.initFields(n.slct.Fields)
}

func (n *scanNode) initFields(fields []mapper.Requestable) error {
	for _, r := range fields {
		// add all the possible base level fields the fetcher is responsible
		// for, including those that are needed by higher level aggregates
		// or grouping alls, which them selves might have further dependents
		switch requestable := r.(type) {
		// field is simple as its just a base level field
		case *mapper.Field:
			n.tryAddField(requestable.GetName())
		// select might have its own select fields and filters fields
		case *mapper.Select:
			n.tryAddField(requestable.Field.Name + "_id") // foreign key for type joins
			err := n.initFields(requestable.Fields)
			if err != nil {
				return err
			}
		// aggregate might have its own target fields and filter fields
		case *mapper.Aggregate:
			for _, target := range requestable.AggregateTargets {
				if target.Filter != nil {
					fieldDescs, err := parser.ParseFilterFieldsForDescription(
						target.Filter.ExternalConditions,
						n.desc.Schema,
					)
					if err != nil {
						return err
					}
					for _, fd := range fieldDescs {
						n.tryAddField(fd.Name)
					}
				}
				if target.ChildTarget.HasValue {
					n.tryAddField(target.ChildTarget.Name)
				} else {
					n.tryAddField(target.Field.Name)
				}
			}
		}
	}
	return nil
}

func (n *scanNode) tryAddField(fieldName string) bool {
	fd, ok := n.desc.Schema.GetField(fieldName)
	if !ok {
		// skip fields that are not part of the
		// schema description. The scanner (and fetcher)
		// is only responsible for basic fields
		return false
	}
	n.fields = append(n.fields, fd)
	return true
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	return nil // no op
}

func (n *scanNode) initScan() error {
	if !n.spans.HasValue {
		start := base.MakeCollectionKey(n.desc)
		n.spans = core.NewSpans(core.NewSpan(start, start.PrefixEnd()))
	}

	err := n.fetcher.Start(n.p.ctx, n.spans)
	if err != nil {
		return err
	}

	return nil
}

// Next gets the next result.
// Returns true, if there is a result,
// and false otherwise.
func (n *scanNode) Next() (bool, error) {
	n.execInfo.iterations++

	if n.spans.HasValue && len(n.spans.Value) == 0 {
		return false, nil
	}

	var err error
	var execInfo fetcher.ExecInfo
	n.docKey, n.currentValue, execInfo, err = n.fetcher.FetchNextDoc(n.p.ctx, n.documentMapping)
	if err != nil {
		return false, err
	}
	n.execInfo.fetches.Add(execInfo)

	if len(n.currentValue.Fields) == 0 {
		return false, nil
	}

	n.documentMapping.SetFirstOfName(
		&n.currentValue,
		request.DeletedFieldName,
		n.currentValue.Status.IsDeleted(),
	)

	return true, nil
}

func (n *scanNode) Spans(spans core.Spans) {
	n.spans = spans
}

func (n *scanNode) Close() error {
	return n.fetcher.Close()
}

func (n *scanNode) Source() planNode { return nil }

// explainSpans explains the spans attribute.
func (n *scanNode) explainSpans() []map[string]any {
	spansExplainer := []map[string]any{}
	for _, span := range n.spans.Value {
		spanExplainer := map[string]any{
			"start": span.Start().ToString(),
			"end":   span.End().ToString(),
		}

		spansExplainer = append(spansExplainer, spanExplainer)
	}

	return spansExplainer
}

func (n *scanNode) simpleExplain() (map[string]any, error) {
	simpleExplainMap := map[string]any{}

	// Add the filter attribute if it exists.
	if n.filter == nil {
		simpleExplainMap[filterLabel] = nil
	} else {
		simpleExplainMap[filterLabel] = n.filter.ToMap(n.documentMapping)
	}

	// Add the collection attributes.
	simpleExplainMap[collectionNameLabel] = n.desc.Name
	simpleExplainMap[collectionIDLabel] = n.desc.IDString()

	// Add the spans attribute.
	simpleExplainMap[spansLabel] = n.explainSpans()

	return simpleExplainMap, nil
}

func (n *scanNode) executeExplain() map[string]any {
	return map[string]any{
		"iterations":   n.execInfo.iterations,
		"docFetches":   n.execInfo.fetches.DocsFetched,
		"fieldFetches": n.execInfo.fetches.FieldsFetched,
	}
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *scanNode) Explain(explainType request.ExplainType) (map[string]any, error) {
	switch explainType {
	case request.SimpleExplain:
		return n.simpleExplain()

	case request.ExecuteExplain:
		return n.executeExplain(), nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

// Merge implements mergeNode
func (n *scanNode) Merge() bool { return true }

func (p *Planner) Scan(parsed *mapper.Select) (*scanNode, error) {
	var f fetcher.Fetcher
	if parsed.Cid.HasValue() {
		f = new(fetcher.VersionedFetcher)
	} else {
		f = new(fetcher.DocumentFetcher)
		f = lens.NewFetcher(f, p.db.LensRegistry())
	}
	scan := &scanNode{
		p:         p,
		fetcher:   f,
		slct:      parsed,
		docMapper: docMapper{parsed.DocumentMapping},
	}

	colDesc, err := p.getCollectionDesc(parsed.CollectionName)
	if err != nil {
		return nil, err
	}
	err = scan.initCollection(colDesc)
	if err != nil {
		return nil, err
	}
	return scan, nil
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
