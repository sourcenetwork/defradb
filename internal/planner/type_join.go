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
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
	"github.com/sourcenetwork/defradb/internal/db/base"
	"github.com/sourcenetwork/defradb/internal/planner/filter"
	"github.com/sourcenetwork/defradb/internal/planner/mapper"
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

	typeFieldDesc, ok := parent.collection.Definition().GetFieldByName(subType.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subType.Name)
	}

	if typeFieldDesc.Kind.IsObject() && !typeFieldDesc.Kind.IsArray() { // One-to-One, or One side of One-to-Many
		joinPlan, err = p.makeTypeJoinOne(parent, source, subType)
	} else if typeFieldDesc.Kind.IsObjectArray() { // Many side of One-to-Many
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

	addExplainData := func(j *invertibleTypeJoin) error {
		// Add the attribute(s).
		simpleExplainMap[joinRootLabel] = immutable.Some(j.getRootTypeName())
		simpleExplainMap[joinSubTypeNameLabel] = j.getSubTypeName()

		subTypeExplainGraph, err := buildSimpleExplainGraph(j.childPlan)
		if err != nil {
			return err
		}

		// Add the joined (subType) type's entire explain graph.
		simpleExplainMap[joinSubTypeLabel] = subTypeExplainGraph
		return nil
	}

	var err error
	switch joinType := n.joinPlan.(type) {
	case *typeJoinOne:
		// Add the direction attribute.
		if joinType.isChildSecondary {
			simpleExplainMap[joinDirectionLabel] = joinDirectionSecondaryLabel
		} else {
			simpleExplainMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		}

		err = addExplainData(&joinType.invertibleTypeJoin)

	case *typeJoinMany:
		err = addExplainData(&joinType.invertibleTypeJoin)

	default:
		err = client.NewErrUnhandledType("join plan", n.joinPlan)
	}

	return simpleExplainMap, err
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
			subScan = getScanNode(joinMany.childPlan)
		}
		if joinOne, isJoinOne := n.joinPlan.(*typeJoinOne); isJoinOne {
			subScan = getScanNode(joinOne.childPlan)
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

// typeJoinOne is the plan node for a type index join
// where the root type is the primary in a one-to-one relation request.
type typeJoinOne struct {
	invertibleTypeJoin
}

func (p *Planner) makeTypeJoinOne(
	parent *selectNode,
	sourcePlan planNode,
	subSelect *mapper.Select,
) (*typeJoinOne, error) {
	prepareScanNodeFilterForTypeJoin(parent, sourcePlan, subSelect)

	subSelectPlan, err := p.Select(subSelect)
	if err != nil {
		return nil, err
	}

	// get the correct sub field schema type (collection)
	parentsRelFieldDef, ok := parent.collection.Definition().GetFieldByName(subSelect.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subSelect.Name)
	}

	subCol, err := p.db.GetCollectionByName(p.ctx, subSelect.CollectionName)
	if err != nil {
		return nil, err
	}

	childsRelFieldDesc, ok := subCol.Description().GetFieldByRelation(
		parentsRelFieldDef.RelationName,
		parent.collection.Name().Value(),
		parentsRelFieldDef.Name,
	)
	if !ok {
		return nil, client.NewErrFieldNotExist(parentsRelFieldDef.RelationName)
	}

	childsRelFieldDef, ok := subCol.Definition().GetFieldByName(childsRelFieldDesc.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subSelect.Name)
	}

	var secondaryFieldIndex immutable.Option[int]
	if !parentsRelFieldDef.IsPrimaryRelation {
		idFieldName := parentsRelFieldDef.Name
		secondaryFieldIndex = immutable.Some(
			parent.documentMapping.FirstIndexOfName(idFieldName + request.RelatedObjectID),
		)
	}

	dir := joinDirection{
		firstNode:               sourcePlan,
		secondNode:              subSelectPlan,
		firstNodesRelFieldName:  parentsRelFieldDef.Name,
		secondNodesRelFieldName: childsRelFieldDesc.Name,
		isInvertable:            true,
	}

	return &typeJoinOne{
		invertibleTypeJoin: invertibleTypeJoin{
			docMapper:           docMapper{parent.documentMapping},
			parentPlan:          sourcePlan,
			childPlan:           subSelectPlan,
			childSelect:         subSelect,
			parentsRelFieldDef:  parentsRelFieldDef,
			childsRelFieldDef:   childsRelFieldDef,
			parentCollection:    parent.collection,
			isChildSecondary:    !parentsRelFieldDef.IsPrimaryRelation,
			secondaryFieldIndex: secondaryFieldIndex,
			secondaryFetchLimit: 1,
			dir:                 dir,
		},
	}, nil
}

func (n *typeJoinOne) Kind() string {
	return "typeJoinOne"
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
		scan.filter = nil
	} else {
		var parentFilter *mapper.Filter
		scan.filter, parentFilter = filter.SplitByFields(scan.filter, subType.Field)
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
	sourcePlan planNode,
	subSelect *mapper.Select,
) (*typeJoinMany, error) {
	prepareScanNodeFilterForTypeJoin(parent, sourcePlan, subSelect)

	subSelectPlan, err := p.Select(subSelect)
	if err != nil {
		return nil, err
	}

	parentsRelFieldDef, ok := parent.collection.Definition().GetFieldByName(subSelect.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subSelect.Name)
	}

	subCol, err := p.db.GetCollectionByName(p.ctx, subSelect.CollectionName)
	if err != nil {
		return nil, err
	}

	childsRelFieldDesc, ok := subCol.Description().GetFieldByRelation(
		parentsRelFieldDef.RelationName,
		parent.collection.Name().Value(),
		parentsRelFieldDef.Name,
	)
	if !ok {
		return nil, client.NewErrFieldNotExist(subSelect.Name)
	}

	childsRelFieldDef, ok := subCol.Definition().GetFieldByName(childsRelFieldDesc.Name)
	if !ok {
		return nil, client.NewErrFieldNotExist(subSelect.Name)
	}

	dir := joinDirection{
		firstNode:              sourcePlan,
		secondNode:             subSelectPlan,
		firstNodesRelFieldName: parentsRelFieldDef.Name,
	}

	if parentsRelFieldDef.RelationName != "" {
		rootField, ok := subCol.Description().GetFieldByRelation(
			parentsRelFieldDef.RelationName,
			parent.collection.Name().Value(),
			parentsRelFieldDef.Name,
		)
		if ok {
			dir.secondNodesRelFieldName = rootField.Name
			dir.isInvertable = true
		}
	}

	return &typeJoinMany{
		invertibleTypeJoin: invertibleTypeJoin{
			docMapper:           docMapper{parent.documentMapping},
			parentPlan:          sourcePlan,
			childPlan:           subSelectPlan,
			childSelect:         subSelect,
			parentsRelFieldDef:  parentsRelFieldDef,
			childsRelFieldDef:   childsRelFieldDef,
			parentCollection:    parent.collection,
			isChildSecondary:    true,
			secondaryFetchLimit: 0,
			dir:                 dir,
		},
	}, nil
}

func (n *typeJoinMany) Kind() string {
	return "typeJoinMany"
}

func getPrimaryDocIDFromSecondaryDoc(subNode planNode, parentProp string) string {
	subDoc := subNode.Value()
	ind := subNode.DocumentMap().FirstIndexOfName(parentProp + request.RelatedObjectID)

	docIDStr, _ := subDoc.Fields[ind].(string)
	return docIDStr
}

func fetchPrimaryDoc(node, subNode planNode, parentProp string) (bool, error) {
	docIDStr := getPrimaryDocIDFromSecondaryDoc(subNode, parentProp)
	if docIDStr == "" {
		return false, nil
	}

	scan := getScanNode(node)
	if scan == nil {
		return false, nil
	}
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(scan.col.Description(), docIDStr)

	spans := core.NewSpans(core.NewSpan(dsKey, dsKey.PrefixEnd()))

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

// joinDirection is a struct that holds the two nodes that are that are executed one after another
// depending on the direction of the join.
type joinDirection struct {
	// firstNode is the node that is executed first (usually an indexed collection).
	firstNode planNode
	// secondNode is the node that is executed second.
	secondNode planNode
	// secondNodesRelFieldName is a field name of a secondary doc that refers to the primary docID (like author_id).
	secondNodesRelFieldName string
	// firstNodesRelFieldName is a field name of the primary doc that refers to the secondary docID (like author_id).
	firstNodesRelFieldName string
	// isInvertable indicates if the join can be inverted.
	isInvertable bool
	// isInverted indicates if the join direction is inverted.
	isInverted bool
}

func (dir *joinDirection) invert() {
	dir.isInverted = !dir.isInverted
	dir.firstNode, dir.secondNode = dir.secondNode, dir.firstNode
	dir.secondNodesRelFieldName, dir.firstNodesRelFieldName = dir.firstNodesRelFieldName, dir.secondNodesRelFieldName
}

type invertibleTypeJoin struct {
	docMapper

	parentPlan planNode
	childPlan  planNode

	childSelect        *mapper.Select
	parentsRelFieldDef client.FieldDefinition
	childsRelFieldDef  client.FieldDefinition
	parentCollection   client.Collection

	isChildSecondary    bool
	secondaryFieldIndex immutable.Option[int]
	secondaryFetchLimit uint

	// docsToYield contains documents read and ready to be yielded by this node.
	docsToYield []core.Doc

	dir joinDirection
}

func (join *invertibleTypeJoin) getRootTypeName() string {
	if join.dir.isInverted {
		return join.dir.firstNodesRelFieldName
	}
	return join.dir.secondNodesRelFieldName
}

func (join *invertibleTypeJoin) getSubTypeName() string {
	if join.dir.isInverted {
		return join.dir.secondNodesRelFieldName
	}
	return join.dir.firstNodesRelFieldName
}

func (join *invertibleTypeJoin) replaceRoot(node planNode) {
	join.parentPlan = node
	if join.dir.isInverted {
		join.dir.secondNode = node
	} else {
		join.dir.firstNode = node
	}
}

func (join *invertibleTypeJoin) Init() error {
	if err := join.childPlan.Init(); err != nil {
		return err
	}
	return join.parentPlan.Init()
}

func (join *invertibleTypeJoin) Start() error {
	if err := join.childPlan.Start(); err != nil {
		return err
	}
	return join.parentPlan.Start()
}

func (join *invertibleTypeJoin) Close() error {
	if err := join.parentPlan.Close(); err != nil {
		return err
	}

	return join.childPlan.Close()
}

func (join *invertibleTypeJoin) Spans(spans core.Spans) {
	join.parentPlan.Spans(spans)
}

func (join *invertibleTypeJoin) Source() planNode { return join.parentPlan }

func (tj *invertibleTypeJoin) invert() {
	tj.dir.invert()
	tj.isChildSecondary = !tj.isChildSecondary
}

// addSecondaryDocsToRootPrimaryDoc adds the secondary docs to the root primary doc.
// If the relations is 1-to-1 a single secondary doc will be added to the root primary doc.
// Otherwise, all secondary docs will be added as an array.
func (join *invertibleTypeJoin) addSecondaryDocsToRootPrimaryDoc(secondDocs []core.Doc) {
	var secondaryResult any
	var secondaryIDResult any
	if join.secondaryFetchLimit == 1 {
		if len(secondDocs) != 0 {
			secondaryResult = secondDocs[0]
			secondaryIDResult = secondDocs[0].GetID()
		}
	} else {
		secondaryResult = secondDocs
		secondDocIDs := make([]string, len(secondDocs))
		for i, doc := range secondDocs {
			secondDocIDs[i] = doc.GetID()
		}
		secondaryIDResult = secondDocIDs
	}
	join.parentPlan.Value().Fields[join.childSelect.Index] = secondaryResult
	if join.secondaryFieldIndex.HasValue() {
		join.parentPlan.Value().Fields[join.secondaryFieldIndex.Value()] = secondaryIDResult
	}
}

type secondaryDocsJoiner struct {
	node          planNode
	primaryDoc    core.Doc
	fieldDef      client.FieldDefinition
	relIDFieldDef client.FieldDefinition
	parentPlan    planNode
	childPlan     planNode

	scan           *scanNode
	origScanFields []client.FieldDefinition

	resultSecondaryDocIDs []string
	resultSecondaryDocs   []core.Doc
}

func newSecondaryDocsJoiner(
	node planNode,
	primaryDoc core.Doc,
	fieldDef client.FieldDefinition,
	relIDFieldDef client.FieldDefinition,
) secondaryDocsJoiner {
	j := secondaryDocsJoiner{
		node:          node,
		primaryDoc:    primaryDoc,
		fieldDef:      fieldDef,
		relIDFieldDef: relIDFieldDef,
	}
	return j
}

func (j *secondaryDocsJoiner) fetchSecondaryDocsForPrimaryDoc() error {
	err := j.fetchSecondaryDocIDs()

	if err != nil {
		return err
	}

	docs, err := j.fetchSecondaryDocsByIDs()

	if err != nil {
		return err
	}

	j.join(docs)

	return nil
}

func (j *secondaryDocsJoiner) prepareScanNodeForDocIDsFetching() error {
	j.scan = getScanNode(j.node)
	if err := j.scan.fetcher.Close(); err != nil {
		return NewErrSubTypeInit(err)
	}
	j.scan.filter = nil
	j.origScanFields = j.scan.fields
	j.scan.fields = []client.FieldDefinition{j.relIDFieldDef}
	j.scan.spans = core.Spans{}
	return nil
}

func (j *secondaryDocsJoiner) collectDocs(numDocs int) ([]core.Doc, error) {
	if err := j.node.Init(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	docs := make([]core.Doc, 0, numDocs)

	for {
		hasValue, err := j.node.Next()

		if err != nil {
			return nil, err
		}

		if !hasValue {
			break
		}

		docs = append(docs, j.node.Value())
	}

	if err := j.node.Close(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	return docs, nil
}

func (j *secondaryDocsJoiner) fetchSecondaryDocIDs() error {
	err := j.prepareScanNodeForDocIDsFetching()
	if err != nil {
		return err
	}

	relIDPropIndex := j.node.DocumentMap().FirstIndexOfName(j.relIDFieldDef.Name)
	setSubTypeFilterToScanNode(j.node, relIDPropIndex, j.primaryDoc.GetID())

	indexOnRelation := findIndexByFieldName(j.scan.col, j.relIDFieldDef.Name)
	j.scan.initFetcher(immutable.None[string](), indexOnRelation)

	docs, err := j.collectDocs(0)

	if err != nil {
		return err
	}

	for i := range docs {
		j.resultSecondaryDocIDs = append(j.resultSecondaryDocIDs, docs[i].Fields[core.DocIDFieldIndex].(string))
	}

	return nil
}

func (j *secondaryDocsJoiner) spansFromDocIDs() core.Spans {
	colRootID := j.scan.col.Description().RootID

	spans := make([]core.Span, 0, len(j.resultSecondaryDocIDs))
	for _, secondaryID := range j.resultSecondaryDocIDs {
		dsKey := core.DataStoreKey{CollectionRootID: colRootID, DocID: secondaryID}
		spans = append(spans, core.NewSpan(dsKey, dsKey.PrefixEnd()))
	}
	return core.NewSpans(spans...)
}

func (j *secondaryDocsJoiner) prepareScanNodeForDocsFetching() error {
	if err := j.scan.fetcher.Close(); err != nil {
		return NewErrSubTypeInit(err)
	}

	j.scan.filter = nil
	for i := range j.origScanFields {
		if j.origScanFields[i].Name == j.relIDFieldDef.Name {
			j.origScanFields = append(j.origScanFields[:i], j.origScanFields[i+1:]...)
			break
		}
	}
	j.scan.fields = j.origScanFields
	return nil
}

func (j *secondaryDocsJoiner) fetchSecondaryDocsByIDs() ([]core.Doc, error) {
	err := j.prepareScanNodeForDocsFetching()

	if err != nil {
		return nil, err
	}

	j.scan.initFetcher(immutable.None[string](), immutable.None[client.IndexDescription]())
	j.node.Spans(j.spansFromDocIDs())

	return j.collectDocs(len(j.resultSecondaryDocIDs))
}

func (j *secondaryDocsJoiner) join(docs []core.Doc) {
	relPropIndex := j.node.DocumentMap().FirstIndexOfName(j.fieldDef.Name)

	for i := range docs {
		docs[i].Fields[relPropIndex] = j.primaryDoc
	}

	j.resultSecondaryDocs = docs
}

func fetchDocsWithFieldValue(plan planNode, fieldName string, val any) ([]core.Doc, error) {
	propIndex := plan.DocumentMap().FirstIndexOfName(fieldName)
	setSubTypeFilterToScanNode(plan, propIndex, val)

	if err := plan.Init(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	var docs []core.Doc
	for {
		next, err := plan.Next()
		if err != nil {
			return nil, err
		}
		if !next {
			break
		}

		docs = append(docs, plan.Value())
	}

	return docs, nil
}

func (j *secondaryDocsJoiner) fetchSecondaryFullScan() error {
	secondDocs, err := fetchDocsWithFieldValue(
		j.node,
		// As the join is from the secondary field, we know that [join.dir.secondaryField] must have a value
		// otherwise the user would not have been able to request it.
		j.relIDFieldDef.Name,
		j.primaryDoc.GetID(),
	)
	if err != nil {
		return err
	}
	if j.node == j.parentPlan {
		if len(secondDocs) == 0 {
			return nil
		}
		for i := range secondDocs {
			//secondDocs[i].Fields[join.subSelect.Index] = j.subType.Value()
			i = i
		}
		j.resultSecondaryDocs = append(j.resultSecondaryDocs, secondDocs...)
		return nil
	} else {
		//j.addSecondaryDocsToRootPrimaryDoc(secondDocs)
		j.resultSecondaryDocs = append(j.resultSecondaryDocs, j.parentPlan.Value())
	}
	return nil
}

func (j *secondaryDocsJoiner) fetchDocsWithFieldValue() error {
	propIndex := j.node.DocumentMap().FirstIndexOfName(j.relIDFieldDef.Name)
	setSubTypeFilterToScanNode(j.node, propIndex, j.primaryDoc.GetID())

	if err := j.node.Init(); err != nil {
		return NewErrSubTypeInit(err)
	}

	j.resultSecondaryDocs = make([]core.Doc, 0, len(j.resultSecondaryDocIDs))
	for {
		hasValue, err := j.node.Next()
		if err != nil {
			return err
		}
		if !hasValue {
			break
		}

		j.resultSecondaryDocs = append(j.resultSecondaryDocs, j.node.Value())
	}

	return nil
}

func (join *invertibleTypeJoin) fetchSecondaryDocsForPrimaryDoc(node planNode, primaryDoc core.Doc) (bool, error) {
	relIDFieldDef, ok := join.parentCollection.Definition().GetFieldByName(join.childSelect.Name + request.RelatedObjectID)
	if !ok {
		return false, client.NewErrFieldNotExist(join.childSelect.Name + request.RelatedObjectID)
	}

	secJoiner := newSecondaryDocsJoiner(node, primaryDoc, join.parentsRelFieldDef, relIDFieldDef)
	secJoiner.parentPlan = join.parentPlan
	secJoiner.childPlan = join.childPlan
	err := secJoiner.fetchSecondaryDocsForPrimaryDoc()
	join.docsToYield = append(join.docsToYield, secJoiner.resultSecondaryDocs...)
	return len(secJoiner.resultSecondaryDocs) > 0, err
}

func (join *invertibleTypeJoin) Next() (bool, error) {
	if len(join.docsToYield) > 0 {
		// If there is one or more documents in the queue, drop the first one -
		// it will have been yielded by the last `Next()` call.
		join.docsToYield = join.docsToYield[1:]
		if len(join.docsToYield) > 0 {
			// If there are still documents in the queue, return true yielding the next
			// one in the queue.
			return true, nil
		}
	}

	hasFirstValue, err := join.dir.firstNode.Next()

	if err != nil || !hasFirstValue {
		return false, err
	}

	if join.isChildSecondary {
		firstDoc := join.dir.firstNode.Value()
		return join.fetchSecondaryDocsForPrimaryDoc(join.dir.secondNode, firstDoc)
	} else {
		hasDoc, err := fetchPrimaryDoc(join.dir.secondNode, join.dir.firstNode, join.dir.firstNodesRelFieldName)
		if err != nil {
			return false, err
		}

		if hasDoc {
			//join.root.Value().Fields[join.subSelect.Index] = join.fetchIndexedRelatedSecondaryDocs(join.dir.firstNode, join.dir.topRelField)
			join.parentPlan.Value().Fields[join.childSelect.Index] = join.childPlan.Value()
		}

		join.docsToYield = append(join.docsToYield, join.parentPlan.Value())
	}

	return true, nil
}

func (join *invertibleTypeJoin) fetchIndexedRelatedSecondaryDocs(node planNode, field string) []any {
	docIDStr := getPrimaryDocIDFromSecondaryDoc(node, field)
	if docIDStr == "" {
		return nil
	}
	var secondaryIDs []string
	var spans []core.Span
	scan := getScanNode(node)

	scan.initFetcher(immutable.None[string](), findIndexByFieldName(scan.col, field))

	colRootID := scan.col.Description().RootID

	for _, secondaryID := range secondaryIDs {
		dsKey := core.DataStoreKey{CollectionRootID: colRootID, DocID: secondaryID}
		spans = append(spans, core.NewSpan(dsKey, dsKey.PrefixEnd()))
	}
	node.Spans(core.NewSpans(spans...))

	if err := node.Init(); err != nil {
		//return false, NewErrSubTypeInit(err)
		return nil
	}

	secondaryDocs := make([]core.Doc, 0, len(secondaryIDs))

	for {
		hasValue, err := node.Next()

		if err != nil {
			//return false, err
			return nil
		}

		if !hasValue {
			break
		}

		secondaryDocs = append(secondaryDocs, node.Value())
	}

	if err := node.Close(); err != nil {
		//return false, NewErrSubTypeInit(err)
		return nil
	}

	//return true, nil
	//subTypeFieldDesc, ok := parent.collection.Definition().GetFieldByName(subType.Name)
	join.parentPlan.Value().Fields[join.childSelect.Index] = secondaryDocs
	if join.secondaryFieldIndex.HasValue() {
		join.parentPlan.Value().Fields[join.secondaryFieldIndex.Value()] = secondaryIDs
	}
	return nil
}

func (join *invertibleTypeJoin) Value() core.Doc {
	if len(join.docsToYield) == 0 {
		return core.Doc{}
	}
	return join.docsToYield[0]
}

func (join *invertibleTypeJoin) invertJoinDirectionWithIndex(
	fieldFilter *mapper.Filter,
	index client.IndexDescription,
) error {
	if !join.dir.isInvertable {
		return nil
	}
	childsRelFieldName := join.dir.secondNodesRelFieldName
	if join.dir.isInverted {
		childsRelFieldName = join.dir.firstNodesRelFieldName
	}
	subScan := getScanNode(join.childPlan)
	subScan.tryAddField(childsRelFieldName + request.RelatedObjectID)
	subScan.filter = fieldFilter
	subScan.initFetcher(immutable.Option[string]{}, immutable.Some(index))

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
