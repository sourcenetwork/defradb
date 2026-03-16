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
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
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
	// based on the relationship of the sub types.
	// May be wrapped by orphan-handling nodes (sequenceNode, orphanNode)
	// during plan expansion when @exhaustive is active.
	joinPlan planNode

	// join is a direct reference to the underlying invertibleTypeJoin,
	// set during plan expansion. Used by simpleExplain to access join
	// metadata (direction, subType) without unwrapping the joinPlan.
	join     *invertibleTypeJoin
	joinKind string

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

	typeFieldDesc, ok := parent.collection.Version().GetFieldByName(subType.Name)
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

	simpleExplainMap[joinTypeLabel] = n.joinKind

	addExplainData := func(j *invertibleTypeJoin) error {
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

	if n.joinKind == typeJoinOneKind {
		if n.join.parentSide.isPrimary() {
			simpleExplainMap[joinDirectionLabel] = joinDirectionPrimaryLabel
		} else {
			simpleExplainMap[joinDirectionLabel] = joinDirectionSecondaryLabel
		}
	}

	err := addExplainData(n.join)

	return simpleExplainMap, err
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

const typeJoinOneKind = "typeJoinOne"

func (n *typeJoinOne) Kind() string {
	return typeJoinOneKind
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

// getFieldsToSplitForTypeJoin returns the fields whose filter conditions should be moved
// from the scan (pre-join) filter to the parent (post-join) filter. This always includes
// the relation field itself. It also includes the secondary FK field (e.g. _publisherID
// on Book where Publisher has @primary) because secondary FK fields are not stored in the
// datastore — they are populated by the join.
func getFieldsToSplitForTypeJoin(parent *selectNode, subType *mapper.Select) []mapper.Field {
	fields := []mapper.Field{subType.Field}
	fkFieldName := request.ToFieldID(subType.Field.Name)
	if fkFieldDesc, ok := parent.collection.Version().GetFieldByName(fkFieldName); ok && !fkFieldDesc.IsPrimary {
		fkFieldIndex := parent.documentMapping.FirstIndexOfName(fkFieldName)
		if fkFieldIndex >= 0 {
			fields = append(fields, mapper.Field{Index: fkFieldIndex, Name: fkFieldName})
		}
	}
	return fields
}

func prepareScanNodeFilterForTypeJoin(
	parent *selectNode,
	source planNode,
	subType *mapper.Select,
) {
	subType.ShowDeleted = parent.selectReq.ShowDeleted

	scan, ok := walkAndFindPlanType[*scanNode](source)
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
		fieldsToSplit := getFieldsToSplitForTypeJoin(parent, subType)
		var parentFilter *mapper.Filter
		scan.filter, parentFilter = filter.SplitByFields(scan.filter, fieldsToSplit...)
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

	parentsRelFieldDef, ok := parent.collection.Version().GetFieldByName(subSelect.Name)
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

	subCol, err := p.db.GetCollectionByName(
		p.ctx,
		subSelect.CollectionName,
		options.WithIdentity(options.GetCollectionByName(), p.identity),
	)
	if err != nil {
		return invertibleTypeJoin{}, err
	}

	var childsRelFieldDef immutable.Option[client.CollectionFieldDescription]
	var childSideRelIDFieldMapIndex immutable.Option[int]
	childsRelFieldDesc, ok := subCol.Version().GetFieldByRelation(
		parentsRelFieldDef.RelationName.Value(),
		parent.collection.Name(),
		parentsRelFieldDef.Name,
	)
	if ok {
		def, ok := subCol.Version().GetFieldByName(childsRelFieldDesc.Name)
		if !ok {
			return invertibleTypeJoin{}, client.NewErrFieldNotExist(subSelect.Name)
		}

		ind := subSelectPlan.DocumentMap().IndexesByName[request.ToFieldID(def.Name)]
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

	ind := parent.documentMapping.IndexesByName[request.ToFieldID(parentsRelFieldDef.Name)]
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

	childScan := getNode[*scanNode](childSide.plan)
	join := invertibleTypeJoin{
		docMapper:  docMapper{parent.documentMapping},
		parentSide: parentSide,
		childSide:  childSide,
		skipChild:  skipChild,
		// we store child's own filter in case an index kicks in and replaces it with it's own filter
		subFilter: childScan.filter,
		// we store child's ordering to apply when fetching child documents
		subOrdering: childScan.ordering,
		exhaustive:  p.joinExpand.exhaustive,
	}

	return join, nil
}

type joinSide struct {
	plan planNode
	// The field definition of the relation-object field on this side of the relation.
	//
	// This will always have a value on the primary side, but it may not have a value on
	// the secondary side, as the secondary half of the relation is optional.
	relFieldDef        immutable.Option[client.CollectionFieldDescription]
	relFieldMapIndex   immutable.Option[int]
	relIDFieldMapIndex immutable.Option[int]
	col                client.Collection
	isFirst            bool
	isParent           bool
}

func (s *joinSide) isPrimary() bool {
	return s.relFieldDef.HasValue() && s.relFieldDef.Value().IsPrimary
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
	ind := node.DocumentMap().FirstIndexOfName(request.ToFieldID(relFieldName))
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

	// Temporarily clear the index for direct docID lookup. When the scan node has an index,
	// the fetcher uses the index keys instead of the docID prefix, which breaks the lookup.
	oldIndex := scan.index
	scan.index = immutable.None[client.IndexDescription]()
	defer func() { scan.index = oldIndex }()

	if err := node.Init(); err != nil {
		return immutable.None[core.Doc](), NewErrSubTypeInit(err)
	}

	hasValue, err := node.Next()

	if err != nil || !hasValue {
		return immutable.None[core.Doc](), err
	}

	return immutable.Some(node.Value()), nil
}

// joinIterationState holds the mutable state that changes during iteration.
// This is separated from configuration to make Init() semantics clearer.
type joinIterationState struct {
	// docsToYield contains documents read and ready to be yielded by this node.
	docsToYield []core.Doc
	// encounteredDocIDs tracks which secondary docs we've already processed
	// to avoid yielding duplicates when multiple primary docs reference the same secondary.
	encounteredDocIDs map[string]struct{}
}

// reset clears all iteration state for a fresh iteration.
func (s *joinIterationState) reset() {
	s.docsToYield = nil
	s.encounteredDocIDs = nil
}

type invertibleTypeJoin struct {
	docMapper

	skipChild  bool
	parentSide joinSide
	childSide  joinSide
	// filter for sub-queries
	subFilter *mapper.Filter
	// ordering for sub-queries
	subOrdering         []mapper.OrderCondition
	secondaryFetchLimit uint
	// exhaustive indicates whether to include orphan documents when ordering
	// by a relation field causes join inversion. When true, documents without
	// related children are fetched separately and merged into results.
	exhaustive bool

	// Iteration state (mutable during execution, reset on Init())
	state joinIterationState
}

func (join *invertibleTypeJoin) replaceRoot(node planNode) {
	join.getFirstSide().plan = node
}

func (join *invertibleTypeJoin) Init() error {
	// Clear iteration state from previous iterations to ensure fresh iteration when reinitializing.
	// This is important for aggregates where we iterate multiple times for different parent docs.
	join.state.reset()

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

func (join *invertibleTypeJoin) parentPlan() planNode { return join.parentSide.plan }
func (join *invertibleTypeJoin) childPlan() planNode  { return join.childSide.plan }

type primaryObjectsRetriever struct {
	relIDFieldDef client.CollectionFieldDescription
	primarySide   *joinSide
	secondarySide *joinSide

	targetSecondaryDoc core.Doc
	filter             *mapper.Filter
	ordering           []mapper.OrderCondition
	exhaustive         bool

	primaryScan *scanNode

	resultPrimaryDocs  []core.Doc
	resultSecondaryDoc core.Doc
}

func (r *primaryObjectsRetriever) retrievePrimaryDocsReferencingSecondaryDoc() error {
	relIDFieldDef, ok := r.primarySide.col.Version().GetFieldByName(
		request.ToFieldID(r.primarySide.relFieldDef.Value().Name))
	if !ok {
		return client.NewErrFieldNotExist(request.ToFieldID(r.primarySide.relFieldDef.Value().Name))
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

func (r *primaryObjectsRetriever) collectDocs() ([]core.Doc, error) {
	p := r.primarySide.plan
	// If the primary side is a multiScanNode, we need to get the source node, as we are the only
	// consumer (one, not multiple) of it.
	if multiScan, ok := p.(*multiScanNode); ok {
		p = multiScan.Source()
	}

	if err := p.Init(); err != nil {
		return nil, NewErrSubTypeInit(err)
	}

	var docs []core.Doc

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

	r.primaryScan.filter = addFilterOnField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
		r.targetSecondaryDoc.GetID())

	oldFetcher := r.primaryScan.fetcher
	oldIndex := r.primaryScan.index
	oldOrdering := r.primaryScan.ordering

	result := selectIndex(selectIndexOptions{
		collection:          r.primaryScan.col,
		filter:              r.primaryScan.filter,
		ordering:            r.ordering,
		relationIDFieldName: r.relIDFieldDef.Name,
		docMapping:          r.primaryScan.documentMapping,
	})

	r.primaryScan.index = result.index

	if result.canSatisfyOrder {
		r.primaryScan.ordering = r.ordering
	} else {
		// Clear ordering so the fetcher doesn't try to use it with an incompatible index.
		// The orderNode (added during plan expansion) will handle in-memory sorting.
		r.primaryScan.ordering = nil
	}

	r.primaryScan.initFetcher(immutable.None[[]string]())

	var docs []core.Doc
	var err error

	if r.exhaustive && r.isOrderingByRelation() {
		if r.orderingRelFieldIsPrimary() {
			orphan := getNode[*orphanNode](r.primarySide.plan)
			if orphan != nil {
				_, relFieldIndex := r.getOrderingInfo()
				relFieldName, _ := r.primaryScan.documentMapping.TryToFindNameFromIndex(relFieldIndex)
				relIDFieldName := request.ToFieldID(relFieldName)
				relIDFieldMapIndex := r.primaryScan.documentMapping.FirstIndexOfName(relIDFieldName)
				parentFilter := addFilterOnField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
					r.targetSecondaryDoc.GetID())
				orphan.setSubQueryContext(parentFilter, relIDFieldName, relIDFieldMapIndex)
			}
		} else {
			orphan := getNode[*orphanPointLookupNode](r.primarySide.plan)
			if orphan != nil {
				parentFilter := addFilterOnField(r.filter, r.primarySide.relIDFieldMapIndex.Value(),
					r.targetSecondaryDoc.GetID())
				orphan.setSubQueryFilter(parentFilter)
			}
		}
	}

	docs, err = r.collectDocs()

	closeErr := r.primaryScan.fetcher.Close()
	r.primaryScan.fetcher = oldFetcher
	r.primaryScan.index = oldIndex
	r.primaryScan.ordering = oldOrdering

	return docs, errors.Join(err, closeErr)
}

// isOrderingByRelation returns true if the ordering involves a relation field.
// This is detected by checking if any order condition has more than one field index
// (indicating traversal through a relation).
func (r *primaryObjectsRetriever) isOrderingByRelation() bool {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return true
		}
	}
	return false
}

// orderingRelFieldIsPrimary returns true if the fetched doc is the primary side of the
// ordering relation (i.e. stores the FK). When true, orphans can be identified directly
// via FK IS NULL. When false, orphans can only be identified by exclusion from join results.
func (r *primaryObjectsRetriever) orderingRelFieldIsPrimary() bool {
	_, relFieldIndex := r.getOrderingInfo()
	fieldName, ok := r.primaryScan.documentMapping.TryToFindNameFromIndex(relFieldIndex)
	if !ok {
		return false
	}
	fieldDef, ok := r.primarySide.col.Version().GetFieldByName(fieldName)
	if !ok {
		return false
	}
	return fieldDef.IsPrimary
}

// getOrderingInfo returns the sort direction and relation field index if the ordering involves a relation field.
func (r *primaryObjectsRetriever) getOrderingInfo() (*mapper.SortDirection, int) {
	for _, order := range r.ordering {
		if len(order.FieldIndexes) > 1 {
			return &order.Direction, order.FieldIndexes[0]
		}
	}
	return nil, 0
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
	ordering []mapper.OrderCondition,
	exhaustive bool,
) ([]core.Doc, core.Doc, error) {
	retriever := primaryObjectsRetriever{
		primarySide:        primarySide,
		secondarySide:      secondarySide,
		targetSecondaryDoc: secondaryDoc,
		filter:             filter,
		ordering:           ordering,
		exhaustive:         exhaustive,
	}
	err := retriever.retrievePrimaryDocsReferencingSecondaryDoc()
	return retriever.resultPrimaryDocs, retriever.resultSecondaryDoc, err
}

func (join *invertibleTypeJoin) Next() (bool, error) {
	if len(join.state.docsToYield) > 0 {
		// If there is one or more documents in the queue, drop the first one -
		// it will have been yielded by the last `Next()` call.
		join.state.docsToYield = join.state.docsToYield[1:]
		if len(join.state.docsToYield) > 0 {
			// If there are still documents in the queue, return true yielding the next
			// one in the queue.
			return true, nil
		}
	}

	firstSide := join.getFirstSide()
	hasFirstValue, err := firstSide.plan.Next()

	if err != nil {
		return false, err
	}

	if !hasFirstValue {
		return false, nil
	}

	if firstSide.isPrimary() {
		return join.fetchRelatedSecondaryDocWithChildren(firstSide.plan.Value())
	} else {
		primaryDocs, secondaryDoc, err := fetchPrimaryDocsReferencingSecondaryDoc(join.getPrimarySide(),
			join.getSecondarySide(), firstSide.plan.Value(), join.subFilter, join.subOrdering, join.exhaustive)
		if err != nil {
			return false, err
		}
		if join.parentSide.isPrimary() {
			join.state.docsToYield = append(join.state.docsToYield, primaryDocs...)
		} else {
			join.state.docsToYield = append(join.state.docsToYield, secondaryDoc)
		}

		// If we reach this line and there are no docs to yield, it likely means that a child
		// document was found but not a parent - this can happen when inverting the join, for
		// example when working with a secondary index.
		if len(join.state.docsToYield) == 0 {
			return join.Next()
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
			join.state.docsToYield = append(join.state.docsToYield, firstSide.plan.Value())
			return true, nil
		}
		return join.Next()
	}

	if secondSide.isParent {
		// child primary docs reference the same secondary parent doc. So if we already encountered
		// the secondary parent doc, we continue to the next primary doc.
		if _, exists := join.state.encounteredDocIDs[secondaryDocID]; exists {
			return join.Next()
		}
		if join.state.encounteredDocIDs == nil {
			join.state.encounteredDocIDs = make(map[string]struct{})
		}
		join.state.encounteredDocIDs[secondaryDocID] = struct{}{}
	}

	secondaryDocOpt, err := fetchDocWithIDAndItsSubDocs(secondSide.plan, secondaryDocID)
	if err != nil {
		return false, err
	}

	if !secondaryDocOpt.HasValue() {
		if firstSide.isParent {
			join.state.docsToYield = append(join.state.docsToYield, firstSide.plan.Value())
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
				join.getPrimarySide(), join.getSecondarySide(), secondaryDoc, join.subFilter, join.subOrdering, join.exhaustive)
			if err != nil {
				return false, err
			}
		}
		secondaryDoc.Fields[join.parentSide.relFieldMapIndex.Value()] = primaryDocs

		join.state.docsToYield = append(join.state.docsToYield, secondaryDoc)
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
		join.state.docsToYield = append(join.state.docsToYield, parentDoc)
	}
	return true, nil
}

func (join *invertibleTypeJoin) Value() core.Doc {
	if len(join.state.docsToYield) == 0 {
		return core.Doc{}
	}
	return join.state.docsToYield[0]
}

func (join *invertibleTypeJoin) invertJoinDirectionWithIndex(
	index client.IndexDescription,
	fieldFilter *mapper.Filter,
	ordering []mapper.OrderCondition,
) {
	childScan := getNode[*scanNode](join.childSide.plan)
	childScan.tryAddFieldWithName(request.ToFieldID(join.childSide.relFieldDef.Value().Name))
	// replace child's filter with the filter that utilizes the index
	// the original child's filter is stored in join.subFilter
	childScan.filter = fieldFilter
	childScan.index = immutable.Some(index)
	childScan.ordering = ordering
	childScan.initFetcher(immutable.Option[[]string]{})

	join.childSide.isFirst = join.parentSide.isFirst
	join.parentSide.isFirst = !join.parentSide.isFirst
}

func getNode[T planNode](plan planNode) T {
	node := plan
	usedFallback := false
	for node != nil {
		if node, ok := node.(T); ok {
			return node
		}
		node = node.Source()
		if node == nil && !usedFallback {
			if topSelect, ok := plan.(*selectTopNode); ok {
				node = topSelect.selectNode
				usedFallback = true
			}
		}
	}
	var zero T
	return zero
}
