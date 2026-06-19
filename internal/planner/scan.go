// Copyright 2025 Democratized Data Foundation
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
	"maps"

	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/fetcher"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
	"github.com/sourcenetwork/defradb/internal/lens"
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

	fields []client.CollectionFieldDescription

	showDeleted bool

	prefixes  []keys.Walkable
	noResults bool

	filter   *mapper.Filter
	ordering []mapper.OrderCondition
	slct     *mapper.Select

	index   immutable.Option[client.IndexDescription]
	fetcher fetcher.Fetcher

	execInfo scanExecInfo
}

func (n *scanNode) Kind() string {
	return "scanNode"
}

func (n *scanNode) Init() error {
	txn := datastore.CtxMustGetTxn(n.p.ctx)
	filter, err := filterWithDocIDAliases(n.p.ctx, n.col, n.documentMapping, n.filter)
	if err != nil {
		return err
	}
	// init the fetcher
	if err := n.fetcher.Init(
		n.p.ctx,
		n.p.identity,
		txn,
		n.p.nodeACP,
		n.p.documentACP,
		n.index,
		n.col,
		n.fields,
		filter,
		n.ordering,
		n.slct.DocumentMapping,
		n.showDeleted,
	); err != nil {
		return err
	}
	return n.initScan()
}

// filterWithDocIDAliases rewrites _docID and DocID-field filters from saved aliases to current public DocIDs.
func filterWithDocIDAliases(
	ctx context.Context,
	col client.Collection,
	mapping *core.DocumentMapping,
	filter *mapper.Filter,
) (*mapper.Filter, error) {
	if filter == nil {
		return nil, nil
	}

	conditions, changed, err := expandDocIDAliasesInConditions(ctx, col, mapping, filter.Conditions)
	if err != nil {
		return nil, err
	}
	if !changed {
		return filter, nil
	}

	result := mapper.NewFilter()
	result.Conditions = conditions
	result.ExternalConditions = make(map[string]any, len(filter.ExternalConditions))
	maps.Copy(result.ExternalConditions, filter.ExternalConditions)
	return result, nil
}

func expandDocIDAliasesInConditions(
	ctx context.Context,
	col client.Collection,
	mapping *core.DocumentMapping,
	conditions map[connor.FilterKey]any,
) (map[connor.FilterKey]any, bool, error) {
	result := make(map[connor.FilterKey]any, len(conditions))
	var changed bool
	for key, value := range conditions {
		switch typedKey := key.(type) {
		case *mapper.PropertyIndex:
			opMap, ok := value.(map[connor.FilterKey]any)
			if !ok {
				result[key] = value
				continue
			}
			if isDocIDFilterField(col, mapping, typedKey.Index) {
				expandedOpMap, opChanged, err := expandDocIDAliasesInOpMap(ctx, opMap)
				if err != nil {
					return nil, false, err
				}
				changed = changed || opChanged
				result[key] = expandedOpMap
				continue
			}

			childMapping := mapping
			if mapping != nil && typedKey.Index < len(mapping.ChildMappings) &&
				mapping.ChildMappings[typedKey.Index] != nil {
				childMapping = mapping.ChildMappings[typedKey.Index]
			}
			expandedConditions, conditionsChanged, err := expandDocIDAliasesInConditions(ctx, col, childMapping, opMap)
			if err != nil {
				return nil, false, err
			}
			changed = changed || conditionsChanged
			result[key] = expandedConditions

		case *mapper.Operator:
			switch typedValue := value.(type) {
			case map[connor.FilterKey]any:
				expandedConditions, conditionsChanged, err := expandDocIDAliasesInConditions(ctx, col, mapping, typedValue)
				if err != nil {
					return nil, false, err
				}
				changed = changed || conditionsChanged
				result[key] = expandedConditions
			case []any:
				expandedList := make([]any, len(typedValue))
				var listChanged bool
				for i, item := range typedValue {
					itemMap, ok := item.(map[connor.FilterKey]any)
					if !ok {
						expandedList[i] = item
						continue
					}
					expandedItem, itemChanged, err := expandDocIDAliasesInConditions(ctx, col, mapping, itemMap)
					if err != nil {
						return nil, false, err
					}
					listChanged = listChanged || itemChanged
					expandedList[i] = expandedItem
				}
				changed = changed || listChanged
				result[key] = expandedList
			default:
				result[key] = value
			}

		default:
			result[key] = value
		}
	}
	return result, changed, nil
}

func expandDocIDAliasesInOpMap(
	ctx context.Context,
	opMap map[connor.FilterKey]any,
) (map[connor.FilterKey]any, bool, error) {
	result := make(map[connor.FilterKey]any, len(opMap))
	var changed bool
	for key, value := range opMap {
		op, ok := key.(*mapper.Operator)
		if !ok {
			result[key] = value
			continue
		}

		switch op.Operation {
		case connor.EqualOp:
			docID, ok := value.(string)
			if !ok {
				result[key] = value
				continue
			}
			values, err := docIDFilterValues(ctx, docID)
			if err != nil {
				return nil, false, err
			}
			if len(values) > 1 {
				result[mapper.FilterInOp] = values
				changed = true
			} else if len(values) == 1 && values[0] != value {
				result[key] = values[0]
				changed = true
			} else {
				result[key] = value
			}
		case connor.InOp:
			values, ok := value.([]any)
			if !ok {
				result[key] = value
				continue
			}
			expandedValues, valuesChanged, err := expandDocIDAliasValues(ctx, values)
			if err != nil {
				return nil, false, err
			}
			result[key] = expandedValues
			changed = changed || valuesChanged
		default:
			result[key] = value
		}
	}
	return result, changed, nil
}

func expandDocIDAliasValues(ctx context.Context, values []any) ([]any, bool, error) {
	expandedValues := make([]any, 0, len(values))
	seenStrings := map[string]struct{}{}
	var changed bool
	for _, value := range values {
		docID, ok := value.(string)
		if !ok {
			expandedValues = append(expandedValues, value)
			continue
		}
		aliases, err := docIDFilterValues(ctx, docID)
		if err != nil {
			return nil, false, err
		}
		for _, alias := range aliases {
			aliasString, ok := alias.(string)
			if !ok {
				expandedValues = append(expandedValues, alias)
				continue
			}
			if _, exists := seenStrings[aliasString]; exists {
				changed = true
				continue
			}
			changed = changed || aliasString != docID
			seenStrings[aliasString] = struct{}{}
			expandedValues = append(expandedValues, alias)
		}
	}
	return expandedValues, changed, nil
}

func isDocIDFilterField(col client.Collection, mapping *core.DocumentMapping, fieldIndex int) bool {
	if col == nil || mapping == nil {
		return false
	}
	fieldName, ok := mapping.TryToFindNameFromIndex(fieldIndex)
	if !ok {
		return false
	}
	if fieldName == request.DocIDFieldName {
		return true
	}
	fieldDef, ok := col.Version().GetFieldByName(fieldName)
	return ok && fieldDef.Kind == client.FieldKind_DocID
}

func docIDFilterValues(ctx context.Context, docID string) ([]any, error) {
	values := []any{docID}
	if docID == "" {
		return values, nil
	}

	localDocID, found, err := id.GetLocalDocID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if !found {
		return values, nil
	}

	publicDocID, found, err := id.GetDocID(ctx, localDocID.CollectionShortID, localDocID.DocShortID)
	if err != nil {
		return nil, err
	}
	if !found {
		return values, nil
	}
	return []any{publicDocID}, nil
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
			n.tryAddFieldWithName(request.ToFieldID(requestable.Field.Name)) // foreign key for type joins
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
						n.col.Version(),
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
		case *mapper.Similarity:
			n.tryAddFieldWithName(requestable.SimilarityTarget.Name)
		}
	}
	return nil
}

func (n *scanNode) tryAddFieldWithName(fieldName string) bool {
	fd, ok := n.col.Version().GetFieldByName(fieldName)
	if !ok {
		// skip fields that are not part of the
		// collection definition. The scanner (and fetcher)
		// is only responsible for basic fields
		return false
	}
	n.addField(fd)
	return true
}

// addField adds a field to the list of fields to be fetched.
// It will not add the field if it is already in the list.
func (n *scanNode) addField(field client.CollectionFieldDescription) {
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

func (n *scanNode) initFetcher(cid immutable.Option[[]string]) {
	var f fetcher.Fetcher
	if cid.HasValue() {
		f = new(fetcher.MultiVersioned)
	} else {
		f = fetcher.NewDocumentFetcher()

		f = lens.NewFetcher(f, n.p.lensStore, n.p.collectionRepository)
	}
	n.fetcher = f
}

// cloneWithFilter creates a new scanNode that shares the collection, fields, select mapping,
// and planner reference from this node but uses the given filter and index.
// A fresh fetcher is initialized. The clone is independent and safe to use concurrently
// with the original (e.g., for orphan fetching while the main scan is in use).
func (n *scanNode) cloneWithFilter(
	filter *mapper.Filter,
	index immutable.Option[client.IndexDescription],
	ordering []mapper.OrderCondition,
) *scanNode {
	clone := &scanNode{
		p:         n.p,
		col:       n.col,
		fields:    n.fields,
		slct:      n.slct,
		docMapper: n.docMapper,
		filter:    filter,
		index:     index,
		ordering:  ordering,
	}
	clone.initFetcher(immutable.Option[[]string]{})
	return clone
}

// Start starts the internal logic of the scanner
// like the DocumentFetcher, and more.
func (n *scanNode) Start() error {
	return nil // no op
}

func (n *scanNode) initScan() error {
	if n.noResults {
		return nil
	}
	if len(n.prefixes) == 0 {
		shortID, err := id.GetShortCollectionID(n.p.ctx, n.col.Version().CollectionID)
		if err != nil {
			return err
		}

		prefix := keys.DataStoreKey{
			CollectionShortID: shortID,
		}
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

	if n.noResults {
		return false, nil
	}

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

	shortID, err := id.GetShortCollectionID(n.p.ctx, n.col.Version().CollectionID)
	if err != nil {
		return false, err
	}

	n.currentValue, err = fetcher.DecodeToDoc(n.p.ctx, shortID, doc, n.documentMapping, false)
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
	n.noResults = false
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
	simpleExplainMap[collectionNameLabel] = n.col.Name()
	simpleExplainMap[collectionIDLabel] = n.col.Version().VersionID

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
) (*scanNode, error) {
	scan := &scanNode{
		p:         p,
		slct:      mapperSelect,
		docMapper: docMapper{mapperSelect.DocumentMapping},
	}

	col, err := p.db.GetCollectionByName(
		p.ctx,
		mapperSelect.CollectionName,
		options.WithIdentity(options.GetCollectionByName(), p.identity),
	)
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
	planNode   planNode
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
		return n.planNode.Init()
	})
	return n.err
}

func (n *multiScanNode) Start() error {
	n.countAndCall(&n.startCount, func() error {
		return n.planNode.Start()
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
		n.nextResult, err = n.planNode.Next()
		return
	})

	return n.nextResult, n.err
}

func (n *multiScanNode) Value() core.Doc {
	return n.planNode.Value()
}

func (n *multiScanNode) Prefixes(prefixes []keys.Walkable) {
	n.planNode.Prefixes(prefixes)
}

func (n *multiScanNode) Source() planNode {
	return n.planNode
}

func (n *multiScanNode) Kind() string {
	return "multiScanNode"
}

func (n *multiScanNode) Close() error {
	n.countAndCall(&n.closeCount, func() error {
		return n.planNode.Close()
	})
	return n.err
}

func (n *multiScanNode) DocumentMap() *core.DocumentMapping {
	return n.planNode.DocumentMap()
}

func (n *multiScanNode) addReader() {
	n.numReaders++
}
