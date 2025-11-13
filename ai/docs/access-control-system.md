---
description: DefraDB Access Control Policy (ACP) system architecture and implementation guidelines
globs: "**/acp/**/*.go"
alwaysApply: true
---

# Access Control System Guide

DefraDB's Access Control Policy (ACP) system provides a sophisticated permissions framework to secure data access. This guide explains the core concepts, architecture, and implementation details of the ACP system.

## Core Concepts

### Relation-Based Access Control (ReBac)

DefraDB uses a Relation-Based Access Control system (ReBac), based on Google's Zanzibar architecture:

1. **Key Components**:
   - **Subjects**: Entities requesting access (e.g., users, services)
   - **Objects**: Resources being protected (e.g., documents, fields)
   - **Relations**: Connections between subjects and objects (e.g., "owner", "reader")
   - **Permissions**: Actions that can be performed (e.g., "read", "update", "delete")

2. **Authorization Model**:
   - Permissions are granted based on relationships
   - Relations form a graph connecting subjects to objects
   - Access checks traverse this graph to determine permissions
   - Policies define which relations grant which permissions

### DefraDB's Implementation

DefraDB implements ReBac through three main layers:

1. **Document Access Control (DAC)**:
   - Controls access to entire documents
   - Documents are the objects being protected
   - Collections define the resource types

2. **Field Access Control (FAC)** (in development):
   - Provides granular field-level permissions
   - Individual fields are the objects
   - Documents become the resources in this context

3. **Admin Access Control (AAC)** (in development):
   - Manages system-level operations
   - Administrative functions are the objects
   - The database itself is the resource

## DefraDB Policy Interface (DPI)

To ensure consistent policy definition, DefraDB defines the DPI (DefraDB Policy Interface):

1. **DPI Rules**:
   - Every resource must include a mandatory "owner" relation
   - Required permissions must include "read", "update", and "delete"
   - Owner must have all required permissions
   - Owner relation must be the first in expressions
   - Only union operations are allowed after the owner

2. **Policy Format**:
   ```yaml
   name: "Example Policy"
   description: "A Valid DefraDB Policy Interface"
   
   actor:
     name: actor
   
   resources:
     users:
       permissions:
         read:
           expr: owner + reader
         update:
           expr: owner + updater
         delete:
           expr: owner + deleter
       
       relations:
         owner:
           types: [actor]
         reader:
           types: [actor]
         updater:
           types: [actor]
         deleter:
           types: [actor]
   ```

## Implementation Architecture

### ACP Module Types

DefraDB supports multiple ACP modules:

1. **Local ACP**:
   - Self-contained within the local node
   - Simpler but lacks distributed capabilities
   - Suitable for single-node deployments

2. **SourceHub ACP**:
   - Uses SourceHub trust protocol
   - Provides decentralized access control
   - Required for multi-node deployments with ACP
   - Stores policies and relationships on a blockchain

### Key Components

1. **Identity System**:
   - Uses secp256k1 keys for authentication
   - Generates DID (Decentralized Identifier) for users
   - Validates JWT tokens for HTTP authentication
   - Verifies signatures for CLI operations

2. **Policy Management**:
   - Validates policy compliance with DPI
   - Stores and retrieves policies
   - Associates policies with collections

3. **Relationship Management**:
   - Creates relations between subjects and objects
   - Verifies authorization for relationship changes
   - Manages relationship lifecycle

4. **Permission Checking**:
   - Evaluates access requests against policies
   - Traverses relationship graph
   - Enforces permission rules during operations

## Integration with GraphQL Schemas

ACP integrates with schema definitions using directives:

```graphql
type User @policy(
  id: "50d354a91ab1b8fce8a0ae4693de7616fb1d82cfc540f25cfbe11eb0195a5765",
  resource: "users"
) {
  name: String
  age: Int
}
```

This associates the User collection with a specific policy and resource.

## Implementation Guidelines

When working with the ACP system:

### Extending ACP Capabilities

1. Implement the required interfaces in `acp/types/types.go`
2. Follow the DAC implementation pattern for new access control types
3. Ensure backward compatibility with existing policies
4. Add comprehensive tests covering:
   - Policy validation
   - Permission checks
   - Relationship management
   - Error handling

### Performance Considerations

1. Cache permission check results when appropriate
2. Optimize relationship graph traversal for common patterns
3. Consider batch operations for permission checks
4. Balance security and performance requirements

### Security Best Practices

1. Always verify identity before performing ACP operations
2. Never bypass permission checks in new features
3. Apply the principle of least privilege in policy design
4. Maintain audit trails for access control changes

## Testing ACP Implementations

When implementing or modifying ACP components:

1. **Verify Policy Validation**:
   - Test both valid and invalid policy structures
   - Verify DPI compliance checking
   - Test edge cases in policy expressions

2. **Test Permission Scenarios**:
   - Verify correct access for various relationship configurations
   - Test with and without relevant permissions
   - Test permissions across relationships
   - Verify cascading permissions work correctly

3. **Identity Verification**:
   - Test correct authentication with valid credentials
   - Verify rejection of invalid credentials
   - Test token expiration and renewal

4. **Integration Testing**:
   - Verify behavior with the complete DefraDB system
   - Test interaction with other components (queries, mutations, P2P)
   - Confirm permissions are enforced at all access points

## Limitations and Caveats

Current limitations to be aware of:

1. Local ACP does not work with P2P for collections with policies
2. Some features have limited ACP integration (listed in acp/README.md)
3. Field-level and admin-level access control are still in development
4. Performance impact increases with policy complexity