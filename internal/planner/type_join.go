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
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/keys"
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

	if !typeFieldDesc.Kind.IsObject() {
		return nil, ErrUnknownRelationType
	}

	if typeFieldDesc.Kind.IsArray() {
		joinPlan, err = p.makeTypeJoinMany(parent, source, subType)
	} else {
		joinPlan, err = p.makeTypeJoinOne(parent, source, subType)
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

func (n *typeIndexJoin) Prefixes(prefixes []keys.Walkable) {
	n.joinPlan.Prefixes(prefixes)
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
		if j.childSide.relFieldDef.HasValue() {
			simpleExplainMap[joinRootLabel] = immutable.Some(j.childSide.relFieldDef.Value().Name)
		}
		simpleExplainMap[joinSubTypeNameLabel] = j.parentSide.relFieldDef.Value().Name

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
			subScan = getNode[*scanNode](joinMany.childSide.plan)
		}
		if joinOne, isJoinOne := n.joinPlan.(*typeJoinOne); isJoinOne {
			subScan = getNode[*scanNode](joinOne.childSide.plan)
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
			parent.filter = filter.Merge(parent.filter, scan.filter)
		}
		scan.filter = nil
	} else {
		var parentFilter *mapper.Filter
		scan.filter, parentFilter = filter.SplitByFields(scan.filter, subType.Field)
		if parentFilter != nil {
			if parent.filter == nil {
				parent.filter = parentFilter
			} else {
				parent.filter = filter.Merge(parent.filter, parentFilter)
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

	var childsRelFieldDef immutable.Option[client.FieldDefinition]
	var childSideRelIDFieldMapIndex immutable.Option[int]
	childsRelFieldDesc, ok := subCol.Version().GetFieldByRelation(
		parentsRelFieldDef.RelationName,
		parent.collection.Name(),
		parentsRelFieldDef.Name,
	)
	if ok {
		def, ok := subCol.Definition().GetFieldByName(childsRelFieldDesc.Name)
		if !ok {
			return invertibleTypeJoin{}, client.NewErrFieldNotExist(subSelect.Name)
		}

		ind := subSelectPlan.DocumentMap().IndexesByName[def.Name+request.RelatedObjectID]
		if len(ind) > 0 {
			childSideRelIDFieldMapIndex = immutable.Some(ind[0])
		}

		childsRelFieldDef = immutable.Some(def)
	}

	parentSide := joinSide{
		plan:             sourcePlan,
		relFieldDef:      immutable.Some(parentsRelFieldDef),
		relFieldMapIndex: immutable.Some(subSelect.Index),
		col:              parent.collection,
		isFirst:          true,
		isParent:         true,
	}

	ind := parent.documentMapping.IndexesByName[parentsRelFieldDef.Name+request.RelatedObjectID]
	if len(ind) > 0 {
		parentSide.relIDFieldMapIndex = immutable.Some(ind[0])
	}

	childSide := joinSide{
		plan:               subSelectPlan,
		relFieldDef:        childsRelFieldDef,
		relIDFieldMapIndex: childSideRelIDFieldMapIndex,
		col:                subCol,
		isFirst:            false,
		isParent:           false,
	}

	join := invertibleTypeJoin{
		docMapper:  docMapper{parent.documentMapping},
		parentSide: parentSide,
		childSide:  childSide,
		skipChild:  skipChild,
		// we store child's own filter in case an index kicks in and replaces it with it's own filter
		subFilter: getNode[*scanNode](childSide.plan).filter,
	}

	return join, nil
}

type joinSide struct {
	plan planNode
	// The field definition of the relation-object field on this side of the relation.
	//
	// This will always have a value on the primary side, but it may not have a value on
	// the secondary side, as the secondary half of the relation is optional.
	relFieldDef        immutable.Option[client.FieldDefinition]
	relFieldMapIndex   immutable.Option[int]
	relIDFieldMapIndex immutable.Option[int]
	col                client.Collection
	isFirst            bool
	isParent           bool
}

func (s *joinSide) isPrimary() bool {
	return s.relFieldDef.HasValue() && s.relFieldDef.Value().IsPrimaryRelation
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

// getForeignKey returns the docID of the related object referenced by the given relation field.
func getForeignKey(node planNode, relFieldName string) string {
	ind := node.DocumentMap().FirstIndexOfName(relFieldName + request.RelatedObjectID)
	docIDStr, _ := node.Value().Fields[ind].(string)
	return docIDStr
}

// fetchDocWithIDAndItsSubDocs fetches a document with the given docID from the given planNode.
func fetchDocWithIDAndItsSubDocs(node planNode, docID string) (immutable.Option[core.Doc], error) {
	scan := getNode[*scanNode](node)
	if scan == nil {
		return immutable.None[core.Doc](), nil
	}

	shortID, err := id.GetShortCollectionID(scan.p.ctx, scan.col.Version().CollectionID)
	if err != nil {
		return immutable.None[core.Doc](), err
	}

	dsKey := keys.DataStoreKey{
		CollectionShortID: shortID,
		DocID:             docID,
	}

	prefixes := []keys.Walkable{dsKey}

	node.Prefixes(prefixes)

	if err := node.Init(); err != nil {
		return immutable.None[core.Doc](), NewErrSubTypeInit(err)
	}

	hasValue, err := node.Next()

	if err != nil || !hasValue {
		return immutable.None[core.Doc](), err
	}

	return immutable.Some(node.Value()), nil
}

type invertibleTypeJoin struct {
	docMapper

	skipChild bool

	parentSide joinSide
	childSide  joinSide

	// the filter of the subnode to store in case it's replaced by an index filter
	subFilter *mapper.Filter

	secondaryFetchLimit uint

	// docsToYield contains documents read and ready to be yielded by this node.
	docsToYield       []core.Doc
	encounteredDocIDs []string
}

func (join *invertibleTypeJoin) replaceRoot(node planNode) {
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

func (join *invertibleTypeJoin) Prefixes(prefixes []keys.Walkable) {
	join.parentSide.plan.Prefixes(prefixes)
}

func (join *invertibleTypeJoin) Source() planNode { return join.parentSide.plan }

type primaryObjectsRetriever struct {
	relIDFieldDef client.FieldDefinition
	primarySide   *joinSide
	secondarySide *joinSide

	targetSecondaryDoc core.Doc
	filter             *mapper.Filter

	primaryScan *scanNode

	resultPrimaryDocs  []core.Doc
	resultSecondaryDoc core.Doc
}

func (r *primaryObjectsRetriever) retrievePrimaryDocsReferencingSecondaryDoc() error {
	relIDFieldDef, ok := r.primarySide.col.Definition().GetFieldByName(
		r.primarySide.relFieldDef.Value().Name + request.RelatedObjectID)
	if !ok {
		return client.NewErrFieldNotExist(r.primarySide.relFieldDef.Value().Name + request.RelatedObjectID)
	}

	r.primaryScan = getNode[*scanNode](r.primarySide.plan)

	r.relIDFieldDef = relIDFieldDef

	primaryDocs, err := r.retrievePrimaryDocs()

	if err != nil {
		return err
	}

	r.resultPrimaryDocs, r.resultSecondaryDoc = joinPrimaryDocs(
		primaryDocs,
		r.targetSecondaryDoc,
		r.primarySide,
		r.secondarySide,
	)

	return nil
}

func (r *primaryObjectsRetriever) collectDocs(numDocs int) ([]core.Doc, error) {
	p := r.primarySide.plan
	// If the primary side is a multiScanNode, we need to get the source node, as we are the only
	// consumer (one, not multiple) of it.
	if multiScan, ok := p.(*multiScanNode); ok {
		p = multiScan.Source()
	}
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

func (r *primaryObjectsRetriever) retrievePrimaryDocs() ([]core.Doc, error) {
	r.primaryScan.addField(r.relIDFieldDef)

	r.primaryScan.filter = addFilterOnIDField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
		r.targetSecondaryDoc.GetID())

	oldFetcher := r.primaryScan.fetcher
	oldIndex := r.primaryScan.index

	r.primaryScan.index = findIndexByFieldName(r.primaryScan.col, r.relIDFieldDef.Name)
	r.primaryScan.initFetcher(immutable.None[string]())

	docs, err := r.collectDocs(0)
	if err != nil {
		return nil, err
	}

	err = r.primaryScan.fetcher.Close()
	if err != nil {
		return nil, err
	}

	r.primaryScan.fetcher = oldFetcher
	r.primaryScan.index = oldIndex

	return docs, nil
}

func docsToDocIDs(docs []core.Doc) []string {
	docIDs := make([]string, len(docs))
	for i, doc := range docs {
		docIDs[i] = doc.GetID()
	}
	return docIDs
}

func joinPrimaryDocs(
	primaryDocs []core.Doc,
	secondaryDoc core.Doc,
	primarySide, secondarySide *joinSide,
) ([]core.Doc, core.Doc) {
	if secondarySide.relFieldMapIndex.HasValue() {
		if !secondarySide.relFieldDef.HasValue() || secondarySide.relFieldDef.Value().Kind.IsArray() {
			secondaryDoc.Fields[secondarySide.relFieldMapIndex.Value()] = primaryDocs
		} else if len(primaryDocs) > 0 {
			secondaryDoc.Fields[secondarySide.relFieldMapIndex.Value()] = primaryDocs[0]
		}
	}

	if secondarySide.relIDFieldMapIndex.HasValue() {
		if !secondarySide.relFieldDef.HasValue() || secondarySide.relFieldDef.Value().Kind.IsArray() {
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

func fetchPrimaryDocsReferencingSecondaryDoc(
	primarySide, secondarySide *joinSide,
	secondaryDoc core.Doc,
	filter *mapper.Filter,
) ([]core.Doc, core.Doc, error) {
	retriever := primaryObjectsRetriever{
		primarySide:        primarySide,
		secondarySide:      secondarySide,
		targetSecondaryDoc: secondaryDoc,
		filter:             filter,
	}
	err := retriever.retrievePrimaryDocsReferencingSecondaryDoc()
	return retriever.resultPrimaryDocs, retriever.resultSecondaryDoc, err
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

	if firstSide.isPrimary() {
		return join.fetchRelatedSecondaryDocWithChildren(firstSide.plan.Value())
	} else {
		primaryDocs, secondaryDoc, err := fetchPrimaryDocsReferencingSecondaryDoc(
			join.getPrimarySide(), join.getSecondarySide(), firstSide.plan.Value(), join.subFilter)
		if err != nil {
			return false, err
		}
		if join.parentSide.isPrimary() {
			join.docsToYield = append(join.docsToYield, primaryDocs...)
		} else {
			join.docsToYield = append(join.docsToYield, secondaryDoc)
		}

		// If we reach this line and there are no docs to yield, it likely means that a child
		// document was found but not a parent - this can happen when inverting the join, for
		// example when working with a secondary index.
		if len(join.docsToYield) == 0 {
			return false, nil
		}
	}

	return true, nil
}

func (join *invertibleTypeJoin) fetchRelatedSecondaryDocWithChildren(primaryDoc core.Doc) (bool, error) {
	firstSide := join.getFirstSide()
	secondSide := join.getSecondSide()

	secondaryDocID := getForeignKey(firstSide.plan, firstSide.relFieldDef.Value().Name)
	if secondaryDocID == "" {
		if firstSide.isParent {
			join.docsToYield = append(join.docsToYield, firstSide.plan.Value())
			return true, nil
		}
		return join.Next()
	}

	if secondSide.isParent {
		// child primary docs reference the same secondary parent doc. So if we already encountered
		// the secondary parent doc, we continue to the next primary doc.
		for i := range join.encounteredDocIDs {
			if join.encounteredDocIDs[i] == secondaryDocID {
				return join.Next()
			}
		}
		join.encounteredDocIDs = append(join.encounteredDocIDs, secondaryDocID)
	}

	secondaryDocOpt, err := fetchDocWithIDAndItsSubDocs(secondSide.plan, secondaryDocID)

	if err != nil {
		return false, err
	}

	if !secondaryDocOpt.HasValue() {
		if firstSide.isParent {
			join.docsToYield = append(join.docsToYield, firstSide.plan.Value())
			return true, nil
		}
		return join.Next()
	}

	secondaryDoc := secondaryDocOpt.Value()

	if join.parentSide.relFieldDef.Value().Kind.IsArray() {
		var primaryDocs []core.Doc
		// if child is not requested as part of the response, we just add the existing one (fetched by the secondary index
		// on a filtered value) so that top select node that runs the filter again can yield it.
		if join.skipChild {
			primaryDocs, secondaryDoc = joinPrimaryDocs(
				[]core.Doc{firstSide.plan.Value()}, secondaryDoc, join.getPrimarySide(), join.getSecondSide())
		} else {
			primaryDocs, secondaryDoc, err = fetchPrimaryDocsReferencingSecondaryDoc(
				join.getPrimarySide(), join.getSecondarySide(), secondaryDoc, join.subFilter)
			if err != nil {
				return false, err
			}
		}
		secondaryDoc.Fields[join.parentSide.relFieldMapIndex.Value()] = primaryDocs

		join.docsToYield = append(join.docsToYield, secondaryDoc)
	} else {
		var parentDoc core.Doc
		var childDoc core.Doc
		if join.getPrimarySide().isParent {
			parentDoc = primaryDoc
			childDoc = secondaryDoc
		} else {
			parentDoc = secondaryDoc
			childDoc = primaryDoc
		}
		parentDoc.Fields[join.parentSide.relFieldMapIndex.Value()] = childDoc
		join.docsToYield = append(join.docsToYield, parentDoc)
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
	index client.IndexDescription,
	fieldFilter *mapper.Filter,
	ordering []mapper.OrderCondition,
) error {
	childScan := getNode[*scanNode](join.childSide.plan)
	childScan.tryAddFieldWithName(join.childSide.relFieldDef.Value().Name + request.RelatedObjectID)
	// replace child's filter with the filter that utilizes the index
	// the original child's filter is stored in join.subFilter
	childScan.filter = fieldFilter
	childScan.index = immutable.Some(index)
	childScan.ordering = ordering
	childScan.initFetcher(immutable.Option[string]{})

	join.childSide.isFirst = join.parentSide.isFirst
	join.parentSide.isFirst = !join.parentSide.isFirst

	return nil
}

func addFilterOnIDField(f *mapper.Filter, propIndex int, docID string) *mapper.Filter {
	if f == nil {
		f = mapper.NewFilter()
	}

	propertyIndex := &mapper.PropertyIndex{Index: propIndex}
	filterConditions := map[connor.FilterKey]any{
		propertyIndex: map[connor.FilterKey]any{
			mapper.FilterEqOp: docID,
		},
	}

	filter.RemoveField(f, mapper.Field{Index: propIndex})
	f.Conditions = filter.MergeConditions(f.Conditions, filterConditions)
	return f
}

func getNode[T planNode](plan planNode) T {
	node := plan
	for node != nil {
		if node, ok := node.(T); ok {
			return node
		}
		node = node.Source()
		if node == nil {
			if topSelect, ok := plan.(*selectTopNode); ok {
				node = topSelect.selectNode
			}
		}
	}
	var zero T
	return zero
}
