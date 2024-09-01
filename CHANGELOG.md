<a name="v0.13.0"></a>
## [v0.13.0](https://github.com/sourcenetwork/defradb/compare/v0.12.0...v0.13.0)

> 2024-08-23

DefraDB v0.13 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.12.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Enable indexing for DateTime fields ([#2933](https://github.com/sourcenetwork/defradb/issues/2933))
* Remove IsObjectArray ([#2859](https://github.com/sourcenetwork/defradb/issues/2859))
* Handle P2P with SourceHub ACP ([#2848](https://github.com/sourcenetwork/defradb/issues/2848))
* Implement SourceHub ACP ([#2657](https://github.com/sourcenetwork/defradb/issues/2657))
* Doc field encryption ([#2817](https://github.com/sourcenetwork/defradb/issues/2817))
* Doc encryption with symmetric key ([#2731](https://github.com/sourcenetwork/defradb/issues/2731))

### Fixes

* Filter with date and document with nil date value ([#2946](https://github.com/sourcenetwork/defradb/issues/2946))
* Handle index queries where child found without parent ([#2942](https://github.com/sourcenetwork/defradb/issues/2942))
* Add ns precision support to time values ([#2940](https://github.com/sourcenetwork/defradb/issues/2940))
* Panic with different composite-indexed child objects ([#2947](https://github.com/sourcenetwork/defradb/issues/2947))
* No panic if filter condition on indexed field is empty ([#2929](https://github.com/sourcenetwork/defradb/issues/2929))
* Create mutation introspection ([#2881](https://github.com/sourcenetwork/defradb/issues/2881))
* Handle multiple child index joins ([#2867](https://github.com/sourcenetwork/defradb/issues/2867))
* Enable filtering doc by fields of JSON and Blob types ([#2841](https://github.com/sourcenetwork/defradb/issues/2841))
* Allow querying of 9th, 19th, 29th, etc collections ([#2819](https://github.com/sourcenetwork/defradb/issues/2819))
* Support one-many self joins without primary directive ([#2799](https://github.com/sourcenetwork/defradb/issues/2799))
* Allow primary field declarations on one-many ([#2796](https://github.com/sourcenetwork/defradb/issues/2796))

### Refactoring

* GQL responses ([#2872](https://github.com/sourcenetwork/defradb/issues/2872))
* Network test sync logic ([#2748](https://github.com/sourcenetwork/defradb/issues/2748))
* Decouple client.DB from net ([#2768](https://github.com/sourcenetwork/defradb/issues/2768))

### Testing

* Add assert on DocIndex for child documents ([#2871](https://github.com/sourcenetwork/defradb/issues/2871))
* Fix refreshing of docs in change detector ([#2832](https://github.com/sourcenetwork/defradb/issues/2832))
* Remove hardcoded test identities ([#2822](https://github.com/sourcenetwork/defradb/issues/2822))
* Allow assertion of AddSchema results ([#2788](https://github.com/sourcenetwork/defradb/issues/2788))

### Chore

* Bump grpc version ([#2824](https://github.com/sourcenetwork/defradb/issues/2824))

### Bot

* Update dependencies (bulk dependabot PRs) 21-08-2024 ([#2939](https://github.com/sourcenetwork/defradb/issues/2939))
* Update dependencies (bulk dependabot PRs) 20-08-2024 ([#2932](https://github.com/sourcenetwork/defradb/issues/2932))
* Update dependencies (bulk dependabot PRs) 12-08-2024 ([#2904](https://github.com/sourcenetwork/defradb/issues/2904))
* Bump eslint from 8.57.0 to 9.9.0 in /playground ([#2903](https://github.com/sourcenetwork/defradb/issues/2903))
* Bump golang.org/x/term from 0.22.0 to 0.23.0 ([#2890](https://github.com/sourcenetwork/defradb/issues/2890))
* Update dependencies (bulk dependabot PRs) 06-08-2024 ([#2889](https://github.com/sourcenetwork/defradb/issues/2889))
* Update dependencies (bulk dependabot PRs) 18-07-2024 ([#2846](https://github.com/sourcenetwork/defradb/issues/2846))
* Update dependencies (bulk dependabot PRs) 05-07-2024 ([#2811](https://github.com/sourcenetwork/defradb/issues/2811))
* Bump github.com/cometbft/cometbft from 0.38.7 to 0.38.8 ([#2794](https://github.com/sourcenetwork/defradb/issues/2794))

<a name="v0.12.0"></a>
## [v0.12.0](https://github.com/sourcenetwork/defradb/compare/v0.11.0...v0.12.0)

> 2024-06-28

DefraDB v0.12 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.11.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Ability to generate a new identity ([#2760](https://github.com/sourcenetwork/defradb/issues/2760))
* Add async transaction callbacks ([#2708](https://github.com/sourcenetwork/defradb/issues/2708))
* Allow lens runtime selection via config ([#2684](https://github.com/sourcenetwork/defradb/issues/2684))
* Sec. indexes on relations ([#2670](https://github.com/sourcenetwork/defradb/issues/2670))
* Add authentication for ACP ([#2649](https://github.com/sourcenetwork/defradb/issues/2649))
* Inject ACP instance into the DB instance ([#2633](https://github.com/sourcenetwork/defradb/issues/2633))
* Keyring ([#2557](https://github.com/sourcenetwork/defradb/issues/2557))
* Enable sec. indexes with ACP ([#2602](https://github.com/sourcenetwork/defradb/issues/2602))

### Fixes

* Race condition when testing CLI ([#2713](https://github.com/sourcenetwork/defradb/issues/2713))
* Remove shared mutable state between database instances ([#2777](https://github.com/sourcenetwork/defradb/issues/2777))
* Change new identity keys to hex format ([#2773](https://github.com/sourcenetwork/defradb/issues/2773))
* Return slice of correct length from db.AddSchema ([#2765](https://github.com/sourcenetwork/defradb/issues/2765))
* Use node representation for Block ([#2746](https://github.com/sourcenetwork/defradb/issues/2746))
* Add version check in basicTxn.Query ([#2742](https://github.com/sourcenetwork/defradb/issues/2742))
* Merge retry logic ([#2719](https://github.com/sourcenetwork/defradb/issues/2719))
* Resolve incorrect merge conflict ([#2723](https://github.com/sourcenetwork/defradb/issues/2723))
* Keyring output ([#2784](https://github.com/sourcenetwork/defradb/issues/2784))
* Incorporate schema root into docID ([#2701](https://github.com/sourcenetwork/defradb/issues/2701))
* Make node options composable ([#2648](https://github.com/sourcenetwork/defradb/issues/2648))
* Remove limit for fetching secondary docs ([#2594](https://github.com/sourcenetwork/defradb/issues/2594))

### Documentation

* Remove reference to client ping from readme ([#2793](https://github.com/sourcenetwork/defradb/issues/2793))
* Add http/openapi documentation & ci workflow ([#2678](https://github.com/sourcenetwork/defradb/issues/2678))
* Streamline cli documentation ([#2646](https://github.com/sourcenetwork/defradb/issues/2646))
* Document Event Update struct ([#2598](https://github.com/sourcenetwork/defradb/issues/2598))

### Refactoring

* Use events to test network logic ([#2700](https://github.com/sourcenetwork/defradb/issues/2700))
* Change local_acp implementation to use acp_core ([#2691](https://github.com/sourcenetwork/defradb/issues/2691))
* Rework definition validation ([#2720](https://github.com/sourcenetwork/defradb/issues/2720))
* Extract definition stuff from collection.go ([#2706](https://github.com/sourcenetwork/defradb/issues/2706))
* Change counters to support encryption ([#2698](https://github.com/sourcenetwork/defradb/issues/2698))
* DAG sync and move merge outside of net package ([#2658](https://github.com/sourcenetwork/defradb/issues/2658))
* Replace subscription events publisher ([#2686](https://github.com/sourcenetwork/defradb/issues/2686))
* Extract Defra specific logic from ACPLocal type ([#2656](https://github.com/sourcenetwork/defradb/issues/2656))
* Change from protobuf to cbor for IPLD ([#2604](https://github.com/sourcenetwork/defradb/issues/2604))
* Reorganize global CLI flags ([#2615](https://github.com/sourcenetwork/defradb/issues/2615))
* Move internal packages to internal dir ([#2599](https://github.com/sourcenetwork/defradb/issues/2599))

### Testing

* Remove duplicate test ([#2787](https://github.com/sourcenetwork/defradb/issues/2787))
* Support asserting on doc index in test results ([#2786](https://github.com/sourcenetwork/defradb/issues/2786))
* Allow test harness to execute benchmarks ([#2740](https://github.com/sourcenetwork/defradb/issues/2740))
* Add relation substitute mechanic to tests ([#2682](https://github.com/sourcenetwork/defradb/issues/2682))
* Test node pkg constructor via integration test suite ([#2641](https://github.com/sourcenetwork/defradb/issues/2641))

### Continuous integration

* Cache dependencies to speed up test runs ([#2732](https://github.com/sourcenetwork/defradb/issues/2732))

### Bot

* Update dependencies (bulk dependabot PRs) 24-06-2024 ([#2761](https://github.com/sourcenetwork/defradb/issues/2761))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 7.13.0 to 7.13.1 in /playground ([#2733](https://github.com/sourcenetwork/defradb/issues/2733))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 7.13.0 to 7.13.1 in /playground ([#2734](https://github.com/sourcenetwork/defradb/issues/2734))
* Update dependencies (bulk dependabot PRs) 06-17-2024 ([#2730](https://github.com/sourcenetwork/defradb/issues/2730))
* Bump braces from 3.0.2 to 3.0.3 in /playground ([#2716](https://github.com/sourcenetwork/defradb/issues/2716))
* Update dependencies (bulk dependabot PRs) 06-10-2024 ([#2705](https://github.com/sourcenetwork/defradb/issues/2705))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 7.11.0 to 7.12.0 in /playground ([#2675](https://github.com/sourcenetwork/defradb/issues/2675))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 7.11.0 to 7.12.0 in /playground ([#2676](https://github.com/sourcenetwork/defradb/issues/2676))
* Update dependencies (bulk dependabot PRs) 03-06-2024 ([#2674](https://github.com/sourcenetwork/defradb/issues/2674))
* Update dependencies (bulk dependabot PRs) 01-06-2024 ([#2660](https://github.com/sourcenetwork/defradb/issues/2660))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 7.9.0 to 7.10.0 in /playground ([#2635](https://github.com/sourcenetwork/defradb/issues/2635))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 7.9.0 to 7.10.0 in /playground ([#2637](https://github.com/sourcenetwork/defradb/issues/2637))
* Bump swagger-ui-react from 5.17.10 to 5.17.12 in /playground ([#2636](https://github.com/sourcenetwork/defradb/issues/2636))
* Bump google.golang.org/protobuf from 1.33.0 to 1.34.1 ([#2607](https://github.com/sourcenetwork/defradb/issues/2607))
* Update dependencies (bulk dependabot PRs) 05-20-2024 ([#2631](https://github.com/sourcenetwork/defradb/issues/2631))
* Update dependencies (bulk dependabot PRs) 05-14-2024 ([#2617](https://github.com/sourcenetwork/defradb/issues/2617))

<a name="v0.11.0"></a>
## [v0.11.0](https://github.com/sourcenetwork/defradb/compare/v0.10.0...v0.11.0)

> 2024-05-03

DefraDB v0.11 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.10.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Update corelog to 0.0.7 ([#2547](https://github.com/sourcenetwork/defradb/issues/2547))
* Move relation field properties onto collection ([#2529](https://github.com/sourcenetwork/defradb/issues/2529))
* Lens runtime config ([#2497](https://github.com/sourcenetwork/defradb/issues/2497))
* Add P Counter CRDT ([#2482](https://github.com/sourcenetwork/defradb/issues/2482))
* Add Access Control Policy ([#2338](https://github.com/sourcenetwork/defradb/issues/2338))
* Force explicit primary decl. in SDL for one-ones ([#2462](https://github.com/sourcenetwork/defradb/issues/2462))
* Allow mutation of col sources via PatchCollection ([#2424](https://github.com/sourcenetwork/defradb/issues/2424))
* Add Defra-Lens support for branching schema ([#2421](https://github.com/sourcenetwork/defradb/issues/2421))
* Add PatchCollection ([#2402](https://github.com/sourcenetwork/defradb/issues/2402))

### Fixes

* Return correct results from one-many indexed filter ([#2579](https://github.com/sourcenetwork/defradb/issues/2579))
* Handle compound filters on related indexed fields ([#2575](https://github.com/sourcenetwork/defradb/issues/2575))
* Add check to filter result for logical ops ([#2573](https://github.com/sourcenetwork/defradb/issues/2573))
* Make all array kinds nillable ([#2534](https://github.com/sourcenetwork/defradb/issues/2534))
* Allow update when updating non-indexed field ([#2511](https://github.com/sourcenetwork/defradb/issues/2511))

### Documentation

* Add data definition document ([#2544](https://github.com/sourcenetwork/defradb/issues/2544))

### Refactoring

* Merge collection UpdateWith and DeleteWith ([#2531](https://github.com/sourcenetwork/defradb/issues/2531))
* DB transactions context ([#2513](https://github.com/sourcenetwork/defradb/issues/2513))
* Add NormalValue ([#2404](https://github.com/sourcenetwork/defradb/issues/2404))
* Clean up client/request package ([#2443](https://github.com/sourcenetwork/defradb/issues/2443))
* Rewrite convertImmutable ([#2445](https://github.com/sourcenetwork/defradb/issues/2445))
* Unify Field Kind and Schema properties ([#2414](https://github.com/sourcenetwork/defradb/issues/2414))
* Replace logging package with corelog ([#2406](https://github.com/sourcenetwork/defradb/issues/2406))

### Testing

* Add flag to skip network tests ([#2495](https://github.com/sourcenetwork/defradb/issues/2495))

### Bot

* Update dependencies (bulk dependabot PRs) 30-04-2024 ([#2570](https://github.com/sourcenetwork/defradb/issues/2570))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 7.7.0 to 7.7.1 in /playground ([#2550](https://github.com/sourcenetwork/defradb/issues/2550))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 7.7.0 to 7.7.1 in /playground ([#2551](https://github.com/sourcenetwork/defradb/issues/2551))
* Bump swagger-ui-react from 5.16.2 to 5.17.0 in /playground ([#2549](https://github.com/sourcenetwork/defradb/issues/2549))
* Update dependencies (bulk dependabot PRs) 23-04-2023 ([#2548](https://github.com/sourcenetwork/defradb/issues/2548))
* Bump go.opentelemetry.io/otel/sdk/metric from 1.24.0 to 1.25.0 ([#2499](https://github.com/sourcenetwork/defradb/issues/2499))
* Bump typescript from 5.4.3 to 5.4.5 in /playground ([#2515](https://github.com/sourcenetwork/defradb/issues/2515))
* Bump swagger-ui-react from 5.14.0 to 5.15.0 in /playground ([#2514](https://github.com/sourcenetwork/defradb/issues/2514))
* Update dependencies (bulk dependabot PRs) 2024-04-09 ([#2509](https://github.com/sourcenetwork/defradb/issues/2509))
* Update dependencies (bulk dependabot PRs) 2024-04-03 ([#2492](https://github.com/sourcenetwork/defradb/issues/2492))
* Update dependencies (bulk dependabot PRs) 03-04-2024 ([#2486](https://github.com/sourcenetwork/defradb/issues/2486))
* Bump github.com/multiformats/go-multiaddr from 0.12.2 to 0.12.3 ([#2480](https://github.com/sourcenetwork/defradb/issues/2480))
* Bump [@types](https://github.com/types)/react from 18.2.66 to 18.2.67 in /playground ([#2427](https://github.com/sourcenetwork/defradb/issues/2427))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 7.2.0 to 7.3.1 in /playground ([#2428](https://github.com/sourcenetwork/defradb/issues/2428))
* Update dependencies (bulk dependabot PRs) 19-03-2024 ([#2426](https://github.com/sourcenetwork/defradb/issues/2426))
* Update dependencies (bulk dependabot PRs) 03-11-2024 ([#2399](https://github.com/sourcenetwork/defradb/issues/2399))

<a name="v0.10.0"></a>
## [v0.10.0](https://github.com/sourcenetwork/defradb/compare/v0.9.0...v0.10.0)

> 2024-03-08

DefraDB v0.10 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.9.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add case insensitive `like` operator ([#2368](https://github.com/sourcenetwork/defradb/issues/2368))
* Reverted order for indexed fields ([#2335](https://github.com/sourcenetwork/defradb/issues/2335))
* Rework GetCollection/SchemaByFoo funcs into single ([#2319](https://github.com/sourcenetwork/defradb/issues/2319))
* Add support for views with Lens transforms ([#2311](https://github.com/sourcenetwork/defradb/issues/2311))
* Model Col. SchemaVersions and migrations on Cols ([#2286](https://github.com/sourcenetwork/defradb/issues/2286))
* Replace FieldDescription.RelationType with IsPrimary ([#2288](https://github.com/sourcenetwork/defradb/issues/2288))
* Multiple docs with nil value on unique-indexed field ([#2276](https://github.com/sourcenetwork/defradb/issues/2276))
* Allow setting null values on doc fields ([#2273](https://github.com/sourcenetwork/defradb/issues/2273))
* Add JSON scalar ([#2254](https://github.com/sourcenetwork/defradb/issues/2254))
* Generate OpenAPI command ([#2235](https://github.com/sourcenetwork/defradb/issues/2235))
* Add composite indexes ([#2226](https://github.com/sourcenetwork/defradb/issues/2226))

### Fixes

* Add `latest` image tag for ghcr ([#2340](https://github.com/sourcenetwork/defradb/issues/2340))
* Move field id off of schema ([#2336](https://github.com/sourcenetwork/defradb/issues/2336))
* Make returned collections respect explicit transactions ([#2385](https://github.com/sourcenetwork/defradb/issues/2385))
* Update GetCollections behaviour ([#2378](https://github.com/sourcenetwork/defradb/issues/2378))
* Add missing directive definitions ([#2369](https://github.com/sourcenetwork/defradb/issues/2369))
* Add validation to JSON fields ([#2375](https://github.com/sourcenetwork/defradb/issues/2375))
* Make peers sync secondary index ([#2390](https://github.com/sourcenetwork/defradb/issues/2390))
* Load root dir before loading config ([#2266](https://github.com/sourcenetwork/defradb/issues/2266))
* Mark docs as deleted when querying in delete mut ([#2298](https://github.com/sourcenetwork/defradb/issues/2298))
* Add missing logs at startup ([#2391](https://github.com/sourcenetwork/defradb/issues/2391))
* Add missing delta payload ([#2306](https://github.com/sourcenetwork/defradb/issues/2306))
* Fix compound relational filters in aggregates ([#2297](https://github.com/sourcenetwork/defradb/issues/2297))

### Refactoring

* Generate field ids using a sequence ([#2339](https://github.com/sourcenetwork/defradb/issues/2339))
* Make config internal to CLI ([#2310](https://github.com/sourcenetwork/defradb/issues/2310))
* Node config ([#2296](https://github.com/sourcenetwork/defradb/issues/2296))
* HTTP config ([#2278](https://github.com/sourcenetwork/defradb/issues/2278))
* Remove unused Delete field from client.Document ([#2275](https://github.com/sourcenetwork/defradb/issues/2275))
* Decouple net config ([#2258](https://github.com/sourcenetwork/defradb/issues/2258))
* Make CollectionDescription.Name Option ([#2223](https://github.com/sourcenetwork/defradb/issues/2223))

### Chore

* Bump to GoLang v1.21 ([#2195](https://github.com/sourcenetwork/defradb/issues/2195))

### Bot

* Update dependencies (bulk dependabot PRs) 05-02-2024 ([#2372](https://github.com/sourcenetwork/defradb/issues/2372))
* Update dependencies (bulk dependabot PRs) 02-27-2024 ([#2353](https://github.com/sourcenetwork/defradb/issues/2353))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.21.0 to 7.0.1 in /playground ([#2331](https://github.com/sourcenetwork/defradb/issues/2331))
* Bump google.golang.org/grpc from 1.61.0 to 1.61.1 ([#2320](https://github.com/sourcenetwork/defradb/issues/2320))
* Update dependencies (bulk dependabot PRs) 2024-02-19 ([#2330](https://github.com/sourcenetwork/defradb/issues/2330))
* Bump vite from 5.1.1 to 5.1.2 in /playground ([#2317](https://github.com/sourcenetwork/defradb/issues/2317))
* Bump golang.org/x/net from 0.20.0 to 0.21.0 ([#2301](https://github.com/sourcenetwork/defradb/issues/2301))
* Update dependencies (bulk dependabot PRs) 2023-02-14 ([#2313](https://github.com/sourcenetwork/defradb/issues/2313))
* Update dependencies (bulk dependabot PRs) 02-07-2024 ([#2294](https://github.com/sourcenetwork/defradb/issues/2294))
* Update dependencies (bulk dependabot PRs) 30-01-2024 ([#2270](https://github.com/sourcenetwork/defradb/issues/2270))
* Update dependencies (bulk dependabot PRs) 23-01-2024 ([#2252](https://github.com/sourcenetwork/defradb/issues/2252))
* Bump vite from 5.0.11 to 5.0.12 in /playground ([#2236](https://github.com/sourcenetwork/defradb/issues/2236))
* Bump github.com/evanphx/json-patch/v5 from 5.7.0 to 5.8.1 ([#2233](https://github.com/sourcenetwork/defradb/issues/2233))

<a name="v0.9.0"></a>
## [v0.9.0](https://github.com/sourcenetwork/defradb/compare/v0.8.0...v0.9.0)

> 2024-01-18

DefraDB v0.9 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.8.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Mutation typed input ([#2167](https://github.com/sourcenetwork/defradb/issues/2167))
* Add PN Counter CRDT type ([#2119](https://github.com/sourcenetwork/defradb/issues/2119))
* Allow users to add Views ([#2114](https://github.com/sourcenetwork/defradb/issues/2114))
* Add unique secondary index ([#2131](https://github.com/sourcenetwork/defradb/issues/2131))
* New cmd for docs auto generation ([#2096](https://github.com/sourcenetwork/defradb/issues/2096))
* Add blob scalar type ([#2091](https://github.com/sourcenetwork/defradb/issues/2091))

### Fixes

* Add entropy to counter CRDT type updates ([#2186](https://github.com/sourcenetwork/defradb/issues/2186))
* Handle multiple nil values on unique indexed fields ([#2178](https://github.com/sourcenetwork/defradb/issues/2178))
* Filtering on unique index if there is no match ([#2177](https://github.com/sourcenetwork/defradb/issues/2177))

### Performance

* Switch LensVM to wasmtime runtime ([#2030](https://github.com/sourcenetwork/defradb/issues/2030))

### Refactoring

* Add strong typing to document creation ([#2161](https://github.com/sourcenetwork/defradb/issues/2161))
* Rename key,id,dockey to docID terminology ([#1749](https://github.com/sourcenetwork/defradb/issues/1749))
* Simplify Merkle CRDT workflow ([#2111](https://github.com/sourcenetwork/defradb/issues/2111))

### Testing

* Add auto-doc generation ([#2051](https://github.com/sourcenetwork/defradb/issues/2051))

### Continuous integration

* Add windows test runner ([#2033](https://github.com/sourcenetwork/defradb/issues/2033))

### Chore

* Update Lens to v0.5 ([#2083](https://github.com/sourcenetwork/defradb/issues/2083))

### Bot

* Bump [@types](https://github.com/types)/react from 18.2.47 to 18.2.48 in /playground ([#2213](https://github.com/sourcenetwork/defradb/issues/2213))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.18.0 to 6.18.1 in /playground ([#2215](https://github.com/sourcenetwork/defradb/issues/2215))
* Update dependencies (bulk dependabot PRs) 15-01-2024 ([#2217](https://github.com/sourcenetwork/defradb/issues/2217))
* Bump follow-redirects from 1.15.3 to 1.15.4 in /playground ([#2181](https://github.com/sourcenetwork/defradb/issues/2181))
* Bump github.com/getkin/kin-openapi from 0.120.0 to 0.122.0 ([#2097](https://github.com/sourcenetwork/defradb/issues/2097))
* Update dependencies (bulk dependabot PRs) 08-01-2024 ([#2173](https://github.com/sourcenetwork/defradb/issues/2173))
* Bump github.com/bits-and-blooms/bitset from 1.12.0 to 1.13.0 ([#2160](https://github.com/sourcenetwork/defradb/issues/2160))
* Bump [@types](https://github.com/types)/react from 18.2.45 to 18.2.46 in /playground ([#2159](https://github.com/sourcenetwork/defradb/issues/2159))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.15.0 to 6.16.0 in /playground ([#2156](https://github.com/sourcenetwork/defradb/issues/2156))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.15.0 to 6.16.0 in /playground ([#2155](https://github.com/sourcenetwork/defradb/issues/2155))
* Update dependencies (bulk dependabot PRs) 27-12-2023 ([#2154](https://github.com/sourcenetwork/defradb/issues/2154))
* Bump github.com/spf13/viper from 1.17.0 to 1.18.2 ([#2145](https://github.com/sourcenetwork/defradb/issues/2145))
* Bump golang.org/x/crypto from 0.16.0 to 0.17.0 ([#2144](https://github.com/sourcenetwork/defradb/issues/2144))
* Update dependencies (bulk dependabot PRs) 18-12-2023 ([#2142](https://github.com/sourcenetwork/defradb/issues/2142))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.13.2 to 6.14.0 in /playground ([#2136](https://github.com/sourcenetwork/defradb/issues/2136))
* Bump [@types](https://github.com/types)/react from 18.2.43 to 18.2.45 in /playground ([#2134](https://github.com/sourcenetwork/defradb/issues/2134))
* Bump vite from 5.0.7 to 5.0.10 in /playground ([#2135](https://github.com/sourcenetwork/defradb/issues/2135))
* Update dependencies (bulk dependabot PRs) 04-12-2023 ([#2133](https://github.com/sourcenetwork/defradb/issues/2133))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.13.1 to 6.13.2 in /playground ([#2109](https://github.com/sourcenetwork/defradb/issues/2109))
* Bump vite from 5.0.2 to 5.0.5 in /playground ([#2112](https://github.com/sourcenetwork/defradb/issues/2112))
* Bump [@types](https://github.com/types)/react from 18.2.41 to 18.2.42 in /playground ([#2108](https://github.com/sourcenetwork/defradb/issues/2108))
* Update dependencies (bulk dependabot PRs) 04-12-2023 ([#2107](https://github.com/sourcenetwork/defradb/issues/2107))
* Bump [@types](https://github.com/types)/react from 18.2.38 to 18.2.39 in /playground ([#2086](https://github.com/sourcenetwork/defradb/issues/2086))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.12.0 to 6.13.0 in /playground ([#2085](https://github.com/sourcenetwork/defradb/issues/2085))
* Update dependencies (bulk dependabot PRs) 27-11-2023 ([#2081](https://github.com/sourcenetwork/defradb/issues/2081))
* Bump swagger-ui-react from 5.10.0 to 5.10.3 in /playground ([#2067](https://github.com/sourcenetwork/defradb/issues/2067))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.11.0 to 6.12.0 in /playground ([#2068](https://github.com/sourcenetwork/defradb/issues/2068))
* Update dependencies (bulk dependabot PRs) 20-11-2023 ([#2066](https://github.com/sourcenetwork/defradb/issues/2066))

<a name="v0.8.0"></a>
## [v0.8.0](https://github.com/sourcenetwork/defradb/compare/v0.7.0...v0.8.0)

> 2023-11-14

DefraDB v0.8 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.7.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add means to fetch schema ([#2006](https://github.com/sourcenetwork/defradb/issues/2006))
* Rename Schema.SchemaID to Schema.Root ([#2005](https://github.com/sourcenetwork/defradb/issues/2005))
* Enable playground in Docker build ([#1986](https://github.com/sourcenetwork/defradb/issues/1986))
* Change GetCollectionBySchemaFoo funcs to return many ([#1984](https://github.com/sourcenetwork/defradb/issues/1984))
* Add Swagger UI to playground ([#1979](https://github.com/sourcenetwork/defradb/issues/1979))
* Add OpenAPI route ([#1960](https://github.com/sourcenetwork/defradb/issues/1960))
* Remove CollectionDescription.Schema ([#1965](https://github.com/sourcenetwork/defradb/issues/1965))
* Remove collection from patch schema ([#1957](https://github.com/sourcenetwork/defradb/issues/1957))
* Make queries utilise secondary indexes ([#1925](https://github.com/sourcenetwork/defradb/issues/1925))
* Allow setting of default schema version ([#1888](https://github.com/sourcenetwork/defradb/issues/1888))
* Add CCIP Support ([#1896](https://github.com/sourcenetwork/defradb/issues/1896))

### Fixes

* Fix test module relying on closed memory leak ([#2037](https://github.com/sourcenetwork/defradb/issues/2037))
* Make return type for FieldKind_INT an int64 ([#1982](https://github.com/sourcenetwork/defradb/issues/1982))
* Node private key requires data directory ([#1938](https://github.com/sourcenetwork/defradb/issues/1938))
* Remove collection name from schema ID generation ([#1920](https://github.com/sourcenetwork/defradb/issues/1920))
* Infinite loop when updating one-one relation ([#1915](https://github.com/sourcenetwork/defradb/issues/1915))

### Refactoring

* CRDT merge direction ([#2016](https://github.com/sourcenetwork/defradb/issues/2016))
* Reorganise collection description storage ([#1988](https://github.com/sourcenetwork/defradb/issues/1988))
* Add peerstore to multistore ([#1980](https://github.com/sourcenetwork/defradb/issues/1980))
* P2P client interface ([#1924](https://github.com/sourcenetwork/defradb/issues/1924))
* Deprecate CollectionDescription.Schema ([#1939](https://github.com/sourcenetwork/defradb/issues/1939))
* Remove net GRPC API ([#1927](https://github.com/sourcenetwork/defradb/issues/1927))
* CLI client interface ([#1839](https://github.com/sourcenetwork/defradb/issues/1839))

### Continuous integration

* Add goreleaser workflow ([#2040](https://github.com/sourcenetwork/defradb/issues/2040))
* Add mac test runner ([#2035](https://github.com/sourcenetwork/defradb/issues/2035))
* Parallelize change detector ([#1871](https://github.com/sourcenetwork/defradb/issues/1871))

### Chore

* Update dependencies ([#2044](https://github.com/sourcenetwork/defradb/issues/2044))

### Bot

* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.10.0 to 6.11.0 in /playground ([#2053](https://github.com/sourcenetwork/defradb/issues/2053))
* Update dependencies (bulk dependabot PRs) 13-11-2023 ([#2052](https://github.com/sourcenetwork/defradb/issues/2052))
* Bump axios from 1.5.1 to 1.6.1 in /playground ([#2041](https://github.com/sourcenetwork/defradb/issues/2041))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.9.1 to 6.10.0 in /playground ([#2042](https://github.com/sourcenetwork/defradb/issues/2042))
* Bump [@vitejs](https://github.com/vitejs)/plugin-react-swc from 3.4.0 to 3.4.1 in /playground ([#2022](https://github.com/sourcenetwork/defradb/issues/2022))
* Update dependencies (bulk dependabot PRs) 08-11-2023 ([#2038](https://github.com/sourcenetwork/defradb/issues/2038))
* Update dependencies (bulk dependabot PRs) 30-10-2023 ([#2015](https://github.com/sourcenetwork/defradb/issues/2015))
* Bump eslint-plugin and parser from 6.8.0 to 6.9.0 in /playground ([#2000](https://github.com/sourcenetwork/defradb/issues/2000))
* Update dependencies (bulk dependabot PRs) 16-10-2023 ([#1998](https://github.com/sourcenetwork/defradb/issues/1998))
* Update dependencies (bulk dependabot PRs) 16-10-2023 ([#1976](https://github.com/sourcenetwork/defradb/issues/1976))
* Bump golang.org/x/net from 0.16.0 to 0.17.0 ([#1961](https://github.com/sourcenetwork/defradb/issues/1961))
* Bump [@types](https://github.com/types)/react-dom from 18.2.11 to 18.2.12 in /playground ([#1952](https://github.com/sourcenetwork/defradb/issues/1952))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.7.4 to 6.7.5 in /playground ([#1953](https://github.com/sourcenetwork/defradb/issues/1953))
* Bump combined dependencies 09-10-2023 ([#1951](https://github.com/sourcenetwork/defradb/issues/1951))
* Bump [@types](https://github.com/types)/react from 18.2.24 to 18.2.25 in /playground ([#1932](https://github.com/sourcenetwork/defradb/issues/1932))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.7.3 to 6.7.4 in /playground ([#1933](https://github.com/sourcenetwork/defradb/issues/1933))
* Bump [@vitejs](https://github.com/vitejs)/plugin-react-swc from 3.3.2 to 3.4.0 in /playground ([#1904](https://github.com/sourcenetwork/defradb/issues/1904))
* Bump combined dependencies 19-09-2023 ([#1931](https://github.com/sourcenetwork/defradb/issues/1931))
* Bump graphql from 16.8.0 to 16.8.1 in /playground ([#1901](https://github.com/sourcenetwork/defradb/issues/1901))
* Update combined dependabot PRs 19-09-2023 ([#1898](https://github.com/sourcenetwork/defradb/issues/1898))

<a name="v0.7.0"></a>
## [v0.7.0](https://github.com/sourcenetwork/defradb/compare/v0.6.0...v0.7.0)

> 2023-09-18

DefraDB v0.7 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

This release has focused on robustness, testing, and schema management. Some highlight new features include notable expansions to the expressiveness of schema migrations.

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.6.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Allow field indexing by name in PatchSchema ([#1810](https://github.com/sourcenetwork/defradb/issues/1810))
* Auto-create relation id fields via PatchSchema ([#1807](https://github.com/sourcenetwork/defradb/issues/1807))
* Support PatchSchema relational field kind substitution ([#1777](https://github.com/sourcenetwork/defradb/issues/1777))
* Add support for adding of relational fields ([#1766](https://github.com/sourcenetwork/defradb/issues/1766))
* Enable downgrading of documents via Lens inverses ([#1721](https://github.com/sourcenetwork/defradb/issues/1721))

### Fixes

* Correctly handle serialisation of nil field values ([#1872](https://github.com/sourcenetwork/defradb/issues/1872))
* Compound filter operators with relations ([#1855](https://github.com/sourcenetwork/defradb/issues/1855))
* Only update updated fields via update requests ([#1817](https://github.com/sourcenetwork/defradb/issues/1817))
* Error when saving a deleted document ([#1806](https://github.com/sourcenetwork/defradb/issues/1806))
* Prevent multiple docs from being linked in one one ([#1790](https://github.com/sourcenetwork/defradb/issues/1790))
* Handle the querying of secondary relation id fields ([#1768](https://github.com/sourcenetwork/defradb/issues/1768))
* Improve the way migrations handle transactions ([#1737](https://github.com/sourcenetwork/defradb/issues/1737))

### Tooling

* Add Akash deployment configuration ([#1736](https://github.com/sourcenetwork/defradb/issues/1736))

### Refactoring

* HTTP client interface ([#1776](https://github.com/sourcenetwork/defradb/issues/1776))
* Simplify fetcher interface ([#1746](https://github.com/sourcenetwork/defradb/issues/1746))

### Testing

* Convert and move out of place explain tests ([#1878](https://github.com/sourcenetwork/defradb/issues/1878))
* Update mutation tests to make use of mutation system ([#1853](https://github.com/sourcenetwork/defradb/issues/1853))
* Test top level agg. with compound relational filter ([#1870](https://github.com/sourcenetwork/defradb/issues/1870))
* Skip unsupported mutation types at test level ([#1850](https://github.com/sourcenetwork/defradb/issues/1850))
* Extend mutation tests with col.Update and Create ([#1838](https://github.com/sourcenetwork/defradb/issues/1838))
* Add tests for multiple one-one joins ([#1793](https://github.com/sourcenetwork/defradb/issues/1793))

### Chore

* Update Badger version to v4 ([#1740](https://github.com/sourcenetwork/defradb/issues/1740))
* Update go-libp2p to 0.29.2 ([#1780](https://github.com/sourcenetwork/defradb/issues/1780))
* Bump golangci-lint to v1.54 ([#1881](https://github.com/sourcenetwork/defradb/issues/1881))
* Bump go.opentelemetry.io/otel/metric from 1.17.0 to 1.18.0 ([#1890](https://github.com/sourcenetwork/defradb/issues/1890))
* Bump [@tanstack](https://github.com/tanstack)/react-query from 4.35.0 to 4.35.3 in /playground ([#1876](https://github.com/sourcenetwork/defradb/issues/1876))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.5.0 to 6.7.0 in /playground ([#1874](https://github.com/sourcenetwork/defradb/issues/1874))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.6.0 to 6.7.0 in /playground ([#1875](https://github.com/sourcenetwork/defradb/issues/1875))
* Combined PRs 2023-09-14 ([#1873](https://github.com/sourcenetwork/defradb/issues/1873))
* Bump [@typescript](https://github.com/typescript)-eslint/eslint-plugin from 6.4.0 to 6.5.0 in /playground ([#1827](https://github.com/sourcenetwork/defradb/issues/1827))
* Bump go.opentelemetry.io/otel/sdk/metric from 0.39.0 to 0.40.0 ([#1829](https://github.com/sourcenetwork/defradb/issues/1829))
* Bump github.com/ipfs/go-block-format from 0.1.2 to 0.2.0 ([#1819](https://github.com/sourcenetwork/defradb/issues/1819))
* Combined PRs ([#1826](https://github.com/sourcenetwork/defradb/issues/1826))
* Bump [@typescript](https://github.com/typescript)-eslint/parser from 6.4.0 to 6.4.1 in /playground ([#1804](https://github.com/sourcenetwork/defradb/issues/1804))
* Combined PRs ([#1803](https://github.com/sourcenetwork/defradb/issues/1803))
* Combined PRs ([#1791](https://github.com/sourcenetwork/defradb/issues/1791))
* Combined PRs ([#1778](https://github.com/sourcenetwork/defradb/issues/1778))
* Bump dependencies ([#1761](https://github.com/sourcenetwork/defradb/issues/1761))
* Bump vite from 4.3.9 to 4.4.8 in /playground ([#1748](https://github.com/sourcenetwork/defradb/issues/1748))
* Bump graphiql from 3.0.4 to 3.0.5 in /playground ([#1730](https://github.com/sourcenetwork/defradb/issues/1730))
* Combined bumps of dependencies under /playground ([#1744](https://github.com/sourcenetwork/defradb/issues/1744))
* Bump github.com/ipfs/boxo from 0.10.2 to 0.11.0 ([#1726](https://github.com/sourcenetwork/defradb/issues/1726))
* Bump github.com/libp2p/go-libp2p-kad-dht from 0.24.2 to 0.24.3 ([#1724](https://github.com/sourcenetwork/defradb/issues/1724))
* Bump google.golang.org/grpc from 1.56.2 to 1.57.0 ([#1725](https://github.com/sourcenetwork/defradb/issues/1725))

<a name="v0.6.0"></a>
## [v0.6.0](https://github.com/sourcenetwork/defradb/compare/v0.5.1...v0.6.0)

> 2023-07-31

DefraDB v0.6 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

There are several new and powerful features, important bug fixes, and notable refactors in this release. Some highlight features include: The initial release of our LensVM based schema migration engine powered by WebAssembly ([#1650](https://github.com/sourcenetwork/defradb/issues/1650)), newly embedded DefraDB Playround which includes a bundled GraphQL client and schema manager, and last but not least a relation field (<type>_id) alias to improve the developer experience ([#1609](https://github.com/sourcenetwork/defradb/issues/1609)).

To get a full outline of the changes, we invite you to review the official changelog below. This release does include a Breaking Change to existing v0.5.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add `_not` operator ([#1631](https://github.com/sourcenetwork/defradb/issues/1631))
* Schema list API ([#1625](https://github.com/sourcenetwork/defradb/issues/1625))
* Add simple data import and export ([#1630](https://github.com/sourcenetwork/defradb/issues/1630))
* Playground ([#1575](https://github.com/sourcenetwork/defradb/issues/1575))
* Add schema migration get and set cmds to CLI ([#1650](https://github.com/sourcenetwork/defradb/issues/1650))
* Allow relation alias on create and update ([#1609](https://github.com/sourcenetwork/defradb/issues/1609))
* Make fetcher calculate docFetches and fieldFetches ([#1713](https://github.com/sourcenetwork/defradb/issues/1713))
* Add lens migration engine to defra ([#1564](https://github.com/sourcenetwork/defradb/issues/1564))
* Add `_keys` attribute to `selectNode` simple explain ([#1546](https://github.com/sourcenetwork/defradb/issues/1546))
* CLI commands for secondary indexes ([#1595](https://github.com/sourcenetwork/defradb/issues/1595))
* Add alias to `groupBy` related object ([#1579](https://github.com/sourcenetwork/defradb/issues/1579))
* Non-unique secondary index (no querying) ([#1450](https://github.com/sourcenetwork/defradb/issues/1450))
* Add ability to explain-debug all nodes ([#1563](https://github.com/sourcenetwork/defradb/issues/1563))
* Include dockey in doc exists err ([#1558](https://github.com/sourcenetwork/defradb/issues/1558))

### Fixes

* Better wait in CLI integration test ([#1415](https://github.com/sourcenetwork/defradb/issues/1415))
* Return error when relation is not defined on both types ([#1647](https://github.com/sourcenetwork/defradb/issues/1647))
* Change `core.DocumentMapping` to pointer ([#1528](https://github.com/sourcenetwork/defradb/issues/1528))
* Fix invalid (badger) datastore state ([#1685](https://github.com/sourcenetwork/defradb/issues/1685))
* Discard index and subscription implicit transactions ([#1715](https://github.com/sourcenetwork/defradb/issues/1715))
* Remove duplicated `peers` in peerstore prefix ([#1678](https://github.com/sourcenetwork/defradb/issues/1678))
* Return errors from typeJoinOne ([#1716](https://github.com/sourcenetwork/defradb/issues/1716))
* Document change detector breaking change ([#1531](https://github.com/sourcenetwork/defradb/issues/1531))
* Standardise `schema migration` CLI errors ([#1682](https://github.com/sourcenetwork/defradb/issues/1682))
* Introspection OrderArg returns null inputFields ([#1633](https://github.com/sourcenetwork/defradb/issues/1633))
* Avoid duplicated requestable fields ([#1621](https://github.com/sourcenetwork/defradb/issues/1621))
* Normalize int field kind ([#1619](https://github.com/sourcenetwork/defradb/issues/1619))
* Change the WriteSyncer to use lock when piping ([#1608](https://github.com/sourcenetwork/defradb/issues/1608))
* Filter splitting and rendering for related types ([#1541](https://github.com/sourcenetwork/defradb/issues/1541))

### Documentation

* Improve CLI command documentation ([#1505](https://github.com/sourcenetwork/defradb/issues/1505))

### Refactoring

* Schema list output to include schemaVersionID ([#1706](https://github.com/sourcenetwork/defradb/issues/1706))
* Reuse lens wasm modules ([#1641](https://github.com/sourcenetwork/defradb/issues/1641))
* Remove redundant txn param from fetcher start ([#1635](https://github.com/sourcenetwork/defradb/issues/1635))
* Remove first CRDT byte from field encoded values ([#1622](https://github.com/sourcenetwork/defradb/issues/1622))
* Merge `node` into `net` and improve coverage ([#1593](https://github.com/sourcenetwork/defradb/issues/1593))
* Fetcher filter and field optimization ([#1500](https://github.com/sourcenetwork/defradb/issues/1500))

### Testing

* Rework transaction test framework capabilities ([#1603](https://github.com/sourcenetwork/defradb/issues/1603))
* Expand backup integration tests ([#1699](https://github.com/sourcenetwork/defradb/issues/1699))
* Disable test ([#1675](https://github.com/sourcenetwork/defradb/issues/1675))
* Add tests for 1-1 group by id ([#1655](https://github.com/sourcenetwork/defradb/issues/1655))
* Remove CLI tests from make test ([#1643](https://github.com/sourcenetwork/defradb/issues/1643))
* Bundle test state into single var ([#1645](https://github.com/sourcenetwork/defradb/issues/1645))
* Convert explain group tests to new explain setup ([#1537](https://github.com/sourcenetwork/defradb/issues/1537))
* Add tests for foo_id field name clashes ([#1521](https://github.com/sourcenetwork/defradb/issues/1521))
* Resume wait correctly following test node restart ([#1515](https://github.com/sourcenetwork/defradb/issues/1515))
* Require no errors when none expected ([#1509](https://github.com/sourcenetwork/defradb/issues/1509))

### Continuous integration

* Add workflows to push, pull, and validate docker images ([#1676](https://github.com/sourcenetwork/defradb/issues/1676))
* Build mocks using make ([#1612](https://github.com/sourcenetwork/defradb/issues/1612))
* Fix terraform plan and merge AMI build + deploy workflow ([#1514](https://github.com/sourcenetwork/defradb/issues/1514))
* Reconfigure CodeCov action to ensure stability ([#1414](https://github.com/sourcenetwork/defradb/issues/1414))

### Chore

* Bump to GoLang v1.20 ([#1689](https://github.com/sourcenetwork/defradb/issues/1689))
* Update to ipfs boxo 0.10.0 ([#1573](https://github.com/sourcenetwork/defradb/issues/1573))


<a name="v0.5.1"></a>
## [v0.5.1](https://github.com/sourcenetwork/defradb/compare/v0.5.0...v0.5.1)

> 2023-05-16

### Features

* Add collection response information on creation ([#1499](https://github.com/sourcenetwork/defradb/issues/1499))
* CLI client request from file ([#1503](https://github.com/sourcenetwork/defradb/issues/1503))
* Add commits fieldName and fieldId fields ([#1451](https://github.com/sourcenetwork/defradb/issues/1451))
* Add allowed origins config ([#1408](https://github.com/sourcenetwork/defradb/issues/1408))
* Add descriptions to all system defined GQL stuff ([#1387](https://github.com/sourcenetwork/defradb/issues/1387))
* Strongly type Request.Errors ([#1364](https://github.com/sourcenetwork/defradb/issues/1364))

### Fixes

* Skip new test packages in change detector ([#1495](https://github.com/sourcenetwork/defradb/issues/1495))
* Make nested joins work correctly from primary direction ([#1491](https://github.com/sourcenetwork/defradb/issues/1491))
* Add reconnection to known peers ([#1482](https://github.com/sourcenetwork/defradb/issues/1482))
* Rename commit field input arg to fieldId ([#1460](https://github.com/sourcenetwork/defradb/issues/1460))
* Reference collectionID in p2p readme ([#1466](https://github.com/sourcenetwork/defradb/issues/1466))
* Handling SIGTERM in CLI `start` command ([#1459](https://github.com/sourcenetwork/defradb/issues/1459))
* Update QL documentation link and replicator command ([#1440](https://github.com/sourcenetwork/defradb/issues/1440))
* Fix typo in readme ([#1419](https://github.com/sourcenetwork/defradb/issues/1419))
* Limit the size of http request bodies that we handle ([#1405](https://github.com/sourcenetwork/defradb/issues/1405))
* Improve P2P event handling ([#1388](https://github.com/sourcenetwork/defradb/issues/1388))
* Serialize DB errors to json in http package ([#1401](https://github.com/sourcenetwork/defradb/issues/1401))
* Do not commit if errors have been returned ([#1390](https://github.com/sourcenetwork/defradb/issues/1390))
* Unlock replicator lock before returning error ([#1369](https://github.com/sourcenetwork/defradb/issues/1369))
* Improve NonNull error message ([#1362](https://github.com/sourcenetwork/defradb/issues/1362))
* Use ring-buffer for WaitForFoo chans ([#1359](https://github.com/sourcenetwork/defradb/issues/1359))
* Guarantee event processing order ([#1352](https://github.com/sourcenetwork/defradb/issues/1352))
* Explain of _group with dockeys filter to be []string ([#1348](https://github.com/sourcenetwork/defradb/issues/1348))

### Refactoring

* Use `int32` for proper gql scalar Int parsing ([#1493](https://github.com/sourcenetwork/defradb/issues/1493))
* Improve rollback on peer P2P collection error ([#1461](https://github.com/sourcenetwork/defradb/issues/1461))
* Improve CLI with test suite and builder pattern ([#928](https://github.com/sourcenetwork/defradb/issues/928))

### Testing

* Add DB/Node Restart tests ([#1504](https://github.com/sourcenetwork/defradb/issues/1504))
* Provide tests for client introspection query ([#1492](https://github.com/sourcenetwork/defradb/issues/1492))
* Convert explain count tests to new explain setup ([#1488](https://github.com/sourcenetwork/defradb/issues/1488))
* Convert explain sum tests to new explain setup ([#1489](https://github.com/sourcenetwork/defradb/issues/1489))
* Convert explain average tests to new explain setup ([#1487](https://github.com/sourcenetwork/defradb/issues/1487))
* Convert explain top-level tests to new explain setup ([#1480](https://github.com/sourcenetwork/defradb/issues/1480))
* Convert explain order tests to new explain setup ([#1478](https://github.com/sourcenetwork/defradb/issues/1478))
* Convert explain join tests to new explain setup ([#1476](https://github.com/sourcenetwork/defradb/issues/1476))
* Convert explain dagscan tests to new explain setup ([#1474](https://github.com/sourcenetwork/defradb/issues/1474))
* Add tests to assert schema id order independence ([#1456](https://github.com/sourcenetwork/defradb/issues/1456))
* Capitalize all integration schema types ([#1445](https://github.com/sourcenetwork/defradb/issues/1445))
* Convert explain limit tests to new explain setup ([#1446](https://github.com/sourcenetwork/defradb/issues/1446))
* Improve change detector performance ([#1433](https://github.com/sourcenetwork/defradb/issues/1433))
* Convert mutation explain tests to new explain setup ([#1416](https://github.com/sourcenetwork/defradb/issues/1416))
* Convert filter explain tests to new explain setup ([#1380](https://github.com/sourcenetwork/defradb/issues/1380))
* Retry test doc mutation on transaction conflict ([#1366](https://github.com/sourcenetwork/defradb/issues/1366))

### Continuous integration

* Remove secret ssh key stuff from change detector wf ([#1438](https://github.com/sourcenetwork/defradb/issues/1438))
* Fix the SSH security issue from AMI scan report ([#1426](https://github.com/sourcenetwork/defradb/issues/1426))
* Add a separate workflow to run the linter ([#1434](https://github.com/sourcenetwork/defradb/issues/1434))
* Allow CI to work from forked repo ([#1392](https://github.com/sourcenetwork/defradb/issues/1392))
* Bump go version within packer for AWS AMI ([#1344](https://github.com/sourcenetwork/defradb/issues/1344))

### Chore

* Enshrine defra logger names ([#1410](https://github.com/sourcenetwork/defradb/issues/1410))
* Remove some dead code ([#1470](https://github.com/sourcenetwork/defradb/issues/1470))
* Update graphql-go ([#1422](https://github.com/sourcenetwork/defradb/issues/1422))
* Improve logging consistency ([#1424](https://github.com/sourcenetwork/defradb/issues/1424))
* Makefile tests with shorter timeout and common flags ([#1397](https://github.com/sourcenetwork/defradb/issues/1397))
* Move to gofrs/uuid ([#1396](https://github.com/sourcenetwork/defradb/issues/1396))
* Move to ipfs boxo ([#1393](https://github.com/sourcenetwork/defradb/issues/1393))
* Document collection.txn ([#1363](https://github.com/sourcenetwork/defradb/issues/1363))

### Bot

* Bump golang.org/x/crypto from 0.8.0 to 0.9.0 ([#1497](https://github.com/sourcenetwork/defradb/issues/1497))
* Bump golang.org/x/net from 0.9.0 to 0.10.0 ([#1496](https://github.com/sourcenetwork/defradb/issues/1496))
* Bump google.golang.org/grpc from 1.54.0 to 1.55.0 ([#1464](https://github.com/sourcenetwork/defradb/issues/1464))
* Bump github.com/ipfs/boxo from 0.8.0 to 0.8.1 ([#1427](https://github.com/sourcenetwork/defradb/issues/1427))
* Bump golang.org/x/crypto from 0.7.0 to 0.8.0 ([#1398](https://github.com/sourcenetwork/defradb/issues/1398))
* Bump github.com/spf13/cobra from 1.6.1 to 1.7.0 ([#1399](https://github.com/sourcenetwork/defradb/issues/1399))
* Bump github.com/ipfs/go-blockservice from 0.5.0 to 0.5.1 ([#1300](https://github.com/sourcenetwork/defradb/issues/1300))
* Bump github.com/ipfs/go-cid from 0.4.0 to 0.4.1 ([#1301](https://github.com/sourcenetwork/defradb/issues/1301))

<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/sourcenetwork/defradb/compare/v0.4.0...v0.5.0)

> 2023-04-12

DefraDB v0.5 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

There many new features in this release, but most importantly, this is the first open source release for DefraDB. As such, this release focused on various quality of life changes and refactors, bug fixes, and overall cleanliness of the repo so it can effectively be used and tested in the public domain.

To get a full outline of the changes, we invite you to review the official changelog below. Some highlights are the first iteration of our schema update system, allowing developers to add new fields to schemas using our JSON Patch based DDL, a new DAG based delete system which will persist "soft-delete" ops into the CRDT Merkle DAG, and a early prototype for our collection level peer-to-peer synchronization.

This release does include a Breaking Change to existing v0.4.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add document delete mechanics ([#1263](https://github.com/sourcenetwork/defradb/issues/1263))
* Ability to explain an executed request ([#1188](https://github.com/sourcenetwork/defradb/issues/1188))
* Add SchemaPatch CLI command ([#1250](https://github.com/sourcenetwork/defradb/issues/1250))
* Add support for one-one mutation from sec. side ([#1247](https://github.com/sourcenetwork/defradb/issues/1247))
* Store only key in DAG instead of dockey path ([#1245](https://github.com/sourcenetwork/defradb/issues/1245))
* Add collectionId field to commit field ([#1235](https://github.com/sourcenetwork/defradb/issues/1235))
* Add field kind substitution for PatchSchema ([#1223](https://github.com/sourcenetwork/defradb/issues/1223))
* Add dockey field for commit field ([#1216](https://github.com/sourcenetwork/defradb/issues/1216))
* Allow new fields to be added locally to schema ([#1139](https://github.com/sourcenetwork/defradb/issues/1139))
* Add `like` sub-string filter ([#1091](https://github.com/sourcenetwork/defradb/issues/1091))
* Add ability for P2P to wait for pushlog by peer ([#1098](https://github.com/sourcenetwork/defradb/issues/1098))
* Add P2P collection topic subscription ([#1086](https://github.com/sourcenetwork/defradb/issues/1086))
* Add support for schema version id in queries ([#1067](https://github.com/sourcenetwork/defradb/issues/1067))
* Add schema version id to commit queries ([#1061](https://github.com/sourcenetwork/defradb/issues/1061))
* Persist schema version at time of commit ([#1055](https://github.com/sourcenetwork/defradb/issues/1055))
* Add ability to input simple explain type arg ([#1039](https://github.com/sourcenetwork/defradb/issues/1039))

### Fixes

* API address parameter validation ([#1311](https://github.com/sourcenetwork/defradb/issues/1311))
* Improve error message for NonNull GQL types ([#1333](https://github.com/sourcenetwork/defradb/issues/1333))
* Handle panics in the rpc server ([#1330](https://github.com/sourcenetwork/defradb/issues/1330))
* Handle returned error in select.go ([#1329](https://github.com/sourcenetwork/defradb/issues/1329))
* Resolve handful of CLI issues ([#1318](https://github.com/sourcenetwork/defradb/issues/1318))
* Only check for events queue on subscription request ([#1326](https://github.com/sourcenetwork/defradb/issues/1326))
* Remove client Create/UpdateCollection ([#1309](https://github.com/sourcenetwork/defradb/issues/1309))
* CLI to display specific command usage help ([#1314](https://github.com/sourcenetwork/defradb/issues/1314))
* Fix P2P collection CLI commands ([#1295](https://github.com/sourcenetwork/defradb/issues/1295))
* Dont double up badger file path ([#1299](https://github.com/sourcenetwork/defradb/issues/1299))
* Update immutable package ([#1290](https://github.com/sourcenetwork/defradb/issues/1290))
* Fix panic on success of Add/RemoveP2PCollections ([#1297](https://github.com/sourcenetwork/defradb/issues/1297))
* Fix deadlock on memory-datastore Close ([#1273](https://github.com/sourcenetwork/defradb/issues/1273))
* Determine if query is introspection query ([#1255](https://github.com/sourcenetwork/defradb/issues/1255))
* Allow newly added fields to sync via p2p ([#1226](https://github.com/sourcenetwork/defradb/issues/1226))
* Expose `ExplainEnum` in the GQL schema ([#1204](https://github.com/sourcenetwork/defradb/issues/1204))
* Resolve aggregates' mapping with deep nested subtypes ([#1175](https://github.com/sourcenetwork/defradb/issues/1175))
* Make sort stable and handle nil comparison ([#1094](https://github.com/sourcenetwork/defradb/issues/1094))
* Change successful schema add status to 200 ([#1106](https://github.com/sourcenetwork/defradb/issues/1106))
* Add delay in P2P test util execution ([#1093](https://github.com/sourcenetwork/defradb/issues/1093))
* Ensure errors test don't hard expect folder name ([#1072](https://github.com/sourcenetwork/defradb/issues/1072))
* Remove potential P2P deadlock ([#1056](https://github.com/sourcenetwork/defradb/issues/1056))
* Rework the P2P integration tests ([#989](https://github.com/sourcenetwork/defradb/issues/989))
* Improve DAG sync with highly concurrent updates ([#1031](https://github.com/sourcenetwork/defradb/issues/1031))

### Documentation

* Update docs for the v0.5 release ([#1320](https://github.com/sourcenetwork/defradb/issues/1320))
* Document client interfaces in client/db.go ([#1305](https://github.com/sourcenetwork/defradb/issues/1305))
* Document client Description types ([#1307](https://github.com/sourcenetwork/defradb/issues/1307))
* Improve security policy ([#1240](https://github.com/sourcenetwork/defradb/issues/1240))
* Add security disclosure policy ([#1194](https://github.com/sourcenetwork/defradb/issues/1194))
* Correct commits query example in readme ([#1172](https://github.com/sourcenetwork/defradb/issues/1172))

### Refactoring

* Improve p2p collection operations on peer ([#1286](https://github.com/sourcenetwork/defradb/issues/1286))
* Migrate gql introspection tests to new framework ([#1211](https://github.com/sourcenetwork/defradb/issues/1211))
* Reorganise client transaction related interfaces ([#1180](https://github.com/sourcenetwork/defradb/issues/1180))
* Config-local viper, rootdir, and logger parsing ([#1132](https://github.com/sourcenetwork/defradb/issues/1132))
* Migrate mutation-relation tests to new framework ([#1109](https://github.com/sourcenetwork/defradb/issues/1109))
* Rework integration test framework ([#1089](https://github.com/sourcenetwork/defradb/issues/1089))
* Generate gql types using col. desc ([#1080](https://github.com/sourcenetwork/defradb/issues/1080))
* Extract config errors to dedicated file ([#1107](https://github.com/sourcenetwork/defradb/issues/1107))
* Change terminology from query to request ([#1054](https://github.com/sourcenetwork/defradb/issues/1054))
* Allow db keys to handle multiple schema versions ([#1026](https://github.com/sourcenetwork/defradb/issues/1026))
* Extract query schema errors to dedicated file ([#1037](https://github.com/sourcenetwork/defradb/issues/1037))
* Extract planner errors to dedicated file ([#1034](https://github.com/sourcenetwork/defradb/issues/1034))
* Extract query parser errors to dedicated file ([#1035](https://github.com/sourcenetwork/defradb/issues/1035))

### Testing

* Remove test reference to DEFRA_ROOTDIR env var ([#1328](https://github.com/sourcenetwork/defradb/issues/1328))
* Expand tests for Peer subscribe actions ([#1287](https://github.com/sourcenetwork/defradb/issues/1287))
* Fix flaky TestCloseThroughContext test ([#1265](https://github.com/sourcenetwork/defradb/issues/1265))
* Add gql introspection tests for patch schema ([#1219](https://github.com/sourcenetwork/defradb/issues/1219))
* Explicitly state change detector split for test ([#1228](https://github.com/sourcenetwork/defradb/issues/1228))
* Add test for successful one-one create mutation ([#1215](https://github.com/sourcenetwork/defradb/issues/1215))
* Ensure that all databases are always closed on exit ([#1187](https://github.com/sourcenetwork/defradb/issues/1187))
* Add P2P tests for Schema Update adding field ([#1182](https://github.com/sourcenetwork/defradb/issues/1182))
* Migrate P2P/state tests to new framework ([#1160](https://github.com/sourcenetwork/defradb/issues/1160))
* Remove sleep from subscription tests ([#1156](https://github.com/sourcenetwork/defradb/issues/1156))
* Fetch documents on test execution start ([#1163](https://github.com/sourcenetwork/defradb/issues/1163))
* Introduce basic testing for the `version` module ([#1111](https://github.com/sourcenetwork/defradb/issues/1111))
* Boost test coverage for collection_update ([#1050](https://github.com/sourcenetwork/defradb/issues/1050))
* Wait between P2P update retry attempts ([#1052](https://github.com/sourcenetwork/defradb/issues/1052))
* Exclude auto-generated protobuf files from codecov ([#1048](https://github.com/sourcenetwork/defradb/issues/1048))
* Add P2P tests for relational docs ([#1042](https://github.com/sourcenetwork/defradb/issues/1042))

### Continuous integration

* Add workflow that builds DefraDB AMI upon tag push ([#1304](https://github.com/sourcenetwork/defradb/issues/1304))
* Allow PR title to end with a capital letter ([#1291](https://github.com/sourcenetwork/defradb/issues/1291))
* Changes for `dependabot` to be well-behaved ([#1165](https://github.com/sourcenetwork/defradb/issues/1165))
* Skip benchmarks for dependabot ([#1144](https://github.com/sourcenetwork/defradb/issues/1144))
* Add workflow to ensure deps build properly ([#1078](https://github.com/sourcenetwork/defradb/issues/1078))
* Runner and Builder Containerfiles ([#951](https://github.com/sourcenetwork/defradb/issues/951))
* Fix go-header linter rule to be any year ([#1021](https://github.com/sourcenetwork/defradb/issues/1021))

### Chore

* Add Islam as contributor ([#1302](https://github.com/sourcenetwork/defradb/issues/1302))
* Update go-libp2p to 0.26.4 ([#1257](https://github.com/sourcenetwork/defradb/issues/1257))
* Improve the test coverage of datastore ([#1203](https://github.com/sourcenetwork/defradb/issues/1203))
* Add issue and discussion templates ([#1193](https://github.com/sourcenetwork/defradb/issues/1193))
* Bump libp2p/go-libp2p-kad-dht from 0.21.0 to 0.21.1 ([#1146](https://github.com/sourcenetwork/defradb/issues/1146))
* Enable dependabot ([#1120](https://github.com/sourcenetwork/defradb/issues/1120))
* Update `opentelemetry` dependencies ([#1114](https://github.com/sourcenetwork/defradb/issues/1114))
* Update dependencies including go-ipfs ([#1112](https://github.com/sourcenetwork/defradb/issues/1112))
* Bump to GoLang v1.19 ([#818](https://github.com/sourcenetwork/defradb/issues/818))
* Remove versionedScan node ([#1049](https://github.com/sourcenetwork/defradb/issues/1049))

### Bot

* Bump github.com/multiformats/go-multiaddr from 0.8.0 to 0.9.0 ([#1277](https://github.com/sourcenetwork/defradb/issues/1277))
* Bump google.golang.org/grpc from 1.53.0 to 1.54.0 ([#1233](https://github.com/sourcenetwork/defradb/issues/1233))
* Bump github.com/multiformats/go-multibase from 0.1.1 to 0.2.0 ([#1230](https://github.com/sourcenetwork/defradb/issues/1230))
* Bump github.com/ipfs/go-libipfs from 0.6.2 to 0.7.0 ([#1231](https://github.com/sourcenetwork/defradb/issues/1231))
* Bump github.com/ipfs/go-cid from 0.3.2 to 0.4.0 ([#1200](https://github.com/sourcenetwork/defradb/issues/1200))
* Bump github.com/ipfs/go-ipfs-blockstore from 1.2.0 to 1.3.0 ([#1199](https://github.com/sourcenetwork/defradb/issues/1199))
* Bump github.com/stretchr/testify from 1.8.1 to 1.8.2 ([#1198](https://github.com/sourcenetwork/defradb/issues/1198))
* Bump github.com/ipfs/go-libipfs from 0.6.1 to 0.6.2 ([#1201](https://github.com/sourcenetwork/defradb/issues/1201))
* Bump golang.org/x/crypto from 0.6.0 to 0.7.0 ([#1197](https://github.com/sourcenetwork/defradb/issues/1197))
* Bump libp2p/go-libp2p-gostream from 0.5.0 to 0.6.0 ([#1152](https://github.com/sourcenetwork/defradb/issues/1152))
* Bump github.com/ipfs/go-libipfs from 0.5.0 to 0.6.1 ([#1166](https://github.com/sourcenetwork/defradb/issues/1166))
* Bump github.com/ugorji/go/codec from 1.2.9 to 1.2.11 ([#1173](https://github.com/sourcenetwork/defradb/issues/1173))
* Bump github.com/libp2p/go-libp2p-pubsub from 0.9.0 to 0.9.3 ([#1183](https://github.com/sourcenetwork/defradb/issues/1183))

<a name="v0.4.0"></a>
## [v0.4.0](https://github.com/sourcenetwork/defradb/compare/v0.3.1...v0.4.0)

> 2023-12-23

DefraDB v0.4 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

There are various new features in this release - some of which are breaking - and we invite you to review the official changelog below. Some highlights are persistence of replicators, DateTime scalars, TLS support, and GQL subscriptions.

This release does include a Breaking Change to existing v0.3.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add basic metric functionality ([#971](https://github.com/sourcenetwork/defradb/issues/971))
* Add thread safe transactional in-memory datastore ([#947](https://github.com/sourcenetwork/defradb/issues/947))
* Persist p2p replicators ([#960](https://github.com/sourcenetwork/defradb/issues/960))
* Add DateTime custom scalars ([#931](https://github.com/sourcenetwork/defradb/issues/931))
* Add GraphQL subscriptions ([#934](https://github.com/sourcenetwork/defradb/issues/934))
* Add support for tls ([#885](https://github.com/sourcenetwork/defradb/issues/885))
* Add group by support for commits ([#887](https://github.com/sourcenetwork/defradb/issues/887))
* Add depth support for commits ([#889](https://github.com/sourcenetwork/defradb/issues/889))
* Make dockey optional for allCommits queries ([#847](https://github.com/sourcenetwork/defradb/issues/847))
* Add WithStack to the errors package ([#870](https://github.com/sourcenetwork/defradb/issues/870))
* Add event system ([#834](https://github.com/sourcenetwork/defradb/issues/834))

### Fixes

* Correct errors.WithStack behaviour ([#984](https://github.com/sourcenetwork/defradb/issues/984))
* Correctly handle nested one to one joins ([#964](https://github.com/sourcenetwork/defradb/issues/964))
* Do not assume parent record exists when joining ([#963](https://github.com/sourcenetwork/defradb/issues/963))
* Change time format for HTTP API log ([#910](https://github.com/sourcenetwork/defradb/issues/910))
* Error if group select contains non-group-by fields ([#898](https://github.com/sourcenetwork/defradb/issues/898))
* Add inspection of values for ENV flags ([#900](https://github.com/sourcenetwork/defradb/issues/900))
* Remove panics from document ([#881](https://github.com/sourcenetwork/defradb/issues/881))
* Add __typename support ([#871](https://github.com/sourcenetwork/defradb/issues/871))
* Handle subscriber close ([#877](https://github.com/sourcenetwork/defradb/issues/877))
* Publish update events post commit ([#866](https://github.com/sourcenetwork/defradb/issues/866))

### Refactoring

* Make rootstore require Batching and TxnDatastore ([#940](https://github.com/sourcenetwork/defradb/issues/940))
* Conceptually clarify schema vs query-language ([#924](https://github.com/sourcenetwork/defradb/issues/924))
* Decouple db.db from gql ([#912](https://github.com/sourcenetwork/defradb/issues/912))
* Merkle clock heads cleanup ([#918](https://github.com/sourcenetwork/defradb/issues/918))
* Simplify dag fetcher ([#913](https://github.com/sourcenetwork/defradb/issues/913))
* Cleanup parsing logic ([#909](https://github.com/sourcenetwork/defradb/issues/909))
* Move planner outside the gql directory ([#907](https://github.com/sourcenetwork/defradb/issues/907))
* Refactor commit nodes ([#892](https://github.com/sourcenetwork/defradb/issues/892))
* Make latest commits syntax sugar ([#890](https://github.com/sourcenetwork/defradb/issues/890))
* Remove commit query ([#841](https://github.com/sourcenetwork/defradb/issues/841))

### Testing

* Add event tests ([#965](https://github.com/sourcenetwork/defradb/issues/965))
* Add new setup for testing explain functionality ([#949](https://github.com/sourcenetwork/defradb/issues/949))
* Add txn relation-type delete and create tests ([#875](https://github.com/sourcenetwork/defradb/issues/875))
* Skip change detection for tests that assert panic ([#883](https://github.com/sourcenetwork/defradb/issues/883))

### Continuous integration

* Bump all gh-action versions to support node16 ([#990](https://github.com/sourcenetwork/defradb/issues/990))
* Bump ssh-agent action to v0.7.0 ([#978](https://github.com/sourcenetwork/defradb/issues/978))
* Add error message format check ([#901](https://github.com/sourcenetwork/defradb/issues/901))

### Chore

* Extract (events, merkle) errors to errors.go ([#973](https://github.com/sourcenetwork/defradb/issues/973))
* Extract (datastore, db) errors to errors.go ([#969](https://github.com/sourcenetwork/defradb/issues/969))
* Extract (connor, crdt, core) errors to errors.go ([#968](https://github.com/sourcenetwork/defradb/issues/968))
* Extract inline (http and client) errors to errors.go ([#967](https://github.com/sourcenetwork/defradb/issues/967))
* Update badger version ([#966](https://github.com/sourcenetwork/defradb/issues/966))
* Move Option and Enumerable to immutables ([#939](https://github.com/sourcenetwork/defradb/issues/939))
* Add configuration of external loggers ([#942](https://github.com/sourcenetwork/defradb/issues/942))
* Strip DSKey prefixes and simplify NewDataStoreKey ([#944](https://github.com/sourcenetwork/defradb/issues/944))
* Include version metadata in cross-building ([#930](https://github.com/sourcenetwork/defradb/issues/930))
* Update to v0.23.2 the libP2P package ([#908](https://github.com/sourcenetwork/defradb/issues/908))
* Remove `ipfslite` dependency ([#739](https://github.com/sourcenetwork/defradb/issues/739))


<a name="v0.3.1"></a>
## [v0.3.1](https://github.com/sourcenetwork/defradb/compare/v0.3.0...v0.3.1)

> 2022-09-23

DefraDB v0.3.1 is a minor release, primarily focusing on additional/extended features and fixes of items added in the `v0.3.0` release.

### Features

* Add cid support for allCommits ([#857](https://github.com/sourcenetwork/defradb/issues/857))
* Add offset support to allCommits ([#859](https://github.com/sourcenetwork/defradb/issues/859))
* Add limit support to allCommits query ([#856](https://github.com/sourcenetwork/defradb/issues/856))
* Add order support to allCommits ([#845](https://github.com/sourcenetwork/defradb/issues/845))
* Display CLI usage on user error ([#819](https://github.com/sourcenetwork/defradb/issues/819))
* Add support for dockey filters in child joins ([#806](https://github.com/sourcenetwork/defradb/issues/806))
* Add sort support for numeric aggregates ([#786](https://github.com/sourcenetwork/defradb/issues/786))
* Allow filtering by nil ([#789](https://github.com/sourcenetwork/defradb/issues/789))
* Add aggregate offset support ([#778](https://github.com/sourcenetwork/defradb/issues/778))
* Remove filter depth limit ([#777](https://github.com/sourcenetwork/defradb/issues/777))
* Add support for and-or inline array aggregate filters ([#779](https://github.com/sourcenetwork/defradb/issues/779))
* Add limit support for aggregates ([#771](https://github.com/sourcenetwork/defradb/issues/771))
* Add support for inline arrays of nillable types ([#759](https://github.com/sourcenetwork/defradb/issues/759))
* Create errors package ([#548](https://github.com/sourcenetwork/defradb/issues/548))
* Add ability to display peer id ([#719](https://github.com/sourcenetwork/defradb/issues/719))
* Add a config option to set the vlog max file size ([#743](https://github.com/sourcenetwork/defradb/issues/743))
* Explain `topLevelNode` like a `MultiNode` plan ([#749](https://github.com/sourcenetwork/defradb/issues/749))
* Make `topLevelNode` explainable ([#737](https://github.com/sourcenetwork/defradb/issues/737))

### Fixes

* Order subtype without selecting the join child ([#810](https://github.com/sourcenetwork/defradb/issues/810))
* Correctly handles nil one-one joins ([#837](https://github.com/sourcenetwork/defradb/issues/837))
* Reset scan node for each join ([#828](https://github.com/sourcenetwork/defradb/issues/828))
* Handle filter input field argument being nil ([#787](https://github.com/sourcenetwork/defradb/issues/787))
* Ensure CLI outputs JSON to stdout when directed to pipe ([#804](https://github.com/sourcenetwork/defradb/issues/804))
* Error if given the wrong side of a one-one relationship ([#795](https://github.com/sourcenetwork/defradb/issues/795))
* Add object marker to enable return of empty docs ([#800](https://github.com/sourcenetwork/defradb/issues/800))
* Resolve the extra `typeIndexJoin`s for `_avg` aggregate ([#774](https://github.com/sourcenetwork/defradb/issues/774))
* Remove _like filter operator ([#797](https://github.com/sourcenetwork/defradb/issues/797))
* Remove having gql types ([#785](https://github.com/sourcenetwork/defradb/issues/785))
* Error if child _group selected without parent groupBy ([#781](https://github.com/sourcenetwork/defradb/issues/781))
* Error nicely on missing field specifier ([#782](https://github.com/sourcenetwork/defradb/issues/782))
* Handle order input field argument being nil ([#701](https://github.com/sourcenetwork/defradb/issues/701))
* Change output to outputpath in config file template for logger ([#716](https://github.com/sourcenetwork/defradb/issues/716))
* Delete mutations not correct persisting all keys ([#731](https://github.com/sourcenetwork/defradb/issues/731))

### Tooling

* Ban the usage of `ioutil` package ([#747](https://github.com/sourcenetwork/defradb/issues/747))
* Migrate from CircleCi to GitHub Actions ([#679](https://github.com/sourcenetwork/defradb/issues/679))

### Documentation

* Clarify meaning of url param, update in-repo CLI docs ([#814](https://github.com/sourcenetwork/defradb/issues/814))
* Disclaimer of exposed to network and not encrypted ([#793](https://github.com/sourcenetwork/defradb/issues/793))
* Update logo to respect theme ([#728](https://github.com/sourcenetwork/defradb/issues/728))

### Refactoring

* Replace all `interface{}` with `any` alias ([#805](https://github.com/sourcenetwork/defradb/issues/805))
* Use fastjson to parse mutation data string ([#772](https://github.com/sourcenetwork/defradb/issues/772))
* Rework limit node flow ([#767](https://github.com/sourcenetwork/defradb/issues/767))
* Make Option immutable ([#769](https://github.com/sourcenetwork/defradb/issues/769))
* Rework sum and count nodes to make use of generics ([#757](https://github.com/sourcenetwork/defradb/issues/757))
* Remove some possible panics from codebase ([#732](https://github.com/sourcenetwork/defradb/issues/732))
* Change logging calls to use feedback in CLI package ([#714](https://github.com/sourcenetwork/defradb/issues/714))

### Testing

* Add tests for aggs with nil filters ([#813](https://github.com/sourcenetwork/defradb/issues/813))
* Add not equals filter tests ([#798](https://github.com/sourcenetwork/defradb/issues/798))
* Fix `cli/peerid_test` to not clash addresses ([#766](https://github.com/sourcenetwork/defradb/issues/766))
* Add change detector summary to test readme ([#754](https://github.com/sourcenetwork/defradb/issues/754))
* Add tests for inline array grouping ([#752](https://github.com/sourcenetwork/defradb/issues/752))

### Continuous integration

* Reduce test resource usage and test with file db ([#791](https://github.com/sourcenetwork/defradb/issues/791))
* Add makefile target to verify the local module cache ([#775](https://github.com/sourcenetwork/defradb/issues/775))
* Allow PR titles to end with a number ([#745](https://github.com/sourcenetwork/defradb/issues/745))
* Add a workflow to validate pull request titles ([#734](https://github.com/sourcenetwork/defradb/issues/734))
* Fix the linter version to `v1.47` ([#726](https://github.com/sourcenetwork/defradb/issues/726))

### Chore

* Remove file system paths from resulting executable ([#831](https://github.com/sourcenetwork/defradb/issues/831))
* Add goimports linter for consistent imports ordering ([#816](https://github.com/sourcenetwork/defradb/issues/816))
* Improve UX by providing more information ([#802](https://github.com/sourcenetwork/defradb/issues/802))
* Change to defra errors and handle errors stacktrace ([#794](https://github.com/sourcenetwork/defradb/issues/794))
* Clean up `go.mod` with pruned module graphs ([#756](https://github.com/sourcenetwork/defradb/issues/756))
* Update to v0.20.3 of libp2p ([#740](https://github.com/sourcenetwork/defradb/issues/740))
* Bump to GoLang `v1.18` ([#721](https://github.com/sourcenetwork/defradb/issues/721))


<a name="v0.3.0"></a>
## [v0.3.0](https://github.com/sourcenetwork/defradb/compare/v0.2.1...v0.3.0)

> 2022-08-02

DefraDB v0.3 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

There are *several* new features in this release, and we invite you to review the official changelog below. Some highlights are various new features for Grouping & Aggregation for the query system, like top-level aggregation and group filtering. Moreover, a brand new Query Explain system was added to introspect the execution plans created by DefraDB. Lastly we introduced a revamped CLI configuration system.

This release does include a Breaking Change to existing v0.2.x databases. If you need help migrating an existing deployment, reach out at [hello@source.network](mailto:hello@source.network) or join our Discord at https://discord.gg/w7jYQVJ/.

### Features

* Add named config overrides ([#659](https://github.com/sourcenetwork/defradb/issues/659))
* Expose color and caller log options, add validation ([#652](https://github.com/sourcenetwork/defradb/issues/652))
* Add ability to explain `groupNode` and it's attribute(s). ([#641](https://github.com/sourcenetwork/defradb/issues/641))
* Add primary directive for schema definitions ([@primary](https://github.com/primary)) ([#650](https://github.com/sourcenetwork/defradb/issues/650))
* Add support for aggregate filters on inline arrays ([#622](https://github.com/sourcenetwork/defradb/issues/622))
* Add explainable renderLimitNode & hardLimitNode attributes. ([#614](https://github.com/sourcenetwork/defradb/issues/614))
* Add support for top level aggregates ([#594](https://github.com/sourcenetwork/defradb/issues/594))
* Update `countNode` explanation to be consistent. ([#600](https://github.com/sourcenetwork/defradb/issues/600))
* Add support for stdin as input in CLI ([#608](https://github.com/sourcenetwork/defradb/issues/608))
* Explain `cid` & `field` attributes for `dagScanNode` ([#598](https://github.com/sourcenetwork/defradb/issues/598))
* Add ability to explain `dagScanNode` attribute(s). ([#560](https://github.com/sourcenetwork/defradb/issues/560))
* Add the ability to send user feedback to the console even when logging to file. ([#568](https://github.com/sourcenetwork/defradb/issues/568))
* Add ability to explain `sortNode` attribute(s). ([#558](https://github.com/sourcenetwork/defradb/issues/558))
* Add ability to explain `sumNode` attribute(s). ([#559](https://github.com/sourcenetwork/defradb/issues/559))
* Introduce top-level config package ([#389](https://github.com/sourcenetwork/defradb/issues/389))
* Add ability to explain `updateNode` attributes. ([#514](https://github.com/sourcenetwork/defradb/issues/514))
* Add `typeIndexJoin` explainable attributes. ([#499](https://github.com/sourcenetwork/defradb/issues/499))
* Add support to explain `countNode` attributes. ([#504](https://github.com/sourcenetwork/defradb/issues/504))
* Add CORS capability to HTTP API ([#467](https://github.com/sourcenetwork/defradb/issues/467))
* Add explaination of spans for `scanNode`. ([#492](https://github.com/sourcenetwork/defradb/issues/492))
* Add ability to Explain the response plan. ([#385](https://github.com/sourcenetwork/defradb/issues/385))
* Add aggregate filter support for groups only ([#426](https://github.com/sourcenetwork/defradb/issues/426))
* Configurable caller option in logger ([#416](https://github.com/sourcenetwork/defradb/issues/416))
* Add Average aggregate support ([#383](https://github.com/sourcenetwork/defradb/issues/383))
* Allow summation of aggregates ([#341](https://github.com/sourcenetwork/defradb/issues/341))
* Add ability to check DefraDB CLI version. ([#339](https://github.com/sourcenetwork/defradb/issues/339))

### Fixes

* Add a check to ensure limit is not 0 when evaluating query limit and offset ([#706](https://github.com/sourcenetwork/defradb/issues/706))
* Support multiple `--logger` flags ([#704](https://github.com/sourcenetwork/defradb/issues/704))
* Return without an error if relation is finalized ([#698](https://github.com/sourcenetwork/defradb/issues/698))
* Logger not correctly applying named config ([#696](https://github.com/sourcenetwork/defradb/issues/696))
* Add content-type media type parsing ([#678](https://github.com/sourcenetwork/defradb/issues/678))
* Remove portSyncLock deadlock condition ([#671](https://github.com/sourcenetwork/defradb/issues/671))
* Silence cobra default errors and usage printing ([#668](https://github.com/sourcenetwork/defradb/issues/668))
* Add stdout validation when setting logging output path ([#666](https://github.com/sourcenetwork/defradb/issues/666))
* Consider `--logoutput` CLI flag properly ([#645](https://github.com/sourcenetwork/defradb/issues/645))
* Handle errors and responses in CLI `client` commands ([#579](https://github.com/sourcenetwork/defradb/issues/579))
* Rename aggregate gql types ([#638](https://github.com/sourcenetwork/defradb/issues/638))
* Error when attempting to insert value into relationship field ([#632](https://github.com/sourcenetwork/defradb/issues/632))
* Allow adding of new schema to database ([#635](https://github.com/sourcenetwork/defradb/issues/635))
* Correctly parse dockey in broadcast log event. ([#631](https://github.com/sourcenetwork/defradb/issues/631))
* Increase system's open files limit in integration tests ([#627](https://github.com/sourcenetwork/defradb/issues/627))
* Avoid populating `order.ordering` with empties. ([#618](https://github.com/sourcenetwork/defradb/issues/618))
* Change to supporting of non-null inline arrays ([#609](https://github.com/sourcenetwork/defradb/issues/609))
* Assert fields exist in collection before saving to them ([#604](https://github.com/sourcenetwork/defradb/issues/604))
* CLI `init` command to reinitialize only config file ([#603](https://github.com/sourcenetwork/defradb/issues/603))
* Add config and registry clearing to TestLogWritesMessagesToFeedbackLog ([#596](https://github.com/sourcenetwork/defradb/issues/596))
* Change `$eq` to `_eq` in the failing test. ([#576](https://github.com/sourcenetwork/defradb/issues/576))
* Resolve failing HTTP API tests via cleanup ([#557](https://github.com/sourcenetwork/defradb/issues/557))
* Ensure Makefile compatibility with macOS ([#527](https://github.com/sourcenetwork/defradb/issues/527))
* Separate out iotas in their own blocks. ([#464](https://github.com/sourcenetwork/defradb/issues/464))
* Use x/cases for titling instead of strings to handle deprecation ([#457](https://github.com/sourcenetwork/defradb/issues/457))
* Handle limit and offset in sub groups ([#440](https://github.com/sourcenetwork/defradb/issues/440))
* Issue preventing DB from restarting with no records ([#437](https://github.com/sourcenetwork/defradb/issues/437))
* log serving HTTP API before goroutine blocks ([#358](https://github.com/sourcenetwork/defradb/issues/358))

### Testing

* Add integration testing for P2P. ([#655](https://github.com/sourcenetwork/defradb/issues/655))
* Fix formatting of tests with no extra brackets ([#643](https://github.com/sourcenetwork/defradb/issues/643))
* Add tests for `averageNode` explain. ([#639](https://github.com/sourcenetwork/defradb/issues/639))
* Add schema integration tests ([#628](https://github.com/sourcenetwork/defradb/issues/628))
* Add tests for default properties ([#611](https://github.com/sourcenetwork/defradb/issues/611))
* Specify which collection to update in test framework ([#601](https://github.com/sourcenetwork/defradb/issues/601))
* Add tests for grouping by undefined value ([#543](https://github.com/sourcenetwork/defradb/issues/543))
* Add test for querying undefined field ([#544](https://github.com/sourcenetwork/defradb/issues/544))
* Expand commit query tests ([#541](https://github.com/sourcenetwork/defradb/issues/541))
* Add cid (time-travel) query tests ([#539](https://github.com/sourcenetwork/defradb/issues/539))
* Restructure and expand filter tests ([#512](https://github.com/sourcenetwork/defradb/issues/512))
* Basic unit testing of `node` package ([#503](https://github.com/sourcenetwork/defradb/issues/503))
* Test filter in filter tests ([#473](https://github.com/sourcenetwork/defradb/issues/473))
* Add test for deletion of records in a relationship ([#329](https://github.com/sourcenetwork/defradb/issues/329))
* Benchmark transaction iteration ([#289](https://github.com/sourcenetwork/defradb/issues/289))

### Refactoring

* Improve CLI error handling and fix small issues ([#649](https://github.com/sourcenetwork/defradb/issues/649))
* Add top-level `version` package ([#583](https://github.com/sourcenetwork/defradb/issues/583))
* Remove extra log levels ([#634](https://github.com/sourcenetwork/defradb/issues/634))
* Change `sortNode` to `orderNode`. ([#591](https://github.com/sourcenetwork/defradb/issues/591))
* Rework update and delete node to remove secondary planner ([#571](https://github.com/sourcenetwork/defradb/issues/571))
* Trim imported connor package  ([#530](https://github.com/sourcenetwork/defradb/issues/530))
* Internal doc restructure ([#471](https://github.com/sourcenetwork/defradb/issues/471))
* Copy-paste connor fork into repo ([#567](https://github.com/sourcenetwork/defradb/issues/567))
* Add safety to the tests, add ability to catch stderr logs and add output path validation ([#552](https://github.com/sourcenetwork/defradb/issues/552))
* Change handler functions implementation and response formatting ([#498](https://github.com/sourcenetwork/defradb/issues/498))
* Improve the HTTP API implementation ([#382](https://github.com/sourcenetwork/defradb/issues/382))
* Use new logger in net/api ([#420](https://github.com/sourcenetwork/defradb/issues/420))
* Rename NewCidV1_SHA2_256 to mixedCaps ([#415](https://github.com/sourcenetwork/defradb/issues/415))
* Remove utils package ([#397](https://github.com/sourcenetwork/defradb/issues/397))
* Rework planNode Next and Value(s) function  ([#374](https://github.com/sourcenetwork/defradb/issues/374))
* Restructure aggregate query syntax ([#373](https://github.com/sourcenetwork/defradb/issues/373))
* Remove dead code from client package and document remaining ([#356](https://github.com/sourcenetwork/defradb/issues/356))
* Restructure datastore keys ([#316](https://github.com/sourcenetwork/defradb/issues/316))
* Add commits lost during github outage ([#303](https://github.com/sourcenetwork/defradb/issues/303))
* Move public members out of core and base packages ([#295](https://github.com/sourcenetwork/defradb/issues/295))
* Make db stuff internal/private ([#291](https://github.com/sourcenetwork/defradb/issues/291))
* Rework client.DB to ensure interface contains only public types ([#277](https://github.com/sourcenetwork/defradb/issues/277))
* Remove GetPrimaryIndexDocKey from collection interface ([#279](https://github.com/sourcenetwork/defradb/issues/279))
* Remove DataStoreKey from (public) dockey struct ([#278](https://github.com/sourcenetwork/defradb/issues/278))
* Renormalize to ensure consistent file line termination. ([#226](https://github.com/sourcenetwork/defradb/issues/226))
* Strongly typed key refactor ([#17](https://github.com/sourcenetwork/defradb/issues/17))

### Documentation

* Use permanent link to BSL license document ([#692](https://github.com/sourcenetwork/defradb/issues/692))
* README update v0.3.0 ([#646](https://github.com/sourcenetwork/defradb/issues/646))
* Improve code documentation ([#533](https://github.com/sourcenetwork/defradb/issues/533))
* Add CONTRIBUTING.md ([#531](https://github.com/sourcenetwork/defradb/issues/531))
* Add package level docs for logging lib ([#338](https://github.com/sourcenetwork/defradb/issues/338))

### Tooling

* Include all touched packages in code coverage ([#673](https://github.com/sourcenetwork/defradb/issues/673))
* Use `gotestsum` over `go test` ([#619](https://github.com/sourcenetwork/defradb/issues/619))
* Update Github pull request template ([#524](https://github.com/sourcenetwork/defradb/issues/524))
* Fix the cross-build script ([#460](https://github.com/sourcenetwork/defradb/issues/460))
* Add test coverage html output ([#466](https://github.com/sourcenetwork/defradb/issues/466))
* Add linter rule for `goconst`. ([#398](https://github.com/sourcenetwork/defradb/issues/398))
* Add github PR template. ([#394](https://github.com/sourcenetwork/defradb/issues/394))
* Disable auto-fixing linter issues by default ([#429](https://github.com/sourcenetwork/defradb/issues/429))
* Fix linting of empty `else` code blocks ([#402](https://github.com/sourcenetwork/defradb/issues/402))
* Add the `gofmt` linter rule. ([#405](https://github.com/sourcenetwork/defradb/issues/405))
* Cleanup linter config file ([#400](https://github.com/sourcenetwork/defradb/issues/400))
* Add linter rule for copyright headers ([#360](https://github.com/sourcenetwork/defradb/issues/360))
* Organize our config files and tooling. ([#336](https://github.com/sourcenetwork/defradb/issues/336))
* Limit line length to 100 characters (linter check) ([#224](https://github.com/sourcenetwork/defradb/issues/224))
* Ignore db/tests folder and the bench marks. ([#280](https://github.com/sourcenetwork/defradb/issues/280))

### Continuous Integration

* Fix circleci cache permission errors. ([#371](https://github.com/sourcenetwork/defradb/issues/371))
* Ban extra elses ([#366](https://github.com/sourcenetwork/defradb/issues/366))
* Fix change-detection to not fail when new tests are added. ([#333](https://github.com/sourcenetwork/defradb/issues/333))
* Update golang-ci linter and explicit go-setup to use v1.17 ([#331](https://github.com/sourcenetwork/defradb/issues/331))
* Comment the benchmarking result comparison to the PR ([#305](https://github.com/sourcenetwork/defradb/issues/305))
* Add benchmark performance comparisons ([#232](https://github.com/sourcenetwork/defradb/issues/232))
* Add caching / storing of bench report on default branch ([#290](https://github.com/sourcenetwork/defradb/issues/290))
* Ensure full-benchmarks are ran on a PR-merge. ([#282](https://github.com/sourcenetwork/defradb/issues/282))
* Add ability to control benchmarks by PR labels. ([#267](https://github.com/sourcenetwork/defradb/issues/267))

### Chore

* Update APL to refer to D2 Foundation ([#711](https://github.com/sourcenetwork/defradb/issues/711))
* Update gitignore to include `cmd` folders ([#617](https://github.com/sourcenetwork/defradb/issues/617))
* Enable random execution order of tests ([#554](https://github.com/sourcenetwork/defradb/issues/554))
* Enable linters exportloopref, nolintlint, whitespace ([#535](https://github.com/sourcenetwork/defradb/issues/535))
* Add utility for generation of man pages ([#493](https://github.com/sourcenetwork/defradb/issues/493))
* Add Dockerfile ([#517](https://github.com/sourcenetwork/defradb/issues/517))
* Enable errorlint linter ([#520](https://github.com/sourcenetwork/defradb/issues/520))
* Binaries in`cmd` folder, examples in `examples` folder ([#501](https://github.com/sourcenetwork/defradb/issues/501))
* Improve log outputs ([#506](https://github.com/sourcenetwork/defradb/issues/506))
* Move testing to top-level `tests` folder ([#446](https://github.com/sourcenetwork/defradb/issues/446))
* Update dependencies ([#450](https://github.com/sourcenetwork/defradb/issues/450))
* Update go-ipfs-blockstore and ipfs-lite ([#436](https://github.com/sourcenetwork/defradb/issues/436))
* Update libp2p dependency to v0.19 ([#424](https://github.com/sourcenetwork/defradb/issues/424))
* Update ioutil package to io / os packages. ([#376](https://github.com/sourcenetwork/defradb/issues/376))
* git ignore vscode ([#343](https://github.com/sourcenetwork/defradb/issues/343))
* Updated README.md contributors section ([#292](https://github.com/sourcenetwork/defradb/issues/292))
* Update changelog v0.2.1 ([#252](https://github.com/sourcenetwork/defradb/issues/252))


<a name="v0.2.1"></a>
## [v0.2.1](https://github.com/sourcenetwork/defradb/compare/v0.2.0...v0.2.1)

> 2022-03-04

### Features

* Add ability to delete multiple documents using filter ([#206](https://github.com/sourcenetwork/defradb/issues/206))
* Add ability to delete multiple documents, using multiple ids ([#196](https://github.com/sourcenetwork/defradb/issues/196))

### Fixes

* Concurrency control of Document using RWMutex ([#213](https://github.com/sourcenetwork/defradb/issues/213))
* Only log errors and above when benchmarking ([#261](https://github.com/sourcenetwork/defradb/issues/261))
* Handle proper type conversion on sort nodes ([#228](https://github.com/sourcenetwork/defradb/issues/228))
* Return empty array if no values found ([#223](https://github.com/sourcenetwork/defradb/issues/223))
* Close fetcher on error ([#210](https://github.com/sourcenetwork/defradb/issues/210))
* Installing binary using defradb name ([#190](https://github.com/sourcenetwork/defradb/issues/190))

### Tooling

* Add short benchmark runner option ([#263](https://github.com/sourcenetwork/defradb/issues/263))

### Documentation

* Add data format changes documentation folder ([#89](https://github.com/sourcenetwork/defradb/issues/89))
* Correcting typos ([#143](https://github.com/sourcenetwork/defradb/issues/143))
* Update generated CLI docs ([#208](https://github.com/sourcenetwork/defradb/issues/208))
* Updated readme with P2P section ([#220](https://github.com/sourcenetwork/defradb/issues/220))
* Update old or missing license headers ([#205](https://github.com/sourcenetwork/defradb/issues/205))
* Update git-chglog config and template ([#195](https://github.com/sourcenetwork/defradb/issues/195))

### Refactoring

* Introduction of logging system ([#67](https://github.com/sourcenetwork/defradb/issues/67))
* Restructure db/txn/multistore structures ([#199](https://github.com/sourcenetwork/defradb/issues/199))
* Initialize database in constructor ([#211](https://github.com/sourcenetwork/defradb/issues/211))
* Purge all println and ban it ([#253](https://github.com/sourcenetwork/defradb/issues/253))

### Testing

* Detect and force breaking filesystem changes to be documented ([#89](https://github.com/sourcenetwork/defradb/issues/89))
* Boost collection test coverage ([#183](https://github.com/sourcenetwork/defradb/issues/183))

### Continuous integration

* Combine the Lint and Benchmark workflows so that the benchmark job depends on the lint job in one workflow ([#209](https://github.com/sourcenetwork/defradb/issues/209))
* Add rule to only run benchmark if other check are successful ([#194](https://github.com/sourcenetwork/defradb/issues/194))
* Increase linter timeout ([#230](https://github.com/sourcenetwork/defradb/issues/230))

### Chore

* Remove commented out code ([#238](https://github.com/sourcenetwork/defradb/issues/238))
* Remove dead code from multi node ([#186](https://github.com/sourcenetwork/defradb/issues/186))


<a name="v0.2.0"></a>
## [v0.2.0](https://github.com/sourcenetwork/defradb/compare/v0.1.0...v0.2.0)

> 2022-02-07

DefraDB v0.2 is a major pre-production release. Until the stable version 1.0 is reached, the SemVer minor patch number will denote notable releases, which will give the project freedom to experiment and explore potentially breaking changes.

This release is jam-packed with new features and a small number of breaking changes. Read the full changelog for a detailed description. Most notable features include a new Peer-to-Peer (P2P) data synchronization system, an expanded query system to support GroupBy & Aggregate operations, and lastly TimeTraveling queries allowing to query previous states of a document.

Much more than just that has been added to ensure we're building reliable software expected of any database, such as expanded test & benchmark suites, automated bug detection, performance gains, and more.

This release does include a Breaking Change to existing v0.1 databases regarding the internal data model, which affects the "Content Identifiers" we use to generate DocKeys and VersionIDs. If you need help migrating an existing deployment, reach out at hello@source.network or join our Discord at https://discord.gg/w7jYQVJ.

### Features

* Added Peer-to-Peer networking data synchronization ([#177](https://github.com/sourcenetwork/defradb/issues/177))
* TimeTraveling (History Traversing) query engine and doc fetcher ([#59](https://github.com/sourcenetwork/defradb/issues/59))
* Add Document Deletion with a Key ([#150](https://github.com/sourcenetwork/defradb/issues/150))
* Add support for sum aggregate ([#121](https://github.com/sourcenetwork/defradb/issues/121))
* Add support for lwwr scalar arrays (full replace on update) ([#115](https://github.com/sourcenetwork/defradb/issues/115))
* Add count aggregate support ([#102](https://github.com/sourcenetwork/defradb/issues/102))
* Add support for named relationships ([#108](https://github.com/sourcenetwork/defradb/issues/108))
* Add multi doc key lookup support ([#76](https://github.com/sourcenetwork/defradb/issues/76))
* Add basic group by functionality ([#43](https://github.com/sourcenetwork/defradb/issues/43))
* Update datastore packages to allow use of context ([#48](https://github.com/sourcenetwork/defradb/issues/48))

### Bug fixes

* Only add join if aggregating child object collection ([#188](https://github.com/sourcenetwork/defradb/issues/188))
* Handle errors generated during input object thunks ([#123](https://github.com/sourcenetwork/defradb/issues/123))
* Remove new types from in-memory cache on generate error ([#122](https://github.com/sourcenetwork/defradb/issues/122))
* Support relationships where both fields have the same name ([#109](https://github.com/sourcenetwork/defradb/issues/109))
* Handle errors generated in fields thunk ([#66](https://github.com/sourcenetwork/defradb/issues/66))
* Ensure OperationDefinition case has at least one selection([#24](https://github.com/sourcenetwork/defradb/pull/24))
* Close datastore iterator on scan close ([#56](https://github.com/sourcenetwork/defradb/pull/56)) (resulted in a panic when using limit)
* Close superseded iterators before orphaning ([#56](https://github.com/sourcenetwork/defradb/pull/56)) (fixes a panic in the join code) 
* Move discard to after error check ([#88](https://github.com/sourcenetwork/defradb/pull/88)) (did result in panic if transaction creation fails)
* Check for nil iterator before closing document fetcher ([#108](https://github.com/sourcenetwork/defradb/pull/108))

### Tooling
* Added benchmark suite ([#160](https://github.com/sourcenetwork/defradb/issues/160))

### Documentation

* Correcting comment typos ([#142](https://github.com/sourcenetwork/defradb/issues/142))
* Correcting README typos ([#140](https://github.com/sourcenetwork/defradb/issues/140))

### Testing

* Add transaction integration tests ([#175](https://github.com/sourcenetwork/defradb/issues/175))
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

### Performance
* Add badger multi scan support ([#85](https://github.com/sourcenetwork/defradb/pull/85))
* Add support for range spans ([#86](https://github.com/sourcenetwork/defradb/pull/86))

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

