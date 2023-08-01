## Verbose Structure of Type Joins

```

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
			[_docID]
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

1. NEXT/VALUES MultiNode.doc = {_docID: bae-KHDFLGHJFLDG, name: "BOB"}
2. NEXT/VALUES TypeJoinOneMany.one {_docID: bae-KHDFLGHJFLDG, name: "BOB"}
3. NEXT/VALUES (many).selectNode.doc = {name: "Eric", date: Oct29}
LOOP
4. NEXT/VALUES TypeJoinNode {_docID: bae-KHDFLGHJFLDG, name: "BOB"} + {friends: [{{name: "Eric", date: Oct29}}]}
5. NEXT/VALUES (many).selectNode.doc = {name: "Jimmy", date: Oct21}
6. NEXT/VALUES TypeJoinNode {_docID: bae-KHDFLGHJFLDG, name: "BOB"} + {friends: [{name: "Eric", date: Oct29}, {name: "Jimmy", date: Oct21}]}
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
			_docID: bae-ALICE
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
			_docID: bae-CHARLIE
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

```
