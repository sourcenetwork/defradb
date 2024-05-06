---
sidebar_label: Schema Migration Guide
sidebar_position: 60
---
# A Guide to Schema Migration in DefraDB

## Overview
In a database system, an application’s requirements can change at any given time, to meet this change, Schema migrations are necessary. This is where Lens comes in, as a migration engine that produces effective schema migration.

This guide will provide an understanding of schema migrations, focusing on the Lens migration engine. Let’s dive in!

Lens is a pipeline for user-defined transformations. It enables users to write their transformations in any programming language and run them through the Lens pipeline, which transforms the cached representation of the data.

## Goals of the Lens Migration System

Here are some of the goals of the Lens schema migration system:

- **Presenting a consistent view of data across nodes**: The Lens schema migration system can present data across nodes consistently, regardless of the schema version being used.

- **Verifiability of data**: Schema migration in the Lens migration system is presented as data, this preserves the user-defined mutations without corrupting system-defined mutations and also allows migrating from one schema version to another.

- **A language-agnostic way of writing schema migrations**: Schema migrations can be written in any programming language and executed properly as Lens is language-agnostic.

- **Safe usage of migrations by others through a sandbox**: Migrations written in Lens are run in a sandbox, which ensures safety and eliminates the concern for remote code executions (RCE).

- **Peer-to-peer sync of schema migrations**: Lens allows peers to write their migrations in different application versions and sync without worrying about the versions other peers are using.

- **Local autonomy of schema migrations**: Lens enables local autonomy in writing schema migrations by giving users control of the schema version they choose to use. The users can stay in a particular schema version and still communicate with peers on different versions, as Lens is not restricted to a particular schema version.

- **Reproducibility and deterministic nature of executing migrations**: When using the Lens migration system, changes to schemas can be written, tagged and shared with other peers regardless of their infrastructure and requirements for deployments.


## Mechanism

In this section, we’ll look at the mechanism behind the Lens migration system and explain how it works.

Lens migration system functions as a bi-directional transformation engine, enabling the migration of data documents in both forward and reverse directions. It allows for the transformation of documents from schema X to Y in the forward direction and Y to X in the reverse direction.

The above process is done foundationally, through a verifiable system powered by WebAssembly (Wasm). Wasm also enables the sandbox safety and language-agnostic feature of Lens.

Internally, schema migrations are evaluated lazily. This avoids the upfront cost of doing a massive migration at once.

*Lazy evaluation is a technique in programming where an expression is only evaluated when its value is needed.*

Adopting lazy evaluation in the migration system also allows rapid toggling between schema versions and representations.

## Usage

The Lens migration system addresses critical use cases related to schema migrations in peer-to-peer, eventually consistent databases. These use cases include:

 

- **Safe Schema Progression**: Ensuring the seamless progression of database schemas is vital for accommodating changing application requirements. Lens facilitates the modification, upgrade, or reversion of schemas while upholding data integrity.

- **Handling Peer-to-Peer Complexity**: In environments where different clients operate on varying application and database versions, Lens offers a solution to address the complexity of schema migrations. It ensures coherence and effectiveness across different networks.

- **Language-Agnostic Flexibility**: Functions in Lens are designed to be language-agnostic, offering the versatility to define schema changes in the preferred programming language. This adaptability makes Lens suitable for diverse development environments and preferences.

- **Lazy Evaluation**: Lens employs a lazy evaluation mechanism, initiating migrations without immediate execution. Schema changes are applied only when documents are read, queried, or updated. This approach reduces the upfront cost of extensive schema migrations while maintaining data consistency.

- **On-Demand Schema Selection**: Lens supports on-demand schema selection during data queries. Users can specify the schema version they wish to work with, facilitating A/B testing and the seamless transition between different schema versions.



These use cases highlight how Lens empowers users to manage schema migrations effectively, ensuring data consistency and adaptability in evolving database systems.


## Example

In this example we will define a collection using a schema with an `emailAddress` field.  We will then patch the schema to add a new field `email`, then define a bi-directional Lens to migrate data to/from the new field.

**Step One**, define the `Users` collection/schema:

```graphql
defradb client schema add '
    type Users {
        emailAddress: String
    }
'
```

**Step Two**, patch the `Users` schema, adding the new field, here we pass in `--set-active=true` to automatically apply the schema change to the `Users` collection:

```graphql
defradb client schema patch '
    [
    	{ "op": "add", "path": "/Users/Fields/-", "value": {"Name": "email", "Kind": "String"} }
    ]
' --set-active=true
```

**Step Three**, fetch the schema ids so that we can later tell Defra which schema versions we wish to migrate to/from:

```graphql
defradb client schema describe --name="Users"
```

**Step Four**, in order to define our Lens module - we need to define 4 functions:

- `next() unsignedInteger8`, this is a host function imported to the module - calling it will return a pointer to a byte array that will either contain
 an error, an EndOfStream identifier (indicating that there are no more source values), or a pointer to the start of a json byte array containing the Defra document to migrate.  It is typically called from within the `transform` and `inverse` functions, and can be called multiple times within them if desired.

 - `alloc(size: unsignedInteger64) unsignedInteger8`​, this is required by all lens modules regardless of language or content - this function should allocate a block of memory of the given `size` , it is used by the Lens engine to pass stuff in to the wasm instance.  The memory needs to remain reserved until the next wasm call, e.g. until `transform` or `set_param` has been called. It's implementation will be different depending on which language you are working with, but it should not need to differ between modules of the same language.  The Rust SDK contains an alloc function that you can call.

- `set_param(ptr: unsignedInteger8) unsignedInteger8`​, this function is only required by modules that accept a set of parameters.  As an input parameter it receives a single pointer that will point to the start of a json byte array containing the parameters defined in the configuration file.  It returns a pointer to either nil, or an error message. It will be called once, when the the migration is defined in Defra (and on restart of the database).  How it is implemented is up to you.

- `transform() unsignedInteger8`​, this function is required by all Lens modules - it is the migration, and within this function you should define what the migration should do, in this example it will copy the data from the `emailAddress` field into the `email` field. Lens Modules can call the `next` function zero to many times to draw documents from the Defra datastore, however modules used in schema migrations should currently limit this to a single call per `transform` call (Lens based views may call it more or less frequently in order to filter or create documents).

- `inverse() unsignedInteger8`​, this function is optional, you only need to define it if you wish to define the inverse migration.  It follows the same pattern as the `transform` function, only you should implement it to do the reverse.  In this example we want this to copy the value from the `email` field into the `emailAddress`​ field.

Here is what our migration would look like if we were to write it in Rust:

```graphql
#[link(wasm_import_module = "lens")]
extern "C" {
    fn next() -> *mut u8;
}

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub src: String,
    pub dst: String,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

#[no_mangle]
pub extern fn alloc(size: usize) -> *mut u8 {
    lens_sdk::alloc(size)
}

#[no_mangle]
pub extern fn set_param(ptr: *mut u8) -> *mut u8 {
    match try_set_param(ptr) {
        Ok(_) => lens_sdk::nil_ptr(),
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_set_param(ptr: *mut u8) -> Result<(), Box<dyn Error>> {
    let parameter = lens_sdk::try_from_mem::<Parameters>(ptr)?;

    let mut dst = PARAMETERS.write()?;
    *dst = Some(parameter);
    Ok(())
}

#[no_mangle]
pub extern fn transform() -> *mut u8 {
    match try_transform() {
        Ok(o) => match o {
            Some(result_json) => lens_sdk::to_mem(lens_sdk::JSON_TYPE_ID, &result_json),
            None => lens_sdk::nil_ptr(),
            EndOfStream => lens_sdk::to_mem(lens_sdk::EOS_TYPE_ID, &[]),
        },
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_transform() -> Result<StreamOption<Vec<u8>>, Box<dyn Error>> {
    let ptr = unsafe { next() };
    let mut input = match lens_sdk::try_from_mem::<HashMap<String, serde_json::Value>>(ptr)? {
        Some(v) => v,
        // Implementations of `transform` are free to handle nil however they like. In this
        // implementation we chose to return nil given a nil input.
        None => return Ok(None),
        EndOfStream => return Ok(EndOfStream)
    };

    let params = PARAMETERS.read()?;

    let value = input.get_mut(&params.src)
        .ok_or(ModuleError::PropertyNotFoundError{requested: params.src.clone()})?
        .clone();

    let mut result = input.clone();
    result.insert(params.dst, value);

    let result_json = serde_json::to_vec(&result)?;
    lens_sdk::free_transport_buffer(ptr)?;
    Ok(Some(result_json))
}

#[no_mangle]
pub extern fn inverse() -> *mut u8 {
    match try_inverse() {
        Ok(o) => match o {
            Some(result_json) => lens_sdk::to_mem(lens_sdk::JSON_TYPE_ID, &result_json),
            None => lens_sdk::nil_ptr(),
            EndOfStream => lens_sdk::to_mem(lens_sdk::EOS_TYPE_ID, &[]),
        },
        Err(e) => lens_sdk::to_mem(lens_sdk::ERROR_TYPE_ID, &e.to_string().as_bytes())
    }
}

fn try_inverse() -> Result<StreamOption<Vec<u8>>, Box<dyn Error>> {
    let ptr = unsafe { next() };
    let mut input = match lens_sdk::try_from_mem::<HashMap<String, serde_json::Value>>(ptr)? {
        Some(v) => v,
        // Implementations of `transform` are free to handle nil however they like. In this
        // implementation we chose to return nil given a nil input.
        None => return Ok(None),
        EndOfStream => return Ok(EndOfStream)
    };

    let params = PARAMETERS.read()?;

	// Note: In this example `inverse` is exactly the same as `transform`, only the useage
    // of `params.dst` and `params.src` is reversed.
    let value = input.get_mut(&params.dst)?;

    let mut result = input.clone();
    result.insert(params.src, value);

    let result_json = serde_json::to_vec(&result)?;
    lens_sdk::free_transport_buffer(ptr)?;
    Ok(Some(result_json))
}
```




More fully coded example modules, including an AssemblyScript example can be found in our integration tests here: https://github.com/sourcenetwork/defradb/tree/develop/tests/lenses

and here: https://github.com/lens-vm/lens/tree/main/tests/modules

We should then compile it to wasm, and copy the resultant `.wasm` file to a location that the Defra node has access to.  Make sure that the file is safe there, at the moment Defra will not copy it and will refer back to that location on database restart.

**Step Five**, now that we have updated the collection, and defined our migration, we need to tell Defra to use it, by providing it the source and destination schema IDs from our earlier `defradb client schema describe`​ call, and a configuration file defining the parameters we wish to pass it:

```graphql
defradb client schema migration set <The source schema ID> <The destination schema ID> '
    {
        "lenses": [
            {
				"path": <The path to your compiled `.wasm` binary from step four>,
				"arguments": {
					"src": "emailAddress",
					"dst": "email"
				}
            }
        ]
    }
'
```


Now the migration has been configured!  Any documents committed under the original schema version will now be returned as if they were committed using the newer schema version.

As we have defined an inverse migration, we can give this migration to other nodes in our peer network still on the original schema version, and they will be able to query our documents committed using the new schema version applying the inverse.

We can also change our active schema version on this node back to the original to see the inverse in action:

```graphql
defradb client schema set-active <Original schema ID>
```

Now when we query Defra, any documents committed after the schema update will be rendered as if they were committed on the original schema version, with `email` field values being copied to the `emailAddress` field at query time.

## Advantages 

Here are some advantages of Lens as a schema migration system:

- Lens is not bound to a particular deployment, programming language, or interaction method. It can be used globally and is accessible to clients regardless of their location or infrastructure. 
- Users can query on-demand even with different schema versions.
- Migration between different schemas is a seamless process.

## Disadvantages

The Lens migration system also has some downsides to schema migration which include:

- Using a Lazy execution approach, errors might be found later when querying through the migration.
- There’s a time constraint as the Lens migration system is a work in progress
- The performance of the system is secondary, with more focus on overall functionality.

## Future Outlook

The core problem we currently have in the Lens schema migration system is the performance issues when migrating schemas, hence for future versions, the following would be considered:

- Increasing the performance of the migration system.
- Making migrations easier to write.
- Expansion of the schema update system to include the removal of fields, not just adding fields.
- Enabling users to query the schema version of their choice on-demand.
- Support for Eager evaluation.
- Implementing dry run testing for development and branching scenarios, and handling divergent schemas.