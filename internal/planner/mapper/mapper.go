// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package mapper

import (
	"context"
	"reflect"
	"strings"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/connor"
	"github.com/sourcenetwork/defradb/internal/core"
)

const (
	// topLevelCollectionName is a dummy collection name to indicate that this item is at the outer most
	// level of the query, typically an aggregate over an entire collection.
	topLevelCollectionName string = "_topLevel"
)

var (
	FilterEqOp = &Operator{Operation: "_eq"}
)

// SelectionType is the type of selection.
type SelectionType int

const (
	ObjectSelection SelectionType = iota
	CommitSelection
)

// ToSelect converts the given [parser.Select] into a [Select].
//
// In the process of doing so it will construct the document map required to access the data
// yielded by the [Select].
func ToSelect(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	selectRequest *request.Select,
) (*Select, error) {
	// the top-level select will always have index=0, and no parent collection name
	return toSelect(ctx, store, rootSelectType, 0, selectRequest, "")
}

// toSelect converts the given [parser.Select] into a [Select].
//
// In the process of doing so it will construct the document map required to access the data
// yielded by the [Select].
func toSelect(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	thisIndex int,
	selectRequest *request.Select,
	parentCollectionName string,
) (*Select, error) {
	if rootSelectType == ObjectSelection && selectRequest.Name == request.VersionFieldName {
		// WARNING: This is a weird quirk upon which some of the mapper code is dependent upon
		// please remove it if/when you have chance to.
		rootSelectType = CommitSelection
	}

	collectionName, err := getCollectionName(ctx, store, rootSelectType, selectRequest, parentCollectionName)
	if err != nil {
		return nil, err
	}

	mapping, definition, err := getTopLevelInfo(ctx, store, rootSelectType, selectRequest, collectionName)
	if err != nil {
		return nil, err
	}

	fields, aggregates, err := getRequestables(ctx, rootSelectType, selectRequest, mapping, collectionName, store)
	if err != nil {
		return nil, err
	}

	// Needs to be done before resolving aggregates, else filter conversion may fail there
	filterDependencies, err := resolveFilterDependencies(
		ctx, store, rootSelectType, collectionName, selectRequest.Filter, mapping, fields)
	if err != nil {
		return nil, err
	}
	fields = append(fields, filterDependencies...)

	// Resolve order dependencies that may have been missed due to not being rendered.
	err = resolveOrderDependencies(
		ctx, store, rootSelectType, collectionName, selectRequest.OrderBy, mapping, &fields)
	if err != nil {
		return nil, err
	}

	aggregates = appendUnderlyingAggregates(aggregates, mapping)
	fields, err = resolveAggregates(
		ctx,
		rootSelectType,
		aggregates,
		fields,
		mapping,
		collectionName,
		definition,
		store,
	)

	if err != nil {
		return nil, err
	}

	if len(definition.Schema.Fields) != 0 {
		fields, err = resolveSecondaryRelationIDs(
			ctx,
			store,
			rootSelectType,
			collectionName,
			definition,
			mapping,
			fields,
		)
		if err != nil {
			return nil, err
		}
	}

	// Resolve groupBy mappings i.e. alias remapping and handle missed inner group.
	if selectRequest.GroupBy.HasValue() {
		groupByFields := selectRequest.GroupBy.Value().Fields
		// Remap all alias field names to use their internal field name mappings.
		for index, groupByField := range groupByFields {
			fieldDesc, ok := definition.GetFieldByName(groupByField)
			if ok && fieldDesc.Kind.IsObject() && !fieldDesc.Kind.IsObjectArray() {
				groupByFields[index] = groupByField + request.RelatedObjectID
			} else if ok && fieldDesc.Kind.IsObjectArray() {
				return nil, NewErrInvalidFieldToGroupBy(groupByField)
			}
		}

		selectRequest.GroupBy = immutable.Some(
			request.GroupBy{
				Fields: groupByFields,
			},
		)

		// If there is a groupBy, and no inner group has been requested, we need to map the property here
		if _, isGroupFieldMapped := mapping.IndexesByName[request.GroupFieldName]; !isGroupFieldMapped {
			index := mapping.GetNextIndex()
			mapping.Add(index, request.GroupFieldName)
		}
	}

	return &Select{
		Targetable:      toTargetable(thisIndex, selectRequest, mapping),
		DocumentMapping: mapping,
		Cid:             selectRequest.CID,
		CollectionName:  collectionName,
		Fields:          fields,
	}, nil
}

// resolveOrderDependencies will map fields that were missed due to them not being requested.
// Modifies the consumed existingFields and mapping accordingly.
func resolveOrderDependencies(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	descName string,
	source immutable.Option[request.OrderBy],
	mapping *core.DocumentMapping,
	existingFields *[]Requestable,
) error {
	if !source.HasValue() {
		return nil
	}

	currentExistingFields := existingFields
	// If there is orderby, and any one of the condition fields that are join fields and have not been
	// requested, we need to map them here.
outer:
	for _, condition := range source.Value().Conditions {
		fields := condition.Fields[:] // copy slice
		for {
			numFields := len(fields)
			// <2 fields: Direct field on the root type: {age: DESC}
			// 2 fields: Single depth related type: {author: {age: DESC}}
			// >2 fields: Multi depth related type: {author: {friends: {age: DESC}}}
			if numFields == 2 {
				joinField := fields[0]

				// ensure the child select is resolved for this order join
				innerSelect, err := resolveChildOrder(
					ctx,
					store,
					rootSelectType,
					descName,
					joinField,
					mapping,
					currentExistingFields,
				)
				if err != nil {
					return err
				}

				// make sure the actual target field inside the join field
				// is included in the select
				targetFieldName := fields[1]
				targetField := &Field{
					Index: innerSelect.FirstIndexOfName(targetFieldName),
					Name:  targetFieldName,
				}
				innerSelect.Fields = append(innerSelect.Fields, targetField)
				continue outer
			} else if numFields > 2 {
				joinField := fields[0]

				// ensure the child select is resolved for this order join
				innerSelect, err := resolveChildOrder(ctx, store, rootSelectType, descName, joinField, mapping, existingFields)
				if err != nil {
					return err
				}
				mapping = innerSelect.DocumentMapping
				currentExistingFields = &innerSelect.Fields
				fields = fields[1:] // chop off the front item, and loop again on inner
			} else { // <= 1
				targetFieldName := fields[0]
				*existingFields = append(*existingFields, &Field{
					Index: mapping.FirstIndexOfName(targetFieldName),
					Name:  targetFieldName,
				})
				// nothing todo, continue the outer for loop
				continue outer
			}
		}
	}

	return nil
}

// given a type join field, ensure its mapping exists
// and add a coorsponding select field(s)
func resolveChildOrder(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	descName string,
	orderChildField string,
	mapping *core.DocumentMapping,
	existingFields *[]Requestable,
) (*Select, error) {
	childFieldIndexes := mapping.IndexesByName[orderChildField]
	// Check if the join field is already mapped, if not then map it.
	if len(childFieldIndexes) == 0 {
		index := mapping.GetNextIndex()
		mapping.Add(index, orderChildField)

		// Resolve the inner child fields and get it's mapping.
		dummyJoinFieldSelect := request.Select{
			Field: request.Field{
				Name: orderChildField,
			},
		}
		innerSelect, err := toSelect(ctx, store, rootSelectType, index, &dummyJoinFieldSelect, descName)
		if err != nil {
			return nil, err
		}
		*existingFields = append(*existingFields, innerSelect)
		mapping.SetChildAt(index, innerSelect.DocumentMapping)
		return innerSelect, nil
	} else {
		for _, field := range *existingFields {
			fieldSelect, ok := field.(*Select)
			if !ok {
				continue
			}
			if fieldSelect.Field.Name == orderChildField {
				return fieldSelect, nil
			}
		}
	}
	return nil, ErrMissingSelect
}

// resolveAggregates figures out which fields the given aggregates are targeting
// and converts the aggregateRequest into an Aggregate, appending it onto the given
// fields slice.
//
// If an aggregate targets a field that doesn't yet exist, it will create it and
// append the new target field as well as the aggregate.  The mapping will also be
// updated with any new fields/aggregates.
func resolveAggregates(
	ctx context.Context,
	rootSelectType SelectionType,
	aggregates []*aggregateRequest,
	inputFields []Requestable,
	mapping *core.DocumentMapping,
	collectionName string,
	def client.CollectionDefinition,
	store client.Store,
) ([]Requestable, error) {
	fields := inputFields
	dependenciesByParentId := map[int][]int{}
	for _, aggregate := range aggregates {
		aggregateTargets := make([]AggregateTarget, len(aggregate.targets))

		for i, target := range aggregate.targets {
			var host Requestable
			var hostTarget *Targetable
			var childTarget OptionalChildTarget

			// If the host has not been requested the child mapping may not yet exist and
			// we must create it before we can convert the filter.
			childIsMapped := len(mapping.IndexesByName[target.hostExternalName]) != 0

			var hasHost bool
			var convertedFilter *Filter
			if childIsMapped {
				fieldDesc, isField := def.GetFieldByName(target.hostExternalName)

				if isField && !fieldDesc.Kind.IsObject() {
					var order *OrderBy
					if target.order.HasValue() && len(target.order.Value().Conditions) > 0 {
						// For inline arrays the order element will consist of just a direction
						order = &OrderBy{
							Conditions: []OrderCondition{
								{
									Direction: SortDirection(target.order.Value().Conditions[0].Direction),
								},
							},
						}
					}

					// If the hostExternalName matches a non-object field
					// we don't have to search for it and can just construct the
					// targeting info here.
					hasHost = true
					host = &Targetable{
						Field: Field{
							Index: int(fieldDesc.ID),
							Name:  target.hostExternalName,
						},
						Filter:  ToFilter(target.filter.Value(), mapping),
						Limit:   target.limit,
						OrderBy: order,
					}
				} else {
					childObjectIndex := mapping.FirstIndexOfName(target.hostExternalName)
					childMapping := mapping.ChildMappings[childObjectIndex]
					convertedFilter = ToFilter(target.filter.Value(), childMapping)
					host, hasHost = tryGetTarget(
						target.hostExternalName,
						convertedFilter,
						target.limit,
						toOrderBy(target.order, childMapping),
						fields,
					)
				}
			}

			if !hasHost {
				// If a matching host is not found, we need to construct and add it.
				index := mapping.GetNextIndex()

				hostSelectRequest := &request.Select{
					Field: request.Field{
						Name: target.hostExternalName,
					},
				}

				if collectionName == topLevelCollectionName {
					collectionName = ""
				}

				childCollectionName, err := getCollectionName(ctx, store, rootSelectType, hostSelectRequest, collectionName)
				if err != nil {
					return nil, err
				}
				mapAggregateNestedTargets(target, hostSelectRequest)

				childMapping, _, err := getTopLevelInfo(ctx, store, rootSelectType, hostSelectRequest, childCollectionName)
				if err != nil {
					return nil, err
				}

				childFields, _, err := getRequestables(
					ctx,
					rootSelectType,
					hostSelectRequest,
					childMapping,
					childCollectionName,
					store,
				)
				if err != nil {
					return nil, err
				}

				err = resolveOrderDependencies(
					ctx, store, rootSelectType, childCollectionName, target.order, childMapping, &childFields)
				if err != nil {
					return nil, err
				}

				childMapping = childMapping.CloneWithoutRender()
				mapping.SetChildAt(index, childMapping)

				filterDependencies, err := resolveFilterDependencies(
					ctx,
					store,
					rootSelectType,
					childCollectionName,
					target.filter,
					mapping.ChildMappings[index],
					childFields,
				)
				if err != nil {
					return nil, err
				}
				childFields = append(childFields, filterDependencies...)

				// If the child was not mapped, the filter will not have been converted yet
				// so we must do that now.
				convertedFilter = ToFilter(target.filter.Value(), mapping.ChildMappings[index])

				dummyJoin := &Select{
					Targetable: Targetable{
						Field: Field{
							Index: index,
							Name:  target.hostExternalName,
						},
						Filter:  convertedFilter,
						Limit:   target.limit,
						OrderBy: toOrderBy(target.order, childMapping),
					},
					CollectionName:  childCollectionName,
					DocumentMapping: childMapping,
					Fields:          childFields,
				}

				fields = append(fields, dummyJoin)
				mapping.Add(index, target.hostExternalName)

				host = dummyJoin
				hostTarget = &dummyJoin.Targetable
			} else {
				var isTargetable bool
				hostTarget, isTargetable = host.AsTargetable()
				if !isTargetable {
					// If the host is not targetable, such as when it is an inline-array field,
					// we don't need to worry about preserving the targetable information and
					// can just take the field properties.
					hostTarget = &Targetable{
						Field: Field{
							Index: host.GetIndex(),
							Name:  host.GetName(),
						},
					}
				}
			}

			if target.childExternalName != "" {
				hostSelect, isHostSelectable := host.AsSelect()
				if !isHostSelectable {
					// I believe this is dead code as the gql library should always catch this error first
					return nil, client.NewErrUnhandledType("host", host)
				}

				if len(hostSelect.IndexesByName[target.childExternalName]) == 0 {
					// I believe this is dead code as the gql library should always catch this error first
					return nil, ErrUnableToIdAggregateChild
				}

				// ensure target aggregate field is included in the type join
				hostSelect.Fields = append(hostSelect.Fields, &Field{
					Index: hostSelect.DocumentMapping.FirstIndexOfName(target.childExternalName),
					Name:  target.childExternalName,
				})

				childTarget = OptionalChildTarget{
					// If there are multiple children of the same name there is no way
					// for us (or the consumer) to identify which one they are hoping for
					// so we take the first.
					Index:    hostSelect.IndexesByName[target.childExternalName][0],
					Name:     target.childExternalName,
					HasValue: true,
				}
			}

			aggregateTargets[i] = AggregateTarget{
				Targetable:  *hostTarget,
				ChildTarget: childTarget,
			}
		}

		newAggregate := Aggregate{
			Field:            aggregate.field,
			DocumentMapping:  mapping,
			AggregateTargets: aggregateTargets,
		}
		fields = append(fields, &newAggregate)
		dependenciesByParentId[aggregate.field.Index] = aggregate.dependencyIndexes
	}

	// Once aggregates have been resolved we pair up their dependencies
	for aggregateId, dependencyIds := range dependenciesByParentId {
		aggregate := fieldAt(fields, aggregateId).(*Aggregate)
		for _, dependencyId := range dependencyIds {
			aggregate.Dependencies = append(aggregate.Dependencies, fieldAt(fields, dependencyId).(*Aggregate))
		}
	}

	return fields, nil
}

func mapAggregateNestedTargets(
	target *aggregateRequestTarget,
	hostSelectRequest *request.Select,
) {
	if target.order.HasValue() {
		for _, cond := range target.order.Value().Conditions {
			if len(cond.Fields) > 1 {
				hostSelectRequest.Fields = append(hostSelectRequest.Fields, &request.Select{
					Field: request.Field{
						Name: cond.Fields[0],
					},
				})
			}
		}
	}

	if target.filter.HasValue() {
		for topKey, topCond := range target.filter.Value().Conditions {
			switch cond := topCond.(type) {
			case map[string]any:
				for _, innerCond := range cond {
					if _, isMap := innerCond.(map[string]any); isMap {
						hostSelectRequest.Fields = append(hostSelectRequest.Fields, &request.Select{
							Field: request.Field{
								Name: topKey,
							},
						})
						break
					}
				}
			}
		}
	}
}

func fieldAt(fields []Requestable, index int) Requestable {
	for _, f := range fields {
		if f.GetIndex() == index {
			return f
		}
	}
	return nil
}

// aggregateDependencies maps aggregate names to the names of any aggregates
// that they may be dependent on.
var aggregateDependencies = map[string][]string{
	request.AverageFieldName: {
		request.CountFieldName,
		request.SumFieldName,
	},
}

// appendUnderlyingAggregates scans the given inputAggregates for any composite aggregates
// (e.g. average), and appends any missing dependencies to the collection and mapping.
//
// It will try and make use of existing aggregates that match the targeting parameters
// before creating new ones.  It will also adjust the target filters if required (e.g.
// average skips nil items).
func appendUnderlyingAggregates(
	inputAggregates []*aggregateRequest,
	mapping *core.DocumentMapping,
) []*aggregateRequest {
	aggregates := inputAggregates

	// Loop through the aggregates slice, including items that may have been appended
	// to the slice whilst looping.
	for i := 0; i < len(aggregates); i++ {
		aggregate := aggregates[i]

		dependencies, hasDependencies := aggregateDependencies[aggregate.field.Name]
		// If the aggregate has no dependencies, then we don't need to do anything and we continue.
		if !hasDependencies {
			continue
		}

		for _, target := range aggregate.targets {
			if target.childExternalName != "" {
				if _, isAggregate := request.Aggregates[target.childExternalName]; isAggregate {
					continue
				}
			}
			// Append a not-nil filter if the target is not an aggregate.
			// If the target has no childExternalName we assume it is an inline-array (and thus not an aggregate).
			// Aggregate-targets are excluded here as they are assumed to always have a value and
			// amending the filter introduces significant complexity for both machine and developer.
			appendNotNilFilter(target, target.childExternalName)
		}

		for _, dependencyName := range dependencies {
			var newAggregate *aggregateRequest
			aggregates, newAggregate = appendIfNotExists(
				dependencyName,
				aggregate.targets,
				aggregates,
				mapping,
			)
			aggregate.dependencyIndexes = append(aggregate.dependencyIndexes, newAggregate.field.Index)
		}
	}
	return aggregates
}

// appendIfNotExists attempts to match the given name and targets against existing
// aggregates, if a match is not found, it will append a new aggregate.
func appendIfNotExists(
	name string,
	targets []*aggregateRequestTarget,
	aggregates []*aggregateRequest,
	mapping *core.DocumentMapping,
) ([]*aggregateRequest, *aggregateRequest) {
	field, exists := tryGetMatchingAggregate(name, targets, aggregates)
	if exists {
		// If a match is found, there is nothing to do so we return the aggregates slice unchanged.
		return aggregates, field
	}

	// If a match is not found, create, map and append the
	// dependency to the aggregates collection.
	index := mapping.GetNextIndex()

	field = &aggregateRequest{
		field: Field{
			Index: index,
			Name:  name,
		},
		targets: targets,
	}

	mapping.Add(index, field.field.Name)
	return append(aggregates, field), field
}

// getRequestables returns a converted slice of consumer-requested Requestables
// and aggregateRequests from the given selectRequest.Fields slice. It also mutates the
// consumed mapping data.
func getRequestables(
	ctx context.Context,
	rootSelectType SelectionType,
	selectRequest *request.Select,
	mapping *core.DocumentMapping,
	collectionName string,
	store client.Store,
) (fields []Requestable, aggregates []*aggregateRequest, err error) {
	for _, field := range selectRequest.Fields {
		switch f := field.(type) {
		case *request.Field:
			// We can map all fields to the first (and only index)
			// as they support no value modifiers (such as filters/limits/etc).
			// All fields should have already been mapped by getTopLevelInfo
			index := mapping.FirstIndexOfName(f.Name)

			fields = append(fields, &Field{
				Index: index,
				Name:  f.Name,
			})

			mapping.RenderKeys = append(mapping.RenderKeys, core.RenderKey{
				Index: index,
				Key:   getRenderKey(f),
			})
		case *request.Select:
			index := mapping.GetNextIndex()

			innerSelect, err := toSelect(ctx, store, rootSelectType, index, f, collectionName)
			if err != nil {
				return nil, nil, err
			}
			fields = append(fields, innerSelect)
			mapping.SetChildAt(index, innerSelect.DocumentMapping)

			mapping.RenderKeys = append(mapping.RenderKeys, core.RenderKey{
				Index: index,
				Key:   getRenderKey(&f.Field),
			})

			mapping.Add(index, f.Name)
		case *request.Aggregate:
			index := mapping.GetNextIndex()
			aggregateRequest, err := getAggregateRequests(index, f)
			if err != nil {
				return nil, nil, err
			}

			aggregates = append(aggregates, &aggregateRequest)

			mapping.RenderKeys = append(mapping.RenderKeys, core.RenderKey{
				Index: index,
				Key:   getRenderKey(&f.Field),
			})

			mapping.Add(index, f.Name)
		default:
			return nil, nil, client.NewErrUnhandledType("field", field)
		}
	}
	return
}

func getRenderKey(field *request.Field) string {
	if field.Alias.HasValue() {
		return field.Alias.Value()
	}
	return field.Name
}

func getAggregateRequests(index int, aggregate *request.Aggregate) (aggregateRequest, error) {
	aggregateTargets, err := getAggregateSources(aggregate)
	if err != nil {
		return aggregateRequest{}, err
	}

	if len(aggregateTargets) == 0 {
		return aggregateRequest{}, ErrAggregateTargetMissing
	}

	return aggregateRequest{
		field: Field{
			Index: index,
			Name:  aggregate.Name,
		},
		targets: aggregateTargets,
	}, nil
}

// getCollectionName returns the name of the selectRequest collection.  This may be empty
// if this is a commit request.
func getCollectionName(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	selectRequest *request.Select,
	parentCollectionName string,
) (string, error) {
	if _, isAggregate := request.Aggregates[selectRequest.Name]; isAggregate {
		return topLevelCollectionName, nil
	}

	if selectRequest.Name == request.GroupFieldName {
		return parentCollectionName, nil
	} else if rootSelectType == CommitSelection {
		return parentCollectionName, nil
	}

	if parentCollectionName != "" {
		parentCollection, err := store.GetCollectionByName(ctx, parentCollectionName)
		if err != nil {
			return "", err
		}

		hostFieldDesc, parentHasField := parentCollection.Definition().GetFieldByName(selectRequest.Name)
		if parentHasField && hostFieldDesc.Kind.IsObject() {
			// If this field exists on the parent, and it is a child object
			// then this collection name is the collection name of the child.
			return hostFieldDesc.Kind.Underlying(), nil
		}
	}

	return selectRequest.Name, nil
}

// getTopLevelInfo returns the collection description and maps the fields directly on the object.
func getTopLevelInfo(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	selectRequest *request.Select,
	collectionName string,
) (*core.DocumentMapping, client.CollectionDefinition, error) {
	mapping := core.NewDocumentMapping()

	if _, isAggregate := request.Aggregates[selectRequest.Name]; isAggregate {
		// If this is a (top-level) aggregate, then it will have no collection
		// description, and no top-level fields, so we return an empty mapping only
		return mapping, client.CollectionDefinition{}, nil
	}

	if rootSelectType == ObjectSelection {
		var definition client.CollectionDefinition
		collection, err := store.GetCollectionByName(ctx, collectionName)
		if err != nil {
			// If the collection is not found, check to see if a schema of that name exists,
			// if so, this must be an embedded object.
			//
			// Note: This is a poor way to check if a collection exists or not, see
			// https://github.com/sourcenetwork/defradb/issues/2146
			schemas, err := store.GetSchemas(
				ctx,
				client.SchemaFetchOptions{
					Name: immutable.Some(collectionName),
				},
			)
			if err != nil {
				return nil, client.CollectionDefinition{}, err
			}
			if len(schemas) == 0 {
				return nil, client.CollectionDefinition{}, NewErrTypeNotFound(collectionName)
			}

			for i, f := range schemas[0].Fields {
				// As embedded objects do not have collections/field-ids, we just take the index
				mapping.Add(int(i), f.Name)
			}

			definition = client.CollectionDefinition{
				// `schemas` will contain all versions of that name, as views cannot be updated atm this should
				// be fine for now
				Schema: schemas[0],
			}
		} else {
			mapping.Add(core.DocIDFieldIndex, request.DocIDFieldName)
			definition = collection.Definition()
			// Map all fields from schema into the map as they are fetched automatically
			for _, f := range definition.GetFields() {
				if f.Kind.IsObject() {
					// Objects are skipped, as they are not fetched by default and
					// have to be requested via selects.
					continue
				}
				mapping.Add(int(f.ID), f.Name)
			}
		}

		// Setting the type name must be done after adding the fields, as
		// the typeName index is dynamic, but the field indexes are not
		mapping.SetTypeName(collectionName)

		mapping.Add(mapping.GetNextIndex(), request.DeletedFieldName)

		return mapping, definition, nil
	}

	if selectRequest.Name == request.LinksFieldName {
		for i, f := range request.LinksFields {
			mapping.Add(i, f)
		}

		// Setting the type name must be done after adding the fields, as
		// the typeName index is dynamic, but the field indexes are not
		mapping.SetTypeName(request.LinksFieldName)
	} else {
		for i, f := range request.VersionFields {
			mapping.Add(i, f)
		}

		// Setting the type name must be done after adding the fields, as
		// the typeName index is dynamic, but the field indexes are not
		mapping.SetTypeName(request.CommitTypeName)
	}

	return mapping, client.CollectionDefinition{}, nil
}

func resolveFilterDependencies(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	parentCollectionName string,
	source immutable.Option[request.Filter],
	mapping *core.DocumentMapping,
	existingFields []Requestable,
) ([]Requestable, error) {
	if !source.HasValue() {
		return nil, nil
	}

	return resolveInnerFilterDependencies(
		ctx,
		store,
		rootSelectType,
		parentCollectionName,
		source.Value().Conditions,
		mapping,
		existingFields,
		nil,
	)
}

func resolveInnerFilterDependencies(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	parentCollectionName string,
	source map[string]any,
	mapping *core.DocumentMapping,
	existingFields []Requestable,
	resolvedFields []Requestable,
) ([]Requestable, error) {
	newFields := []Requestable{}

	for key := range source {
		if key == request.FilterOpAnd || key == request.FilterOpOr {
			compoundFilter := source[key].([]any)
			for _, innerFilter := range compoundFilter {
				innerFields, err := resolveInnerFilterDependencies(
					ctx,
					store,
					rootSelectType,
					parentCollectionName,
					innerFilter.(map[string]any),
					mapping,
					existingFields,
					resolvedFields,
				)
				if err != nil {
					return nil, err
				}

				resolvedFields = append(resolvedFields, innerFields...)
				newFields = append(newFields, innerFields...)
			}
			continue
		} else if key == request.FilterOpNot {
			notFilter := source[key].(map[string]any)
			innerFields, err := resolveInnerFilterDependencies(
				ctx,
				store,
				rootSelectType,
				parentCollectionName,
				notFilter,
				mapping,
				existingFields,
				resolvedFields,
			)
			if err != nil {
				return nil, err
			}

			resolvedFields = append(resolvedFields, innerFields...)
			newFields = append(newFields, innerFields...)
			continue
		}

		propertyMapped := len(mapping.IndexesByName[key]) != 0

		var childSelect *Select
		if propertyMapped {
			var field Requestable
			for _, f := range existingFields {
				if f.GetIndex() == mapping.FirstIndexOfName(key) {
					field = f
					break
				}
			}
			for _, f := range resolvedFields {
				if f.GetIndex() == mapping.FirstIndexOfName(key) {
					field = f
					break
				}
			}
			if field == nil {
				newFields = append(newFields, &Field{Index: mapping.FirstIndexOfName(key), Name: key})
				continue
			}
			var isSelect bool
			childSelect, isSelect = field.(*Select)
			if !isSelect {
				continue
			}
		} else {
			var err error
			childSelect, err = constructEmptyJoin(ctx, store, rootSelectType, parentCollectionName, mapping, key)
			if err != nil {
				return nil, err
			}

			childSelect.SkipResolve = true
			newFields = append(newFields, childSelect)
		}

		childSource := source[key]
		childFilter, isChildFilter := childSource.(map[string]any)
		if !isChildFilter {
			// If the filter is not a child filter then the will be no inner dependencies to add and
			// we can continue.
			continue
		}

		dummyParsed := &request.Select{Field: request.Field{Name: key}}
		childCollectionName, err := getCollectionName(ctx, store, rootSelectType, dummyParsed, parentCollectionName)
		if err != nil {
			return nil, err
		}

		childFields, err := resolveInnerFilterDependencies(
			ctx,
			store,
			rootSelectType,
			childCollectionName,
			childFilter,
			childSelect.DocumentMapping,
			childSelect.Fields,
			nil,
		)
		if err != nil {
			return nil, err
		}

		childSelect.Fields = append(childSelect.Fields, childFields...)
	}

	return newFields, nil
}

// constructEmptyJoin constructs a valid empty join with no requested fields.
func constructEmptyJoin(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	parentCollectionName string,
	parentMapping *core.DocumentMapping,
	name string,
) (*Select, error) {
	index := parentMapping.GetNextIndex()

	dummyParsed := &request.Select{
		Field: request.Field{
			Name: name,
		},
	}

	childCollectionName, err := getCollectionName(ctx, store, rootSelectType, dummyParsed, parentCollectionName)
	if err != nil {
		return nil, err
	}

	childMapping, _, err := getTopLevelInfo(ctx, store, rootSelectType, dummyParsed, childCollectionName)
	if err != nil {
		return nil, err
	}
	childMapping = childMapping.CloneWithoutRender()
	parentMapping.SetChildAt(index, childMapping)
	parentMapping.Add(index, name)

	return &Select{
		Targetable: Targetable{
			Field: Field{
				Index: index,
				Name:  name,
			},
		},
		CollectionName:  childCollectionName,
		DocumentMapping: childMapping,
	}, nil
}

// resolveSecondaryRelationIDs constructs the required stuff needed to resolve secondary relation ids.
//
// They are handled by joining (if not already done so) the related object and copying its key into the
// secondary relation id field.
//
// They copying itself is handled within [typeJoinOne].
func resolveSecondaryRelationIDs(
	ctx context.Context,
	store client.Store,
	rootSelectType SelectionType,
	collectionName string,
	schema client.CollectionDefinition,
	mapping *core.DocumentMapping,
	requestables []Requestable,
) ([]Requestable, error) {
	fields := requestables

	for _, requestable := range requestables {
		existingField, isField := requestable.(*Field)
		if !isField {
			continue
		}

		fieldDesc, descFound := schema.GetFieldByName(existingField.Name)
		if !descFound {
			continue
		}

		if fieldDesc.Kind != client.FieldKind_DocID {
			continue
		}

		objectFieldName := strings.TrimSuffix(existingField.Name, request.RelatedObjectID)

		var siblingFound bool
		for _, siblingRequestable := range requestables {
			if siblingRequestable.GetName() == objectFieldName {
				siblingFound = true
				break
			}
		}

		if !siblingFound {
			objectFieldName := strings.TrimSuffix(existingField.Name, request.RelatedObjectID)

			// We only require the docID of the related object, so an empty join is all we need.
			join, err := constructEmptyJoin(
				ctx,
				store,
				rootSelectType,
				collectionName,
				mapping,
				objectFieldName,
			)
			if err != nil {
				return nil, err
			}

			fields = append(fields, join)
		}
	}

	return fields, nil
}

// ToCommitSelect converts the given [request.CommitSelect] into a [CommitSelect].
//
// In the process of doing so it will construct the document map required to access the data
// yielded by the [Select] embedded in the [CommitSelect].
func ToCommitSelect(
	ctx context.Context,
	store client.Store,
	selectRequest *request.CommitSelect,
) (*CommitSelect, error) {
	underlyingSelect, err := ToSelect(ctx, store, CommitSelection, selectRequest.ToSelect())
	if err != nil {
		return nil, err
	}
	return &CommitSelect{
		Select:  *underlyingSelect,
		DocID:   selectRequest.DocID,
		FieldID: selectRequest.FieldID,
		Depth:   selectRequest.Depth,
		Cid:     selectRequest.CID,
	}, nil
}

// ToMutation converts the given [request.Mutation] into a [Mutation].
//
// In the process of doing so it will construct the document map required to access the data
// yielded by the [Select] embedded in the [Mutation].
func ToMutation(ctx context.Context, store client.Store, mutationRequest *request.ObjectMutation) (*Mutation, error) {
	underlyingSelect, err := ToSelect(ctx, store, ObjectSelection, mutationRequest.ToSelect())
	if err != nil {
		return nil, err
	}

	return &Mutation{
		Select:        *underlyingSelect,
		Type:          MutationType(mutationRequest.Type),
		Input:         mutationRequest.Input,
		Inputs:        mutationRequest.Inputs,
		Encrypt:       mutationRequest.Encrypt,
		EncryptFields: mutationRequest.EncryptFields,
	}, nil
}

func toTargetable(index int, selectRequest *request.Select, docMap *core.DocumentMapping) Targetable {
	return Targetable{
		Field:       toField(index, selectRequest),
		DocIDs:      selectRequest.DocIDs,
		Filter:      ToFilter(selectRequest.Filter.Value(), docMap),
		Limit:       toLimit(selectRequest.Limit, selectRequest.Offset),
		GroupBy:     toGroupBy(selectRequest.GroupBy, docMap),
		OrderBy:     toOrderBy(selectRequest.OrderBy, docMap),
		ShowDeleted: selectRequest.ShowDeleted,
	}
}

func toField(index int, selectRequest *request.Select) Field {
	return Field{
		Index: index,
		Name:  selectRequest.Name,
	}
}

// ToFilter converts the given `source` request filter to a Filter using the given mapping.
//
// Any requestables identified by name will be converted to being identified by index instead.
func ToFilter(source request.Filter, mapping *core.DocumentMapping) *Filter {
	if len(source.Conditions) == 0 {
		return nil
	}
	conditions := make(map[connor.FilterKey]any, len(source.Conditions))

	for sourceKey, sourceClause := range source.Conditions {
		key, clause := toFilterMap(sourceKey, sourceClause, mapping)
		conditions[key] = clause
	}

	return &Filter{
		Conditions:         conditions,
		ExternalConditions: source.Conditions,
	}
}

// toFilterMap converts a consumer-defined filter key-value into a filter clause
// keyed by field index.
//
// Return key will either be an int (field index), or a string (operator).
func toFilterMap(
	sourceKey string,
	sourceClause any,
	mapping *core.DocumentMapping,
) (connor.FilterKey, any) {
	if strings.HasPrefix(sourceKey, "_") && sourceKey != request.DocIDFieldName {
		key := &Operator{
			Operation: sourceKey,
		}
		switch typedClause := sourceClause.(type) {
		case []any:
			// If the clause is an array then we need to convert any inner maps.
			returnClauses := []any{}
			for _, innerSourceClause := range typedClause {
				var returnClause any
				switch typedInnerSourceClause := innerSourceClause.(type) {
				case map[string]any:
					innerMapClause := map[connor.FilterKey]any{}
					for innerSourceKey, innerSourceValue := range typedInnerSourceClause {
						rKey, rValue := toFilterMap(innerSourceKey, innerSourceValue, mapping)
						innerMapClause[rKey] = rValue
					}
					returnClause = innerMapClause
				default:
					returnClause = innerSourceClause
				}
				returnClauses = append(returnClauses, returnClause)
			}
			return key, returnClauses
		case map[string]any:
			innerMapClause := map[connor.FilterKey]any{}
			for innerSourceKey, innerSourceValue := range typedClause {
				rKey, rValue := toFilterMap(innerSourceKey, innerSourceValue, mapping)
				innerMapClause[rKey] = rValue
			}
			return key, innerMapClause
		default:
			return key, typedClause
		}
	} else {
		// If there are multiple properties of the same name we can just take the first as
		// we have no other reasonable way of identifying which property they mean if multiple
		// consumer specified requestables are available.  Aggregate dependencies should not
		// impact this as they are added after selects.
		index := mapping.FirstIndexOfName(sourceKey)
		key := &PropertyIndex{
			Index: index,
		}
		switch typedClause := sourceClause.(type) {
		case map[string]any:
			returnClause := map[connor.FilterKey]any{}
			for innerSourceKey, innerSourceValue := range typedClause {
				var innerMapping *core.DocumentMapping
				switch innerSourceValue.(type) {
				case map[string]any:
					// If the innerSourceValue is also a map, then we should parse the nested clause
					// using the child mapping, as this key must refer to a host property in a join
					// and deeper keys must refer to properties on the child items.
					innerMapping = mapping.ChildMappings[index]
				default:
					innerMapping = mapping
				}
				rKey, rValue := toFilterMap(innerSourceKey, innerSourceValue, innerMapping)
				returnClause[rKey] = rValue
			}
			return key, returnClause
		default:
			return key, sourceClause
		}
	}
}

func toLimit(limit immutable.Option[uint64], offset immutable.Option[uint64]) *Limit {
	var limitValue uint64
	var offsetValue uint64
	if !limit.HasValue() && !offset.HasValue() {
		return nil
	}

	if limit.HasValue() {
		limitValue = limit.Value()
	}

	if offset.HasValue() {
		offsetValue = offset.Value()
	}

	return &Limit{
		Limit:  limitValue,
		Offset: offsetValue,
	}
}

func toGroupBy(source immutable.Option[request.GroupBy], mapping *core.DocumentMapping) *GroupBy {
	if !source.HasValue() {
		return nil
	}

	fields := make([]Field, len(source.Value().Fields))
	for i, fieldName := range source.Value().Fields {
		// If there are multiple properties of the same name we can just take the first as
		// we have no other reasonable way of identifying which property they mean if multiple
		// consumer specified requestables are available.  Aggregate dependencies should not
		// impact this as they are added after selects.
		key := mapping.FirstIndexOfName(fieldName)

		fields[i] = Field{
			Index: key,
			Name:  fieldName,
		}
	}

	return &GroupBy{
		Fields: fields,
	}
}

func toOrderBy(source immutable.Option[request.OrderBy], mapping *core.DocumentMapping) *OrderBy {
	if !source.HasValue() {
		return nil
	}

	conditions := make([]OrderCondition, len(source.Value().Conditions))
	for conditionIndex, condition := range source.Value().Conditions {
		fieldIndexes := make([]int, len(condition.Fields))
		currentMapping := mapping
		for fieldIndex, field := range condition.Fields {
			// If there are multiple properties of the same name we can just take the first as
			// we have no other reasonable way of identifying which property they mean if multiple
			// consumer specified requestables are available.  Aggregate dependencies should not
			// impact this as they are added after selects.
			firstFieldIndex := currentMapping.FirstIndexOfName(field)
			fieldIndexes[fieldIndex] = firstFieldIndex
			if fieldIndex != len(condition.Fields)-1 {
				// no need to do this for the last (and will panic)
				currentMapping = currentMapping.ChildMappings[firstFieldIndex]
			}
		}

		conditions[conditionIndex] = OrderCondition{
			FieldIndexes: fieldIndexes,
			Direction:    SortDirection(condition.Direction),
		}
	}

	return &OrderBy{
		Conditions: conditions,
	}
}

// RunFilter runs the given filter expression
// using the document, and evaluates.
func RunFilter(doc any, filter *Filter) (bool, error) {
	if filter == nil {
		return true, nil
	}

	return connor.Match(filter.Conditions, doc)
}

// equal compares the given Targetables and returns true if they can be considered equal.
func (s Targetable) equal(other Targetable) bool {
	if s.Index != other.Index &&
		s.Name != other.Name {
		return false
	}

	if !s.Filter.equal(other.Filter) {
		return false
	}

	if !s.Limit.equal(other.Limit) {
		return false
	}

	if !s.OrderBy.equal(other.OrderBy) {
		return false
	}

	return true
}

func (l *Limit) equal(other *Limit) bool {
	if l == nil {
		return other == nil
	}

	if other == nil {
		return l == nil
	}

	return l.Limit == other.Limit && l.Offset == other.Offset
}

func (f *Filter) equal(other *Filter) bool {
	if f == nil {
		return other == nil
	}

	if other == nil {
		return f == nil
	}

	return deepEqualConditions(f.Conditions, other.Conditions)
}

// deepEqualConditions performs a deep equality of two conditions.
// Handles: map[0xc00069cfd0:map[0xc005eda8c0:<nil>]] -> map[{5}:map[{_ne}:<nil>]]
func deepEqualConditions(x map[connor.FilterKey]any, y map[connor.FilterKey]any) bool {
	if len(x) != len(y) {
		return false
	}

	for xKey, xValue := range x {
		var isFoundInY bool

		// Ensure a matching key exists in the other map.
		for yKey, yValue := range y {
			if !xKey.Equal(yKey) {
				continue
			}

			xValueConditions, xOk := xValue.(map[connor.FilterKey]any)
			yValueConditions, yOk := yValue.(map[connor.FilterKey]any)
			if xOk && yOk {
				if deepEqualConditions(xValueConditions, yValueConditions) {
					isFoundInY = true
					break
				}
			} else if !xOk && !yOk {
				// Both are some basic values.
				if reflect.DeepEqual(xValue, yValue) {
					isFoundInY = true
					break
				}
			}
		}

		// No matching key (including matching data, of the pointer keys) found, so exit early.
		if !isFoundInY {
			return false
		}
	}

	return true
}

func (o *OrderBy) equal(other *OrderBy) bool {
	if o == nil {
		return other == nil
	}

	if other == nil {
		return o == nil
	}

	if len(o.Conditions) != len(other.Conditions) {
		return false
	}

	for i, conditionA := range o.Conditions {
		conditionB := other.Conditions[i]
		if conditionA.Direction != conditionB.Direction {
			return false
		}

		if len(conditionA.FieldIndexes) != len(conditionB.FieldIndexes) {
			return false
		}

		for j, fieldIndexA := range conditionA.FieldIndexes {
			fieldIndexB := conditionB.FieldIndexes[j]
			if fieldIndexA != fieldIndexB {
				return false
			}
		}
	}

	return true
}

// aggregateRequest is an intermediary struct defining a consumer-requested
// aggregate. These are defined before it can be determined as to which exact
// fields they target and so only specify the names of the target properties
// as they are know to the consumer.
type aggregateRequest struct {
	// This field.
	//
	// The Index and Name of *this* aggregate are known, and are specified here.
	field Field

	// The targets of this aggregate, as defined by the consumer.
	targets           []*aggregateRequestTarget
	dependencyIndexes []int
}

// aggregateRequestTarget contains the user defined information for an aggregate
// target before the actual underlying target is identified and/or created.
type aggregateRequestTarget struct {
	// The name of the host target as known by the consumer.
	//
	// This name may match zero to many field names requested by the consumer.
	hostExternalName string

	// The name of the child target as known by the consumer. This property is
	// optional and may be default depending on aggregate type and the type of
	// the host property.
	//
	// This name may match zero to many field names requested by the consumer.
	childExternalName string

	// The aggregate filter specified by the consumer for this target. Optional.
	filter immutable.Option[request.Filter]

	// The aggregate limit-offset specified by the consumer for this target. Optional.
	limit *Limit

	// The order in which items should be aggregated. Affects results when used with
	// limit. Optional.
	order immutable.Option[request.OrderBy]
}

// Returns the source of the aggregate as requested by the consumer
func getAggregateSources(field *request.Aggregate) ([]*aggregateRequestTarget, error) {
	targets := make([]*aggregateRequestTarget, len(field.Targets))

	for i, target := range field.Targets {
		targets[i] = &aggregateRequestTarget{
			hostExternalName:  target.HostName,
			childExternalName: target.ChildName.Value(),
			filter:            target.Filter,
			limit:             toLimit(target.Limit, target.Offset),
			order:             target.OrderBy,
		}
	}

	return targets, nil
}

// tryGetMatchingAggregate scans the given collection for aggregates with the given name and targets.
//
// Will return the matching target and true if one is found, otherwise will return false.
func tryGetMatchingAggregate(
	name string,
	targets []*aggregateRequestTarget,
	collection []*aggregateRequest,
) (*aggregateRequest, bool) {
collectionLoop:
	for _, aggregate := range collection {
		if aggregate.field.Name != name {
			continue
		}
		if len(aggregate.targets) != len(targets) {
			continue
		}

		for i, target := range targets {
			potentialMatchingTarget := aggregate.targets[i]

			if target.hostExternalName != potentialMatchingTarget.hostExternalName {
				continue collectionLoop
			}

			if target.childExternalName != potentialMatchingTarget.childExternalName {
				continue collectionLoop
			}

			if !target.filter.HasValue() && potentialMatchingTarget.filter.HasValue() {
				continue collectionLoop
			}

			if !potentialMatchingTarget.filter.HasValue() && target.filter.HasValue() {
				continue collectionLoop
			}

			if !target.filter.HasValue() && !potentialMatchingTarget.filter.HasValue() {
				// target matches, so continue the `target` loop and check the remaining.
				continue
			}

			if !reflect.DeepEqual(target.filter.Value().Conditions, potentialMatchingTarget.filter.Value().Conditions) {
				continue collectionLoop
			}
		}

		return aggregate, true
	}
	return nil, false
}

// tryGetTarget scans the given collection of Requestables for an item that matches the given
// name and filter.
//
// If a match is found the matching field will be returned along with true. If a match is not
// found, nil and false will be returned.
func tryGetTarget(
	name string,
	filter *Filter,
	limit *Limit,
	order *OrderBy,
	collection []Requestable,
) (Requestable, bool) {
	dummyTarget := Targetable{
		Field: Field{
			Name: name,
		},
		Filter:  filter,
		Limit:   limit,
		OrderBy: order,
	}

	for _, field := range collection {
		if field == nil {
			continue
		}
		targetable, isTargetable := field.AsTargetable()
		if isTargetable && targetable.equal(dummyTarget) {
			// Return the original field in order to preserve type specific info
			return field, true
		}
	}
	return nil, false
}

// appendNotNilFilter appends a not nil filter for the given child field
// to the given Select.
func appendNotNilFilter(field *aggregateRequestTarget, childField string) {
	if !field.filter.HasValue() || field.filter.Value().Conditions == nil {
		field.filter = immutable.Some(
			request.Filter{
				Conditions: map[string]any{},
			},
		)
	}

	var childBlock any
	var hasChildBlock bool
	if childField == "" {
		childBlock = field.filter.Value().Conditions
	} else {
		childBlock, hasChildBlock = field.filter.Value().Conditions[childField]
		if !hasChildBlock {
			childBlock = map[string]any{}
			field.filter.Value().Conditions[childField] = childBlock
		}
	}

	typedChildBlock := childBlock.(map[string]any)
	typedChildBlock["_ne"] = nil
}
