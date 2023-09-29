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
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/connor"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/planner/filter"
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

	primary             bool
	secondaryFieldIndex immutable.Option[int]

	spans     core.Spans
	subSelect *mapper.Select
}

func (p *Planner) makeTypeJoinOne(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinOne, error) {
	prepareScanNodeFilterForTypeJoin(parent, source, subType)

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
	isPrimary := subTypeFieldDesc.RelationType.IsSet(client.Relation_Type_Primary)

	subTypeCollectionDesc, err := p.getCollectionDesc(subType.CollectionName)
	if err != nil {
		return nil, err
	}

	subTypeField, subTypeFieldNameFound := subTypeCollectionDesc.GetFieldByRelation(
		subTypeFieldDesc.RelationName,
		parent.sourceInfo.collectionDescription.Name,
		subTypeFieldDesc.Name,
		&subTypeCollectionDesc.Schema,
	)
	if !subTypeFieldNameFound {
		return nil, client.NewErrFieldNotExist(subTypeFieldDesc.RelationName)
	}

	var secondaryFieldIndex immutable.Option[int]
	if !isPrimary {
		idFieldName := subTypeFieldDesc.Name + request.RelatedObjectID
		secondaryFieldIndex = immutable.Some(
			parent.documentMapping.FirstIndexOfName(idFieldName),
		)
	}

	return &typeJoinOne{
		p:                   p,
		root:                source,
		subSelect:           subType,
		subTypeName:         subType.Name,
		subTypeFieldName:    subTypeField.Name,
		subType:             selectPlan,
		primary:             isPrimary,
		secondaryFieldIndex: secondaryFieldIndex,
		docMapper:           docMapper{parent.documentMapping},
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
	propIndex := n.subType.DocumentMap().FirstIndexOfName(n.subTypeFieldName + request.RelatedObjectID)
	// using the doc._key as a filter
	setSubTypeFilterToScanNode(n.subType, propIndex, doc.GetKey())

	// We have to reset the scan node after appending the new key-filter
	if err := n.subType.Init(); err != nil {
		return doc, NewErrSubTypeInit(err)
	}

	next, err := n.subType.Next()
	if !next || err != nil {
		return doc, err
	}

	subDoc := n.subType.Value()
	doc.Fields[n.subSelect.Index] = subDoc

	if n.secondaryFieldIndex.HasValue() {
		doc.Fields[n.secondaryFieldIndex.Value()] = subDoc.GetKey()
	}

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
	// with an empty map for the subDoc
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

func prepareScanNodeFilterForTypeJoin(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) {
	subType.ShowDeleted = parent.selectReq.ShowDeleted

	scan, ok := source.(*scanNode)
	if !ok || scan.filter == nil {
		return
	}

	if filter.IsComplex(scan.filter) {
		if parent.filter == nil {
			parent.filter = mapper.NewFilter()
			parent.filter.Conditions = filter.Copy(scan.filter.Conditions)
		} else {
			parent.filter.Conditions = filter.Merge(
				parent.filter.Conditions, scan.filter.Conditions)
		}
		filter.RemoveField(scan.filter, subType.Field)
	} else {
		var parentFilter *mapper.Filter
		scan.filter, parentFilter = filter.SplitByField(scan.filter, subType.Field)
		if parentFilter != nil {
			if parent.filter == nil {
				parent.filter = parentFilter
			} else {
				parent.filter.Conditions = filter.Merge(
					parent.filter.Conditions, parentFilter.Conditions)
			}
		}
	}
}

func (p *Planner) makeTypeJoinMany(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinMany, error) {
	prepareScanNodeFilterForTypeJoin(parent, source, subType)

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

	rootField, rootNameFound := subTypeCollectionDesc.GetFieldByRelation(
		subTypeFieldDesc.RelationName,
		parent.sourceInfo.collectionDescription.Name,
		subTypeFieldDesc.Name,
		&subTypeCollectionDesc.Schema,
	)

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
	// if there is, scan and aggregate results
	// if not, then manually scan the subtype table
	subDocs := make([]core.Doc, 0)
	if n.index != nil {
		// @todo: handle index for one-to-many setup
	} else {
		propIndex := n.subSelect.FirstIndexOfName(n.rootName + request.RelatedObjectID)
		// using the doc._key as a filter
		setSubTypeFilterToScanNode(n.subType, propIndex, n.currentValue.GetKey())

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

			subDoc := n.subType.Value()
			subDocs = append(subDocs, subDoc)
		}
	}

	n.currentValue.Fields[n.subSelect.Index] = subDocs
	return true, nil
}

func (n *typeJoinMany) Close() error {
	if err := n.root.Close(); err != nil {
		return err
	}

	return n.subType.Close()
}

func (n *typeJoinMany) Source() planNode { return n.root }

func setSubTypeFilterToScanNode(plan planNode, propIndex int, key string) {
	scan := getScanNode(plan)
	if scan == nil {
		return
	}

	if scan.filter == nil {
		scan.filter = mapper.NewFilter()
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: key,
		},
	}

	filter.RemoveField(scan.filter, mapper.Field{Index: propIndex})
	scan.filter.Conditions = filter.Merge(scan.filter.Conditions, filterConditions)
}

func getScanNode(plan planNode) *scanNode {
	node := plan
	for node != nil {
		scanNode, ok := node.(*scanNode)
		if ok {
			return scanNode
		}
		node = node.Source()
	}
	return nil
}
