---
sidebar_label: Explain Systems Guide
sidebar_position: 20
---
# A Guide to Explain Systems in DefraDB

## Overview

The DefraDB Explain System is a powerful tool designed to introspect requests, examine plan graphs, and deliver insights into the execution of queries and mutations in DefraDB. These requests can range from basic information queries to highly intricate multi-step operations, all enabled with a single directive added to the request.

### Regular Request

```graphql
query {
    Author {
      _key
      name
      age
    }
}
```

### Explain Request

```graphql
query @explain {
    Author {
      _key
      name
      age
    }
}
```

As application demand grows and schemas expand, requests often become more complex. This could involve adding a type-join or sorting large data sets, which can significantly increase the workload for the database.  This often requires tweaking the query or the schema to ensure requests run as faster. However, without the capability to introspect and understand the request's execution flow, the database will be a black box to developers, limiting their capacity to optimize. This is why DefraDB allows developers to ask for an explanation of the request execution, plan graph, and runtime metrics.

DefraDB provides the option to explain or analyze requests to gain insight into query resolution. Instead of directly requesting data, these queries ask the database to outline the steps it would take to resolve the request and execute all necessary operations before generating the result. This provides transparency into potential bottlenecks, such as inefficient scans or redundant sorting operations. Explain requests enable developers to better understand the database's inner workings and clarify the operations required for request resolution.

Explain requests interact directly with the request planner, executor, and the resulting Plan Graph.

## Planner and Plan Graph

The request planner plays a crucial role in DefraDB as it is responsible for executing and presenting request results. When a database receives a request, it converts it into a series of operations planned and implemented by the request planner. These operations are represented as a Plan Graph, which is a directed graph of operations the database must perform to deliver the requested information.

The Plan Graph is beneficial because it offers a structured request representation, allowing concurrent traversal, branch exploration, and independent subgraph optimization. Each Plan Graph node represents a specific work unit and consists of smaller graphs. For instance, the Plan Graph may contain scan nodes, index nodes, sorting nodes, and filter nodes, among others. The Plan Graph's order and structure are hierarchical, with each node relying on the previous node's output. For example, the final output may depend on the state rendering, which in turn relies on the state limiting, state sorting, and state scanning.



The Plan Graph is a vital component of request processing as it enables the database to simplify complex operations into smaller, more manageable units. In this way, the Plan Graph contributes to the database's performance and scalability enhancement.

The Explain System and Plan Graph collectively provide structured, accessible insights and transparency into the steps a database takes to execute a request.

## Benefits

At its core, the Explain System is a tool that assists developers in optimizing database queries and enhancing performance. Here is an example that emphasizes its advantages.

Quick scans - Most queries begin with a scan node, which is a brute-force method of searching the entire key-value collection. This can be slow for large data sets. However, by using a secondary index, a space-time tradeoff can be made to improve query performance and avoid full scans.

Use of Secondary Indexes- Determining the performance benefits of adding a secondary index can be challenging. Fortunately, the Explain System offers valuable insights into DefraDB's internal processing and Plan Graph, helping to identify the impact. Most importantly, developers can run a simple Explain request and obtain these insights without actually executing the request or building the index, as it only operates on the plan graph.

Improved transparency- Submitting an explain request informs developers whether a full table scan or an index scan will be conducted, and which other elements will be involved in the process. This information enables developers to understand the steps required to execute their queries and create more efficient ones.

Query Optimization- For example, it is more efficient to query from primary to secondary than from secondary to primary. The Explain System can also accurately demonstrate the inefficiency of certain queries, such as a simple point lookup compared to an efficient join index. Overall, the Explain System helps developers gain insight into the inner workings of the database and queries, allowing for greater introspection and understanding.

## How it works

When you send a request to the database, it can either execute the request or explain it. By default, the database will execute the request as expected. This will compile the request, construct a plan, and evaluate the nodes of the plan to render the results. 

Conversely, an Explain will compile the request, construct a plan, and finally walk the plan graph, collecting node attributes and execution metrics. The goal is to gather details about each part of the plan and show this information to the developer in a clear and organized way.

Having the plan arranged as parts in a graph is helpful because it's both fast to process and simple to understand. When a request is changed into an Explain request, it creates an organized view of the plan graph that developers can make sense of. Some smaller details might be left out, but the main points and important features give a clear link between the internal and external views of the graph. By gathering the structure and features of the plan graph, developers can learn the steps needed to run their requests and make them work better and faster.

## Types of Explain Requests

### Simple Explain

Simple Explain Requests is the default mode for explanation, only requiring the additional `@explain` directive. You can also be explicit and provide a type argument to the directive like this `@explain(type: simple)`. 

This mode of explanation returns only the syntactic and structural information of the Plan Graph, its nodes, and their attributes.

The following example shows a Simple Explain request applies to an `Author` query request.

```graphql
query @explain {
    Author {
        name
        age
    }
}
```

```json
// Response
{
    "explain": {
        "select TopNode": {
            "selectNode": {
                "filter": null,
                "scanNode": {
                    "filter":null,
                    "collectionID": "3",
                    "collectionName": "Author",
                    "spans": [{
                        "start": "/3",
                        "end": "/4"
                    }]
                }
            }
        }
    }
}
```

With the corresponding Plan Graph:

Simple Explain requests are extremely fast, since it does not actually execute the constructed Plan Graph. It is intended to give transparency back to the developer, and to understand the structure and operations of how the database would resolve their request.

### Execute Explain

Execute explanation differs from Simple mode because it actually executes the constructed plan graph from the request. However, it doesn't return the results, but instead collects various metrics and runtime information about how the request was executed, and returns it using using the same rendered plan graph structure that the Simple Explain does. This is similar to EXPLAIN ANALYZE from PostgreSQL or MySQL

You can create an Execute Explain by specifying the explain type using the directive type​arguments @explain(type: execute).

The following example shows a Execute Explain request applies to an author query request.

```graphql
query @explain(type: execute) {
	Author {
		name
		age
	}
}
```

```json
// Response
[
	{
		"explain": {
			"executionSuccess": true,
			"sizeOfResult":     1,
			"planExecutions":   2,
			"selectTopNode": {
				"selectNode": {
					"iterations":    2,
					"filterMatches": 1,
					"scanNode": {
						"iterations":    2,
						"docFetches":    2,
						"filterMatches": 1
					}
				}
			}
		}
	}
]
```

Because Execute Explain actually executes the plan, it will of course take more time to complete and return results than the Simple Explain. It will actually take slightly longer to execute than the non-explain counterpart, as it has the overhead of measuring and collecting information.

## Limitations

One disadvantage of the Explain System is that it violates the formal specification of the GraphQL API. This means that certain guarantees, such as the symmetry between the structure of the request and result, is not maintained. 

For example, if a request is sent to a user collection, the GraphQL Schema specifies that it will return an array of users. However, if the explain directive is added, the structure of the result will not match the schema specified and will instead be the plan graph representation. While this violation is considered acceptable in order to improve the developer experience, it is important to be aware of this limitation.

## Next Steps

A future feature called Prediction Explain aims to provide a balance between speed and information. These requests do not execute the plan graph, but instead make educated guesses about the potential impact of the query based on attributes and metrics. Prediction Explain Requests take longer than the Simple Explain System, but not as long as Execution Explain Requests.

The Explain System is being developed with additional tooling in mind. Currently, it returns a structured JSON object that represents the plan graph. In the future, the aim is for the tool to provide different representations of the Plan Graph, including a text output that is more easily readable by humans and a visual graph that displays the top-down structure of the graph. In addition to the Simple and Execution Explain Requests that the Explain System currently supports or will support in the future, the team is also working on serializing and representing the returned object in various ways. This will provide developers with more options for understanding and analyzing the database and queries.
