# Document Discovery Learnings

This file will capture new insights and knowledge discovered during the implementation that are not documented elsewhere in the codebase.

## Codebase Insights

### Event Bus Subscription Pattern
- DefraDB uses a subscription-based event system where `Events().Subscribe()` returns a `Subscription` object
- Event handlers run in separate goroutines that listen on `sub.Message()` channels
- Events must be wrapped in `event.NewMessage()` with proper event names when publishing

### P2P DB Interface Limitations
- The P2P DB interface is intentionally limited and doesn't include all DB methods
- Adding new methods requires careful consideration of the interface boundaries
- Query execution from P2P context requires extending the interface

### GraphQL Schema Generation Integration Points
- New query types must be added both to generation logic and to query type registration
- Discovery queries follow the same pattern as encrypted queries but with simpler arguments
- Schema generation happens during type map construction and field registration

### Query Planning and Mapping
- SelectionType enum controls query routing in the planner
- Mapper handles conversion from request types to planner types
- Each query type (normal, encrypted, discovery) needs its own selection type and mapping logic

### Filter and Parameter Handling
- Query parameters (limit, offset) are embedded in the Limit struct within Targetable
- Filter building for network requests requires careful serialization of complex filter structures
- Collection name/ID conversion is critical for proper query execution across network boundaries