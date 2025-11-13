# Searchable Encryption Retry Mechanism Specifications

## Overview
Implement a unified retry mechanism for both Searchable Encryption (SE) and document replication failures in DefraDB.

## Requirements

### Functional Requirements
1. **Unified Retry Coordinator**
   - Create a generic retry coordinator that can handle both SE and document replication retries
   - Support different data types for retry items (type-safe with generics)
   - Implement two-level retry structure (peer level + item level)

2. **SE Retry Migration**
   - Migrate SE from single-level to two-level retry structure
   - Follow the document replication pattern as reference implementation
   - Store SE-specific data (CollectionID, DocID, FieldNames, PublicKey, KeyType)

3. **Key Reuse**
   - Reuse existing key types with different prefixes
   - No new key types should be created
   - Use SE_RETRY_ID and SE_RETRY_ITEM constants

4. **Handler Interface**
   - Minimal handler interface with only essential methods
   - ProcessItem for retry logic
   - UpdateStatus for status changes (moved from config)

## Acceptance Criteria
- [ ] Generic retry coordinator implemented with type parameter T
- [ ] SE retry handler stores SERetryInfo struct with all necessary fields
- [ ] Document retry handler uses empty struct{} for data
- [ ] All existing tests pass
- [ ] No regression in functionality
- [ ] Code builds successfully without errors
- [ ] Retry logic is properly extracted and deduplicated

## Technical Constraints
- Must follow existing document replication patterns
- Cannot break existing retry functionality
- Must maintain backward compatibility (no migration needed as requested)
- Logger should be passed as parameter, not in config
- Config should be a simple data structure

## Out of Scope
- Migration of existing retry data
- Changes to retry intervals or timing
- Changes to the core retry algorithm
- Performance optimizations beyond deduplication