# Introduction

In the realm of information technology (IT) and cybersecurity, **access control** plays a pivotal role in ensuring the confidentiality, integrity, and availability of sensitive resources. Let's delve into why access control policies are crucial for protecting your valuable data.

## What Is Access Control?

**Access control** is a mechanism that regulates who or what can view, use, or access a specific resource within a computing environment. Its primary goal is to minimize security risks by ensuring that only **authorized users**, systems, or services have access to the resources they need. But it's more than just granting or denying access, it involves several key components:

1. **Authentication**: Verifying the identity of an individual or system.
2. **Authorization**: Determining what actions or operations an actor is allowed to perform.
3. **Access**: Granting or denying access based on authorization.
4. **Management**: Administering access rights and permissions.
5. **Audit**: Tracking and monitoring access patterns for accountability.

## Why Is Access Control Important?

1. **Mitigating Security Risks**: Cybercriminals are becoming increasingly sophisticated, employing advanced techniques to breach security systems. By controlling who has access to your database, you significantly reduce the risk of unauthorized access, both from external attackers and insider threats.

2. **Compliance with Regulations**: Various regulatory requirements, such as the **General Data Protection Regulation (GDPR)** and the **Health Insurance Portability and Accountability Act (HIPAA)**, mandate stringent access control measures to protect personal data. Implementing access control ensures compliance with these regulations.

3. **Preventing Data Breaches**: Access control acts as a proactive measure to deter, detect, and prevent unauthorized access. It ensures that only those with the necessary permissions can access sensitive data or services.

4. **Managing Complexity**: Modern IT infrastructure, including cloud computing and mobile devices, has exponentially increased the number of access points. Technologies like **identity and access management (IAM)** and approaches like **zero trust** help manage this complexity effectively.

## Types of Security Access Controls

Several access control models exist, including:

- **Role-Based Access Control (RBAC)**: Assigns permissions to roles, roles then are granted to users. A user's active role then defines their access. (e.g., admin, user, manager).
- **Attribute-Based Access Control (ABAC)**: Considers various attributes (e.g., user attributes, resource attributes) for access decisions.
- **Discretionary Access Control (DAC)**: Users with sufficient permissions (resource owners) are to grant / share an object with other users.
- **Mandatory Access Control (MAC)**: Users are not allowed to grant access to other users. Permissions are granted based on a minimum role / hierarchy (security labels and clearances) that must be met.
- **Policy-Based Access Control (PBAC)**: Enforces access based on defined policies.
- **Relation-Based Access Control (ReBac)**: Relations between objects and users in the system are used to derive their permissions.

- Note: **DefraDB** access control rules strongly resembles **Discretionary Access Control (DAC)**, which is implemented through a **Relation-Based Access Control System (ReBac) Engine**

## Challenges of Access Control in Cybersecurity

- **Distributed IT Environments**: Cloud computing and remote work create new challenges.
- **Rise of Mobility**: Mobile devices in the workplace add complexity.
- **Password Fatigue**: Balancing security with usability.
- **Data Governance**: Ensuring visibility and control.
- **Multi-Tenancy**: Managing complex permissions in SaaS applications.

## Key takeaway
A robust access control policy system is your first line of defense against unauthorized access and data breaches.


# DefraDB's Access Control System

## ReBac Authorization Model

### Zanzibar
In 2019, Google published their [Zanzibar](https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/) paper, a paper explaining how they handle authorization across their many services. It uses access control lists but with relationship-based access control rather than role-based access control. Relationship-Based Access Control (ReBAC) establishes an authorization model where a subject's permission to access an object is defined by the presence of relationships between those subjects and objects.
The way Zanzibar works is it exposes an API with (mainly) operations to manage `Relationships` (`tuples`) and Verify Access Requests (can Bob do X) through the `Check` call. A `tuple` includes subject, relation, and object. The Check call performs Graph Search over the `tuples` to find a path between the user and the object, if such a path exist then according to `RelBAC` the user has the queried permission. It operates as a Consistent and Partition-Tolerant System.

### Zanzi
However the Zanzibar API is centralized, so we (Source Network) created a decentralized implementation of Zanzibar called **Zanzi**. Which is powered by our SourceHub trust protocol. Zanzi is a general purpose Zanzibar implementation which operates over a KV persistence layer.

### SourceHub ACP Module
DefraDB wraps the `local` and `remote` SourceHub ACP Modules to bring all that magic to DefraDB.

In order to setup the relation based access control, SourceHub requires an agreed upon contract which models the `relations`, `permissions`, and `actors`. That contract is refered to as a `SourceHub Policy`. The policy model's all the `relations` and `permissions` under a `resource`.
A `resource` corresponds to that "thing" that we want to gate the access control around. This can be a `Type`, `Container`, `Schema`, `Shape` or anything that has Objects that need access control. Once the policy is finalized, it has to be uploaded to the `SourceHub Module` so it can be used.
Once the `Policy` is uploaded to the `SourceHub Module` then an `Actor` can begin registering the `Object` for access control by linking to a `Resource` that exists on the uploaded `Policy`.
After the `Object` is registered successfully, the `Actor` will then get a special built-in relation with that `Object` called the `"owner"` relation. This relation is given to the `Registerer` of an `Object`.
Then an `Actor` can issue `Check` calls to see if they have access to an `Object`.

## Document Access Control (DAC)
In DefraDB's case we wanted to gate access control around the `Documents` that belonged to a specific `Collection`. Here, the `Collection` (i.e. the type/shape of the `Object`) can be thought of as the `Resource`, and the `Documents` are the `Objects`.


## Field Access Control (FAC) (coming soon)
We also want the ability to do a more granular access control than just DAC. Therefore we have `Field` level access control for situations where some fields of a `Document` need to be private, while others do not. In this case the `Document` becomes the `Resource` and the `Fields` are the `Objects` being gated.


## Admin Access Control (AAC) (coming soon)
We also want to model access control around the `Admin Level Operations` that exist in `DefraDB`. In this case the entire `Database` would be the `Resource` and the `Admin Level Operations` are the `Objects` being gated.

A non-exhastive list of some operations only admins should have access for:
- Ability to turnoff ACP
- Ability to interact with the P2P system

## SourceHub Policies Are Too Flexible
SourceHub Policies are too flexible (atleast until the ability to define `Meta Policies` is implemented). This is because SourceHub leaves it up to the user to specify any type of `Permissions` and `Relations`. However for DefraDB, there are certain guarantees that **MUST** be maintained in order for the `Policy` to be effective. For example the user can input any name for a `Permission`, or `Relation` that DefraDB has no knowledge of. Another example is when a user might make a `Policy` that does not give any `Permission` to the `owner`. Which means in the case of DAC no one will have any access to the `Document` they created.
Therefore There was a very clear need to define some rules while writing a `Resource` in a `Policy` which will be used with DefraDB's DAC, FAC, or AAC. These rules will guarantee that certain `Required Permissions` will always be there on a `Resource` and that `Owner` has the correct `Permissions`.

We call these rules DPI A.K.A DefraDB Policy Interface.

## Terminologies
- 'SourceHub Address' is a `Bech32` Address with a specific SourceHub prefix.
- 'Identity' is a combination of SourceHub Address and a Key-Pair Signature.
- 'DPI' means 'DefraDB Policy Interface'.
- 'Partially-DPI' policy means a policy with at least one DPI compliant resource.
- 'Permissioned Collection' means to have a policy on the collection, like: `@policy(id:".." resource: "..")`
- 'Permissioned Request' means to have a request with a SourceHub Identity.


## DAC DPI Rules

To qualify as a DPI-compliant `resource`, the following rules **MUST** be satisfied:
- The resource **must include** the mandatory `registerer` (`owner`) relation within the `relations` attribute.
- The resource **must encompass** all the required permissions under the `permissions` attribute.
- Every required permission must have the required registerer relation (`owner`) in `expr`.
- The required registerer relation **must be positioned** as the leading (first) relation in `expr` (see example below).
- Any relation after the required registerer relation must only be a union set operation (`+`).

For a `Policy` to be `DPI` compliant for DAC, all of its `resources` must be DPI compliant.
To be `Partially-DPI` at least one of its `resource` must be DPI compliant.

### More Into The Weeds:

All mandatory permissions are:
- Specified in the `dpi.go` file within the variable `dpiRequiredPermissions`.

The name of the required 'registerer' relation is:
- Specified in the `dpi.go` file within the variable `requiredRegistererRelationName`.

### DPI Resource Examples:
- Check out tests here: [tests/integration/acp/schema/add_dpi](/tests/integration/acp/schema/add_dpi)
- The tests linked are broken into `accept_*_test.go` and `reject_*_test.go` files.
- Accepted tests document the valid DPIs (as the schema is accepted).
- Rejected tests document invalid DPIs (as the schema is rejected).
- There are also some Partially-DPI tests that are both accepted and rejected depending on the resource.

### Required Permission's Expression:
Even though the following expressions are valid generic policy expressions, they will make a
DPI compliant resource lose its DPI status as these expressions are not in accordance to
our DPI [rules](#dac-dpi-rules). Assuming these `expr` are under a required permission label:
- `expr: owner-owner`
- `expr: owner-reader`
- `expr: owner&reader`
- `expr: owner - reader`
- `expr: ownerMalicious + owner`
- `expr: ownerMalicious`
- `expr: owner_new`
- `expr: reader+owner`
- `expr: reader-owner`
- `expr: reader - owner`

Here are some valid expression examples. Assuming these `expr` are under a required permission label:
- `expr: owner`
- `expr: owner + reader`
- `expr: owner +reader`
- `expr: owner+reader`

## DAC Usage CLI:

### Authentication

To perform authenticated operations you will need to generate a `secp256k1` key pair.

The command below will generate a new secp256k1 private key and print the 256 bit X coordinate as a hexadecimal value.

```sh
openssl ecparam -name secp256k1 -genkey | openssl ec -text -noout | head -n5 | tail -n3 | tr -d '\n:\ '
```

Copy the private key hex from the output.

```sh
read EC key
e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Use the private key to generate authentication tokens for each request.

```sh
defradb client ... --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

### Adding a Policy:

We have in `examples/dpi_policy/user_dpi_policy.yml`:
```yaml
description: A Valid DefraDB Policy Interface (DPI)

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader
      write:
        expr: owner

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
```

CLI Command:
```sh
defradb client acp policy add -f examples/dpi_policy/user_dpi_policy.yml --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "PolicyID": "50d354a91ab1b8fce8a0ae4693de7616fb1d82cfc540f25cfbe11eb0195a5765"
}
```

### Add schema, linking to a resource within the policy we added:

We have in `examples/schema/permissioned/users.graphql`:
```graphql
type Users @policy(
    id: "50d354a91ab1b8fce8a0ae4693de7616fb1d82cfc540f25cfbe11eb0195a5765",
    resource: "users"
) {
    name: String
    age: Int
}
```

CLI Command:
```sh
defradb client schema add -f examples/schema/permissioned/users.graphql
```

Result:
```json
[
  {
    "Name": "Users",
    "ID": 1,
    "RootID": 1,
    "SchemaVersionID": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
    "Sources": [],
    "Fields": [
      {
        "Name": "_docID",
        "ID": 0
      },
      {
        "Name": "age",
        "ID": 1
      },
      {
        "Name": "name",
        "ID": 2
      }
    ],
    "Indexes": [],
    "Policy": {
      "ID": "50d354a91ab1b8fce8a0ae4693de7616fb1d82cfc540f25cfbe11eb0195a5765",
      "ResourceName": "users"
    }
  }
]

```

### Create private documents (with identity)

CLI Command:
```sh
defradb client collection create --name Users '[{ "name": "SecretShahzad" }, { "name": "SecretLone" }]' --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

### Create public documents (without identity)

CLI Command:
```sh
defradb client collection create  --name Users '[{ "name": "PublicShahzad" }, { "name": "PublicLone" }]'
```

### Get all docIDs without an identity (shows only public):
CLI Command:
```sh
defradb client collection docIDs --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "docID": "bae-63ba68c9-78cb-5060-ab03-53ead1ec5b83",
  "error": ""
}
{
  "docID": "bae-ba315e98-fb37-5225-8a3b-34a1c75cba9e",
  "error": ""
}
```


### Get all docIDs with an identity (shows public and owned documents):
```sh
defradb client collection docIDs --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "docID": "bae-63ba68c9-78cb-5060-ab03-53ead1ec5b83",
  "error": ""
}
{
  "docID": "bae-a5830219-b8e7-5791-9836-2e494816fc0a",
  "error": ""
}
{
  "docID": "bae-ba315e98-fb37-5225-8a3b-34a1c75cba9e",
  "error": ""
}
{
  "docID": "bae-eafad571-e40c-55a7-bc41-3cf7d61ee891",
  "error": ""
}
```


### Access the private document (including field names):
CLI Command:
```sh
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "_docID": "bae-a5830219-b8e7-5791-9836-2e494816fc0a",
  "name": "SecretShahzad"
}
```

### Accessing the private document without an identity:
CLI Command:
```sh
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a"
```

Error:
```
    Error: document not found or not authorized to access
```

### Accessing the private document with wrong identity:
CLI Command:
```sh
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5
```

Error:
```
    Error: document not found or not authorized to access
```

### Update private document:
CLI Command:
```sh
defradb client collection update --name Users --docID "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --updater '{ "name": "SecretUpdatedShahzad" }' --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "Count": 1,
  "DocIDs": [
    "bae-a5830219-b8e7-5791-9836-2e494816fc0a"
  ]
}
```

#### Check if it actually got updated:
CLI Command:
```sh
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "_docID": "bae-a5830219-b8e7-5791-9836-2e494816fc0a",
  "name": "SecretUpdatedShahzad"
}
```

### Update With Filter example (coming soon)

### Delete private document:
CLI Command:
```sh
defradb client collection delete --name Users --docID "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "Count": 1,
  "DocIDs": [
    "bae-a5830219-b8e7-5791-9836-2e494816fc0a"
  ]
}
```

#### Check if it actually got deleted:
CLI Command:
```sh
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Error:
```
    Error: document not found or not authorized to access
```

### Delete With Filter example (coming soon)

### Typejoin example (coming soon)

### View example (coming soon)

### P2P example (coming soon)

### Backup / Import example (coming soon)

### Secondary Indexes example (coming soon)

### Execute Explain example (coming soon)

### Sharing Private Documents With Others

To share a document (or grant a more restricted access) with another actor, we must add a relationship between the
actor and the document. Inorder to make the relationship we require all of the following:

1) **Target DocID**: The `docID` of the document we want to make a relationship for.
2) **Collection Name**: The name of the collection that has the `Target DocID`.
3) **Relation Name**: The type of relation (name must be defined within the linked policy on collection).
4) **Target Identity**: The identity of the actor the relationship is being made with.
5) **Requesting Identity**: The identity of the actor that is making the request.

Note:
  - ACP must be available (i.e. ACP can not be disabled).
  - The collection with the target document must have a valid policy and resource linked.
  - The target document must be registered with ACP already (private document).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the specified relation was not granted the miminum DPI permissions (read or write) within the policy,
  and a relationship is formed, the subject/actor will still not be able to access (read or write) the resource.
  - If the relationship already exists, then it will just be a no-op.

Consider the following policy that we have under `examples/dpi_policy/user_dpi_policy_with_manages.yml`:

```yaml
name: An Example Policy

description: A Policy

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader + writer

      write:
        expr: owner + writer

      nothing:
        expr: dummy

    relations:
      owner:
        types:
          - actor

      reader:
        types:
          - actor

      writer:
        types:
          - actor

      admin:
        manages:
          - reader
        types:
          - actor

      dummy:
        types:
          - actor
```

Add the policy:
```sh
defradb client acp policy add -f examples/dpi_policy/user_dpi_policy_with_manages.yml \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "PolicyID": "ec11b7e29a4e195f95787e2ec9b65af134718d16a2c9cd655b5e04562d1cabf9"
}
```

Add schema, linking to the users resource and our policyID:
```sh
defradb client schema add '
type Users @policy(
    id: "ec11b7e29a4e195f95787e2ec9b65af134718d16a2c9cd655b5e04562d1cabf9",
    resource: "users"
) {
    name: String
    age: Int
}
'
```

Result:
```json
[
  {
    "Name": "Users",
    "ID": 1,
    "RootID": 1,
    "SchemaVersionID": "bafkreihhd6bqrjhl5zidwztgxzeseveplv3cj3fwtn3unjkdx7j2vr2vrq",
    "Sources": [],
    "Fields": [
      {
        "Name": "_docID",
        "ID": 0,
        "Kind": null,
        "RelationName": null,
        "DefaultValue": null
      },
      {
        "Name": "age",
        "ID": 1,
        "Kind": null,
        "RelationName": null,
        "DefaultValue": null
      },
      {
        "Name": "name",
        "ID": 2,
        "Kind": null,
        "RelationName": null,
        "DefaultValue": null
      }
    ],
    "Indexes": [],
    "Policy": {
      "ID": "ec11b7e29a4e195f95787e2ec9b65af134718d16a2c9cd655b5e04562d1cabf9",
      "ResourceName": "users"
    },
    "IsMaterialized": true
  }
]
```

Create a private document:
```sh
defradb client collection create --name Users '[{ "name": "SecretShahzadLone" }]' \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Only the owner can see it:
```sh
defradb client collection docIDs --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "docID": "bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c",
  "error": ""
}
```

Another actor can not:
```sh
defradb client collection docIDs --identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5
```

**Result is empty from the above command**


Now let's make the other actor a reader of the document by adding a relationship:
```sh
defradb client acp relationship add \
--collection Users \
--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
--relation reader \
--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "ExistedAlready": false
}
```

**Note: If the same relationship is created again the `ExistedAlready` would then be true, indicating no-op**

Now the other actor can read:
```sh
defradb client collection docIDs --identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5
```

Result:
```json
{
  "docID": "bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c",
  "error": ""
}
```

But, they still can not perform an update as they were only granted a read permission (through `reader` relation):
```sh
defradb client collection update --name Users --docID "bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c" \
--identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5 '{ "name": "SecretUpdatedShahzad" }'
```

Result:
```sh
Error: document not found or not authorized to access
```

Sometimes we might want to give a specific access (i.e. form a relationship) not just with one identity, but with
any identity (includes even requests with no-identity).
In that case we can specify "*" instead of specifying an explicit `actor`:
```sh
defradb client acp relationship add \
--collection Users \
--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
--relation reader \
--actor "*" \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "ExistedAlready": false
}
```

**Note: specifying `*` does not overwrite any previous formed relationships, they will remain as is **

### Revoking Access To Private Documents

To revoke access to a document for an actor, we must delete the relationship between the
actor and the document. Inorder to delete the relationship we require all of the following:

1) Target DocID: The docID of the document we want to delete a relationship for.
2) Collection Name: The name of the collection that has the Target DocID.
3) Relation Name: The type of relation (name must be defined within the linked policy on collection).
4) Target Identity: The identity of the actor the relationship is being deleted for.
5) Requesting Identity: The identity of the actor that is making the request.

Notes:
  - ACP must be available (i.e. ACP can not be disabled).
  - The target document must be registered with ACP already (policy & resource specified).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the relationship record was not found, then it will be a no-op.

Consider the same policy and added relationship from the previous example in the section above where we learnt
how to share the document with other actors. 

We made the document accessible to an actor by adding a relationship:
```sh
defradb client acp relationship add \
--collection Users \
--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
--relation reader \
--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "ExistedAlready": false
}
```

Similarly, inorder to revoke access to a document we have the following command to delete the relationship:
```sh
defradb client acp relationship delete \
--collection Users \
--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
--relation reader \
--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "RecordFound": true
}
```

**Note: If the same relationship is deleted again (or a record for a relationship does not exist) then the `RecordFound`
would be false, indicating no-op**

Now the other actor can no longer read:
```sh
defradb client collection docIDs --identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5
```

**Result is empty from the above command**

We can also revoke the previously granted implicit relationship which gave all actors access using the "*" actor.
Similarly we can just specify "*" to revoke all access given to actors implicitly through this relationship:
```sh
defradb client acp relationship delete \
--collection Users \
--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
--relation reader \
--actor "*" \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "RecordFound": true
}
```

**Note: Deleting with`*` does not remove any explicitly formed relationships, they will remain as they were **

## DAC Usage HTTP:

### Authentication

To perform authenticated operations you will need to build and sign a JWT token with the following required fields:

- `sub` public key of the identity
- `aud` host name of the defradb api
- The `exp` and `nbf` fields should also be set to short-lived durations.

Additionally, if using SourceHub ACP, the following must be set:
- `iss` should be set to the user's DID, e.g. `"did:key:z6MkkHsQbp3tXECqmUJoCJwyuxSKn1BDF1RHzwDGg9tHbXKw"`
- `iat` should be set to the current unix timestamp
- `authorized_account` should be set to the SourceHub address of the account signing SourceHub transactions on your
  behalf - WARNING - this will currently enable this account to make any SourceHub as your user for the lifetime of the
  token, so please only set this if you fully trust the node/account.

The JWT must be signed with the `secp256k1` private key of the identity you wish to perform actions as.

The signed token must be set on the `Authorization` header of the HTTP request with the `bearer ` prefix prepended to it.

If authentication fails for any reason a `403` forbidden response will be returned.

## _AAC DPI Rules (coming soon)_
## _AAC Usage: (coming soon)_

## _FAC DPI Rules (coming soon)_
## _FAC Usage: (coming soon)_

## Warning / Caveats
- If using Local ACP, P2P will only work with collections that do not have a policy assigned.  If you wish to use ACP
on collections connected to a multi-node network, please use SourceHub ACP.

The following features currently don't work with ACP, they are being actively worked on.
- [Adding Secondary Indexes](https://github.com/sourcenetwork/defradb/issues/2365)
- [Backing/Restoring Private Documents](https://github.com/sourcenetwork/defradb/issues/2430)

The following features may have undefined/unstable behavior until they are properly tested:
- [Views](https://github.com/sourcenetwork/defradb/issues/2018)
- [Average Operations](https://github.com/sourcenetwork/defradb/issues/2475)
- [Count Operations](https://github.com/sourcenetwork/defradb/issues/2474)
- [Group Operations](https://github.com/sourcenetwork/defradb/issues/2473)
- [Limit Operations](https://github.com/sourcenetwork/defradb/issues/2472)
- [Order Operations](https://github.com/sourcenetwork/defradb/issues/2471)
- [Sum Operations](https://github.com/sourcenetwork/defradb/issues/2470)
- [Dag/Commit Operations](https://github.com/sourcenetwork/defradb/issues/2469)
- [Delete With Filter Operations](https://github.com/sourcenetwork/defradb/issues/2468)
- [Update With Filter Operations](https://github.com/sourcenetwork/defradb/issues/2467)
- [Type Join Many Operations](https://github.com/sourcenetwork/defradb/issues/2466)
- [Type Join One Operations](https://github.com/sourcenetwork/defradb/issues/2466)
- [Parallel Operations](https://github.com/sourcenetwork/defradb/issues/2465)
- [Execute Explain](https://github.com/sourcenetwork/defradb/issues/2464)
