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
		if joinType.isSecondary {
			simpleExplainMap[joinDirectionLabel] = joinDirectionSecondaryLabel
		} else {
			simpleExplainMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		}

		// Add the attribute(s).
		simpleExplainMap[joinRootLabel] = joinType.rootName
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
		result := map[string]any{
			"iterations": n.execInfo.iterations,
		}
		var subScan *scanNode
		if joinMany, isJoinMany := n.joinPlan.(*typeJoinMany); isJoinMany {
			subScan = getScanNode(joinMany.subType)
		}
		if joinOne, isJoinOne := n.joinPlan.(*typeJoinOne); isJoinOne {
			subScan = getScanNode(joinOne.subType)
		}
		if subScan != nil {
			subScanExplain, err := subScan.Explain(explainType)
			if err != nil {
				return nil, err
			}
			result["subTypeScanNode"] = subScanExplain
		}
		return result, nil

	default:
		return nil, ErrUnknownExplainRequestType
	}
}

// Merge implements mergeNode
func (n *typeIndexJoin) Merge() bool { return true }

// typeJoinOne is the plan node for a type index join
// where the root type is the primary in a one-to-one relation request.
type typeJoinOne struct {
	invertibleTypeJoin
}

func (p *Planner) makeTypeJoinOne(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) (*typeJoinOne, error) {
	prepareScanNodeFilterForTypeJoin(parent, source, subType)

	selectPlan, err := p.Select(subType)
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

	subTypeCol, err := p.db.GetCollectionByName(p.ctx, subType.CollectionName)
	if err != nil {
		return nil, err
	}

	subTypeField, subTypeFieldNameFound := subTypeCol.Description().GetFieldByRelation(
		subTypeFieldDesc.RelationName,
		parent.sourceInfo.collectionDescription.Name,
		subTypeFieldDesc.Name,
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

	dir := joinDirection{
		firstNode:      source,
		secondNode:     selectPlan,
		secondaryField: subTypeField.Name + request.RelatedObjectID,
		primaryField:   subTypeFieldDesc.Name + request.RelatedObjectID,
	}

	return &typeJoinOne{
		invertibleTypeJoin: invertibleTypeJoin{
			docMapper:           docMapper{parent.documentMapping},
			root:                source,
			subType:             selectPlan,
			subSelect:           subType,
			rootName:            subTypeField.Name,
			subTypeName:         subType.Name,
			isSecondary:         !isPrimary,
			secondaryFieldIndex: secondaryFieldIndex,
			secondaryFetchLimit: 1,
			dir:                 dir,
		},
	}, nil
}

func (n *typeJoinOne) Kind() string {
	return "typeJoinOne"
}

func fetchDocsWithFieldValue(plan planNode, fieldName string, val any, limit uint) ([]core.Doc, error) {
	propIndex := plan.DocumentMap().FirstIndexOfName(fieldName)
	setSubTypeFilterToScanNode(plan, propIndex, val)

	if err := plan.Init(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	docs := make([]core.Doc, 0, limit)
	for {
		next, err := plan.Next()
		if err != nil {
			return nil, err
		}
		if !next {
			break
		}

		docs = append(docs, plan.Value())

		if limit > 0 && len(docs) >= int(limit) {
			break
		}
	}

	return docs, nil
}

type typeJoinMany struct {
	invertibleTypeJoin
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

	selectPlan, err := p.Select(subType)
	if err != nil {
		return nil, err
	}

	subTypeFieldDesc, ok := parent.sourceInfo.collectionDescription.Schema.GetField(subType.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subType.Name)
	}

	subTypeCol, err := p.db.GetCollectionByName(p.ctx, subType.CollectionName)
	if err != nil {
		return nil, err
	}

	rootField, rootNameFound := subTypeCol.Description().GetFieldByRelation(
		subTypeFieldDesc.RelationName,
		parent.sourceInfo.collectionDescription.Name,
		subTypeFieldDesc.Name,
	)

	if !rootNameFound {
		return nil, client.NewErrFieldNotExist(subTypeFieldDesc.RelationName)
	}

	dir := joinDirection{
		firstNode:      source,
		secondNode:     selectPlan,
		secondaryField: rootField.Name + request.RelatedObjectID,
		primaryField:   subTypeFieldDesc.Name + request.RelatedObjectID,
	}

	return &typeJoinMany{
		invertibleTypeJoin: invertibleTypeJoin{
			docMapper:           docMapper{parent.documentMapping},
			root:                source,
			subType:             selectPlan,
			subSelect:           subType,
			rootName:            rootField.Name,
			isSecondary:         true,
			subTypeName:         subType.Name,
			secondaryFetchLimit: 0,
			dir:                 dir,
		},
	}, nil
}

func (n *typeJoinMany) Kind() string {
	return "typeJoinMany"
}

func fetchPrimaryDoc(node, subNode planNode, parentProp string) (bool, error) {
	subDoc := subNode.Value()
	ind := subNode.DocumentMap().FirstIndexOfName(parentProp)

	docKeyStr, isStr := subDoc.Fields[ind].(string)
	if !isStr {
		return false, nil
	}

	scan := getScanNode(node)
	if scan == nil {
		return false, nil
	}
	rootDocKey := base.MakeDocKey(scan.desc, docKeyStr)

	spans := core.NewSpans(core.NewSpan(rootDocKey, rootDocKey.PrefixEnd()))

	node.Spans(spans)

	if err := node.Init(); err != nil {
		return false, NewErrSubTypeInit(err)
	}

	hasValue, err := node.Next()

	if err != nil || !hasValue {
		return false, err
	}

	return true, nil
}

type joinDirection struct {
	firstNode      planNode
	secondNode     planNode
	secondaryField string
	primaryField   string
	isInverted     bool
}

func (dir *joinDirection) invert() {
	dir.isInverted = !dir.isInverted
	dir.firstNode, dir.secondNode = dir.secondNode, dir.firstNode
	dir.secondaryField, dir.primaryField = dir.primaryField, dir.secondaryField
}

type invertibleTypeJoin struct {
	documentIterator
	docMapper

	root        planNode
	subType     planNode
	rootName    string
	subTypeName string

	subSelect *mapper.Select

	isSecondary         bool
	secondaryFieldIndex immutable.Option[int]
	secondaryFetchLimit uint

	dir joinDirection
}

func (join *invertibleTypeJoin) replaceRoot(node planNode) {
	join.root = node
	if join.dir.isInverted {
		join.dir.secondNode = node
	} else {
		join.dir.firstNode = node
	}
}

func (join *invertibleTypeJoin) Init() error {
	if err := join.subType.Init(); err != nil {
		return err
	}
	return join.root.Init()
}

func (join *invertibleTypeJoin) Start() error {
	if err := join.subType.Start(); err != nil {
		return err
	}
	return join.root.Start()
}

func (join *invertibleTypeJoin) Close() error {
	if err := join.root.Close(); err != nil {
		return err
	}

	return join.subType.Close()
}

func (join *invertibleTypeJoin) Spans(spans core.Spans) {
	join.root.Spans(spans)
}

func (join *invertibleTypeJoin) Source() planNode { return join.root }

func (tj *invertibleTypeJoin) invert() {
	tj.dir.invert()
	tj.isSecondary = !tj.isSecondary
}

func (join *invertibleTypeJoin) processSecondResult(secondDocs []core.Doc) (any, any) {
	var secondResult any
	var secondIDResult any
	if join.secondaryFetchLimit == 1 {
		if len(secondDocs) != 0 {
			secondResult = secondDocs[0]
			secondIDResult = secondDocs[0].GetKey()
		}
	} else {
		secondResult = secondDocs
		secondDocKeys := make([]string, len(secondDocs))
		for i, doc := range secondDocs {
			secondDocKeys[i] = doc.GetKey()
		}
		secondIDResult = secondDocKeys
	}
	join.root.Value().Fields[join.subSelect.Index] = secondResult
	if join.secondaryFieldIndex.HasValue() {
		join.root.Value().Fields[join.secondaryFieldIndex.Value()] = secondIDResult
	}
	return secondResult, secondIDResult
}

func (join *invertibleTypeJoin) Next() (bool, error) {
	hasFirstValue, err := join.dir.firstNode.Next()

	if err != nil || !hasFirstValue {
		return false, err
	}

	firstDoc := join.dir.firstNode.Value()

	if join.isSecondary {
		secondDocs, err := fetchDocsWithFieldValue(join.dir.secondNode, join.dir.secondaryField, firstDoc.GetKey(), join.secondaryFetchLimit)
		if err != nil {
			return false, err
		}
		if join.dir.secondNode == join.root {
			join.root.Value().Fields[join.subSelect.Index] = join.subType.Value()
		} else {
			secondResult, secondIDResult := join.processSecondResult(secondDocs)
			join.dir.firstNode.Value().Fields[join.subSelect.Index] = secondResult
			if join.secondaryFieldIndex.HasValue() {
				join.dir.firstNode.Value().Fields[join.secondaryFieldIndex.Value()] = secondIDResult
			}
		}
	} else {
		hasDoc, err := fetchPrimaryDoc(join.dir.secondNode, join.dir.firstNode, join.dir.primaryField)
		if err != nil {
			return false, err
		}

		if hasDoc {
			join.root.Value().Fields[join.subSelect.Index] = join.subType.Value()
		}
	}

	join.currentValue = join.root.Value()

	return true, nil
}

func (join *invertibleTypeJoin) invertJoinDirectionWithIndex(fieldFilter *mapper.Filter, field client.FieldDescription) error {
	subScan := getScanNode(join.subType)
	subScan.tryAddField(join.rootName + request.RelatedObjectID)
	subScan.filter = fieldFilter
	subScan.initFetcher(immutable.Option[string]{}, immutable.Some(field))

	join.invert()

	return nil
}

func setSubTypeFilterToScanNode(plan planNode, propIndex int, val any) {
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
			mapper.FilterEqOp: val,
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
		if node == nil {
			if topSelect, ok := plan.(*selectTopNode); ok {
				node = topSelect.selectNode
			}
		}
	}
	return nil
}
