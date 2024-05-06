---
sidebar_label: Execution Flow
sidebar_position: 140
---
# Execution Flow

Understanding the execution flow of a query can help understand its structure, and help you with your queries. Query execution is broken down into the following three phases: 
- Parsing 
- Planning
- Executing

## Parsing Phase

The parsing phase parses the query as a string and returns a structured Abstract Syntax Tree (AST) representation. It also does a semantic validation of the structure against the schema.

## Planning Phase

The planning phase analyzes the query, the storage structure, and any additional indexes to determine query execution. This phase is highly dependant on the deployment environment and underlying storage engine as it uses available features and structure to provide optimal performance. Specific schemas automatically create certain secondary indexes. The planning phase automatically uses available custom secondary indexes created by you.

## Execution Phase

The execution phase does data scanning, filtering, and formatting. This phase has a deterministic process towards the steps taken to produce results. This is due to the priority an argument and its parameters have over another.

The priority order of arguments is as follows:

1. filter -> groupBy: Filtered Data
1. groupBy -> aggregate: Subgroups
1. aggregate -> having: Subgroups
1. having -> order: Filtered Data
1. order -> limit: Ordered Data

Each step has a specific purpose as described here.

1. `filter` argument breaks down the target collection (based on provided parameters and fields) into the output result set.
1. `groupBy` argument divides the result set further into subgroups across potentially several dimensions.
1. `aggregate` phase processes a subgroup's given fields.
1.  `having` argument filters the data based on the grouped fields or aggregate results.
1. `order` argument structures the result set based on the ordering (ascending or descending) of one or more field values.
1. `limit` argument and its associated arguments restrict the number of the finalized, filtered, ordered result set.

See the image below for an example of the execution order:

![](https://i.imgur.com/Yf0KJ5A.png)