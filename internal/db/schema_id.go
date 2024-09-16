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
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/internal/core/cid"
)

const schemaSetDeliminator string = "-"

// setSchemaIDs sets all ID fields on a schema description, mutating the input parameter.
//
// This includes RootID (if not already set), VersionID, and relational fields.
func setSchemaIDs(newSchemas []client.SchemaDescription) error {
	// We need to group the inputs and then mutate them, so we temporarily
	// map them to pointers.
	newSchemaPtrs := make([]*client.SchemaDescription, len(newSchemas))
	for i := range newSchemas {
		schema := newSchemas[i]
		newSchemaPtrs[i] = &schema
	}

	schemaSets := getSchemaSets(newSchemaPtrs)

	for _, schemaSet := range schemaSets {
		setID, err := generateSetID(schemaSet)
		if err != nil {
			return err
		}

		assignIDs(setID, schemaSet)
	}

	for i := range newSchemaPtrs {
		newSchemas[i] = *newSchemaPtrs[i]
	}

	substituteRelationFieldKinds(newSchemas)

	return nil
}

// schemaRelations is a trimmed down [client.SchemaDescription] containing
// only the useful information to the functions in this file.
type schemaRelations struct {
	// The name of this schema
	name string

	// The schema names of the primary relations from this schema.
	relations []string
}

// getSchemaSets groups schemas into sets.
//
// Most sets will contain a single schema, however if a circular dependency chain is found
// all elements within that chain will be grouped together into a single set.
//
// For example if User contains a relation *to* Dog, and Dog contains a relationship *to*
// User, they will be grouped into the same set.
func getSchemaSets(newSchemas []*client.SchemaDescription) [][]*client.SchemaDescription {
	schemasWithRelations := map[string]schemaRelations{}
	for _, schema := range newSchemas {
		relations := []string{}
		for _, field := range schema.Fields {
			switch kind := field.Kind.(type) {
			case *client.NamedKind:
				// We only need to worry about use provided `NamedKind` relations in this scope.
				// Other relation kinds can either not be circular, or are relative to the host.
				relations = append(relations, kind.Name)
			default:
				// no-op
			}
		}

		if len(relations) == 0 {
			// If a schema is defined with no relations, then it is not relevant to this function
			// and can be skipped.
			continue
		}

		schemasWithRelations[schema.Name] = schemaRelations{
			name:      schema.Name,
			relations: relations,
		}
	}

	changedInLoop := true
	for changedInLoop {
		// This loop strips out schemas from `schemasWithRelations` that do not form circular
		// schema sets (e.g. User=>Dog=>User).  This allows later logic that figures out the
		// exact path that circles forms to operate on a minimal set of data, reducing its cost
		// and complexity.
		//
		// Some non circular relations may still remain after this first pass, for example
		// one-directional relations between two circles.
		changedInLoop = false
		for _, schema := range schemasWithRelations {
			i := 0
			relation := ""
			deleteI := false
			for i, relation = range schema.relations {
				if _, ok := schemasWithRelations[relation]; !ok {
					// If the related schema is not in `schemasWithRelations` it must have been removed
					// in a previous iteration of the schemasWithRelations loop, this will have been
					// done because it had no relevant remaining relations and thus could not be part
					// of a circular schema set.  If this is the case, this `relation` is also irrelevant
					// here and can be removed as it too cannot form part of a circular schema set.
					changedInLoop = true
					deleteI = true
					break
				}
			}

			if deleteI {
				old := schema.relations
				schema.relations = make([]string, len(schema.relations)-1)
				if i > 0 {
					copy(schema.relations, old[:i-1])
				}
				copy(schema.relations[i:], old[i+1:])
				schemasWithRelations[schema.name] = schema
			}

			if len(schema.relations) == 0 {
				// If there are no relevant relations from this schema, remove the schema from
				// `schemasWithRelations` as the schema cannot form part of a circular schema
				// set.
				changedInLoop = true
				delete(schemasWithRelations, schema.name)
				break
			}
		}
	}

	// If len(schemasWithRelations) > 0 here there are circular relations.
	// We then need to traverse them all to break the remaing set down into
	// sub sets of non-overlapping circles - we want this as the self-referencing
	// set must be as small as possible, so that users providing multiple SDL/schema operations
	// will result in the same IDs as a single large operation, provided that the individual schema
	// declarations remain the same.

	circularSchemaNames := make([]string, len(schemasWithRelations))
	for name := range schemasWithRelations {
		circularSchemaNames = append(circularSchemaNames, name)
	}
	// The order in which ID indexes are assigned must be deterministic, so
	// we must loop through a sorted slice instead of the map.
	slices.Sort(circularSchemaNames)

	var i int
	schemaSetIds := map[string]int{}
	schemasHit := map[string]struct{}{}
	for _, name := range circularSchemaNames {
		schema := schemasWithRelations[name]
		mapSchemaSetIDs(&i, schema, schemaSetIds, schemasWithRelations, schemasHit)
	}

	schemaSetsByID := map[int][]*client.SchemaDescription{}
	for _, schema := range newSchemas {
		schemaSetId, ok := schemaSetIds[schema.Name]
		if !ok {
			// In most cases, if a schema does not form a circular set then it will not be in
			// schemaSetIds, and we can assign it a new, unused setID
			i++
			schemaSetId = i
		}

		schemaSet, ok := schemaSetsByID[schemaSetId]
		if !ok {
			schemaSet = make([]*client.SchemaDescription, 0, 1)
		}

		schemaSet = append(schemaSet, schema)
		schemaSetsByID[schemaSetId] = schemaSet
	}

	schemaSets := [][]*client.SchemaDescription{}
	for _, schemaSet := range schemaSetsByID {
		schemaSets = append(schemaSets, schemaSet)
	}

	return schemaSets
}

// mapSchemaSetIDs recursively scans through a schema and its relations, assigning each schema to a temporary setID.
//
// If a set of schemas form a circular dependency, all involved schemas will be assigned the same setID. Assigned setIDs
// will be added to the input param `schemaSetIds`.
//
// This function will return when all descendents of the initial schema have been processed.
//
// Parameters:
//   - i: The largest setID so far assigned. This parameter is mutated by this function.
//   - schema: The current schema to process
//   - schemaSetIds: The set of already assigned setIDs mapped by schema name - this parameter will be mutated by this
//     function
//   - schemasRelationsBySchemaName: The full set of relevant schemas/relations mapped by schema name
//   - schemasFullyProcessed: The set of schema names that have already been completely processed.  If `schema` is in
//     this set the function will return.  This parameter is mutated by this function.
func mapSchemaSetIDs(
	i *int,
	schema schemaRelations,
	schemaSetIds map[string]int,
	schemasRelationsBySchemaName map[string]schemaRelations,
	schemasFullyProcessed map[string]struct{},
) {
	if _, ok := schemasFullyProcessed[schema.name]; ok {
		// we've circled all the way through and already processed this schema
		return
	}
	schemasFullyProcessed[schema.name] = struct{}{}

	for _, relation := range schema.relations {
		// if more than one relation, need to find out if the relation loops back here! It might connect to a separate circle
		circlesBackHere := circlesBack(schema.name, relation, schemasRelationsBySchemaName, map[string]struct{}{})

		var circleID int
		if circlesBackHere {
			if id, ok := schemaSetIds[relation]; ok {
				// If this schema has already been assigned a setID, use that
				circleID = id
			} else {
				schemaSetId, ok := schemaSetIds[schema.name]
				if !ok {
					// If this schema has not already been assigned a setID, it must be
					// the first discovered node in a new circle.  Assign it a new setID,
					// this will be picked up by its circle-forming descendents.
					*i = *i + 1
					schemaSetId = *i
				}
				schemaSetIds[schema.name] = schemaSetId
				circleID = schemaSetId
			}
		} else {
			// If this schema and its relations does not circle back to itself, we
			// increment `i` and assign the new value to this schema *only*
			*i = *i + 1
			circleID = *i
		}

		schemaSetIds[relation] = circleID
		mapSchemaSetIDs(
			i,
			schemasRelationsBySchemaName[relation],
			schemaSetIds,
			schemasRelationsBySchemaName,
			schemasFullyProcessed,
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
	schemasWithRelations map[string]schemaRelations,
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

func generateSetID(schemaSet []*client.SchemaDescription) (string, error) {
	// The schemas within each set must be in a deterministic order to ensure that
	// their IDs are deterministic.
	slices.SortFunc(schemaSet, func(a, b *client.SchemaDescription) int {
		return strings.Compare(a.Name, b.Name)
	})

	var cidComponents any
	if len(schemaSet) == 1 {
		cidComponents = schemaSet[0]
	} else {
		cidComponents = schemaSet
	}

	buf, err := json.Marshal(cidComponents)
	if err != nil {
		return "", err
	}

	scid, err := cid.NewSHA256CidV1(buf)
	if err != nil {
		return "", err
	}
	return scid.String(), nil
}

func assignIDs(baseID string, schemaSet []*client.SchemaDescription) {
	if len(schemaSet) == 1 {
		schemaSet[0].VersionID = baseID
		if schemaSet[0].Root == "" {
			// Schema Root remains constant through all versions, if it is set at this point
			// do not update it.
			schemaSet[0].Root = baseID
		}
		return
	}

	for i := range schemaSet {
		id := fmt.Sprintf("%s%v%v", baseID, schemaSetDeliminator, i)

		schemaSet[i].VersionID = id
		if schemaSet[i].Root == "" {
			// Schema Root remains constant through all versions, if it is set at this point
			// do not update it.
			schemaSet[i].Root = id
		}
	}
}

// substituteRelationFieldKinds substitutes relations defined using [NamedKind]s to their long-term
// types.
//
// Using names to reference other types is unsuitable as the names may change over time.
func substituteRelationFieldKinds(schemas []client.SchemaDescription) {
	schemasByName := map[string]client.SchemaDescription{}
	for _, schema := range schemas {
		schemasByName[schema.Name] = schema
	}

	for i := range schemas {
		rootComponents := strings.Split(schemas[i].Root, schemaSetDeliminator)
		rootBase := rootComponents[0]

		for j := range schemas[i].Fields {
			switch kind := schemas[i].Fields[j].Kind.(type) {
			case *client.NamedKind:
				relationSchema, ok := schemasByName[kind.Name]
				if !ok {
					// Continue, and let the validation step pick up whatever went wrong later
					continue
				}

				relationRootComponents := strings.Split(relationSchema.Root, schemaSetDeliminator)
				if relationRootComponents[0] == rootBase {
					if len(relationRootComponents) == 2 {
						schemas[i].Fields[j].Kind = client.NewSelfKind(relationRootComponents[1], kind.IsArray())
					} else {
						// If the relation root is simple and does not contain a relative index, then this relation
						// must point to the host schema (self-reference, e.g. User=>User).
						schemas[i].Fields[j].Kind = client.NewSelfKind("", kind.IsArray())
					}
				} else {
					schemas[i].Fields[j].Kind = client.NewSchemaKind(relationSchema.Root, kind.IsArray())
				}

			default:
				// no-op
			}
		}
	}
}
