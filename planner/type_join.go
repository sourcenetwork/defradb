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
			result["subTypeScan"] = subScanExplain
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
	twoWayFetchDirector
	p *Planner

	primary             bool
	secondaryFieldIndex immutable.Option[int]
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

	return &typeJoinOne{
		twoWayFetchDirector: twoWayFetchDirector{
			docMapper:   docMapper{parent.documentMapping},
			root:        source,
			subType:     selectPlan,
			subSelect:   subType,
			rootName:    subTypeField.Name,
			subTypeName: subType.Name,
		},
		p:                   p,
		primary:             isPrimary,
		secondaryFieldIndex: secondaryFieldIndex,
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

func (n *typeJoinOne) valuesSecondary(doc core.Doc) (core.Doc, error) {
	fieldName := n.rootName + request.RelatedObjectID
	subDocs, err := fetchDocsWithFieldValue(n.subType, fieldName, doc.GetKey(), 1)
	if err != nil {
		return core.Doc{}, err
	}

	if len(subDocs) > 0 {
		doc.Fields[n.subSelect.Index] = subDocs[0]
		if n.secondaryFieldIndex.HasValue() {
			doc.Fields[n.secondaryFieldIndex.Value()] = subDocs[0].GetKey()
		}
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

	// do a point lookup with the new span (index key)
	n.subType.Spans(core.NewSpans(core.NewSpan(subKeyIndexKey, subKeyIndexKey.PrefixEnd())))

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

type typeJoinMany struct {
	twoWayFetchDirector

	p *Planner
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

	return &typeJoinMany{
		twoWayFetchDirector: twoWayFetchDirector{
			docMapper:   docMapper{parent.documentMapping},
			root:        source,
			subType:     selectPlan,
			subSelect:   subType,
			rootName:    rootField.Name,
			subTypeName: subType.Name,
		},
		p: p,
	}, nil
}

func (n *typeJoinMany) Kind() string {
	return "typeJoinMany"
}

func fetchPrimaryDoc(node, subNode planNode, parentProp string) (bool, error) {
	subDoc := subNode.Value()
	ind := subNode.DocumentMap().FirstIndexOfName(parentProp)

	rootDocKey := base.MakeDocKey(node.(*scanNode).desc, subDoc.Fields[ind].(string))

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

type twoWayFetchDirector struct {
	documentIterator
	docMapper

	root        planNode
	subType     planNode
	rootName    string
	subTypeName string

	subSelect *mapper.Select

	isInverted bool
}

func (n *twoWayFetchDirector) Init() error {
	if err := n.subType.Init(); err != nil {
		return err
	}
	return n.root.Init()
}

func (n *twoWayFetchDirector) Start() error {
	if err := n.subType.Start(); err != nil {
		return err
	}
	return n.root.Start()
}

func (d *twoWayFetchDirector) Close() error {
	if err := d.root.Close(); err != nil {
		return err
	}

	return d.subType.Close()
}

func (n *twoWayFetchDirector) Spans(spans core.Spans) {
	n.root.Spans(spans)
}

func (n *twoWayFetchDirector) Source() planNode { return n.root }

func (d *twoWayFetchDirector) invert() {
	d.isInverted = !d.isInverted
}

func (d *twoWayFetchDirector) Next() (bool, error) {
	if d.isInverted {
		return d.fetchInverted()
	} else {
		return d.fetchDefault()
	}
}

func (d *twoWayFetchDirector) fetchDefault() (bool, error) {
	hasNext, err := d.root.Next()
	if err != nil || !hasNext {
		return hasNext, err
	}

	d.currentValue = d.root.Value()

	fieldName := d.rootName + request.RelatedObjectID
	subDocs, err := fetchDocsWithFieldValue(d.subType, fieldName, d.currentValue.GetKey(), 0)
	if err != nil {
		return false, err
	}

	d.currentValue.Fields[d.subSelect.Index] = subDocs
	return true, nil
}

func (d *twoWayFetchDirector) fetchInverted() (bool, error) {
	for {
		hasValue, err := d.subType.Next()

		if err != nil {
			return false, err
		}

		if !hasValue {
			return false, nil
		}

		hasPrimaryDoc, err := fetchPrimaryDoc(d.root, d.subType, d.rootName+request.RelatedObjectID)
		if err != nil {
			return false, err
		}

		if !hasPrimaryDoc {
			continue
		}

		d.currentValue = d.root.Value()

		doc := d.root.Value()
		subDoc := d.subType.Value()
		doc.Fields[d.subSelect.Index] = subDoc
		//if n.secondaryFieldIndex.HasValue() {
		//doc.Fields[n.secondaryFieldIndex.Value()] = subDoc.GetKey()
		//}

		return true, nil
	}
}

func (n *twoWayFetchDirector) invertJoinDirectionWithIndex(fieldFilter *mapper.Filter, field client.FieldDescription) error {
	subScan := getScanNode(n.subType)
	subScan.tryAddField(n.rootName + request.RelatedObjectID)
	subScan.filter = fieldFilter
	subScan.initFetcher(immutable.Option[string]{}, immutable.Some(field))

	n.invert()

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
