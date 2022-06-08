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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/mapper"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
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

	// root        planNode
	// subType     planNode
	// subTypeName string

	// actual join plan, could be one of several strategies
	// based on the relationship of the sub types
	joinPlan planNode

	// doc map[string]interface{}

	// spans core.Spans
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
	typeFieldDesc, ok := desc.GetField(subType.Name)
	if !ok {
		// return nil, fmt.Errorf("Unknown field on sub selection")
		return nil, fmt.Errorf("Unknown field %s on sub selection", subType.Name)
	}

	meta := typeFieldDesc.RelationType
	if schema.IsOne(meta) { // One-to-One, or One side of One-to-Many
		joinPlan, err = p.makeTypeJoinOne(parent, source, subType)
	} else if schema.IsOneToMany(meta) { // Many side of One-to-Many
		joinPlan, err = p.makeTypeJoinMany(parent, source, subType)
	} else { // more to come, Many-to-Many, Embedded?
		return nil, fmt.Errorf("Failed sub selection, unknown relation type")
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

func (n *typeIndexJoin) Spans(spans core.Spans) { /* todo */ }

func (n *typeIndexJoin) Next() (bool, error) {
	return n.joinPlan.Next()
}

func (n *typeIndexJoin) Value() core.Doc {
	return n.joinPlan.Value()
}

func (n *typeIndexJoin) Close() error {
	return n.joinPlan.Close()
}

func (n *typeIndexJoin) Source() planNode { return n.joinPlan }

// Explain method returns a map containing all attributes of this node that
// are to be explained, subscribes / opts-in this node to be an explainablePlanNode.
func (n *typeIndexJoin) Explain() (map[string]interface{}, error) {
	const (
		joinTypeLabel               = "joinType"
		joinDirectionLabel          = "direction"
		joinDirectionPrimaryLabel   = "primary"
		joinDirectionSecondaryLabel = "secondary"
		joinSubTypeLabel            = "subType"
		joinRootLabel               = "rootName"
	)

	explainerMap := map[string]interface{}{}

	// Add the type attribute.
	explainerMap[joinTypeLabel] = n.joinPlan.Kind()

	switch joinType := n.joinPlan.(type) {

	case *typeJoinOne:
		// Add the direction attribute.
		if joinType.primary {
			explainerMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		} else {
			explainerMap[joinDirectionLabel] = joinDirectionSecondaryLabel
		}

		// Add the attribute(s).
		explainerMap[joinRootLabel] = joinType.subTypeFieldName

		subTypeExplainGraph, err := buildExplainGraph(joinType.subType)
		if err != nil {
			return nil, err
		}

		// Add the joined (subType) type's entire explain graph.
		explainerMap[joinSubTypeLabel] = subTypeExplainGraph

	case *typeJoinMany:
		// Add the attribute(s).
		explainerMap[joinRootLabel] = joinType.rootName

		subTypeExplainGraph, err := buildExplainGraph(joinType.subType)
		if err != nil {
			return nil, err
		}

		// Add the joined (subType) type's entire explain graph.
		explainerMap[joinSubTypeLabel] = subTypeExplainGraph

	default:
		return explainerMap, fmt.Errorf("Unknown type of an index join to explain.")
	}

	return explainerMap, nil
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
		return filter, &mapper.Filter{}
	}

	// create new splitup filter
	// our schema ensures that if sub exists, its of type map[string]interface{}
	splitF := &mapper.Filter{Conditions: map[connor.FilterKey]interface{}{conditionKey: sub}}
	return filter, splitF
}

// typeJoinOne is the plan node for a type index join
// where the root type is the primary in a one-to-one relation
// query.
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
		scan.filter, parent.filter = splitFilterByType(scan.filter, subType.Index)
	}

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}

	subTypeFieldName, err := p.db.GetRelationshipIdField(
		subType.Name,
		subType.CollectionName,
		parent.parsed.CollectionName,
	)
	if err != nil {
		return nil, err
	}

	// get the correct sub field schema type (collection)
	subTypeFieldDesc, ok := parent.sourceInfo.collectionDescription.GetField(subType.Name)
	if !ok {
		return nil, fmt.Errorf("couldn't find subtype field description for typeJoin node")
	}

	// determine relation direction (primary or secondary?)
	// check if the field we're querying is the primary side of the relation
	isPrimary := subTypeFieldDesc.RelationType&client.Relation_Type_Primary > 0

	return &typeJoinOne{
		p:                p,
		root:             source,
		subSelect:        subType,
		subTypeName:      subType.Name,
		subTypeFieldName: subTypeFieldName,
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

func (n *typeJoinOne) Spans(spans core.Spans) { /* todo */ }

func (n *typeJoinOne) Next() (bool, error) {
	hasNext, err := n.root.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	doc := n.root.Value()
	if n.primary {
		n.currentValue = n.valuesPrimary(doc)
	} else {
		n.currentValue = n.valuesSecondary(doc)
	}
	return true, nil
}

func (n *typeJoinOne) valuesSecondary(doc core.Doc) core.Doc {
	fkIndex := &mapper.PropertyIndex{
		Index: n.subType.DocumentMap().FirstIndexOfName(n.subTypeFieldName + "_id"),
	}
	filter := map[connor.FilterKey]interface{}{
		fkIndex: doc.GetKey(),
	}
	// using the doc._key as a filter
	err := appendFilterToScanNode(n.subType, filter)
	if err != nil {
		return core.Doc{}
	}

	doc.Fields[n.subSelect.Index] = n.subSelect.DocumentMapping.NewDoc()
	next, err := n.subType.Next()
	if !next || err != nil {
		return doc
	}

	subdoc := n.subType.Value()
	doc.Fields[n.subSelect.Index] = subdoc
	return doc
}

func (n *typeJoinOne) valuesPrimary(doc core.Doc) core.Doc {
	// get the subtype doc key
	subDocKey := n.docMapper.documentMapping.FirstOfName(doc, n.subTypeName+"_id")

	subDocKeyStr, ok := subDocKey.(string)
	if !ok {
		return doc
	}

	doc.Fields[n.subSelect.Index] = n.subSelect.DocumentMapping.NewDoc()

	// create the collection key for the sub doc
	slct := n.subType.(*selectTopNode).selectnode
	desc := slct.sourceInfo.collectionDescription
	subKeyIndexKey := base.MakeDocKey(desc, subDocKeyStr)

	n.spans = core.Spans{} // reset span
	n.spans = append(n.spans, core.NewSpan(subKeyIndexKey, subKeyIndexKey.PrefixEnd()))

	// do a point lookup with the new span (index key)
	n.subType.Spans(n.spans)

	// re-initialize the sub type plan
	if err := n.subType.Init(); err != nil {
		log.ErrorE(n.p.ctx, "Sub-type initialization error at scan node reset", err)
		return doc
	}

	// if we don't find any docs from our point span lookup
	// or if we encounter an error just return the base doc,
	// with an empty map for the subdoc
	next, err := n.subType.Next()

	if err != nil {
		log.ErrorE(n.p.ctx, "Sub-type initialization error at scan node reset", err)
		return doc
	}

	if !next {
		return doc
	}

	subDoc := n.subType.Value()
	doc.Fields[n.subSelect.Index] = subDoc

	return doc
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

	// the main type that is a the parent level of the query.
	root     planNode
	rootName string
	// the index to use to gather the subtype IDs
	index *scanNode
	// the subtype plan to get the subtype docs
	subType planNode

	subSelect *mapper.Select
}

func (p *Planner) makeTypeJoinMany(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinMany, error) {
	// split filter
	if scan, ok := source.(*scanNode); ok {
		scan.filter, parent.filter = splitFilterByType(scan.filter, subType.Index)
	}

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}

	rootName, err := p.db.GetRelationshipIdField(
		subType.Name,
		subType.CollectionName,
		parent.parsed.CollectionName,
	)
	if err != nil {
		return nil, err
	}

	return &typeJoinMany{
		p:         p,
		root:      source,
		subSelect: subType,
		rootName:  rootName,
		subType:   selectPlan,
		docMapper: docMapper{parent.documentMapping},
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

func (n *typeJoinMany) Spans(spans core.Spans) { /* todo */ }

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
			Index: n.subSelect.FirstIndexOfName(n.rootName + "_id"),
		}
		filter := map[connor.FilterKey]interface{}{
			fkIndex: n.currentValue.GetKey(), // user_id: "bae-ALICE" |  user_id: "bae-CHARLIE"
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

func appendFilterToScanNode(plan planNode, filterCondition map[connor.FilterKey]interface{}) error {
	switch node := plan.(type) {
	case *scanNode:
		filter := node.filter
		if filter == nil {
			filter = mapper.NewFilter()
		}

		// merge filter conditions
		for k, v := range filterCondition {
			indexKey, isIndexKey := k.(*mapper.PropertyIndex)
			if !isIndexKey {
				continue
			}
			removeConditionIndex(indexKey, filter.Conditions)
			filter.Conditions[k] = v
		}

		node.filter = filter
	case nil:
		return nil
	default:
		return appendFilterToScanNode(node.Source(), filterCondition)
	}
	return nil
}

func removeConditionIndex(
	key *mapper.PropertyIndex,
	filterConditions map[connor.FilterKey]interface{},
) (bool, interface{}) {
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
