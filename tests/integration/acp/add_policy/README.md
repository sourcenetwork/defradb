## This directory tests the `Adding of a Policy` through DefraDB.

### These are NOT DefraDB Policy Interface (DPI) Tests
There are certain requirements for DPI. A policy must be a valid DPI to link to a collection.
However it's important to note that DefraDB does allow uploading / adding policies that aren't
DPI compliant as long as sourcehub (acp module) deems them to be valid. There are various reasons
for this, mostly because DefraDB is a tool that can be used to upload policies to sourcehub that
might not be only for use with collections / schema. Nonetheless we still need a way to validate
that the policy linked within a collection within the schema that is being added/loading is valid.
Therefore, when a schema is being loaded, and it has policyID and resource defined on the
collection with the appropriate directive. At that point before we accept that schema the
validation occurs. Inotherwords, we do not allow a non-DPI compliant policy to be specified
on a collection schema, if it is, then the schema would be rejected.

### Non-DPI Compliant Policies Documented In Tests
These test files document some cases where DefraDB would upload policies that aren't DPI compliant,
but are sourcehub compatible, might be worthwhile to look at the documented tests and notes there:
- `./with_no_perms_test.go`
- `./with_no_resources_test.go`
- `./with_permissionless_owner_test.go`
