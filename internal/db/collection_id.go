// Copyright 2024 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"fmt"
	"slices"
	"strings"

	cidlink "github.com/ipld/go-ipld-prime/linking/cid"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	coreblock "github.com/sourcenetwork/defradb/internal/core/block"
	"github.com/sourcenetwork/defradb/internal/core/crdt"
	"github.com/sourcenetwork/defradb/internal/db/description"
)

// setCollectionIDs saves the given collections to the blockstore and sets the resultant ids on given
// collections, mutating the input parameter.
//
// This includes CollectionID (if not already set), VersionID, FieldID, and relational field Kinds.
func setCollectionIDs(
	ctx context.Context,
	newCollections []client.CollectionVersion,
	existingCollections []client.CollectionVersion,
) error {
	// We need to group the inputs and then mutate them, so we temporarily
	// map them to pointers.
	newCollectionPtrs := make([]*client.CollectionVersion, len(newCollections))
	for i := range newCollections {
		collection := newCollections[i]
		newCollectionPtrs[i] = &collection
	}

	collectionSets := getCollectionSets(newCollectionPtrs)
	collectionSets = sortCollectionSets(collectionSets)

	for _, collectionSet := range collectionSets {
		// The schemas within each set must be in a deterministic order to ensure that
		// their IDs are deterministic.
		sortSet(collectionSet)

		substituteRelationFieldKinds(collectionSet, collectionSets, existingCollections)
		err := saveBlocks(ctx, collectionSet)
		if err != nil {
			return err
		}
	}

	for _, collectionSet := range collectionSets {
		// Secondary fields are not saved in the blockstore, thus they do not contribute to the collection IDs.
		// The Kinds do however need to reference by CollectionID, which need to be substituted after the
		// CollectionIDs have been generated.
		substituteSecondaryRelationFieldKinds(collectionSet, collectionSets, existingCollections)
	}

	for i := range newCollectionPtrs {
		newCollections[i] = *newCollectionPtrs[i]
	}

	return nil
}

// collectionRelations is a trimmed down [client.CollectionVersion] containing
// only the useful information to the functions in this file.
type collectionRelations struct {
	// The name of this collection
	name string

	// The collection names of the primary relations from this collection.
	relations []string
}

// getCollectionSets groups collections into sets.
//
// Most sets will contain a single collection, however if a circular dependency chain is found
// all elements within that chain will be grouped together into a single set.
//
// For example if User contains a relation *to* Dog, and Dog contains a relationship *to*
// User, they will be grouped into the same set.
func getCollectionSets(
	newCollections []*client.CollectionVersion,
) [][]*client.CollectionVersion {
	// Build name-level relation graph
	type collInfo struct {
		relations []string
	}

	graph := map[string]collInfo{}

	for _, col := range newCollections {
		var rels []string
		for _, f := range col.Fields {
			if !f.IsPrimary {
				continue
			}
			if kind, ok := f.Kind.(*client.NamedKind); ok {
				rels = append(rels, kind.Name)
			}
		}
		if len(rels) > 0 {
			graph[col.Name] = collInfo{relations: rels}
		}
	}

	// Perform Tarjan SCC on the collection names
	index := 0
	stack := []string{}
	onStack := map[string]bool{}
	indices := map[string]int{}
	lowlink := map[string]int{}
	circularNames := map[string]struct{}{}

	var strongConnect func(string)
	strongConnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++

		stack = append(stack, v)
		onStack[v] = true

		for _, w := range graph[v].relations {
			if _, seen := indices[w]; !seen {
				strongConnect(w)
				lowlink[v] = min(lowlink[v], lowlink[w])
			} else if onStack[w] {
				lowlink[v] = min(lowlink[v], indices[w])
			}
		}

		if lowlink[v] == indices[v] {
			var scc []string
			for {
				n := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[n] = false
				scc = append(scc, n)
				if n == v {
					break
				}
			}
			if len(scc) > 1 {
				for _, name := range scc {
					circularNames[name] = struct{}{}
				}
			}
		}
	}

	for name := range graph {
		if _, seen := indices[name]; !seen {
			strongConnect(name)
		}
	}

	// Assign set IDs per version
	setIDByVersion := map[string]int{}
	nextSetID := 0

	// First: circular collections (grouped by name)
	circularSetIDByName := map[string]int{}

	for _, col := range newCollections {
		if _, isCircular := circularNames[col.Name]; !isCircular {
			continue
		}

		setID, exists := circularSetIDByName[col.Name]
		if !exists {
			nextSetID++
			setID = nextSetID
			circularSetIDByName[col.Name] = setID
		}

		setIDByVersion[col.VersionID] = setID
	}

	// Then: non-circular versions (each gets its own set)
	for _, col := range newCollections {
		if _, ok := setIDByVersion[col.VersionID]; ok {
			continue
		}
		nextSetID++
		setIDByVersion[col.VersionID] = nextSetID
	}

	// Materialize sets
	collectionSetsByID := map[int][]*client.CollectionVersion{}

	for _, col := range newCollections {
		setID := setIDByVersion[col.VersionID]
		collectionSetsByID[setID] = append(collectionSetsByID[setID], col)
	}

	collectionSets := make([][]*client.CollectionVersion, 0, nextSetID)
	for i := 1; i <= nextSetID; i++ {
		if set, ok := collectionSetsByID[i]; ok {
			collectionSets = append(collectionSets, set)
		}
	}

	return collectionSets
}

// min is a Helper function needed for the Tarjan SCC algorithm. Returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// mapCollectionSetIDs recursively scans through a collection and its relations, assigning each collection to a
// temporary setID.
//
// If a set of collections form a circular dependency, all involved collections will be assigned the same setID.
// Assigned setIDs will be added to the input param `collectionSetIds`.
//
// This function will return when all descendents of the initial collection have been processed.
//
// Parameters:
//   - i: The largest setID so far assigned. This parameter is mutated by this function.
//   - collection: The current collection to process
//   - collectionSetIds: The set of already assigned setIDs mapped by collection name - this parameter will be mutated
//     by this function
//   - collectionRelationsByCollectionName: The full set of relevant collections/relations mapped by collection name
//   - collectionsFullyProcessed: The set of collection names that have already been completely processed.  If
//     `collection` is in this set the function will return.  This parameter is mutated by this function.
func mapCollectionSetIDs(
	i *int,
	collection collectionRelations,
	collectionSetIds map[string]int,
	collectionRelationsByCollectionName map[string]collectionRelations,
	collectionsFullyProcessed map[string]struct{},
) {
	if _, ok := collectionsFullyProcessed[collection.name]; ok {
		// we've circled all the way through and already processed this collection
		return
	}
	collectionsFullyProcessed[collection.name] = struct{}{}

	for _, relation := range collection.relations {
		// if more than one relation, need to find out if the relation loops back here! It might connect to a separate circle
		circlesBackHere := circlesBack(collection.name, relation, collectionRelationsByCollectionName, map[string]struct{}{})

		var circleID int
		if circlesBackHere {
			if id, ok := collectionSetIds[relation]; ok {
				// If this collection has already been assigned a setID, use that
				circleID = id
			} else {
				collectionSetId, ok := collectionSetIds[collection.name]
				if !ok {
					// If this collection has not already been assigned a setID, it must be
					// the first discovered node in a new circle.  Assign it a new setID,
					// this will be picked up by its circle-forming descendents.
					*i = *i + 1
					collectionSetId = *i
				}
				collectionSetIds[collection.name] = collectionSetId
				circleID = collectionSetId
			}
		} else {
			// If this collection and its relations does not circle back to itself, we
			// increment `i` and assign the new value to this collection *only*
			*i = *i + 1
			circleID = *i
		}

		collectionSetIds[relation] = circleID
		mapCollectionSetIDs(
			i,
			collectionRelationsByCollectionName[relation],
			collectionSetIds,
			collectionRelationsByCollectionName,
			collectionsFullyProcessed,
		)
	}
}

// circlesBack returns true if any path from this schema through it's relations (and their relations) circles
// back to this schema.
//
// Parameters:
//   - originalSchemaName: The original start schema of this recursive check - this will not change as this function
//     recursively checks the relations on `currentSchemaName`.
//   - currentSchemaName: The current schema to process.
//   - schemasWithRelations: The full set of relevant schemas that may be referenced by this schema or its descendents.
//   - schemasFullyProcessed: The set of schema names that have already been completely processed.  If `schema` is in
//     this set the function will return.  This parameter is mutated by this function.
func circlesBack(
	originalSchemaName string,
	currentSchemaName string,
	schemasWithRelations map[string]collectionRelations,
	schemasFullyProcessed map[string]struct{},
) bool {
	if _, ok := schemasFullyProcessed[currentSchemaName]; ok {
		// we've circled all the way through and not found the original
		return false
	}

	if currentSchemaName == originalSchemaName {
		return true
	}

	schemasFullyProcessed[currentSchemaName] = struct{}{}

	for _, relation := range schemasWithRelations[currentSchemaName].relations {
		ciclesBackToOriginal := circlesBack(originalSchemaName, relation, schemasWithRelations, schemasFullyProcessed)
		if ciclesBackToOriginal {
			return true
		}
	}

	return false
}

// sortCollectionSets orders the given collection sets based on the order in which they must be written
// to the block store, based on the relations between them.
//
// This is required so that the CIDs can be properly formed - if a field on `Book` references `Author`, then `Author`
// needs to be inserted first so that the referencing field can reference it by `Author`'s CID, and in turn form part
// of `Book`'s own CID.
func sortCollectionSets(collectionSets [][]*client.CollectionVersion) [][]*client.CollectionVersion {
	return sortCollectionSetsFrom(0, collectionSets)
}

// sortCollectionSetsFrom sorts collection sets from the given index onwards.
func sortCollectionSetsFrom(index int, collectionSets [][]*client.CollectionVersion) [][]*client.CollectionVersion {
	// This function is recursive, each call will add the collection sets that it can (no relations, or relations
	// to sets already sorted), the rest get added to deferredSets so that they can be sorted in later passes.
	deferredSets := make([][]*client.CollectionVersion, 0, len(collectionSets))
	result := make([][]*client.CollectionVersion, 0, len(collectionSets))

	allColNames := make(map[string]struct{}, len(collectionSets))
	for _, set := range collectionSets {
		for _, col := range set {
			allColNames[col.Name] = struct{}{}
		}
	}

	sortedColNames := map[string]struct{}{}
	for i := 0; i < index && i < len(collectionSets); i++ {
		for _, col := range collectionSets[i] {
			sortedColNames[col.Name] = struct{}{}
		}

		if i != index {
			// Append any sets already sorted to the result.
			result = append(result, collectionSets[i])
		}
	}

setLoop:
	for i := index; i < len(collectionSets); i++ {
		collectionSet := collectionSets[i]
		colNamesInSet := map[string]struct{}{}
		for _, col := range collectionSet {
			colNamesInSet[col.Name] = struct{}{}
		}

		for _, col := range collectionSet {
			for _, field := range col.Fields {
				switch kind := field.Kind.(type) {
				case *client.NamedKind:
					_, relationInternalToSet := colNamesInSet[kind.Name]
					_, relationToSortedSet := sortedColNames[kind.Name]
					_, relationToKnown := allColNames[kind.Name]

					if !field.IsPrimary {
						// Only primary relation fields get saved in the collection block DAG - secondary fields
						// have no impact on the collection CIDs and can be ignored by the sorting.
						continue
					}

					if !relationInternalToSet && !relationToSortedSet && relationToKnown {
						// If the collection referenced by the field is:
						// - Within this set it is self containing, and thus can be sorted (no external dependency).
						// - Related to a set that has already been sorted, the external dependency has been sorted
						//   and will be fully formed by the time this set is finalized.
						// - Unknown then there is nothing that this code can do to help it, and we must avoid it and
						//   let the validation code return a human readable error to the user.
						deferredSets = append(deferredSets, collectionSet)
						continue setLoop
					}
				}
			}
		}

		result = append(result, collectionSet)
	}

	if len(deferredSets) > 0 {
		return sortCollectionSetsFrom(len(result), append(result, deferredSets...))
	}

	return result
}

// saveBlocks saves the collection set to the block and headstore.  It mutates the given collectionSet
// setting the ids and migrations.
func saveBlocks(
	ctx context.Context,
	collectionSet []*client.CollectionVersion,
) error {
	colIds := make([]cidlink.Link, 0, len(collectionSet))
	hasSetUpdated := false

	for _, collection := range collectionSet {
		if collection.VersionID == "" && collection.CollectionID != "" {
			// If the VersionID is empty, but the CollectionID is not, the user has patched one
			// of these properties - continue, and let the validation code return an error.
			continue
		}

		var oldCol client.CollectionVersion
		if collection.VersionID != "" {
			var err error
			oldCol, err = description.GetCollectionByID(ctx, collection.VersionID)
			if err != nil {
				if errors.Is(err, corekv.ErrNotFound) {
					// If the key does not exist, continue, and let the validation code handle it
					// in a user friendly way.
					continue
				}
				return err
			}
		}

		var hasFieldsChanged bool
		newFieldLevelCIDs := []coreblock.DAGLink{}
		for i, newField := range collection.Fields {
			fieldCRDT := crdt.NewFieldDefinition(collection.Name, newField.Name)
			delta, hasFieldChanged, err := fieldCRDT.Delta(
				newField,
				// We cheat here for now, as users cannot yet mutate fields.  When they can,
				// we will need to pass in the old version here.
				client.CollectionFieldDescription{},
			)
			if err != nil {
				return err
			}

			if !hasFieldChanged {
				continue
			}
			hasFieldsChanged = true

			cid, _, err := coreblock.AddDelta(ctx, fieldCRDT, delta)
			if err != nil {
				return err
			}

			collection.Fields[i].FieldID = cid.String()
			newFieldLevelCIDs = append(newFieldLevelCIDs, coreblock.DAGLink{Link: cid})
		}

		colCRDT := crdt.NewCollectionDefinition(collection.Name)
		delta, hasCollectionChanged, err := colCRDT.Delta(*collection, oldCol)
		if err != nil {
			return err
		}

		if !hasFieldsChanged && !hasCollectionChanged {
			// If the global collection state has not changed, there is nothing to do here and we
			// move on to the next collection.
			continue
		}
		hasSetUpdated = true

		cid, _, err := coreblock.AddDelta(ctx, colCRDT, delta, newFieldLevelCIDs...)
		if err != nil {
			return err
		}

		collection.VersionID = cid.String()
		if collection.CollectionID == "" {
			collection.CollectionID = collection.VersionID
		}

		colIds = append(colIds, cid)

		if oldCol.VersionID != "" {
			var migration immutable.Option[string]
			if oldCol.PreviousVersion.Value().Transform.Value() != collection.PreviousVersion.Value().Transform.Value() {
				// If the patch has updated the migration, use it, otherwise assume it was the old version migration,
				// and ignore it.
				migration = collection.PreviousVersion.Value().Transform
			}

			collection.PreviousVersion = immutable.Some(
				client.CollectionSource{
					SourceCollectionID: oldCol.VersionID,
					Transform:          migration,
				},
			)
		}
	}

	if hasSetUpdated && len(collectionSet) > 1 {
		colSetCRDT := crdt.NewCollectionSet(collectionSet[0].CollectionID)
		delta := colSetCRDT.Delta()

		links := make([]coreblock.DAGLink, 0, len(colIds))
		for _, colId := range colIds {
			links = append(links, coreblock.DAGLink{Link: colId})
		}

		cid, _, err := coreblock.AddDelta(ctx, colSetCRDT, delta, links...)
		if err != nil {
			return err
		}

		collectionSetID := cid.String()

		for i := range collectionSet {
			collectionSet[i].CollectionSet = immutable.Some(client.CollectionSetDescription{
				CollectionSetID: collectionSetID,
				RelativeID:      i,
			})
		}
	}

	return nil
}

// substituteRelationFieldKinds substitutes relations defined using [NamedKind]s to their long-term
// types.
//
// Using names to reference other types is unsuitable as the names may change over time.
func substituteRelationFieldKinds(
	collectionSet []*client.CollectionVersion,
	newCollectionSets [][]*client.CollectionVersion,
	existingCollections []client.CollectionVersion,
) {
	collectionsByName := map[string]client.CollectionVersion{}
	for _, set := range newCollectionSets {
		for _, collection := range set {
			collectionsByName[collection.Name] = *collection
		}
	}
	for _, collection := range existingCollections {
		collectionsByName[collection.Name] = collection
	}

	setIndexesByName := map[string]int{}
	for i, col := range collectionSet {
		setIndexesByName[col.Name] = i
	}

	for i := range collectionSet {
		for j := range collectionSet[i].Fields {
			switch kind := collectionSet[i].Fields[j].Kind.(type) {
			case *client.NamedKind:
				relationCollection, ok := collectionsByName[kind.Name]
				if !ok {
					// Continue, and let the validation step pick up whatever went wrong later
					continue
				}

				relativeIndex, referencesHostSet := setIndexesByName[kind.Name]

				if referencesHostSet {
					// The CollectionID will not exist until the field and collection blocks have been saved for the entire set
					// due to a circular relation(s), so any fields that reference collections within this set must use the
					// `SelfKind` kind instead of a normal `CollectionKind`.

					// SelfKind fields do not care about primary/secondary in this context as they do not reference by VersionID
					// so we might as well handle the secondary side conversion here too.

					if len(setIndexesByName) > 1 {
						collectionSet[i].Fields[j].Kind = client.NewSelfKind(fmt.Sprint(relativeIndex), kind.IsArray())
					} else {
						// If the relation root is simple and does not contain a relative index, then this relation
						// must point to the host schema (self-reference, e.g. User=>User).
						collectionSet[i].Fields[j].Kind = client.NewSelfKind("", kind.IsArray())
					}
				} else {
					if !collectionSet[i].Fields[j].IsPrimary {
						continue
					}

					collectionSet[i].Fields[j].Kind = client.NewCollectionKind(relationCollection.CollectionID, kind.IsArray())
				}

			default:
				// no-op
			}
		}
	}
}

func substituteSecondaryRelationFieldKinds(
	collectionSet []*client.CollectionVersion,
	newCollectionSets [][]*client.CollectionVersion,
	existingCollections []client.CollectionVersion,
) {
	collectionsByName := map[string]client.CollectionVersion{}
	for _, set := range newCollectionSets {
		for _, collection := range set {
			collectionsByName[collection.Name] = *collection
		}
	}
	for _, collection := range existingCollections {
		collectionsByName[collection.Name] = collection
	}

	for i := range collectionSet {
		for j := range collectionSet[i].Fields {
			switch kind := collectionSet[i].Fields[j].Kind.(type) {
			case *client.NamedKind:
				relationCollection, ok := collectionsByName[kind.Name]
				if !ok {
					// Continue, and let the validation step pick up whatever went wrong later
					continue
				}

				// SelfKind fields do not care about primary/secondary in this context as they do not reference by VersionID,
				// they will have already been converted from NamedKinds earlier.

				collectionSet[i].Fields[j].Kind = client.NewCollectionKind(relationCollection.CollectionID, kind.IsArray())

			default:
				// no-op
			}
		}
	}
}

func sortSet(collectionSet []*client.CollectionVersion) {
	slices.SortFunc(collectionSet, func(a, b *client.CollectionVersion) int {
		// Because the set is as small as possible, as it only includes circular collections, which by definition
		// must all be present, sorting by Name is globally consistent.
		return strings.Compare(a.Name, b.Name)
	})
}
