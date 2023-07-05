# Tests 

This directory contains two types of tests: benchmark tests (located in the bench directory) and integration tests (located in the integration directory). 
In addition to these, unit tests are also distributed among the other directories of the source code.

## Test Types

### Benchmark Tests

The bench directory contains benchmark tests that are used to measure and monitor the performance of the database.

### Integration Tests

The integration directory contains integration tests that ensure different components of the system work together correctly.

### Unit Tests

Unit tests are spread throughout the source code and are located in the same directories as the code they are testing. 
These tests focus on small, isolated parts of the code to ensure each part is working as expected.

## Mocks

For unit tests, we sometimes use mocks. Mocks are automatically generated from Go interfaces using the mockery tool. 
This helps to isolate the code being tested and provide more focused and reliable tests.

To regenerate the mocks, run `make mock`.  `make test:ci` will also do this.
