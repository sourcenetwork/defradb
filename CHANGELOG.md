
<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/sourcenetwork/defradb/compare/v0.1.0...v0.2.0)

> 2022-01-31

### Features

* Add support for sum aggregate ([#121](https://github.com/sourcenetwork/defradb/issues/121))
* Add support for lwwr scalar arrays (full replace on update) ([#115](https://github.com/sourcenetwork/defradb/issues/115))
* Add count aggregate support ([#102](https://github.com/sourcenetwork/defradb/issues/102))
* Add support for named relationships ([#108](https://github.com/sourcenetwork/defradb/issues/108))
* Add multi doc key lookup support ([#76](https://github.com/sourcenetwork/defradb/issues/76))
* Add basic group by functionality ([#43](https://github.com/sourcenetwork/defradb/issues/43))
* Update datastore packages to allow use of context ([#48](https://github.com/sourcenetwork/defradb/issues/48))

### Bug fixes

* Handle errors generated during input object thunks ([#123](https://github.com/sourcenetwork/defradb/issues/123))
* Remove new types from in-memory cache on generate error ([#122](https://github.com/sourcenetwork/defradb/issues/122))
* Support relationships where both fields have the same name ([#109](https://github.com/sourcenetwork/defradb/issues/109))
* Handle errors generated in fields thunk ([#66](https://github.com/sourcenetwork/defradb/issues/66))
* Ensure OperationDefinition case has at least one selection([#24](https://github.com/sourcenetwork/defradb/pull/24))
* Close datastore iterator on scan close ([#56](https://github.com/sourcenetwork/defradb/pull/56)) (resulted in a panic when using limit)
* Close superseded iterators before orphaning ([#56](https://github.com/sourcenetwork/defradb/pull/56)) (fixes a panic in the join code) 
* Move discard to after error check ([#88](https://github.com/sourcenetwork/defradb/pull/88)) (did result in panic if transaction creation fails)
* Check for nil iterator before closing document fetcher ([#108](https://github.com/sourcenetwork/defradb/pull/108))

### Documentation

* Correcting comment typos ([#142](https://github.com/sourcenetwork/defradb/issues/142))
* Correcting README typos ([#140](https://github.com/sourcenetwork/defradb/issues/140))

### Testing

* Allow running of tests using badger-file as well as IM options ([#128](https://github.com/sourcenetwork/defradb/issues/128))
* Add test datastore selection support ([#88](https://github.com/sourcenetwork/defradb/issues/88))

### Refactoring

* Datatype modification protection ([#138](https://github.com/sourcenetwork/defradb/issues/138))
* Cleanup Linter Complaints and Setup Makefile ([#63](https://github.com/sourcenetwork/defradb/issues/63))
* Rework document rendering to avoid data duplication and mutation ([#68](https://github.com/sourcenetwork/defradb/issues/68))
* Remove dependency on concrete datastore implementations from db package ([#51](https://github.com/sourcenetwork/defradb/issues/51))
* Remove all `errors.Wrap` and update them with `fmt.Errorf`. ([#41](https://github.com/sourcenetwork/defradb/issues/41))
* Restructure integration tests to provide better visibility ([#15](https://github.com/sourcenetwork/defradb/pull/15))
* Remove schemaless code branches ([#23](https://github.com/sourcenetwork/defradb/pull/23))

### Continous integration

* Use more accurate test coverage. ([#134](https://github.com/sourcenetwork/defradb/issues/134))
* Disable Codecov's Patch Check
* Make codcov less strict for now to unblock development ([#125](https://github.com/sourcenetwork/defradb/issues/125))
* Add codecov config file. ([#118](https://github.com/sourcenetwork/defradb/issues/118))
* Add workflow that runs a job on AWS EC2 instance. ([#110](https://github.com/sourcenetwork/defradb/issues/110))
* Add Code Test Coverage with CodeCov ([#116](https://github.com/sourcenetwork/defradb/issues/116))
* Integrate GitHub Action for golangci-lint Annotations ([#106](https://github.com/sourcenetwork/defradb/issues/106))
* Add Linter Check to CircleCi ([#92](https://github.com/sourcenetwork/defradb/issues/92))

### Chore

* Remove the S1038 rule of the gosimple linter. ([#129](https://github.com/sourcenetwork/defradb/issues/129))
* Update to badger v3, and use badger as default in memory store ([#56](https://github.com/sourcenetwork/defradb/issues/56))
* Make Cid versions consistent ([#57](https://github.com/sourcenetwork/defradb/issues/57))


<a name="v0.1.0"></a>
## v0.1.0

> 2021-03-15
