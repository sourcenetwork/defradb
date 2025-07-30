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
	"github.com/sourcenetwork/lens/host-go/config/model"

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
	migration immutable.Option[model.Lens],
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

		substituteRelationFieldKinds(collectionSet, collectionSets)
		err := saveBlocks(ctx, collectionSet, migration)
		if err != nil {
			return err
		}
	}

	for _, collectionSet := range collectionSets {
		// Secondary fields are not saved in the blockstore, thus they do not contribute to the collection IDs.
		// The Kinds do however need to reference by CollectionID, which need to be substituted after the
		// CollectionIDs have been generated.
		substituteSecondaryRelationFieldKinds(collectionSet, collectionSets)
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
func getCollectionSets(newCollections []*client.CollectionVersion) [][]*client.CollectionVersion {
	collectionsByName := make(map[string]client.CollectionVersion, len(newCollections))
	for _, col := range newCollections {
		collectionsByName[col.Name] = *col
	}

	collectionsWithRelations := map[string]collectionRelations{}
	for _, collection := range newCollections {
		relations := []string{}
		for _, field := range collection.Fields {
			if !field.IsPrimary {
				continue
			}

			switch kind := field.Kind.(type) {
			case *client.NamedKind:
				if otherCol, ok := collectionsByName[kind.Name]; ok {
					// We only need to worry about bi-directional relationships here, single sided relationships cannot be
					// circular.
					if _, ok := otherCol.GetFieldByRelation(field.RelationName.Value(), collection.Name, field.Name); ok {
						// We only need to worry about user provided `NamedKind` relations in this scope.
						// Other relation kinds can either not be circular, or are relative to the host.
						relations = append(relations, kind.Name)
					}
				}

			default:
				// no-op
			}
		}

		if len(relations) == 0 {
			// If a collection is defined with no relations, then it is not relevant to this function
			// and can be skipped.
			continue
		}

		collectionsWithRelations[collection.Name] = collectionRelations{
			name:      collection.Name,
			relations: relations,
		}
	}

	changedInLoop := true
	for changedInLoop {
		// This loop strips out collections from `collectionsWithRelations` that do not form circular
		// collection sets (e.g. User=>Dog=>User).  This allows later logic that figures out the
		// exact path that circles forms to operate on a minimal set of data, reducing its cost
		// and complexity.
		//
		// Some non circular relations may still remain after this first pass, for example
		// one-directional relations between two circles.
		changedInLoop = false
		for _, collection := range collectionsWithRelations {
			i := 0
			relation := ""
			deleteI := false
			for i, relation = range collection.relations {
				if _, ok := collectionsWithRelations[relation]; !ok {
					// If the related collection is not in `collectionsWithRelations` it must have been removed
					// in a previous iteration of the collectionsWithRelations loop, this will have been
					// done because it had no relevant remaining relations and thus could not be part
					// of a circular collection set.  If this is the case, this `relation` is also irrelevant
					// here and can be removed as it too cannot form part of a circular collection set.
					changedInLoop = true
					deleteI = true
					break
				}
			}

			if deleteI {
				collection.relations = append(collection.relations[:i], collection.relations[i+1:]...)
				collectionsWithRelations[collection.name] = collection
			}

			if len(collection.relations) == 0 {
				// If there are no relevant relations from this collection, remove the collection from
				// `collectionsWithRelations` as the collection cannot form part of a circular collection
				// set.
				changedInLoop = true
				delete(collectionsWithRelations, collection.name)
				break
			}
		}
	}

	// If len(collectionsWithRelations) > 0 here there are circular relations.
	// We then need to traverse them all to break the remaing set down into
	// sub sets of non-overlapping circles - we want this as the self-referencing
	// set must be as small as possible, so that users providing multiple SDL/collection operations
	// will result in the same IDs as a single large operation, provided that the individual collection
	// declarations remain the same.

	circularCollectionNames := make([]string, len(collectionsWithRelations))
	for name := range collectionsWithRelations {
		circularCollectionNames = append(circularCollectionNames, name)
	}
	// The order in which ID indexes are assigned must be deterministic, so
	// we must loop through a sorted slice instead of the map.
	slices.Sort(circularCollectionNames)

	var i int
	collectionSetIds := map[string]int{}
	collectionsHit := map[string]struct{}{}
	for _, name := range circularCollectionNames {
		collection := collectionsWithRelations[name]
		mapCollectionSetIDs(&i, collection, collectionSetIds, collectionsWithRelations, collectionsHit)
	}

	collectionSetsByID := map[int][]*client.CollectionVersion{}
	for _, collection := range newCollections {
		collectionSetId, ok := collectionSetIds[collection.Name]
		if !ok {
			// In most cases, if a collection does not form a circular set then it will not be in
			// collectionSetIds, and we can assign it a new, unused setID
			i++
			collectionSetId = i
		}

		collectionSet, ok := collectionSetsByID[collectionSetId]
		if !ok {
			collectionSet = make([]*client.CollectionVersion, 0, 1)
		}

		collectionSet = append(collectionSet, collection)
		collectionSetsByID[collectionSetId] = collectionSet
	}

	collectionSets := [][]*client.CollectionVersion{}
	for _, collectionSet := range collectionSetsByID {
		collectionSets = append(collectionSets, collectionSet)
	}

	return collectionSets
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
	skippedSets := make([][]*client.CollectionVersion, 0, len(collectionSets))
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
						skippedSets = append(skippedSets, collectionSet)
						continue setLoop
					}
				}
			}
		}

		result = append(result, collectionSet)
	}

	if len(skippedSets) > 0 {
		return sortCollectionSetsFrom(len(result), append(result, skippedSets...))
	}

	return result
}

// saveBlocks saves the collection set to the block and headstore.  It mutates the given collectionSet
// setting the ids and migrations.
func saveBlocks(
	ctx context.Context,
	collectionSet []*client.CollectionVersion,
	migration immutable.Option[model.Lens],
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
		delta, hasCollectionChanged := colCRDT.Delta(*collection, oldCol)

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

		newColSources := collection.CollectionSources()
		oldColSources := oldCol.CollectionSources()
		if oldCol.VersionID != "" && len(newColSources) == len(oldColSources) {
			newSource := &client.CollectionSource{
				SourceCollectionID: oldCol.VersionID,
				Transform:          migration,
			}

			if len(newColSources) == 0 {
				collection.Sources = append(collection.Sources, newSource)
			} else if newColSources[0].SourceCollectionID == oldColSources[0].SourceCollectionID {
				oldSourceFound := false
				for i, source := range oldCol.Sources {
					if _, ok := source.(*client.CollectionSource); ok {
						collection.Sources[i] = newSource
						oldSourceFound = true
						break
					}
				}

				if !oldSourceFound {
					collection.Sources = append(collection.Sources, newSource)
				}
			}
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
	allCollectionSets [][]*client.CollectionVersion,
) {
	collectionsByName := map[string]client.CollectionVersion{}
	for _, collectionSet := range allCollectionSets {
		for _, collection := range collectionSet {
			collectionsByName[collection.Name] = *collection
		}
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
	allCollectionSets [][]*client.CollectionVersion,
) {
	collectionsByName := map[string]client.CollectionVersion{}
	for _, collectionSet := range allCollectionSets {
		for _, collection := range collectionSet {
			collectionsByName[collection.Name] = *collection
		}
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
