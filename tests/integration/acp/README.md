## More Information on ACP test directories.


1) `./defradb/tests/integration/acp/dac/add_policy`
    - This directory tests ONLY the `Adding of a Policy` through DefraDB's Document ACP System.
    - Does NOT assert the schema.
    - While this uses document ACP system, the tests DO NOT test document resource interface (DRI) validation.

2) `./defradb/tests/integration/acp/dac/link_collection`
    - This directory tests the loading/adding/linking of a collection that has `@policy(id, resource)`
      specified. The tests ensure that only a collection linking to
      a valid DRI policy is accepted. Naturally these tests will also be `Adding a Policy`
      through DefraDB like in (1) before actually adding the collection. If a collection has a
      policy specified that doesn't exist (or wasn't added yet), that collection WILL/MUST
      be rejected in these tests.
    - The tests assert the collection after to ensure rejection/acceptance.
    - Tests DRI validation.

3) `./defradb/tests/integration/acp/dac/relationship/doc_actor`
    - This directory tests adding document and actor relationships.
    - This directory tests deleting document and actor relationships.

4) `./defradb/tests/integration/acp/dac/index`
    - This directory tests document acp with index.

5) `./defradb/tests/integration/acp/dac/p2p`
    - This directory tests document acp with p2p.

### Learn more about DefraDB [ACP System](/acp/README.md)
