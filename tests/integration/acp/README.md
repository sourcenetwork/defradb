## More Information on ACP test directories.


1) `./defradb/tests/integration/acp/add_policy`
    - This directory tests ONLY the `Adding of a Policy` through DefraDB.
    - Does NOT assert the schema.
    - Does NOT test DPI validation.

2) `./defradb/tests/integration/acp/schema/add_dpi`
    - This directory tests the loading/adding of a schema that has `@policy(id, resource)`
      specified (i.e. permissioned schema). The tests ensure that only a schema linking to
      a valid DPI policy is accepted. Naturally these tests will also be `Adding a Policy`
      through DefraDB like in (1) before actually adding the schema. If a schema has a
      policy specified that doesn't exist (or wasn't added yet), that schema WILL/MUST
      be rejected in these tests.
    - The tests assert the schema after to ensure rejection/acceptance.
    - Tests DPI validation.
