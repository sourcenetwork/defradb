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

		subTypeExplainGraph, err := buildSimpleExplainGraph(j.childSide.plan)
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
		if joinType.parentSide.isPrimary() {
			simpleExplainMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		} else {
			simpleExplainMap[joinDirectionLabel] = joinDirectionSecondaryLabel
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
			subScan = getScanNode(joinMany.childSide.plan)
		}
		if joinOne, isJoinOne := n.joinPlan.(*typeJoinOne); isJoinOne {
			subScan = getScanNode(joinOne.childSide.plan)
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
	invertibleTypeJoin, err := p.newInvertableTypeJoin(parent, sourcePlan, subSelect)
	if err != nil {
		return nil, err
	}
	invertibleTypeJoin.secondaryFetchLimit = 1
	return &typeJoinOne{invertibleTypeJoin: invertibleTypeJoin}, nil
}

func (n *typeJoinOne) Kind() string {
	return "typeJoinOne"
}

type typeJoinMany struct {
	invertibleTypeJoin
}

func (p *Planner) makeTypeJoinMany(
	parent *selectNode,
	sourcePlan planNode,
	subSelect *mapper.Select,
) (*typeJoinMany, error) {
	invertibleTypeJoin, err := p.newInvertableTypeJoin(parent, sourcePlan, subSelect)
	if err != nil {
		return nil, err
	}
	invertibleTypeJoin.secondaryFetchLimit = 0
	return &typeJoinMany{invertibleTypeJoin: invertibleTypeJoin}, nil
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

func (p *Planner) newInvertableTypeJoin(
	parent *selectNode,
	sourcePlan planNode,
	subSelect *mapper.Select,
) (invertibleTypeJoin, error) {
	prepareScanNodeFilterForTypeJoin(parent, sourcePlan, subSelect)

	subSelectPlan, err := p.Select(subSelect)
	if err != nil {
		return invertibleTypeJoin{}, err
	}

	parentsRelFieldDef, ok := parent.collection.Definition().GetFieldByName(subSelect.Name)
	if !ok {
		return invertibleTypeJoin{}, client.NewErrFieldNotExist(subSelect.Name)
	}

	skipChild := false
	for _, field := range parent.selectReq.Fields {
		if field.GetName() == subSelect.Name {
			if childSelect, ok := field.AsSelect(); ok {
				if childSelect.SkipResolve {
					skipChild = true
				}
			}
			break
		}
	}

	subCol, err := p.db.GetCollectionByName(p.ctx, subSelect.CollectionName)
	if err != nil {
		return invertibleTypeJoin{}, err
	}

	childsRelFieldDesc, ok := subCol.Description().GetFieldByRelation(
		parentsRelFieldDef.RelationName,
		parent.collection.Name().Value(),
		parentsRelFieldDef.Name,
	)
	if !ok {
		return invertibleTypeJoin{}, client.NewErrFieldNotExist(parentsRelFieldDef.Name)
	}

	childsRelFieldDef, ok := subCol.Definition().GetFieldByName(childsRelFieldDesc.Name)
	if !ok {
		return invertibleTypeJoin{}, client.NewErrFieldNotExist(subSelect.Name)
	}

	parentSide := joinSide{
		plan:             sourcePlan,
		relFieldDef:      parentsRelFieldDef,
		relFieldMapIndex: immutable.Some(subSelect.Index),
		col:              parent.collection,
		requestedFields:  getRequestedFields(sourcePlan),
		isFirst:          true,
		isParent:         true,
	}

	ind := parent.documentMapping.IndexesByName[parentsRelFieldDef.Name+request.RelatedObjectID]
	if len(ind) > 0 {
		parentSide.relIDFieldMapIndex = immutable.Some(ind[0])
	}

	childSide := joinSide{
		plan:            subSelectPlan,
		relFieldDef:     childsRelFieldDef,
		col:             subCol,
		requestedFields: getRequestedFields(subSelectPlan),
		isFirst:         false,
		isParent:        false,
	}

	ind = subSelectPlan.DocumentMap().IndexesByName[childsRelFieldDef.Name+request.RelatedObjectID]
	if len(ind) > 0 {
		childSide.relIDFieldMapIndex = immutable.Some(ind[0])
	}

	return invertibleTypeJoin{
		docMapper:   docMapper{parent.documentMapping},
		parentSide:  parentSide,
		childSide:   childSide,
		childSelect: subSelect,
		skipChild:   skipChild,
	}, nil
}

func getRequestedFields(sourcePlan planNode) []string {
	scan := getScanNode(sourcePlan)
	if scan == nil {
		return nil
	}
	fields := make([]string, len(scan.fields))
	for i := range scan.fields {
		fields[i] = scan.fields[i].Name
	}
	return fields
}

type joinSide struct {
	plan               planNode
	relFieldDef        client.FieldDefinition
	relFieldMapIndex   immutable.Option[int]
	relIDFieldMapIndex immutable.Option[int]
	col                client.Collection
	requestedFields    []string
	isFirst            bool
	isParent           bool
}

func (s *joinSide) isPrimary() bool {
	return s.relFieldDef.IsPrimaryRelation
}

func (join *invertibleTypeJoin) getFirstSide() *joinSide {
	if join.parentSide.isFirst {
		return &join.parentSide
	}
	return &join.childSide
}

func (join *invertibleTypeJoin) getSecondSide() *joinSide {
	if !join.parentSide.isFirst {
		return &join.parentSide
	}
	return &join.childSide
}

func (join *invertibleTypeJoin) getPrimarySide() *joinSide {
	if join.parentSide.isPrimary() {
		return &join.parentSide
	}
	return &join.childSide
}

func (join *invertibleTypeJoin) getSecondarySide() *joinSide {
	if !join.parentSide.isPrimary() {
		return &join.parentSide
	}
	return &join.childSide
}

func (n *typeJoinMany) Kind() string {
	return "typeJoinMany"
}

func getForeignKey(node planNode, relFieldName string) string {
	ind := node.DocumentMap().FirstIndexOfName(relFieldName + request.RelatedObjectID)
	docIDStr, _ := node.Value().Fields[ind].(string)
	return docIDStr
}

func fetchDocWithID(node planNode, docID string) (bool, error) {
	scan := getScanNode(node)
	if scan == nil {
		return false, nil
	}
	dsKey := base.MakeDataStoreKeyWithCollectionAndDocID(scan.col.Description(), docID)

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

type invertibleTypeJoin struct {
	docMapper

	childSelect *mapper.Select
	skipChild   bool

	parentSide joinSide
	childSide  joinSide

	secondaryFetchLimit uint

	// docsToYield contains documents read and ready to be yielded by this node.
	docsToYield       []core.Doc
	encounteredDocIDs []string
}

func (join *invertibleTypeJoin) getRootTypeName() string {
	return join.getSecondSide().relFieldDef.Name
}

func (join *invertibleTypeJoin) getSubTypeName() string {
	return join.getFirstSide().relFieldDef.Name
}

func (join *invertibleTypeJoin) replaceRoot(node planNode) {
	join.parentSide.plan = node
	join.getFirstSide().plan = node
}

func (join *invertibleTypeJoin) Init() error {
	if err := join.childSide.plan.Init(); err != nil {
		return err
	}
	return join.parentSide.plan.Init()
}

func (join *invertibleTypeJoin) Start() error {
	if err := join.childSide.plan.Start(); err != nil {
		return err
	}
	return join.parentSide.plan.Start()
}

func (join *invertibleTypeJoin) Close() error {
	if err := join.parentSide.plan.Close(); err != nil {
		return err
	}

	return join.childSide.plan.Close()
}

func (join *invertibleTypeJoin) Spans(spans core.Spans) {
	join.parentSide.plan.Spans(spans)
}

func (join *invertibleTypeJoin) Source() planNode { return join.parentSide.plan }

func (join *invertibleTypeJoin) invert() {
	join.childSide.isFirst = join.parentSide.isFirst
	join.parentSide.isFirst = !join.parentSide.isFirst
}

type docsJoiner struct {
	relIDFieldDef client.FieldDefinition
	primarySide   *joinSide
	secondarySide *joinSide

	primaryScan *scanNode

	resultPrimaryDocs  []core.Doc
	resultSecondaryDoc core.Doc
}

func newSecondaryDocsJoiner(
	primarySide, secondarySide *joinSide,
) docsJoiner {
	j := docsJoiner{
		primarySide:   primarySide,
		secondarySide: secondarySide,
	}
	return j
}

func (j *docsJoiner) fetchPrimaryDocsReferencingSecondaryDoc() error {
	relIDFieldDef, ok := j.primarySide.col.Definition().GetFieldByName(
		j.primarySide.relFieldDef.Name + request.RelatedObjectID)
	if !ok {
		return client.NewErrFieldNotExist(j.primarySide.relFieldDef.Name + request.RelatedObjectID)
	}

	j.primaryScan = getScanNode(j.primarySide.plan)

	j.relIDFieldDef = relIDFieldDef

	primaryDocs, err := j.fetchPrimaryDocs()

	if err != nil {
		return err
	}

	j.resultPrimaryDocs, j.resultSecondaryDoc = joinPrimaryDocs(primaryDocs, j.secondarySide, j.primarySide)

	return nil
}

func (j *docsJoiner) addIDFieldToScanner() {
	found := false
	for i := range j.primaryScan.fields {
		if j.primaryScan.fields[i].Name == j.relIDFieldDef.Name {
			found = true
			break
		}
	}
	if !found {
		j.primaryScan.fields = append(j.primaryScan.fields, j.relIDFieldDef)
	}
}

func (j *docsJoiner) collectDocs(numDocs int) ([]core.Doc, error) {
	p := j.primarySide.plan
	if err := p.Init(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	docs := make([]core.Doc, 0, numDocs)

	for {
		hasValue, err := p.Next()

		if err != nil {
			return nil, err
		}

		if !hasValue {
			break
		}

		docs = append(docs, p.Value())
	}

	return docs, nil
}

func (j *docsJoiner) fetchPrimaryDocs() ([]core.Doc, error) {
	j.addIDFieldToScanner()

	secondaryDoc := j.secondarySide.plan.Value()
	addFilterOnIDField(j.primaryScan, j.primarySide.relIDFieldMapIndex.Value(), secondaryDoc.GetID())

	oldFetcher := j.primaryScan.fetcher
	// TODO: check if spans are necessary to be saved
	oldSpans := j.primaryScan.spans

	indexOnRelation := findIndexByFieldName(j.primaryScan.col, j.relIDFieldDef.Name)
	j.primaryScan.initFetcher(immutable.None[string](), indexOnRelation)

	docs, err := j.collectDocs(0)

	j.primaryScan.fetcher.Close()

	j.primaryScan.spans = oldSpans
	j.primaryScan.fetcher = oldFetcher

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func docsToDocIDs(docs []core.Doc) []string {
	docIDs := make([]string, len(docs))
	for i, doc := range docs {
		docIDs[i] = doc.GetID()
	}
	return docIDs
}

func joinPrimaryDocs(primaryDocs []core.Doc, secondarySide, primarySide *joinSide) ([]core.Doc, core.Doc) {
	secondaryDoc := secondarySide.plan.Value()

	if secondarySide.relFieldMapIndex.HasValue() {
		if secondarySide.relFieldDef.Kind.IsArray() {
			secondaryDoc.Fields[secondarySide.relFieldMapIndex.Value()] = primaryDocs
		} else if len(primaryDocs) > 0 {
			secondaryDoc.Fields[secondarySide.relFieldMapIndex.Value()] = primaryDocs[0]
		}
	}

	if secondarySide.relIDFieldMapIndex.HasValue() {
		if secondarySide.relFieldDef.Kind.IsArray() {
			secondaryDoc.Fields[secondarySide.relIDFieldMapIndex.Value()] = docsToDocIDs(primaryDocs)
		} else if len(primaryDocs) > 0 {
			secondaryDoc.Fields[secondarySide.relIDFieldMapIndex.Value()] = primaryDocs[0].GetID()
		}
	}

	if primarySide.relFieldMapIndex.HasValue() {
		for i := range primaryDocs {
			primaryDocs[i].Fields[primarySide.relFieldMapIndex.Value()] = secondaryDoc
		}
	}

	if primarySide.relIDFieldMapIndex.HasValue() {
		for i := range primaryDocs {
			primaryDocs[i].Fields[primarySide.relIDFieldMapIndex.Value()] = secondaryDoc.GetID()
		}
	}

	return primaryDocs, secondaryDoc
}

func (join *invertibleTypeJoin) fetchPrimaryDocsReferencingSecondaryDoc() ([]core.Doc, core.Doc, error) {
	secJoiner := newSecondaryDocsJoiner(join.getPrimarySide(), join.getSecondarySide())
	err := secJoiner.fetchPrimaryDocsReferencingSecondaryDoc()
	return secJoiner.resultPrimaryDocs, secJoiner.resultSecondaryDoc, err
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

	firstSide := join.getFirstSide()
	hasFirstValue, err := firstSide.plan.Next()

	if err != nil || !hasFirstValue {
		return false, err
	}

	secondSide := join.getSecondSide()

	if firstSide.isPrimary() {
		secondaryDocID := getForeignKey(firstSide.plan, firstSide.relFieldDef.Name)
		// TODO: add some tests with filter on nil relation
		if secondaryDocID == "" {
			if firstSide.isParent {
				join.docsToYield = append(join.docsToYield, firstSide.plan.Value())
				return true, nil
			}
			return join.Next()
		}

		if !firstSide.isParent {
			for i := range join.encounteredDocIDs {
				if join.encounteredDocIDs[i] == secondaryDocID {
					return join.Next()
				}
			}
			join.encounteredDocIDs = append(join.encounteredDocIDs, secondaryDocID)
		}

		// check if there can ever be false (a.k.a. hasDoc = false)
		hasDoc, err := fetchDocWithID(secondSide.plan, secondaryDocID)
		if err != nil {
			return false, err
		}

		// TODO: add some tests that either return error if the doc is not found or return
		// the related doc (without this one) and let it be filtered.
		if !hasDoc {
			if firstSide.isParent {
				join.docsToYield = append(join.docsToYield, firstSide.plan.Value())
				return true, nil
			}
			return join.Next()
		}

		if join.parentSide.relFieldDef.Kind.IsArray() {
			var primaryDocs []core.Doc
			var secondaryDoc core.Doc
			if join.skipChild {
				primaryDocs, secondaryDoc = joinPrimaryDocs([]core.Doc{firstSide.plan.Value()}, secondSide, firstSide)
			} else {
				primaryDocs, secondaryDoc, err = join.fetchPrimaryDocsReferencingSecondaryDoc()
				if err != nil {
					return false, err
				}
			}
			secondaryDoc.Fields[join.childSelect.Index] = primaryDocs

			join.docsToYield = append(join.docsToYield, secondaryDoc)
		} else {
			parentDoc := join.parentSide.plan.Value()
			parentDoc.Fields[join.parentSide.relFieldMapIndex.Value()] = join.childSide.plan.Value()
			join.docsToYield = append(join.docsToYield, parentDoc)
		}
	} else {
		primaryDocs, secondaryDoc, err := join.fetchPrimaryDocsReferencingSecondaryDoc()
		if err != nil {
			return false, err
		}
		if join.parentSide.isPrimary() {
			join.docsToYield = append(join.docsToYield, primaryDocs...)
		} else {
			join.docsToYield = append(join.docsToYield, secondaryDoc)
		}
	}

	return true, nil
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
	p := join.childSide.plan
	s := getScanNode(p)
	s.tryAddField(join.childSide.relFieldDef.Name + request.RelatedObjectID)
	s.filter = fieldFilter
	s.initFetcher(immutable.Option[string]{}, immutable.Some(index))

	join.invert()

	return nil
}

func addFilterOnIDField(scan *scanNode, propIndex int, val any) {
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
