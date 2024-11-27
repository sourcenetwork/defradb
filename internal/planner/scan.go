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

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/lens"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
	"github.com/sourcenetwork/defradb/internal/request/graphql/parser"
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

	p   *Planner
	col client.Collection

	fields []client.FieldDefinition

	showDeleted bool

	prefixes []keys.Walkable

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
		n.p.identity,
		n.p.txn,
		n.p.acp,
		n.col,
		n.fields,
		n.filter,
		n.slct.DocumentMapping,
		n.showDeleted,
	); err != nil {
		return err
	}
	return n.initScan()
}

func (n *scanNode) initCollection(col client.Collection) error {
	n.col = col
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
			n.tryAddFieldWithName(requestable.GetName())
		// select might have its own select fields and filters fields
		case *mapper.Select:
			n.tryAddFieldWithName(requestable.Field.Name + request.RelatedObjectID) // foreign key for type joins
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
						n.col.Definition(),
					)
					if err != nil {
						return err
					}
					for _, fd := range fieldDescs {
						n.tryAddFieldWithName(fd.Name)
					}
				}
				if target.ChildTarget.HasValue {
					n.tryAddFieldWithName(target.ChildTarget.Name)
				} else {
					n.tryAddFieldWithName(target.Field.Name)
				}
			}
		}
	}
	return nil
}

func (n *scanNode) tryAddFieldWithName(fieldName string) bool {
	fd, ok := n.col.Definition().GetFieldByName(fieldName)
	if !ok {
		// skip fields that are not part of the
		// schema description. The scanner (and fetcher)
		// is only responsible for basic fields
		return false
	}
	n.addField(fd)
	return true
}

// addField adds a field to the list of fields to be fetched.
// It will not add the field if it is already in the list.
func (n *scanNode) addField(field client.FieldDefinition) {
	found := false
	for i := range n.fields {
		if n.fields[i].Name == field.Name {
			found = true
			break
		}
	}
	if !found {
		n.fields = append(n.fields, field)
	}
}

func (scan *scanNode) initFetcher(
	cid immutable.Option[string],
	index immutable.Option[client.IndexDescription],
) {
	var f fetcher.Fetcher
	if cid.HasValue() {
		f = new(fetcher.VersionedFetcher)
	} else {
		f = fetcher.NewDocumentFetcher()

		if index.HasValue() {
			fieldsToMove := make([]mapper.Field, 0, len(index.Value().Fields))
			fieldsToCopy := make([]mapper.Field, 0, len(index.Value().Fields))
			for _, field := range index.Value().Fields {
				fieldName := field.Name
				typeIndex := scan.documentMapping.FirstIndexOfName(fieldName)
				indexField := mapper.Field{Index: typeIndex, Name: fieldName}
				fd, _ := scan.col.Definition().Schema.GetFieldByName(fieldName)
				// if the field is an array, we need to copy it instead of moving so that the
				// top select node can do final filter check on the whole array of the document
				if fd.Kind.IsArray() {
					fieldsToCopy = append(fieldsToCopy, indexField)
				} else {
					fieldsToMove = append(fieldsToMove, indexField)
				}
			}
			var indexFilter *mapper.Filter
			scan.filter, indexFilter = filter.SplitByFields(scan.filter, fieldsToMove...)
			for i := range fieldsToCopy {
				indexFilter = filter.Merge(indexFilter, filter.CopyField(scan.filter, fieldsToCopy[i]))
			}
			if indexFilter != nil {
				f = fetcher.NewIndexFetcher(f, index.Value(), indexFilter)
			}
		}

		f = lens.NewFetcher(f, scan.p.db.LensRegistry())
	}
	scan.fetcher = f
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	return nil // no op
}

func (n *scanNode) initScan() error {
	if len(n.prefixes) == 0 {
		prefix := base.MakeDataStoreKeyWithCollectionDescription(n.col.Description())
		n.prefixes = []keys.Walkable{prefix}
	}

	err := n.fetcher.Start(n.p.ctx, n.prefixes...)
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

	if len(n.prefixes) == 0 {
		return false, nil
	}

	doc, execInfo, err := n.fetcher.FetchNext(n.p.ctx)
	if err != nil {
		return false, err
	}
	n.execInfo.fetches.Add(execInfo)

	if doc == nil {
		return false, nil
	}

	n.currentValue, err = fetcher.DecodeToDoc(doc, n.documentMapping, false)
	if err != nil {
		return false, err
	}

	n.documentMapping.SetFirstOfName(
		&n.currentValue,
		request.DeletedFieldName,
		n.currentValue.Status.IsDeleted(),
	)

	return true, nil
}

func (n *scanNode) Prefixes(prefixes []keys.Walkable) {
	n.prefixes = prefixes
}

func (n *scanNode) Close() error {
	return n.fetcher.Close()
}

func (n *scanNode) Source() planNode { return nil }

// explainPrefixes explains the prefixes attribute.
func (n *scanNode) explainPrefixes() []string {
	prefixes := make([]string, len(n.prefixes))
	for i, prefix := range n.prefixes {
		prefixes[i] = keys.PrettyPrint(prefix)
	}
	return prefixes
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
	simpleExplainMap[collectionNameLabel] = n.col.Name().Value()
	simpleExplainMap[collectionIDLabel] = n.col.Description().IDString()

	// Add the prefixes attribute.
	simpleExplainMap[prefixesLabel] = n.explainPrefixes()

	return simpleExplainMap, nil
}

func (n *scanNode) executeExplain() map[string]any {
	return map[string]any{
		"iterations":   n.execInfo.iterations,
		"docFetches":   n.execInfo.fetches.DocsFetched,
		"fieldFetches": n.execInfo.fetches.FieldsFetched,
		"indexFetches": n.execInfo.fetches.IndexesFetched,
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

func (p *Planner) Scan(
	mapperSelect *mapper.Select,
	colDesc client.CollectionDescription,
) (*scanNode, error) {
	scan := &scanNode{
		p:         p,
		slct:      mapperSelect,
		docMapper: docMapper{mapperSelect.DocumentMapping},
	}

	col, err := p.db.GetCollectionByName(p.ctx, mapperSelect.CollectionName)
	if err != nil {
		return nil, err
	}
	err = scan.initCollection(col)
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
//
// NOTE: calling Init() on multiScanNode is subject to counting as well and as such
// doesn't not provide idempotency guarantees. Counting is purely for performance
// reasons and removing it should be safe.
type multiScanNode struct {
	scanNode   *scanNode
	numReaders int
	nextCount  int
	initCount  int
	startCount int
	closeCount int

	nextResult bool
	err        error
}

// Init initializes the multiScanNode.
// NOTE: this function is subject to counting based on the number of readers and as such
// doesn't not provide idempotency guarantees. Counting is purely for performance
// reasons and removing it should be safe.
func (n *multiScanNode) Init() error {
	n.countAndCall(&n.initCount, func() error {
		return n.scanNode.Init()
	})
	return n.err
}

func (n *multiScanNode) Start() error {
	n.countAndCall(&n.startCount, func() error {
		return n.scanNode.Start()
	})
	return n.err
}

// countAndCall keeps track of number of requests to call a given function by checking a
// function's count.
// The function is only called when the count is 0.
// If the count is equal to the number of readers, the count is reset.
// If the function returns an error, the error is stored in the multiScanNode.
func (n *multiScanNode) countAndCall(count *int, f func() error) {
	if *count == 0 {
		err := f()
		if err != nil {
			n.err = err
		}
	}
	*count++

	// if the number of calls equals the numbers of readers
	// reset the counter, so our next call actually executes the function
	if *count == n.numReaders {
		*count = 0
	}
}

// Next only calls Next() on the underlying
// scanNode every numReaders.
func (n *multiScanNode) Next() (bool, error) {
	n.countAndCall(&n.nextCount, func() (err error) {
		n.nextResult, err = n.scanNode.Next()
		return
	})

	return n.nextResult, n.err
}

func (n *multiScanNode) Value() core.Doc {
	return n.scanNode.documentIterator.Value()
}

func (n *multiScanNode) Prefixes(prefixes []keys.Walkable) {
	n.scanNode.Prefixes(prefixes)
}

func (n *multiScanNode) Source() planNode {
	return n.scanNode
}

func (n *multiScanNode) Kind() string {
	return "multiScanNode"
}

func (n *multiScanNode) Close() error {
	n.countAndCall(&n.closeCount, func() error {
		return n.scanNode.Close()
	})
	return n.err
}

func (n *multiScanNode) DocumentMap() *core.DocumentMapping {
	return n.scanNode.DocumentMap()
}

func (n *multiScanNode) addReader() {
	n.numReaders++
}
