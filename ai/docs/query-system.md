---
description: DefraDB Query Language (DQL) and query execution system
globs: "**/request/**/*.go, **/planner/**/*.go"
alwaysApply: true
---

# DefraDB Query System Guide

DefraDB uses a query language (DQL) compatible with GraphQL but with powerful extensions for document databases. Understanding this system is crucial for implementing features that interact with queries.

## Query Language Overview

DQL supports standard GraphQL syntax plus extensions for:

1. **Document Operations**:
   - Create, update, and delete documents
   - Get documents by ID or filter conditions

2. **Filtering**:
   - Comparison operators: `_eq`, `_ne`, `_gt`, `_lt`, `_ge`, `_le`
   - Membership tests: `_in`, `_nin`
   - String pattern matching: `_like`, `_nlike`, `_ilike`, `_nilike`
   - Logical operators: `_and`, `_or`, `_not`

3. **Relationships**:
   - One-to-one, one-to-many, many-to-many
   - Nested filtering across relationships
   - Bidirectional navigation

4. **Aggregations**:
   - `count`, `sum`, `avg`, `min`, `max`
   - Group by specific fields

5. **Sorting and Pagination**:
   - `order` for sorting by multiple fields (ascending/descending)
   - `limit` and `offset` for pagination

6. **Time-Travel Queries**:
   - Query historical document states
   - Access document commit history

## Query Execution Flow

The query system follows this execution flow:

1. **Parsing**: Convert GraphQL/JSON to internal representation
2. **Planning**: Generate an optimized execution plan
3. **Execution**: Process the plan against data stores
4. **Result Formation**: Structure results according to query

### Key Components

#### Request Package (`internal/request`)

Handles incoming requests and converts them to internal representations:
- `request.Request`: Core interface for all request types
- `request.Select`: Fields to retrieve from documents
- `request.Filter`: Conditions for document selection
- `request.Order`: Sorting specifications
- `request.Limit`/`request.Offset`: Pagination controls

#### Planner Package (`internal/planner`)

Creates and optimizes execution plans:
- `planner.Planner`: Converts requests to execution plans
- `planner.Plan`: Represents a complete execution strategy
- `planner.Operation`: Individual execution steps (scan, filter, join, etc.)
- `planner.Pipe`: Links operations in execution sequence

#### Operation Types

The query planner generates operations such as:
- `Scan`: Retrieve documents from collections
- `Filter`: Apply predicates to narrow results
- `TypeJoin`: Connect documents across collections
- `Order`: Sort result sets
- `Limit`: Cap result count
- `Group`: Organize results by field values
- `Aggregate`: Calculate aggregate values
- `Update`/`Create`/`Delete`: Modify document data

## Query Optimization Patterns

When implementing features that interact with the query system:

1. **Use Indexes Effectively**:
   - Check if required fields are indexed
   - Leverage index-accelerated operations
   - Consider index implications for new filter types

2. **Minimize Document Scans**:
   - Use direct document access when IDs are known
   - Leverage filtered iterators instead of post-filtering

3. **Optimize Join Operations**:
   - Push down filters to smallest possible result sets first
   - Use indexed relationships when available
   - Consider cardinality in join planning

4. **Handle Large Result Sets**:
   - Implement chunked processing for large collections
   - Use iterators for memory-efficient processing
   - Apply limits early in the execution pipeline

5. **Leverage Type Information**:
   - Use schema knowledge to optimize operations
   - Apply type-specific optimizations for comparisons
   - Consider field cardinality for operation ordering

## Key Interfaces for Extensions

When extending the query system, focus on these interfaces:

1. **Adding New Operators**:
   - Extend filter package with new predicates
   - Implement matching logic in fetcher components
   - Update parser to recognize new syntax

2. **New Aggregation Functions**:
   - Add new operation types in planner
   - Implement state tracking for incremental computation
   - Update result formation for new aggregate types

3. **Query Transformation**:
   - Use `filter.Normalization` for standardizing expressions
   - Apply transformations at plan optimization phase
   - Consider both local and distributed execution impacts

## Testing Query Features

When implementing query-related features:

1. Use both unit tests and integration tests
2. Test with various collection sizes for performance characteristics
3. Verify behavior with indexes and without indexes
4. Include explain-plan tests to verify optimization
5. Test edge cases like empty collections and null values
6. For complex queries, add explicit performance benchmarks