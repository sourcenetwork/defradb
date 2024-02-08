## This directory tests the `Adding of a Policy` through DefraDB.

### These are NOT DefraDB Policy Interface (DPI) Tests
There are certain requirements for a DPI. A resource must be a valid DPI to link to a collection.
However it's important to note that DefraDB does allow adding policies that might not have DPI
compliant resources. But as long as sourcehub (acp module) deems them to be valid they are allowed
to be added. There are various reasons for this, mostly because DefraDB is a tool that can be used
to upload policies to sourcehub that might not be only for use with collections / schema. Nonetheless
we still need a way to validate that the resource specified on the schema that is being added is DPI
compliant resource on a already registered policy. Therefore, when a schema is being added, and it has
the policyID and resource defined using the `@policy` directive, then during the 'adding of the schema'
the validation occurs. Inotherwords, we do not allow a non-DPI compliant resource to be specified on a
schema, if it is, then the schema is rejected.

### Non-DPI Compliant Policies Documented In Tests
These test files document some cases where DefraDB would upload policies that aren't DPI compliant,
but are sourcehub compatible, might be worthwhile to look at the documented tests and notes there:
- `./with_no_perms_test.go`
- `./with_no_resources_test.go`
- `./with_permissionless_owner_test.go`
