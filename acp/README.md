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

To perform authenticated operations you will need to generate a `secp256k1` key pair:

```sh
openssl ecparam -name secp256k1 -genkey | openssl ec -text -noout
```

Copy the private key hex from the output:

```sh
read EC key
Private-Key: (256 bit)
priv:
    e3:b7:22:90:6e:e4:e5:63:68:f5:81:cd:8b:18:ab:
    0f:48:af:1e:a5:3e:63:5e:3f:7b:8a:cd:07:66:76:
    f6:ac
pub:
    04:03:96:9a:de:33:20:ec:fe:46:fb:ee:3e:d2:d8:
    45:d8:a2:eb:ba:07:0c:50:51:37:13:5b:22:ca:d0:
    14:1e:40:b8:75:62:04:8e:dd:a0:7d:41:32:c6:10:
    7d:5a:9f:c9:e5:8f:6a:e3:95:88:88:ef:86:e1:86:
    45:d6:84:dc:79
ASN1 OID: secp256k1
```

Authenticate with the identity:

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
defradb client acp policy add -f examples/dpi_policy/user_dpi_policy.yml \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

Result:
```json
{
  "PolicyID": "24ab8cba6d6f0bcfe4d2712c7d95c09dd1b8076ea5a8896476413fd6c891c18c"
}
```

### Add schema, linking to a resource within the policy we added:

We have in `examples/schema/permissioned/users.graphql`:
```graphql
type Users @policy(
    id: "24ab8cba6d6f0bcfe4d2712c7d95c09dd1b8076ea5a8896476413fd6c891c18c",
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
      "ID": "24ab8cba6d6f0bcfe4d2712c7d95c09dd1b8076ea5a8896476413fd6c891c18c",
      "ResourceName": "users"
    }
  }
]

```

### Create private documents (with identity)

CLI Command:
```sh
defradb client collection create --name Users \
'[{ "name": "SecretShahzad" }, { "name": "SecretLone" }]' \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
```

### Create public documents (without identity)

CLI Command:
```sh
defradb client collection create  --name Users '[{ "name": "PublicShahzad" }, { "name": "PublicLone" }]'
```

### Get all docIDs without an identity (shows only public):
CLI Command:
```sh
defradb client collection docIDs \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
defradb client collection docIDs \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" \
--identity 4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5
```

Error:
```
    Error: document not found or not authorized to access
```

### Update private document:
CLI Command:
```sh
defradb client collection update --name Users \
--docID "bae-a5830219-b8e7-5791-9836-2e494816fc0a" \
--updater '{ "name": "SecretUpdatedShahzad" }' \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
defradb client collection delete -i e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac --name Users --docID "bae-a5830219-b8e7-5791-9836-2e494816fc0a"
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
defradb client collection get --name Users "bae-a5830219-b8e7-5791-9836-2e494816fc0a" \
--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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


## DAC Usage HTTP:
HTTP requests work similar to their CLI counter parts, the main difference is that the identity will just be specified within the Auth Header like so: `Authorization: Basic <identity>`.

Note: The `Basic` label will change to `Bearer ` after JWS Authentication Tokens are supported.

## _AAC DPI Rules (coming soon)_
## _AAC Usage: (coming soon)_

## _FAC DPI Rules (coming soon)_
## _FAC Usage: (coming soon)_

## Warning / Caveats
The following features currently don't work with ACP, they are being actively worked on.
- [P2P: Adding a replicator with permissioned collection](https://github.com/sourcenetwork/defradb/issues/2366)
- [P2P: Subscription to a permissioned collection](https://github.com/sourcenetwork/defradb/issues/2366)
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
