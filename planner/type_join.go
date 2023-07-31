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
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/mapper"
	"github.com/sourcenetwork/defradb/request/graphql/schema"
)

/*
 Some verbose structure and rough visualization of type joins
 can be found in the file: `type_join.md` in the same directory.
*/

// typeIndexJoin provides the needed join functionality
// for querying relationship based sub types.
// It constructs a new plan node, which queries the
// root node, then does primary key point lookups
// based on the type index key in the root.
//
// It will grab batches of docs from the root graph
// before it does the point lookups (indexJoinBatchSize).
//
// Additionally, we may need to split the provided filter
// into the root and subType components.
// Eg. (filter: {age: 10, name: "bob", author: {birthday: "June 26, 1990"}})
//
// The root filter is the conditions that apply to the main
// type ie: {age: 10, name: "bob"}.
//
// The subType filter is the conditions that apply to the
// queried sub type ie: {birthday: "June 26, 1990"}.
//
// The typeIndexJoin works by using a basic scanNode for the
// root, and recursively creates a new selectNode for the
// subType.
type typeIndexJoin struct {
	docMapper

	p *Planner

	// actual join plan, could be one of several strategies
	// based on the relationship of the sub types
	joinPlan planNode

	execInfo typeIndexJoinExecInfo
}

type typeIndexJoinExecInfo struct {
	// Total number of times typeIndexJoin node was executed.
	iterations uint64
}

func (p *Planner) makeTypeIndexJoin(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeIndexJoin, error) {
	typeJoin := &typeIndexJoin{
		p:         p,
		docMapper: docMapper{parent.documentMapping},
	}

	// handle join relation strategies
	var joinPlan planNode
	var err error

	desc := parent.sourceInfo.collectionDescription
	typeFieldDesc, ok := desc.Schema.GetField(subType.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subType.Name)
	}

	meta := typeFieldDesc.RelationType
	if schema.IsOne(meta) { // One-to-One, or One side of One-to-Many
		joinPlan, err = p.makeTypeJoinOne(parent, source, subType)
	} else if schema.IsOneToMany(meta) { // Many side of One-to-Many
		joinPlan, err = p.makeTypeJoinMany(parent, source, subType)
	} else { // more to come, Many-to-Many, Embedded?
		return nil, ErrUnknownRelationType
	}
	if err != nil {
		return nil, err
	}

	typeJoin.joinPlan = joinPlan
	return typeJoin, nil
}

func (n *typeIndexJoin) Kind() string {
	return "typeIndexJoin"
}

func (n *typeIndexJoin) Init() error {
	return n.joinPlan.Init()
}

func (n *typeIndexJoin) Start() error {
	return n.joinPlan.Start()
}

func (n *typeIndexJoin) Spans(spans core.Spans) {
	n.joinPlan.Spans(spans)
}

func (n *typeIndexJoin) Next() (bool, error) {
	n.execInfo.iterations++

	return n.joinPlan.Next()
}

func (n *typeIndexJoin) Value() core.Doc {
	return n.joinPlan.Value()
}

func (n *typeIndexJoin) Close() error {
	return n.joinPlan.Close()
}

func (n *typeIndexJoin) Source() planNode { return n.joinPlan }

func (n *typeIndexJoin) simpleExplain() (map[string]any, error) {
	const (
		joinTypeLabel               = "joinType"
		joinDirectionLabel          = "direction"
		joinDirectionPrimaryLabel   = "primary"
		joinDirectionSecondaryLabel = "secondary"
		joinSubTypeNameLabel        = "subTypeName"
		joinRootLabel               = "rootName"
	)

	simpleExplainMap := map[string]any{}

	// Add the type attribute.
	simpleExplainMap[joinTypeLabel] = n.joinPlan.Kind()

	switch joinType := n.joinPlan.(type) {
	case *typeJoinOne:
		// Add the direction attribute.
		if joinType.primary {
			simpleExplainMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		} else {
			simpleExplainMap[joinDirectionLabel] = joinDirectionSecondaryLabel
		}

		// Add the attribute(s).
		simpleExplainMap[joinRootLabel] = joinType.subTypeFieldName
		simpleExplainMap[joinSubTypeNameLabel] = joinType.subTypeName

		subTypeExplainGraph, err := buildSimpleExplainGraph(joinType.subType)
		if err != nil {
			return nil, err
		}

		// Add the joined (subType) type's entire explain graph.
		simpleExplainMap[joinSubTypeLabel] = subTypeExplainGraph

	case *typeJoinMany:
		// Add the attribute(s).
		simpleExplainMap[joinRootLabel] = joinType.rootName
		simpleExplainMap[joinSubTypeNameLabel] = joinType.subTypeName

		subTypeExplainGraph, err := buildSimpleExplainGraph(joinType.subType)
		if err != nil {
			return nil, err
		}

		// Add the joined (subType) type's entire explain graph.
		simpleExplainMap[joinSubTypeLabel] = subTypeExplainGraph

	default:
		return simpleExplainMap, client.NewErrUnhandledType("join plan", n.joinPlan)
	}

	return simpleExplainMap, nil
}

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *typeIndexJoin) Explain(explainType request.ExplainType) (map[string]any, error) {
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

// Merge implements mergeNode
func (n *typeIndexJoin) Merge() bool { return true }

// split the provided filter
// into the root and subType components.
// Eg. (filter: {age: 10, name: "bob", author: {birthday: "June 26, 1990", ...}, ...})
//
// The root filter is the conditions that apply to the main
// type ie: {age: 10, name: "bob", ...}.
//
// The subType filter is the conditions that apply to the
// queried sub type ie: {birthday: "June 26, 1990", ...}.
func splitFilterByType(filter *mapper.Filter, subType int) (*mapper.Filter, *mapper.Filter) {
	if filter == nil {
		return nil, nil
	}
	conditionKey := &mapper.PropertyIndex{
		Index: subType,
	}

	keyFound, sub := removeConditionIndex(conditionKey, filter.Conditions)
	if !keyFound {
		return filter, nil
	}

	// create new splitup filter
	// our schema ensures that if sub exists, its of type map[string]any
	splitF := &mapper.Filter{Conditions: map[connor.FilterKey]any{conditionKey: sub}}

	// check if we have any remaining filters
	if len(filter.Conditions) == 0 {
		return nil, splitF
	}
	return filter, splitF
}

// typeJoinOne is the plan node for a type index join
// where the root type is the primary in a one-to-one relation request.
type typeJoinOne struct {
	documentIterator
	docMapper

	p *Planner

	root    planNode
	subType planNode

	subTypeName      string
	subTypeFieldName string

	primary bool

	spans     core.Spans
	subSelect *mapper.Select
}

func (p *Planner) makeTypeJoinOne(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinOne, error) {
	// split filter
	if scan, ok := source.(*scanNode); ok {
		var parentfilter *mapper.Filter
		scan.filter, parentfilter = splitFilterByType(scan.filter, subType.Index)
		if parentfilter != nil {
			if parent.filter == nil {
				parent.filter = new(mapper.Filter)
			}
			parent.filter.Conditions = mergeFilterConditions(
				parent.filter.Conditions, parentfilter.Conditions)
		}
		subType.ShowDeleted = parent.selectReq.ShowDeleted
	}

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}

	// get the correct sub field schema type (collection)
	subTypeFieldDesc, ok := parent.sourceInfo.collectionDescription.Schema.GetField(subType.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subType.Name)
	}

	// determine relation direction (primary or secondary?)
	// check if the field we're querying is the primary side of the relation
	isPrimary := subTypeFieldDesc.RelationType&client.Relation_Type_Primary > 0

	subTypeCollectionDesc, err := p.getCollectionDesc(subType.CollectionName)
	if err != nil {
		return nil, err
	}

	subTypeField, subTypeFieldNameFound := subTypeCollectionDesc.GetRelation(subTypeFieldDesc.RelationName)
	if !subTypeFieldNameFound {
		return nil, client.NewErrFieldNotExist(subTypeFieldDesc.RelationName)
	}

	return &typeJoinOne{
		p:                p,
		root:             source,
		subSelect:        subType,
		subTypeName:      subType.Name,
		subTypeFieldName: subTypeField.Name,
		subType:          selectPlan,
		primary:          isPrimary,
		docMapper:        docMapper{parent.documentMapping},
	}, nil
}

func (n *typeJoinOne) Kind() string {
	return "typeJoinOne"
}

func (n *typeJoinOne) Init() error {
	if err := n.subType.Init(); err != nil {
		return err
	}
	return n.root.Init()
}

func (n *typeJoinOne) Start() error {
	if err := n.subType.Start(); err != nil {
		return err
	}
	return n.root.Start()
}

func (n *typeJoinOne) Spans(spans core.Spans) {
	n.root.Spans(spans)
}

func (n *typeJoinOne) Next() (bool, error) {
	hasNext, err := n.root.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	doc := n.root.Value()
	if n.primary {
		n.currentValue, err = n.valuesPrimary(doc)
	} else {
		n.currentValue, err = n.valuesSecondary(doc)
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (n *typeJoinOne) valuesSecondary(doc core.Doc) (core.Doc, error) {
	fkIndex := &mapper.PropertyIndex{
		Index: n.subType.DocumentMap().FirstIndexOfName(n.subTypeFieldName + request.RelatedObjectID),
	}
	filter := map[connor.FilterKey]any{
		fkIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: doc.GetKey(),
		},
	}

	// using the doc._key as a filter
	err := appendFilterToScanNode(n.subType, filter)
	if err != nil {
		return core.Doc{}, err
	}

	// We have to reset the scan node after appending the new key-filter
	if err := n.subType.Init(); err != nil {
		return doc, NewErrSubTypeInit(err)
	}

	next, err := n.subType.Next()
	if !next || err != nil {
		return doc, err
	}

	subdoc := n.subType.Value()
	doc.Fields[n.subSelect.Index] = subdoc
	return doc, nil
}

func (n *typeJoinOne) valuesPrimary(doc core.Doc) (core.Doc, error) {
	// get the subtype doc key
	subDocKey := n.docMapper.documentMapping.FirstOfName(doc, n.subTypeName+request.RelatedObjectID)

	subDocKeyStr, ok := subDocKey.(string)
	if !ok {
		return doc, nil
	}

	// create the collection key for the sub doc
	slct := n.subType.(*selectTopNode).selectNode
	desc := slct.sourceInfo.collectionDescription
	subKeyIndexKey := base.MakeDocKey(desc, subDocKeyStr)

	// reset span
	n.spans = core.NewSpans(core.NewSpan(subKeyIndexKey, subKeyIndexKey.PrefixEnd()))

	// do a point lookup with the new span (index key)
	n.subType.Spans(n.spans)

	// re-initialize the sub type plan
	if err := n.subType.Init(); err != nil {
		return doc, NewErrSubTypeInit(err)
	}

	// if we don't find any docs from our point span lookup
	// or if we encounter an error just return the base doc,
	// with an empty map for the subdoc
	next, err := n.subType.Next()

	if err != nil {
		return doc, err
	}

	if !next {
		return doc, nil
	}

	subDoc := n.subType.Value()
	doc.Fields[n.subSelect.Index] = subDoc

	return doc, nil
}

func (n *typeJoinOne) Close() error {
	err := n.root.Close()
	if err != nil {
		return err
	}
	return n.subType.Close()
}

func (n *typeJoinOne) Source() planNode { return n.root }

type typeJoinMany struct {
	documentIterator
	docMapper

	p *Planner

	// the main type that is at the parent level of the request.
	root     planNode
	rootName string
	// the index to use to gather the subtype IDs
	index *scanNode
	// the subtype plan to get the subtype docs
	subType     planNode
	subTypeName string

	subSelect *mapper.Select
}

func (p *Planner) makeTypeJoinMany(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinMany, error) {
	// split filter
	if scan, ok := source.(*scanNode); ok {
		var parentfilter *mapper.Filter
		scan.filter, parentfilter = splitFilterByType(scan.filter, subType.Index)
		if parentfilter != nil {
			if parent.filter == nil {
				parent.filter = new(mapper.Filter)
			}
			parent.filter.Conditions = mergeFilterConditions(
				parent.filter.Conditions, parentfilter.Conditions)
		}
		subType.ShowDeleted = parent.selectReq.ShowDeleted
	}

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}

	subTypeFieldDesc, ok := parent.sourceInfo.collectionDescription.Schema.GetField(subType.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subType.Name)
	}

	subTypeCollectionDesc, err := p.getCollectionDesc(subType.CollectionName)
	if err != nil {
		return nil, err
	}

	rootField, rootNameFound := subTypeCollectionDesc.GetRelation(subTypeFieldDesc.RelationName)
	if !rootNameFound {
		return nil, client.NewErrFieldNotExist(subTypeFieldDesc.RelationName)
	}

	return &typeJoinMany{
		p:           p,
		root:        source,
		subSelect:   subType,
		subTypeName: subType.Name,
		rootName:    rootField.Name,
		subType:     selectPlan,
		docMapper:   docMapper{parent.documentMapping},
	}, nil
}

func (n *typeJoinMany) Kind() string {
	return "typeJoinMany"
}

func (n *typeJoinMany) Init() error {
	if err := n.subType.Init(); err != nil {
		return err
	}
	return n.root.Init()
}

func (n *typeJoinMany) Start() error {
	if err := n.subType.Start(); err != nil {
		return err
	}
	return n.root.Start()
}

func (n *typeJoinMany) Spans(spans core.Spans) {
	n.root.Spans(spans)
}

func (n *typeJoinMany) Next() (bool, error) {
	hasNext, err := n.root.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	n.currentValue = n.root.Value()

	// check if theres an index
	// if there is, scan and aggregate resuts
	// if not, then manually scan the subtype table
	subdocs := make([]core.Doc, 0)
	if n.index != nil {
		// @todo: handle index for one-to-many setup
	} else {
		fkIndex := &mapper.PropertyIndex{
			Index: n.subSelect.FirstIndexOfName(n.rootName + request.RelatedObjectID),
		}
		filter := map[connor.FilterKey]any{
			fkIndex: map[connor.FilterKey]any{
				mapper.FilterEqOp: n.currentValue.GetKey(),
			},
		}

		// using the doc._key as a filter
		err := appendFilterToScanNode(n.subType, filter)
		if err != nil {
			return false, err
		}

		// reset scan node
		if err := n.subType.Init(); err != nil {
			return false, err
		}

		for {
			next, err := n.subType.Next()
			if err != nil {
				return false, err
			}
			if !next {
				break
			}

			subdoc := n.subType.Value()
			subdocs = append(subdocs, subdoc)
		}
	}

	n.currentValue.Fields[n.subSelect.Index] = subdocs
	return true, nil
}

func (n *typeJoinMany) Close() error {
	if err := n.root.Close(); err != nil {
		return err
	}

	return n.subType.Close()
}

func (n *typeJoinMany) Source() planNode { return n.root }

func appendFilterToScanNode(plan planNode, filterCondition map[connor.FilterKey]any) error {
	switch node := plan.(type) {
	case *scanNode:
		filter := node.filter
		if filter == nil && len(filterCondition) > 0 {
			filter = mapper.NewFilter()
		}

		filter.Conditions = mergeFilterConditions(filter.Conditions, filterCondition)

		node.filter = filter
	case nil:
		return nil
	default:
		return appendFilterToScanNode(node.Source(), filterCondition)
	}
	return nil
}

// merge into dest with src, return dest
func mergeFilterConditions(dest map[connor.FilterKey]any, src map[connor.FilterKey]any) map[connor.FilterKey]any {
	if dest == nil {
		dest = make(map[connor.FilterKey]any)
	}
	// merge filter conditions
	for k, v := range src {
		indexKey, isIndexKey := k.(*mapper.PropertyIndex)
		if !isIndexKey {
			continue
		}
		removeConditionIndex(indexKey, dest)
		dest[k] = v
	}
	return dest
}

func removeConditionIndex(
	key *mapper.PropertyIndex,
	filterConditions map[connor.FilterKey]any,
) (bool, any) {
	for targetKey, clause := range filterConditions {
		if indexKey, isIndexKey := targetKey.(*mapper.PropertyIndex); isIndexKey {
			if key.Index == indexKey.Index {
				delete(filterConditions, targetKey)
				return true, clause
			}
		}
	}
	return false, nil
}
