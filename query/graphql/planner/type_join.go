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
	"fmt"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/core"
	"github.com/sourcenetwork/defradb/db/base"
	"github.com/sourcenetwork/defradb/query/graphql/parser"
	"github.com/sourcenetwork/defradb/query/graphql/schema"
)

/*

type User {
	name: String
	age: Int
	friends: [Friend]
}

type Friend {
	name: String
	friendsDate: DateTime
	user_id: DocKey
}

- >

/graphql
/explain


{
	query {
		user { selectTopNode -> (source) selectNode -> (source) scanNode(user) -> filter: NIL
			[_key]
			name

			// key = bae-KHDFLGHJFLDG
			friends selectNode -> (source) scanNode(friend) -> filter: {user_id: {_eq: "bae-KHDFLGHJFLDG"}} {
				name
				date: friendsDate
			}
		}
	}
}

selectTopNode - > selectNode -> MultiNode.children: []planNode  -> multiScanNode(scanNode(user)**)											-> } -> scanNode(user).Next() -> FETCHER_STUFF + FILTER_STUFF + OTHER_STUFF
										  						-> TypeJoinNode(merge**) -> TypeJoinOneMany -> (one) multiScanNode(scanNode(user)**)	-> } -> scanNode(user).Value() -> doc
																			 					   -> (many) selectNode - > scanNode(friend)

1. NEXT/VALUES MultiNode.doc = {_key: bae-KHDFLGHJFLDG, name: "BOB"}
2. NEXT/VALUES TypeJoinOneMany.one {_key: bae-KHDFLGHJFLDG, name: "BOB"}
3. NEXT/VALUES (many).selectNode.doc = {name: "Eric", date: Oct29}
LOOP
4. NEXT/VALUES TypeJoinNode {_key: bae-KHDFLGHJFLDG, name: "BOB"} + {friends: [{{name: "Eric", date: Oct29}}]}
5. NEXT/VALUES (many).selectNode.doc = {name: "Jimmy", date: Oct21}
6. NEXT/VALUES TypeJoinNode {_key: bae-KHDFLGHJFLDG, name: "BOB"} + {friends: [{name: "Eric", date: Oct29}, {name: "Jimmy", date: Oct21}]}
GOTO LOOP

// SPLIT FILTER
query {
		user {
			age
			name
			points

			friends {
				name
				points
		}
	}
}

{
	data: [
		{
			_key: bae-ALICE
			age: 22,
			name: "Alice",
			points: 45,

			friends: [
				{
					name: "Bob",
					points:  11
					user_id: "bae-ALICE"
				},
			]
		},

		{
			_key: bae-CHARLIE
			age: 22,
			name: "Charlie",
			points: 45,

			friends: [
				// {
				// 	name: "Mickey",
				// 	points:  6
				// }
			]
		},
	]
}

ALL EMPTY
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = NIL -> ... -> scanNode.filter = NIL

ROOT EMPTY / SUB FULL
{friends: {points: {_gt: 10}}}
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = {friends: {points: {_gt: 10}}} -> ... -> scanNode.filter = NIL

ROOT FULL / SUB EMPTY
{age: {_gte: 21}}
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = NIL -> ... -> scanNode(user).filter = {age: {_gte: 21}}

ROOT FULL / SUB FULL
{age: {_gte: 21}, friends: {points: {_gt: 10}}}
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = {friends: {points: {_gt: 10}}} -> ... -> scanNode(user).filter = {age: {_gte: 21}}
																																-> scanNode(friends).filter = NIL

ROOT FULL / SUB EMPTY / SUB SUB FULL
{age: {_gte: 21}}
friends: {points: {_gt: 10}}
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = NIL -> ... -> scanNode(user).filter = {age: {_gte: 21}}
																									 -> scanNode(friends).filter = {points: {_gt: 10}}

ROOT FULL / SUB FULL / SUB SUB FULL
{age: {_gte: 21}}
friends: {points: {_gt: 10}}
PLAN -> selectTopNode.plan -> limit (optional) -> order (optional) -> selectNode.filter = {friends: {points: {_gt: 10}}} -> ... -> scanNode(user).filter = {age: {_gte: 21}}
																									 							-> scanNode(friends).filter = {points: {_gt: 10}}


ONE-TO-ONE EXAMPLE WITH FILTER TRACKING
type user {
	age: Int
	points: Float
	name: String

	address: Address @primary
	address_id: bae-address-VALUE
}

type Address: {
	street_name: String
	house_number: Int
	city: String
	country: String
	...

	user: user
	# user_id: DocKey
}

query {
	user {
		age
		points
		name

		address {
			street_name
			city
			country
		}
	}
}

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
	p *Planner

	// root        planNode
	// subType     planNode
	// subTypeName string

	// actual join plan, could be one of several strategies
	// based on the relationship of the sub types
	joinPlan planNode

	// doc map[string]interface{}

	// spans core.Spans
}

func (p *Planner) makeTypeIndexJoin(parent *selectNode, source planNode, subType *parser.Select) (*typeIndexJoin, error) {
	typeJoin := &typeIndexJoin{
		p: p,
	}

	// handle join relation strategies
	var joinPlan planNode
	var err error

	desc := parent.sourceInfo.collectionDescription
	typeFieldDesc, ok := desc.GetField(subType.Name)
	if !ok {
		// return nil, fmt.Errorf("Unknown field on sub selection")
		return nil, fmt.Errorf("Unknown field %s on sub selection", subType.Name)
	}

	meta := typeFieldDesc.RelationType
	if schema.IsOne(meta) { // One-to-One, or One side of One-to-Many
		joinPlan, err = p.makeTypeJoinOne(parent, source, subType)
	} else if schema.IsOneToMany(meta) { // Many side of One-to-Many
		joinPlan, err = p.makeTypeJoinMany(parent, source, subType)
	} else { // more to come, Many-to-Many, Embedded?
		return nil, fmt.Errorf("Failed sub selection, unknown relation type")
	}
	if err != nil {
		return nil, err
	}

	typeJoin.joinPlan = joinPlan
	return typeJoin, nil
}

func (n *typeIndexJoin) Init() error {
	return n.joinPlan.Init()
}

func (n *typeIndexJoin) Start() error {
	return n.joinPlan.Start()
}

func (n *typeIndexJoin) Spans(spans core.Spans) { /* todo */ }

func (n *typeIndexJoin) Next() (bool, error) {
	return n.joinPlan.Next()
}

func (n *typeIndexJoin) Values() map[string]interface{} {
	return n.joinPlan.Values()
}

func (n *typeIndexJoin) Close() error {
	return n.joinPlan.Close()
}

func (n *typeIndexJoin) Source() planNode { return n.joinPlan }

// Merge implements mergeNode
func (n *typeIndexJoin) Merge() bool { return true }

// split the provided filter
// into the root and subType components.
// Eg. (filter: {age: 10, name: "bob", author: {birthday: "June 26, 1990", ...}, ...})
//
// The root filter is the conditions that apply to the main
// type ie: {age: 10, name: "bob", ...}.
//
// The subType filter is the conditions that apply to the
// queried sub type ie: {birthday: "June 26, 1990", ...}.
func splitFilterByType(filter *parser.Filter, subType string) (*parser.Filter, *parser.Filter) {
	if filter == nil {
		return nil, nil
	}
	sub, ok := filter.Conditions[subType]
	if !ok {
		return filter, &parser.Filter{}
	}

	// delete old filter value
	delete(filter.Conditions, subType)
	// create new splitup filter
	// our schema ensures that if sub exists, its of type map[string]interface{}
	splitF := &parser.Filter{Conditions: map[string]interface{}{subType: sub}}
	return filter, splitF
}

// typeJoinOne is the plan node for a type index join
// where the root type is the primary in a one-to-one relation
// query.
type typeJoinOne struct {
	p *Planner

	root    planNode
	subType planNode

	subTypeName      string
	subTypeFieldName string

	primary bool

	spans core.Spans
}

func (p *Planner) makeTypeJoinOne(parent *selectNode, source planNode, subType *parser.Select) (*typeJoinOne, error) {
	//ignore recurse for now.
	typeJoin := &typeJoinOne{
		p:    p,
		root: source,
	}

	desc := parent.sourceInfo.collectionDescription
	// get the correct sub field schema type (collection)
	subTypeFieldDesc, ok := desc.GetField(subType.Name)
	if !ok {
		return nil, fmt.Errorf("couldn't find subtype field description for typeJoin node")
	}

	subType.CollectionName = subTypeFieldDesc.Schema

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}
	typeJoin.subType = selectPlan

	typeJoin.subTypeName = subTypeFieldDesc.Name
	typeJoin.subTypeFieldName, err = p.db.GetRelationshipIdField(subType.Name, subTypeFieldDesc.Schema, desc.Name)
	if err != nil {
		return nil, err
	}

	// split filter
	if scan, ok := source.(*scanNode); ok {
		scan.filter, parent.filter = splitFilterByType(scan.filter, typeJoin.subTypeName)
	}
	// source.filter, parent.filter = splitFilterByType(source.filter, typeJoin.subTypeName)

	// determine relation direction (primary or secondary?)
	// check if the field we're querying is the primary side of the relation
	if subTypeFieldDesc.RelationType&client.Relation_Type_Primary > 0 {
		typeJoin.primary = true
	} else {
		typeJoin.primary = false
	}

	return typeJoin, nil
}

func (n *typeJoinOne) Init() error {
	if err := n.subType.Init(); err != nil {
		return err
	}
	return n.root.Init()
}

func (n *typeJoinOne) Start() error {
	if err := n.subType.Start(); err != nil {
		return err
	}
	return n.root.Start()
}

func (n *typeJoinOne) Spans(spans core.Spans) { /* todo */ }

func (n *typeJoinOne) Next() (bool, error) {
	return n.root.Next()
}

func (n *typeJoinOne) Values() map[string]interface{} {
	doc := n.root.Values()
	if n.primary {
		return n.valuesPrimary(doc)
	}
	return n.valuesSecondary(doc)
}

func (n *typeJoinOne) valuesSecondary(doc map[string]interface{}) map[string]interface{} {
	docKey := doc["_key"].(string)
	filter := map[string]interface{}{
		n.subTypeFieldName + "_id": docKey,
	}
	// using the doc._key as a filter
	err := appendFilterToScanNode(n.subType, filter)
	if err != nil {
		return nil
	}

	doc[n.subTypeName] = make(map[string]interface{})
	next, err := n.subType.Next()
	if !next || err != nil {
		return doc
	}

	subdoc := n.subType.Values()
	doc[n.subTypeName] = subdoc
	return doc
}

func (n *typeJoinOne) valuesPrimary(doc map[string]interface{}) map[string]interface{} {
	// get the subtype doc key
	subDocKey, ok := doc[n.subTypeName+"_id"]
	if !ok {
		return doc
	}

	subDocKeyStr, ok := subDocKey.(string)
	if !ok {
		return doc
	}

	subDocField := n.subTypeName
	doc[subDocField] = map[string]interface{}{}

	// create the index key for the sub doc
	slct := n.subType.(*selectTopNode).source.(*selectNode)
	desc := slct.sourceInfo.collectionDescription
	subKeyIndexKey := base.MakeIndexKey(desc, &desc.Indexes[0], subDocKeyStr)

	n.spans = core.Spans{} // reset span
	n.spans = append(n.spans, core.NewSpan(subKeyIndexKey, subKeyIndexKey.PrefixEnd()))

	// do a point lookup with the new span (index key)
	n.subType.Spans(n.spans)

	// re-initialize the sub type plan
	if err := n.subType.Init(); err != nil {
		log.ErrorE(n.p.ctx, "Sub-type initalization error at scan node reset", err)
		return doc
	}

	// if we don't find any docs from our point span lookup
	// or if we encounter an error just return the base doc,
	// with an empty map for the subdoc
	next, err := n.subType.Next()

	if err != nil {
		log.ErrorE(n.p.ctx, "Sub-type initalization error at scan node reset", err)
		return doc
	}

	if !next {
		return doc
	}

	subDoc := n.subType.Values()
	doc[subDocField] = subDoc

	return doc
}

func (n *typeJoinOne) Close() error {
	err := n.root.Close()
	if err != nil {
		return err
	}
	return n.subType.Close()
}

func (n *typeJoinOne) Source() planNode { return n.root }

type typeJoinMany struct {
	p *Planner

	// the main type that is a the parent level of the query.
	root     planNode
	rootName string
	// the index to use to gather the subtype IDs
	index *scanNode
	// the subtype plan to get the subtype docs
	subType     planNode
	subTypeName string
}

func (p *Planner) makeTypeJoinMany(parent *selectNode, source planNode, subType *parser.Select) (*typeJoinMany, error) {
	//ignore recurse for now.
	typeJoin := &typeJoinMany{
		p:    p,
		root: source,
	}

	desc := parent.sourceInfo.collectionDescription
	// get the correct sub field schema type (collection)
	subTypeFieldDesc, ok := desc.GetField(subType.Name)
	if !ok {
		return nil, fmt.Errorf("couldn't find subtype field description for typeJoin node")
	}
	subType.CollectionName = subTypeFieldDesc.Schema

	selectPlan, err := p.SubSelect(subType)
	if err != nil {
		return nil, err
	}
	typeJoin.subType = selectPlan
	typeJoin.subTypeName = subTypeFieldDesc.Name
	typeJoin.rootName, err = p.db.GetRelationshipIdField(subType.Name, subTypeFieldDesc.Schema, desc.Name)
	if err != nil {
		return nil, err
	}

	// split filter
	if scan, ok := source.(*scanNode); ok {
		scan.filter, parent.filter = splitFilterByType(scan.filter, typeJoin.subTypeName)
	}
	// source.filter, parent.filter = splitFilterByType(source.filter, typeJoin.subTypeName)
	return typeJoin, nil
}

func (n *typeJoinMany) Init() error {
	if err := n.subType.Init(); err != nil {
		return err
	}
	return n.root.Init()
}

func (n *typeJoinMany) Start() error {
	if err := n.subType.Start(); err != nil {
		return err
	}
	return n.root.Start()
}

func (n *typeJoinMany) Spans(spans core.Spans) { /* todo */ }

func (n *typeJoinMany) Next() (bool, error) {
	return n.root.Next()
}

func (n *typeJoinMany) Values() map[string]interface{} {
	doc := n.root.Values()

	// check if theres an index
	// if there is, scan and aggregate resuts
	// if not, then manually scan the subtype table
	subdocs := make([]map[string]interface{}, 0)
	if n.index != nil {
		// @todo: handle index for one-to-many setup
	} else {
		docKey := doc["_key"].(string)
		filter := map[string]interface{}{
			n.rootName + "_id": docKey, // user_id: "bae-ALICE" |  user_id: "bae-CHARLIE"
		}
		// using the doc._key as a filter
		err := appendFilterToScanNode(n.subType, filter)
		if err != nil {
			return nil
		}

		// reset scan node
		if err := n.subType.Init(); err != nil {
			log.ErrorE(n.p.ctx, "Sub-type initialization error at scan node reset", err)
		}

		for {
			next, err := n.subType.Next()
			if !next || err != nil {
				break
			}

			subdoc := n.subType.Values()
			subdocs = append(subdocs, subdoc)
		}
	}

	doc[n.subTypeName] = subdocs
	return doc
}

func (n *typeJoinMany) Close() error {
	if err := n.root.Close(); err != nil {
		return err
	}

	return n.subType.Close()
}

func (n *typeJoinMany) Source() planNode { return n.root }

func appendFilterToScanNode(plan planNode, filterCondition map[string]interface{}) error {
	switch node := plan.(type) {
	case *scanNode:
		var err error
		filter := node.filter
		if filter == nil {
			filter, err = parser.NewFilter(nil)
			if err != nil {
				return err
			}
		}

		// merge filter conditions
		for k, v := range filterCondition {
			filter.Conditions[k] = v
		}

		node.filter = filter
	case nil:
		return nil
	default:
		return appendFilterToScanNode(node.Source(), filterCondition)
	}
	return nil
}
