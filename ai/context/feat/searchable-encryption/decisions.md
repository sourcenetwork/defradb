# Implementation Decisions

## 1. Generic Type Parameter for Coordinator
**Decision**: Made the Coordinator generic with type parameter T
**Initial Approach**: Tried to keep non-generic with interface{} 
**Why Changed**: Type safety and avoiding runtime type assertions. Each handler can define its own data structure.
**Result**: `Coordinator[T any]` with SE using `SERetryInfo` and documents using `struct{}`

## 2. Removed GetCleanupFunc from Handler
**Decision**: Removed GetCleanupFunc method from Handler interface
**Initial Approach**: Handler could provide custom cleanup function
**Why Changed**: All implementations used the same cleanup pattern. Better to centralize in coordinator.
**Result**: Coordinator has built-in `deleteRetryAndItems` method used by all handlers

## 3. UpdateStatus Moved to Handler Interface
**Decision**: Moved UpdateStatus from Config to Handler interface
**Initial Approach**: UpdateStatusFunc was a field in Config struct
**Why Changed**: Config should only contain data, not behavior. Custom logic belongs in handlers.
**Result**: Each handler implements its own UpdateStatus logic

## 4. Logger as Parameter
**Decision**: Pass logger as parameter to NewCoordinator
**Initial Approach**: Logger was part of Config struct
**Why Changed**: Config should be simple data. Logger is infrastructure concern.
**Result**: `NewCoordinator[T](..., logger *corelog.Logger)`

## 5. Extract Common SE Logic
**Decision**: Created publishSEReplication helper method
**Initial Approach**: Duplicated logic in handleUpdateEvent and ProcessItem
**Why Changed**: DRY principle - same logic for generating and publishing SE artifacts
**Result**: Both methods now use common publishSEReplication method

## 6. No Migration Implementation
**Decision**: Removed migration code as requested
**Initial Approach**: Started implementing migration from old to new structure
**Why Changed**: User specified no migration needed
**Result**: Clean implementation without migration complexity