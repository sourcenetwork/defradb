// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/matchers"
	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/datastore"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema/types"
	netConfig "github.com/sourcenetwork/defradb/net/config"
	"github.com/sourcenetwork/defradb/node"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/gen"
	"github.com/sourcenetwork/defradb/tests/predefined"
)

const (
	mutationTypeEnvName     = "DEFRA_MUTATION_TYPE"
	viewTypeEnvName         = "DEFRA_VIEW_TYPE"
	skipNetworkTestsEnvName = "DEFRA_SKIP_NETWORK_TESTS"
	vectorEmbeddingEnvName  = "DEFRA_VECTOR_EMBEDDING"
)

// The MutationType that tests will run using.
//
// For example if set to [CollectionSaveMutationType], all supporting
// actions (such as [UpdateDoc]) will execute via [Collection.Save].
//
// Defaults to CollectionSaveMutationType.
type MutationType string

const (
	// CollectionSaveMutationType will cause all supporting actions
	// to run their mutations via [Collection.Save].
	CollectionSaveMutationType MutationType = "collection-save"

	// CollectionNamedMutationType will cause all supporting actions
	// to run their mutations via their corresponding named [Collection]
	// call.
	//
	// For example, CreateDoc will call [Collection.Create], and
	// UpdateDoc will call [Collection.Update].
	CollectionNamedMutationType MutationType = "collection-named"

	// GQLRequestMutationType will cause all supporting actions to
	// run their mutations using GQL requests, typically these will
	// include a `id` parameter to target the specified document.
	GQLRequestMutationType MutationType = "gql"
)

type ViewType string

const (
	CachelessViewType    ViewType = "cacheless"
	MaterializedViewType ViewType = "materialized"
)

var (
	log          = corelog.NewLogger("tests.integration")
	mutationType MutationType
	viewType     ViewType
	// skipNetworkTests will skip any tests that involve network actions
	skipNetworkTests = false
	// skipBackupTests will skip any tests that involve backup actions
	skipBackupTests = false
	// runVectorEmbeddingTests will whether tests with vector embedding generation should be executed.
	runVectorEmbeddingTests = false
)

const (
	// subscriptionTimeout is the maximum time to wait for subscription results to be returned.
	subscriptionTimeout = 1 * time.Second
	// Instantiating lenses is expensive, and our tests do not benefit from a large number of them,
	// so we explicitly set it to a low value.
	lensPoolSize = 2
)

const testJSONFile = "/test.json"

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	if value, ok := os.LookupEnv(mutationTypeEnvName); ok {
		mutationType = MutationType(value)
	} else {
		// Default to testing mutations via Collection.Save - it should be simpler and
		// faster. We assume this is desirable when not explicitly testing any particular
		// mutation type.
		mutationType = CollectionSaveMutationType
	}

	if value, ok := os.LookupEnv(viewTypeEnvName); ok {
		viewType = ViewType(value)
	} else {
		viewType = CachelessViewType
	}

	if value, ok := os.LookupEnv(skipNetworkTestsEnvName); ok {
		skipNetworkTests, _ = strconv.ParseBool(value)
	}

	if value, ok := os.LookupEnv(vectorEmbeddingEnvName); ok {
		runVectorEmbeddingTests, _ = strconv.ParseBool(value)
	}
}

// AssertPanic asserts that the code inside the specified PanicTestFunc panics.
//
// This function is not supported by either the change detector, or the http-client.
// Calling this within either of them will result in the test being skipped.
//
// Usage: AssertPanic(t, func() { executeTestCase(t, test) })
func AssertPanic(t *testing.T, f assert.PanicTestFunc) bool {
	if changeDetector.Enabled {
		// The `assert.Panics` call will falsely fail if this test is executed during
		// a detect changes test run.
		t.Skip("Assert panic with the change detector is not currently supported.")
	}

	if httpClient || cliClient {
		// The http / cli client will return an error instead of panicking at the moment.
		t.Skip("Assert panic with the http client is not currently supported.")
	}

	return assert.Panics(t, f, "expected a panic, but none found.")
}

// ExecuteTestCase executes the given TestCase against the configured database
// instances.
//
// Will also attempt to detect incompatible changes in the persisted data if
// configured to do so (the CI will do so, but disabled by default as it is slow).
func ExecuteTestCase(
	t testing.TB,
	testCase TestCase,
) {
	flattenActions(&testCase)
	collectionNames := getCollectionNames(testCase)
	changeDetector.PreTestChecks(t, collectionNames)
	skipIfMutationTypeUnsupported(t, testCase.SupportedMutationTypes)
	skipIfDocumentACPTypeUnsupported(t, testCase.SupportedDocumentACPTypes)
	skipIfNetworkTest(t, testCase.Actions)
	skipIfBackupTest(t, testCase.Actions)
	skipIfViewCacheTypeUnsupported(t, testCase.SupportedViewTypes)
	skipIfVectorEmbeddingTest(t, testCase.Actions)

	var clients []ClientType
	if httpClient {
		clients = append(clients, HTTPClientType)
	}
	if goClient {
		clients = append(clients, GoClientType)
	}
	if cliClient {
		clients = append(clients, CLIClientType)
	}
	if jsClient {
		clients = append(clients, JSClientType)
	}

	var databases []DatabaseType
	if badgerInMemory {
		databases = append(databases, BadgerIMType)
	}
	if badgerFile {
		databases = append(databases, BadgerFileType)
	}
	if inMemoryStore {
		databases = append(databases, DefraIMType)
	}

	var kmsList []KMSType
	if testCase.KMS.Activated {
		kmsList = getKMSTypes()
		for _, excluded := range testCase.KMS.ExcludedTypes {
			kmsList = slices.DeleteFunc(kmsList, func(t KMSType) bool { return t == excluded })
		}
	}
	if len(kmsList) == 0 {
		kmsList = []KMSType{NoneKMSType}
	}

	// Assert that these are not empty to protect against accidental mis-configurations,
	// otherwise an empty set would silently pass all the tests.
	require.NotEmpty(t, databases)
	require.NotEmpty(t, clients)

	databases = skipIfDatabaseTypeUnsupported(t, databases, testCase.SupportedDatabaseTypes)
	clients = skipIfClientTypeUnsupported(t, clients, testCase.SupportedClientTypes)

	ctx := context.Background()
	for _, ct := range clients {
		for _, dbt := range databases {
			for _, kms := range kmsList {
				executeTestCase(ctx, t, collectionNames, testCase, kms, dbt, ct)
			}
		}
	}
}

func executeTestCase(
	ctx context.Context,
	t testing.TB,
	collectionNames []string,
	testCase TestCase,
	kms KMSType,
	dbt DatabaseType,
	clientType ClientType,
) {
	logAttrs := []slog.Attr{
		corelog.Any("database", dbt),
		corelog.Any("client", clientType),
		corelog.Any("mutationType", mutationType),
		corelog.String("databaseDir", databaseDir),
		corelog.Bool("badgerEncryption", badgerEncryption),
		corelog.Bool("skipNetworkTests", skipNetworkTests),
		corelog.Bool("changeDetector.Enabled", changeDetector.Enabled),
		corelog.Bool("changeDetector.SetupOnly", changeDetector.SetupOnly),
		corelog.String("changeDetector.SourceBranch", changeDetector.SourceBranch),
		corelog.String("changeDetector.TargetBranch", changeDetector.TargetBranch),
		corelog.String("changeDetector.Repository", changeDetector.Repository),
	}

	if kms != NoneKMSType {
		logAttrs = append(logAttrs, corelog.Any("KMS", kms))
	}

	log.InfoContext(ctx, testCase.Description, logAttrs...)

	startActionIndex, endActionIndex := getActionRange(t, testCase)

	s := newState(ctx, t, testCase, kms, dbt, clientType, collectionNames)
	setStartingNodes(s)

	// It is very important that the databases are always closed, otherwise resources will leak
	// as tests run.  This is particularly important for file based datastores.
	defer closeNodes(s, Close{})

	// Documents and Collections may already exist in the database if actions have been split
	// by the change detector so we should fetch them here at the start too (if they exist).
	// collections are by node (index), as they are specific to nodes.
	refreshCollections(s)
	refreshDocuments(s, startActionIndex)

	for i := startActionIndex; i <= endActionIndex; i++ {
		performAction(s, i, testCase.Actions[i])
	}

	// matchers can be instantiated not as part of the test state, but as a variable for Test... function scope
	// which will outlive all test runs (test instance of type [testUtils.TestCase]) and will be reused
	// by them. So the matchers need to be reset between the test runs.
	resetMatchers(s)

	// Notify any active subscriptions that all requests have been sent.
	close(s.allActionsDone)

	for _, resultsChan := range s.subscriptionResultsChans {
		select {
		case subscriptionAssert := <-resultsChan:
			// We want to assert back in the main thread so failures get recorded properly
			subscriptionAssert()

		// a safety in case the stream hangs - we don't want the tests to run forever.
		case <-time.After(subscriptionTimeout):
			assert.Fail(t, "timeout occurred while waiting for data stream", testCase.Description)
		}
	}
}

func performAction(
	s *state,
	actionIndex int,
	act any,
) {
	switch action := act.(type) {
	case ConfigureNode:
		configureNode(s, action)

	case Restart:
		restartNodes(s)

	case Close:
		closeNodes(s, action)

	case Start:
		startNodes(s, action)

	case ConnectPeers:
		connectPeers(s, action)

	case ConfigureReplicator:
		configureReplicator(s, action)

	case DeleteReplicator:
		deleteReplicator(s, action)

	case SubscribeToCollection:
		subscribeToCollection(s, action)

	case UnsubscribeToCollection:
		unsubscribeToCollection(s, action)

	case GetAllP2PCollections:
		getAllP2PCollections(s, action)

	case SchemaUpdate:
		updateSchema(s, action)

	case SchemaPatch:
		patchSchema(s, action)

	case PatchCollection:
		patchCollection(s, action)

	case GetSchema:
		getSchema(s, action)

	case GetCollections:
		getCollections(s, action)

	case SetActiveSchemaVersion:
		setActiveSchemaVersion(s, action)

	case CreateView:
		createView(s, action)

	case RefreshViews:
		refreshViews(s, action)

	case ConfigureMigration:
		configureMigration(s, action)

	case AddDACPolicy:
		addDACPolicy(s, action)

	case AddDACActorRelationship:
		addDACActorRelationship(s, action)

	case DeleteDACActorRelationship:
		deleteDACActorRelationship(s, action)

	case CreateDoc:
		createDoc(s, action)

	case DeleteDoc:
		deleteDoc(s, action)

	case UpdateDoc:
		updateDoc(s, action)

	case UpdateWithFilter:
		updateWithFilter(s, action)

	case CreateIndex:
		createIndex(s, action)

	case DropIndex:
		dropIndex(s, action)

	case GetIndexes:
		getIndexes(s, action)

	case BackupExport:
		backupExport(s, action)

	case BackupImport:
		backupImport(s, action)

	case TransactionCommit:
		commitTransaction(s, action)

	case SubscriptionRequest:
		executeSubscriptionRequest(s, action)

	case Request:
		executeRequest(s, action)

	case ExplainRequest:
		executeExplainRequest(s, action)

	case IntrospectionRequest:
		assertIntrospectionResults(s, action)

	case ClientIntrospectionRequest:
		assertClientIntrospectionResults(s, action)

	case WaitForSync:
		waitForSync(s, action)

	case Wait:
		<-time.After(action.Duration)

	case Benchmark:
		benchmarkAction(s, actionIndex, action)

	case GenerateDocs:
		generateDocs(s, action)

	case CreatePredefinedDocs:
		generatePredefinedDocs(s, action)

	case GetNodeIdentity:
		performGetNodeIdentityAction(s, action)

	case VerifyBlockSignature:
		performVerifySignatureAction(s, action)

	case SetupComplete:
		// no-op, just continue.

	default:
		s.t.Fatalf("Unknown action type %T", action)
	}
}

func createGenerateDocs(s *state, docs []gen.GeneratedDoc, nodeID immutable.Option[int]) {
	nameToInd := make(map[string]int)
	for i, name := range s.collectionNames {
		nameToInd[name] = i
	}
	for _, doc := range docs {
		docJSON, err := doc.Doc.String()
		if err != nil {
			s.t.Fatalf("Failed to generate docs %s", err)
		}
		createDoc(s, CreateDoc{CollectionID: nameToInd[doc.Col.Version.Name], Doc: docJSON, NodeID: nodeID})
	}
}

func generateDocs(s *state, action GenerateDocs) {
	nodeIDs, _ := getNodesWithIDs(action.NodeID, s.nodes)
	firstNodesID := nodeIDs[0]
	collections := s.nodes[firstNodesID].collections
	defs := make([]client.CollectionDefinition, 0, len(collections))
	for _, collection := range collections {
		if len(action.ForCollections) == 0 || slices.Contains(action.ForCollections, collection.Name()) {
			defs = append(defs, collection.Definition())
		}
	}
	docs, err := gen.AutoGenerate(defs, action.Options...)
	if err != nil {
		s.t.Fatalf("Failed to generate docs %s", err)
	}
	createGenerateDocs(s, docs, action.NodeID)
}

func generatePredefinedDocs(s *state, action CreatePredefinedDocs) {
	nodeIDs, _ := getNodesWithIDs(action.NodeID, s.nodes)
	firstNodesID := nodeIDs[0]
	collections := s.nodes[firstNodesID].collections
	defs := make([]client.CollectionDefinition, 0, len(collections))
	for _, col := range collections {
		defs = append(defs, col.Definition())
	}
	docs, err := predefined.Create(defs, action.Docs)
	if err != nil {
		s.t.Fatalf("Failed to generate docs %s", err)
	}
	createGenerateDocs(s, docs, action.NodeID)
}

func benchmarkAction(
	s *state,
	actionIndex int,
	bench Benchmark,
) {
	if s.dbt == DefraIMType {
		// Benchmarking makes no sense for test in-memory storage
		return
	}
	if len(bench.FocusClients) > 0 {
		isFound := false
		for _, clientType := range bench.FocusClients {
			if s.clientType == clientType {
				isFound = true
				break
			}
		}
		if !isFound {
			return
		}
	}

	runBench := func(benchCase any) time.Duration {
		startTime := time.Now()
		for i := 0; i < bench.Reps; i++ {
			performAction(s, actionIndex, benchCase)
		}
		return time.Since(startTime)
	}

	s.isBench = true
	defer func() { s.isBench = false }()

	baseElapsedTime := runBench(bench.BaseCase)
	optimizedElapsedTime := runBench(bench.OptimizedCase)

	factoredBaseTime := int64(float64(baseElapsedTime) / bench.Factor)
	assert.Greater(s.t, factoredBaseTime, optimizedElapsedTime,
		"Optimized case should be faster at least by factor of %.2f than the base case. Base: %d, Optimized: %d (μs)",
		bench.Factor, optimizedElapsedTime.Microseconds(), baseElapsedTime.Microseconds())
}

// getCollectionNames gets an ordered, unique set of collection names across all nodes
// from the action set within the given test case.
//
// It preserves the order in which they are declared, and shares indexes across all nodes, so
// if a second node adds a collection of a name that was previously declared in another node
// the new node will respect the index originally assigned.  This allows collections to be
// referenced across multiple nodes by a consistent, predictable index - allowing a single
// action to target the same collection across multiple nodes.
//
// WARNING: This will not work with schemas ending in `type`, e.g. `user_type`
func getCollectionNames(testCase TestCase) []string {
	nextIndex := 0
	collectionIndexByName := map[string]int{}

	for _, a := range testCase.Actions {
		switch action := a.(type) {
		case SchemaUpdate:
			if action.ExpectedError != "" {
				// If an error is expected then no collections should result from this action
				continue
			}

			nextIndex = getCollectionNamesFromSchema(collectionIndexByName, action.Schema, nextIndex)

		case CreateView:
			if action.ExpectedError != "" {
				// If an error is expected then no collections should result from this action
				continue
			}

			nextIndex = getCollectionNamesFromSchema(collectionIndexByName, action.SDL, nextIndex)
		}
	}

	collectionNames := make([]string, len(collectionIndexByName))
	for name, index := range collectionIndexByName {
		collectionNames[index] = name
	}

	return collectionNames
}

func getCollectionNamesFromSchema(result map[string]int, schema string, nextIndex int) int {
	// WARNING: This will not work with schemas ending in `type`, e.g. `user_type`
	splitByType := strings.Split(schema, "type ")
	// Skip the first, as that precede `type ` if `type ` is present,
	// else there are no types.
	for i := 1; i < len(splitByType); i++ {
		wipSplit := strings.TrimLeft(splitByType[i], " ")
		indexOfLastChar := strings.IndexAny(wipSplit, " {")
		if indexOfLastChar <= 0 {
			// This should never happen
			continue
		}

		collectionName := wipSplit[:indexOfLastChar]
		if _, ok := result[collectionName]; ok {
			// Collection name has already been added, possibly via another node
			continue
		}

		result[collectionName] = nextIndex
		nextIndex++
	}
	return nextIndex
}

// closeNodes closes all the given nodes, ensuring that resources are properly released.
func closeNodes(
	s *state,
	action Close,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		node.Close()
		node.closed = true
	}
}

// getNodesWithIDs gets the applicable node(s) and their ID(s) for the given target nodeID.
//
// If nodeID has a value it will return that node and it's ID only. Otherwise all nodes will
// be returned with their corresponding IDs in a list.
//
// WARNING:
// The caller must not assume the returned node's ID is in order of the node's index if the specified nodeID is
// greater than 0. For example if requesting a node with nodeID=2 then the resulting output will contain only
// one element (at index 0) caller might accidentally assume that this node belongs to node 0. Therefore, the
// caller should always use the returned IDs, instead of guessing the IDs based on node indexes.
func getNodesWithIDs(nodeID immutable.Option[int], nodes []*nodeState) ([]int, []*nodeState) {
	if !nodeID.HasValue() {
		indexes := make([]int, len(nodes))
		for i := range nodes {
			indexes[i] = i
		}
		return indexes, nodes
	}

	return []int{nodeID.Value()}, []*nodeState{nodes[nodeID.Value()]}
}

func calculateLenForFlattenedActions(testCase *TestCase) int {
	newLen := 0
	for _, a := range testCase.Actions {
		actionGroup := reflect.ValueOf(a)
		switch actionGroup.Kind() {
		case reflect.Array, reflect.Slice:
			newLen += actionGroup.Len()
		default:
			newLen++
		}
	}
	return newLen
}

func flattenActions(testCase *TestCase) {
	newLen := calculateLenForFlattenedActions(testCase)
	if newLen == len(testCase.Actions) {
		return
	}
	newActions := make([]any, 0, newLen)

	for _, a := range testCase.Actions {
		actionGroup := reflect.ValueOf(a)
		switch actionGroup.Kind() {
		case reflect.Array, reflect.Slice:
			for i := 0; i < actionGroup.Len(); i++ {
				newActions = append(
					newActions,
					actionGroup.Index(i).Interface(),
				)
			}
		default:
			newActions = append(newActions, a)
		}
	}
	testCase.Actions = newActions
}

// getActionRange returns the index of the first action to be run, and the last.
//
// Not all processes will run all actions - if this is a change detector run they
// will be split.
//
// If a SetupComplete action is provided, the actions will be split there, if not
// they will be split at the first non SchemaUpdate/CreateDoc/UpdateDoc action.
func getActionRange(t testing.TB, testCase TestCase) (int, int) {
	startIndex := 0
	endIndex := len(testCase.Actions) - 1

	if !changeDetector.Enabled {
		return startIndex, endIndex
	}

	setupCompleteIndex := -1
	firstNonSetupIndex := -1

ActionLoop:
	for i := range testCase.Actions {
		switch testCase.Actions[i].(type) {
		case SetupComplete:
			setupCompleteIndex = i
			// We don't care about anything else if this has been explicitly provided
			break ActionLoop

		case SchemaUpdate, CreateDoc, UpdateDoc, Restart:
			continue

		default:
			firstNonSetupIndex = i
			break ActionLoop
		}
	}

	if changeDetector.SetupOnly {
		if setupCompleteIndex > -1 {
			endIndex = setupCompleteIndex
		} else if firstNonSetupIndex > -1 {
			// -1 to exclude this index
			endIndex = firstNonSetupIndex - 1
		}
	} else {
		if setupCompleteIndex > -1 {
			// +1 to exclude the SetupComplete action
			startIndex = setupCompleteIndex + 1
		} else if firstNonSetupIndex > -1 {
			// We must not set this to -1 :)
			startIndex = firstNonSetupIndex
		} else {
			// if we don't have any non-mutation actions and the change detector is enabled
			// skip this test as we will not gain anything from running (change detector would
			// run an identical profile to a normal test run)
			t.Skipf("no actions to execute")
		}
	}

	return startIndex, endIndex
}

// setStartingNodes adds a set of initial Defra nodes for the test to execute against.
//
// If a node(s) has been explicitly configured via a `ConfigureNode` action then no new
// nodes will be added.
func setStartingNodes(
	s *state,
) {
	for _, action := range s.testCase.Actions {
		switch action.(type) {
		case ConfigureNode:
			s.isNetworkEnabled = true
		}
	}

	// If nodes have not been explicitly configured via actions, setup a default one.
	if !s.isNetworkEnabled {
		st, err := setupNode(s, db.WithNodeIdentity(getIdentity(s, NodeIdentity(0))))
		require.Nil(s.t, err)
		s.nodes = append(s.nodes, st)
	}
}

func startNodes(s *state, action Start) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	// We need to restart the nodes in reverse order, to avoid dial backoff issues.
	for i := len(nodes) - 1; i >= 0; i-- {
		nodeIndex := nodeIDs[i]
		originalPath := databaseDir
		databaseDir = s.nodes[nodeIndex].dbPath
		opts := []node.Option{db.WithNodeIdentity(getIdentity(s, NodeIdentity(nodeIndex)))}
		for _, opt := range s.nodes[nodeIndex].netOpts {
			opts = append(opts, opt)
		}
		var addresses []string
		for _, addr := range s.nodes[nodeIndex].peerInfo.Addrs {
			addresses = append(addresses, addr.String())
		}
		opts = append(opts, netConfig.WithListenAddresses(addresses...))
		node, err := setupNode(s, opts...)
		require.NoError(s.t, err)
		databaseDir = originalPath
		node.p2p = s.nodes[nodeIndex].p2p
		s.nodes[nodeIndex] = node

		waitForNetworkSetupEvents(s, nodeIndex)
	}

	// If the db was restarted we need to refresh the collection definitions as the old instances
	// will reference the old (closed) database instances.
	refreshCollections(s)
}

func restartNodes(
	s *state,
) {
	if s.dbt == BadgerIMType || s.dbt == DefraIMType {
		return
	}
	closeNodes(s, Close{})
	startNodes(s, Start{})
	reconnectPeers(s)
}

// refreshCollections refreshes all the collections of the given names, preserving order.
//
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func refreshCollections(
	s *state,
) {
	for _, node := range s.nodes {
		node.collections = make([]client.Collection, len(s.collectionNames))
		allCollections, err := node.GetCollections(s.ctx, client.CollectionFetchOptions{})
		require.Nil(s.t, err)

		for i, collectionName := range s.collectionNames {
			for _, collection := range allCollections {
				if collection.Name() == collectionName {
					if _, ok := s.collectionIndexesByCollectionID[collection.Version().CollectionID]; !ok {
						// If the root is not found here this is likely the first refreshCollections
						// call of the test, we map it by root in case the collection is renamed -
						// we still wish to preserve the original index so test maintainers can reference
						// them in a convenient manner.
						s.collectionIndexesByCollectionID[collection.Version().CollectionID] = i
					}
					break
				}
			}
		}

		for _, collection := range allCollections {
			if index, ok := s.collectionIndexesByCollectionID[collection.Version().CollectionID]; ok {
				node.collections[index] = collection
			}
		}
	}
}

// configureNode configures and starts a new Defra node using the provided configuration.
//
// It returns the new node, and its peer address. Any errors generated during configuration
// will result in a test failure.
func configureNode(
	s *state,
	action ConfigureNode,
) {
	if changeDetector.Enabled {
		// We do not yet support the change detector for tests running across multiple nodes.
		s.t.SkipNow()
		return
	}

	privateKey, err := crypto.GenerateEd25519()
	require.NoError(s.t, err)

	netNodeOpts := action()
	netNodeOpts = append(netNodeOpts, netConfig.WithPrivateKey(privateKey))

	nodeOpts := []node.Option{db.WithRetryInterval([]time.Duration{time.Millisecond * 1})}
	for _, opt := range netNodeOpts {
		nodeOpts = append(nodeOpts, opt)
	}
	nodeOpts = append(nodeOpts, db.WithNodeIdentity(getIdentity(s, NodeIdentity(len(s.nodes)))))

	node, err := setupNode(s, nodeOpts...) //disable change detector, or allow it?
	require.NoError(s.t, err)

	s.nodes = append(s.nodes, node)
}

func refreshDocuments(
	s *state,
	startActionIndex int,
) {
	if len(s.nodes) == 0 {
		// This should only be possible at the moment for P2P testing, for which the
		// change detector is currently disabled.  We'll likely need some fancier logic
		// here if/when we wish to enable it.
		return
	}

	// For now just do the initial setup using the collections on the first node,
	// this may need to become more involved at a later date depending on testing
	// requirements.
	s.docIDs = make([][]client.DocID, len(s.nodes[0].collections))

	for i := range s.nodes[0].collections {
		s.docIDs[i] = []client.DocID{}
	}

	for i := 0; i < startActionIndex; i++ {
		// We need to add the existing documents in the order in which the test case lists them
		// otherwise they cannot be referenced correctly by other actions.
		switch action := s.testCase.Actions[i].(type) {
		case CreateDoc:
			nodeIDs, _ := getNodesWithIDs(action.NodeID, s.nodes)
			// Just use the collection from the first relevant node, as all will be the same for this
			// purpose.
			firstNodesID := nodeIDs[0]
			collection := s.nodes[firstNodesID].collections[action.CollectionID]

			if action.DocMap != nil {
				substituteRelations(s, action)
			}
			docs, err := parseCreateDocs(action, collection)
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			for _, doc := range docs {
				s.docIDs[action.CollectionID] = append(s.docIDs[action.CollectionID], doc.ID())
			}
		}
	}
}

func getIndexes(
	s *state,
	action GetIndexes,
) {
	if len(s.nodes) == 0 {
		return
	}

	var expectedErrorRaised bool

	nodeIDs, _ := getNodesWithIDs(action.NodeID, s.nodes)
	for _, nodeID := range nodeIDs {
		collections := s.nodes[nodeID].collections
		err := withRetryOnNode(
			s.nodes[nodeID],
			func() error {
				actualIndexes, err := collections[action.CollectionID].GetIndexes(s.ctx)
				if err != nil {
					return err
				}

				assertIndexesListsEqual(action.ExpectedIndexes,
					actualIndexes, s.t, s.testCase.Description)

				return nil
			},
		)
		expectedErrorRaised = expectedErrorRaised ||
			AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

func assertIndexesListsEqual(
	expectedIndexes []client.IndexDescription,
	actualIndexes []client.IndexDescription,
	t testing.TB,
	testDescription string,
) {
	toNames := func(indexes []client.IndexDescription) []string {
		names := make([]string, len(indexes))
		for i, index := range indexes {
			names[i] = index.Name
		}
		return names
	}

	require.ElementsMatch(t, toNames(expectedIndexes), toNames(actualIndexes), testDescription)

	toMap := func(indexes []client.IndexDescription) map[string]client.IndexDescription {
		resultMap := map[string]client.IndexDescription{}
		for _, index := range indexes {
			resultMap[index.Name] = index
		}
		return resultMap
	}

	expectedMap := toMap(expectedIndexes)
	actualMap := toMap(actualIndexes)
	for key := range expectedMap {
		assertIndexesEqual(expectedMap[key], actualMap[key], t)
	}
}

func assertIndexesEqual(expectedIndex, actualIndex client.IndexDescription, t testing.TB) {
	assert.Equal(t, expectedIndex.Name, actualIndex.Name, "index name mismatch")
	assert.Equal(t, expectedIndex.ID, actualIndex.ID, "index id mismatch")

	toNames := func(fields []client.IndexedFieldDescription) []string {
		names := make([]string, len(fields))
		for i, field := range fields {
			names[i] = field.Name
		}
		return names
	}

	require.ElementsMatch(t, toNames(expectedIndex.Fields), toNames(actualIndex.Fields), "index fields' names mismatch")

	toMap := func(fields []client.IndexedFieldDescription) map[string]client.IndexedFieldDescription {
		resultMap := map[string]client.IndexedFieldDescription{}
		for _, field := range fields {
			resultMap[field.Name] = field
		}
		return resultMap
	}

	expectedMap := toMap(expectedIndex.Fields)
	actualMap := toMap(actualIndex.Fields)
	for key := range expectedMap {
		assert.Equal(t, expectedMap[key], actualMap[key], "index fields' values mismatch")
	}
}

// updateSchema updates the schema using the given details.
func updateSchema(
	s *state,
	action SchemaUpdate,
) {
	// Do some sanitation checks if PolicyIDs are to be substituted, and error out early if invalid usage.
	if len(action.Replace) > 0 {
		for substituteLabel := range action.Replace {
			if substituteLabel == "" {
				require.Fail(s.t, "Empty substitution label.", s.testCase.Description)
			}

			howManyLabelsToSub := strings.Count(action.Schema, substituteLabel)
			if howManyLabelsToSub == 0 {
				require.Fail(
					s.t,
					"Can't do substitution because no label: "+substituteLabel,
					s.testCase.Description,
				)
			}
		}
	}

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		// This schema might be modified if the caller needs some substitution magic done.
		var modifiedSchema = action.Schema

		// We need to substitute the policyIDs into the `%policyID% place holders.
		if len(action.Replace) > 0 {
			nodeID := nodeIDs[index]
			nodesPolicyIDs := s.policyIDs[nodeID]
			templateData := map[string]string{}
			// Build template with the replacing values.
			for substituteLabel, replaceWith := range action.Replace {
				replacer, err := replaceWith.Replacer(nodesPolicyIDs)
				require.NoError(s.t, err)
				templateData[substituteLabel] = replacer
			}

			// Template should be built now, so execute it.
			tmpl := template.Must(template.New("schema").Parse(modifiedSchema))
			var renderedSchema bytes.Buffer
			err := tmpl.Execute(&renderedSchema, templateData)
			if err != nil {
				require.Fail(s.t, "Template execution for schema update failed.", s.testCase.Description)
			}

			modifiedSchema = renderedSchema.String()
		}

		results, err := node.AddSchema(s.ctx, modifiedSchema)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if action.ExpectedResults != nil {
			assertCollectionVersions(s, action.ExpectedResults, results)
		}
	}

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(s)
}

func patchSchema(
	s *state,
	action SchemaPatch,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		var setAsDefaultVersion bool
		if action.SetAsDefaultVersion.HasValue() {
			setAsDefaultVersion = action.SetAsDefaultVersion.Value()
		} else {
			setAsDefaultVersion = true
		}

		err := node.PatchSchema(s.ctx, action.Patch, action.Lens, setAsDefaultVersion)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(s)
}

func patchCollection(
	s *state,
	action PatchCollection,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		err := node.PatchCollection(s.ctx, action.Patch)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(s)
}

func getSchema(
	s *state,
	action GetSchema,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		var results []client.SchemaDescription
		var err error
		switch {
		case action.VersionID.HasValue():
			result, e := node.GetSchemaByVersionID(s.ctx, action.VersionID.Value())
			err = e
			results = []client.SchemaDescription{result}
		default:
			results, err = node.GetSchemas(
				s.ctx,
				client.SchemaFetchOptions{
					Root: action.Root,
					Name: action.Name,
				},
			)
		}

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, action.ExpectedResults, results)
		}
	}
}

func getCollections(
	s *state,
	action GetCollections,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		txn := getTransaction(s, node, action.TransactionID, "")
		ctx := db.InitContext(s.ctx, txn)
		results, err := node.GetCollections(ctx, action.FilterOptions)
		resultDescriptions := make([]client.CollectionVersion, len(results))
		for i, col := range results {
			resultDescriptions[i] = col.Version()
		}

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			assertCollectionVersions(s, action.ExpectedResults, resultDescriptions)
		}
	}
}

func setActiveSchemaVersion(
	s *state,
	action SetActiveSchemaVersion,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		err := node.SetActiveSchemaVersion(s.ctx, action.SchemaVersionID)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}

	refreshCollections(s)
}

func createView(
	s *state,
	action CreateView,
) {
	if viewType == MaterializedViewType {
		typeIndex := strings.Index(action.SDL, "\ttype ")
		subStrSquigglyIndex := strings.Index(action.SDL[typeIndex:], "{")
		squigglyIndex := typeIndex + subStrSquigglyIndex
		action.SDL = strings.Join([]string{
			action.SDL[:squigglyIndex],
			"@",
			types.MaterializedDirectiveLabel,
			action.SDL[squigglyIndex:],
			"",
		}, "")
	}

	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		_, err := node.AddView(s.ctx, action.Query, action.SDL, action.Transform)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

func refreshViews(
	s *state,
	action RefreshViews,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		err := node.RefreshViews(s.ctx, action.FilterOptions)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}
}

// createDoc creates a document using the chosen [mutationType] and caches it in the
// test state object.
func createDoc(
	s *state,
	action CreateDoc,
) {
	if action.DocMap != nil {
		substituteRelations(s, action)
	}

	var mutation func(*state, CreateDoc, client.DB, int, client.Collection) ([]client.DocID, error)
	switch mutationType {
	case CollectionSaveMutationType:
		mutation = createDocViaColSave
	case CollectionNamedMutationType:
		mutation = createDocViaColCreate
	case GQLRequestMutationType:
		mutation = createDocViaGQL
	default:
		s.t.Fatalf("invalid mutationType: %v", mutationType)
	}

	var expectedErrorRaised bool
	var docIDs []client.DocID

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]
		err := withRetryOnNode(
			node,
			func() error {
				var err error
				docIDs, err = mutation(
					s,
					action,
					node,
					nodeID,
					collection,
				)
				return err
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.CollectionID >= len(s.docIDs) {
		// Expand the slice if required, so that the document can be accessed by collection index
		s.docIDs = append(s.docIDs, make([][]client.DocID, action.CollectionID-len(s.docIDs)+1)...)
	}
	s.docIDs[action.CollectionID] = append(s.docIDs[action.CollectionID], docIDs...)

	docIDMap := make(map[string]struct{})
	for _, docID := range docIDs {
		docIDMap[docID.String()] = struct{}{}
	}

	if action.ExpectedError == "" {
		waitForUpdateEvents(s, action.NodeID, action.CollectionID, docIDMap, action.Identity)
	}
}

func createDocViaColSave(
	s *state,
	action CreateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
	docs, err := parseCreateDocs(action, collection)
	if err != nil {
		return nil, err
	}

	txn := getTransaction(s, node, immutable.None[int](), action.ExpectedError)
	ctx := makeContextForDocCreate(s, db.InitContext(s.ctx, txn), nodeIndex, &action)

	docIDs := make([]client.DocID, len(docs))
	for i, doc := range docs {
		err := collection.Save(ctx, doc, makeDocCreateOptions(&action)...)
		if err != nil {
			return nil, err
		}
		docIDs[i] = doc.ID()
	}
	return docIDs, nil
}

func makeContextForDocCreate(s *state, ctx context.Context, nodeIndex int, action *CreateDoc) context.Context {
	ctx = getContextWithIdentity(ctx, s, action.Identity, nodeIndex)
	return ctx
}

func makeDocCreateOptions(action *CreateDoc) []client.DocCreateOption {
	return []client.DocCreateOption{
		client.CreateDocEncrypted(action.IsDocEncrypted),
		client.CreateDocWithEncryptedFields(action.EncryptedFields),
	}
}

func createDocViaColCreate(
	s *state,
	action CreateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
	docs, err := parseCreateDocs(action, collection)
	if err != nil {
		return nil, err
	}

	txn := getTransaction(s, node, immutable.None[int](), action.ExpectedError)
	ctx := makeContextForDocCreate(s, db.InitContext(s.ctx, txn), nodeIndex, &action)

	switch {
	case len(docs) > 1:
		err := collection.CreateMany(ctx, docs, makeDocCreateOptions(&action)...)
		if err != nil {
			return nil, err
		}

	default:
		err := collection.Create(ctx, docs[0], makeDocCreateOptions(&action)...)
		if err != nil {
			return nil, err
		}
	}

	docIDs := make([]client.DocID, len(docs))
	for i, doc := range docs {
		docIDs[i] = doc.ID()
	}
	return docIDs, nil
}

func createDocViaGQL(
	s *state,
	action CreateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) ([]client.DocID, error) {
	var input string

	paramName := request.Input

	var err error
	if action.DocMap != nil {
		input, err = valueToGQL(action.DocMap)
	} else if client.IsJSONArray([]byte(action.Doc)) {
		var docMaps []map[string]any
		err = json.Unmarshal([]byte(action.Doc), &docMaps)
		require.NoError(s.t, err)
		input, err = arrayToGQL(docMaps)
	} else {
		input, err = jsonToGQL(action.Doc)
	}
	require.NoError(s.t, err)

	params := paramName + ": " + input

	if action.IsDocEncrypted {
		params = params + ", " + request.EncryptDocArgName + ": true"
	}
	if len(action.EncryptedFields) > 0 {
		params = params + ", " + request.EncryptFieldsArgName + ": [" +
			strings.Join(action.EncryptedFields, ", ") + "]"
	}

	key := fmt.Sprintf("create_%s", collection.Name())
	req := fmt.Sprintf(`mutation { %s(%s) { _docID } }`, key, params)

	txn := getTransaction(s, node, immutable.None[int](), action.ExpectedError)
	ctx := getContextWithIdentity(db.InitContext(s.ctx, txn), s, action.Identity, nodeIndex)

	result := node.ExecRequest(ctx, req)
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	resultData := result.GQL.Data.(map[string]any)
	resultDocs := ConvertToArrayOfMaps(s.t, resultData[key])

	docIDs := make([]client.DocID, len(resultDocs))
	for i, docMap := range resultDocs {
		docIDString := docMap[request.DocIDFieldName].(string)
		docID, err := client.NewDocIDFromString(docIDString)
		require.NoError(s.t, err)
		docIDs[i] = docID
	}

	return docIDs, nil
}

// substituteRelations scans the fields defined in [action.DocMap], if any are of type [DocIndex]
// it will substitute the [DocIndex] for the the corresponding document ID found in the state.
//
// If a document at that index is not found it will panic.
func substituteRelations(
	s *state,
	action CreateDoc,
) {
	for k, v := range action.DocMap {
		index, isIndex := v.(DocIndex)
		if !isIndex {
			continue
		}

		docID := s.docIDs[index.CollectionIndex][index.Index]
		action.DocMap[k] = docID.String()
	}
}

// deleteDoc deletes a document using the collection api and caches it in the
// given documents slice.
func deleteDoc(
	s *state,
	action DeleteDoc,
) {
	docID := s.docIDs[action.CollectionID][action.DocID]

	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		err := withRetryOnNode(
			node,
			func() error {
				_, err := collection.Delete(ctx, docID)
				return err
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.ExpectedError == "" {
		expect := map[string]struct{}{
			docID.String(): {},
		}

		waitForUpdateEvents(s, action.NodeID, action.CollectionID, expect, immutable.None[Identity]())
	}
}

// updateDoc updates a document using the chosen [mutationType].
func updateDoc(
	s *state,
	action UpdateDoc,
) {
	var mutation func(*state, UpdateDoc, client.DB, int, client.Collection) error
	switch mutationType {
	case CollectionSaveMutationType:
		mutation = updateDocViaColSave
	case CollectionNamedMutationType:
		mutation = updateDocViaColUpdate
	case GQLRequestMutationType:
		mutation = updateDocViaGQL
	default:
		s.t.Fatalf("invalid mutationType: %v", mutationType)
	}

	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]
		err := withRetryOnNode(
			node,
			func() error {
				return mutation(
					s,
					action,
					node,
					nodeID,
					collection,
				)
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.ExpectedError == "" && !action.SkipLocalUpdateEvent {
		waitForUpdateEvents(
			s,
			action.NodeID,
			action.CollectionID,
			getEventsForUpdateDoc(s, action),
			immutable.None[Identity](),
		)
	}
}

func updateDocViaColSave(
	s *state,
	action UpdateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) error {
	ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeIndex)

	doc, err := collection.Get(ctx, s.docIDs[action.CollectionID][action.DocID], true)
	if err != nil {
		return err
	}
	err = doc.SetWithJSON([]byte(action.Doc))
	if err != nil {
		return err
	}
	return collection.Save(ctx, doc)
}

func updateDocViaColUpdate(
	s *state,
	action UpdateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) error {
	ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeIndex)

	doc, err := collection.Get(ctx, s.docIDs[action.CollectionID][action.DocID], true)
	if err != nil {
		return err
	}
	err = doc.SetWithJSON([]byte(action.Doc))
	if err != nil {
		return err
	}
	return collection.Update(ctx, doc)
}

func updateDocViaGQL(
	s *state,
	action UpdateDoc,
	node client.DB,
	nodeIndex int,
	collection client.Collection,
) error {
	docID := s.docIDs[action.CollectionID][action.DocID]

	input, err := jsonToGQL(action.Doc)
	require.NoError(s.t, err)

	request := fmt.Sprintf(
		`mutation {
			update_%s(docID: "%s", input: %s) {
				_docID
			}
		}`,
		collection.Name(),
		docID.String(),
		input,
	)

	ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeIndex)

	result := node.ExecRequest(ctx, request)
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}
	return nil
}

// updateWithFilter updates the set of matched documents.
func updateWithFilter(s *state, action UpdateWithFilter) {
	var res *client.UpdateResult
	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, nodeID)
		err := withRetryOnNode(
			node,
			func() error {
				var err error
				res, err = collection.UpdateWithFilter(ctx, action.Filter, action.Updater)
				return err
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.ExpectedError == "" && !action.SkipLocalUpdateEvent {
		waitForUpdateEvents(
			s,
			action.NodeID,
			action.CollectionID,
			getEventsForUpdateWithFilter(s, action, res),
			immutable.None[Identity](),
		)
	}
}

// createIndex creates a secondary index using the collection api.
func createIndex(
	s *state,
	action CreateIndex,
) {
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]
		indexDesc := client.IndexCreateRequest{
			Name: action.IndexName,
		}
		if action.FieldName != "" {
			indexDesc.Fields = []client.IndexedFieldDescription{
				{
					Name: action.FieldName,
				},
			}
		} else if len(action.Fields) > 0 {
			for i := range action.Fields {
				indexDesc.Fields = append(indexDesc.Fields, client.IndexedFieldDescription{
					Name:       action.Fields[i].Name,
					Descending: action.Fields[i].Descending,
				})
			}
		}

		indexDesc.Unique = action.Unique
		err := withRetryOnNode(
			node,
			func() error {
				_, err := collection.CreateIndex(s.ctx, indexDesc)
				return err
			},
		)
		if AssertError(s.t, s.testCase.Description, err, action.ExpectedError) {
			return
		}
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, false)
}

// dropIndex drops the secondary index using the collection api.
func dropIndex(
	s *state,
	action DropIndex,
) {
	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		collection := s.nodes[nodeID].collections[action.CollectionID]

		err := withRetryOnNode(
			node,
			func() error {
				return collection.DropIndex(s.ctx, action.IndexName)
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// backupExport generates a backup using the db api.
func backupExport(
	s *state,
	action BackupExport,
) {
	if action.Config.Filepath == "" {
		action.Config.Filepath = s.t.TempDir() + testJSONFile
	}

	var expectedErrorRaised bool

	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		err := withRetryOnNode(
			node,
			func() error { return node.BasicExport(s.ctx, &action.Config) },
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		if !expectedErrorRaised {
			assertBackupContent(s.t, action.ExpectedContent, action.Config.Filepath)
		}
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// backupImport imports data from a backup using the db api.
func backupImport(
	s *state,
	action BackupImport,
) {
	if action.Filepath == "" {
		action.Filepath = s.t.TempDir() + testJSONFile
	}

	// we can avoid checking the error here as this would mean the filepath is invalid
	// and we want to make sure that `BasicImport` fails in this case.
	_ = os.WriteFile(action.Filepath, []byte(action.ImportContent), 0664)

	var expectedErrorRaised bool

	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		err := withRetryOnNode(
			node,
			func() error { return node.BasicImport(s.ctx, action.Filepath) },
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// withRetryOnNode attempts to perform the given action, retrying up to a DB-defined
// maximum attempt count if a transaction conflict error is returned.
//
// If a P2P-sync commit for the given document is already in progress this
// Save call can fail as the transaction will conflict. We dont want to worry
// about this in our tests so we just retry a few times until it works (or the
// retry limit is breached - important incase this is a different error)
func withRetryOnNode(
	node client.DB,
	action func() error,
) error {
	for i := 0; i < node.MaxTxnRetries(); i++ {
		err := action()
		if errors.Is(err, corekv.ErrTxnConflict) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return nil
}

func getTransaction(
	s *state,
	db client.DB,
	transactionSpecifier immutable.Option[int],
	expectedError string,
) datastore.Txn {
	if !transactionSpecifier.HasValue() {
		return nil
	}

	transactionID := transactionSpecifier.Value()

	if transactionID >= len(s.txns) {
		// Extend the txn slice so this txn can fit and be accessed by TransactionId
		s.txns = append(s.txns, make([]datastore.Txn, transactionID-len(s.txns)+1)...)
	}

	if s.txns[transactionID] == nil {
		// Create a new transaction if one does not already exist.
		txn, err := db.NewTxn(s.ctx, false)
		if AssertError(s.t, s.testCase.Description, err, expectedError) {
			txn.Discard(s.ctx)
			return nil
		}

		s.txns[transactionID] = txn
	}

	return s.txns[transactionID]
}

// commitTransaction commits the given transaction.
//
// Will panic if the given transaction does not exist. Discards the transaction if
// an error is returned on commit.
func commitTransaction(
	s *state,
	action TransactionCommit,
) {
	err := s.txns[action.TransactionID].Commit(s.ctx)
	if err != nil {
		s.txns[action.TransactionID].Discard(s.ctx)
	}

	expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// executeRequest executes the given request.
func executeRequest(
	s *state,
	action Request,
) {
	var expectedErrorRaised bool
	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.nodes)
nodeLoop:
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		txn := getTransaction(s, node, action.TransactionID, action.ExpectedError)

		ctx := getContextWithIdentity(db.InitContext(s.ctx, txn), s, action.Identity, nodeID)

		var options []client.RequestOption
		if action.OperationName.HasValue() {
			options = append(options, client.WithOperationName(action.OperationName.Value()))
		}
		if action.Variables.HasValue() {
			options = append(options, client.WithVariables(action.Variables.Value()))
		}

		if !expectedErrorRaised && viewType == MaterializedViewType {
			for _, colName := range s.collectionNames {
				// Refresh the views in the order in which they were declared, this way
				// any views of views should be based off of refreshed data, assuming they were declared in
				// an intuitive order.
				err := node.RefreshViews(
					s.ctx,
					client.CollectionFetchOptions{
						Name: immutable.Some(colName),
					},
				)
				expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
				if expectedErrorRaised {
					continue nodeLoop
				}
			}
		}

		result := node.ExecRequest(ctx, action.Request, options...)

		expectedErrorRaised = assertRequestResults(
			s,
			&result.GQL,
			action.Results,
			action.ExpectedError,
			action.Asserter,
			nodeID,
		)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// executeSubscriptionRequest executes the given subscription request, returning
// a channel that will receive a single event once the subscription has been completed.
//
// The returned channel will receive a function that asserts that
// the subscription received all its expected results and no more.
// It should be called from the main test routine to ensure that
// failures are recorded properly. It will only yield once, once
// the subscription has terminated.
func executeSubscriptionRequest(
	s *state,
	action SubscriptionRequest,
) {
	subscriptionAssert := make(chan func())

	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		result := node.ExecRequest(s.ctx, action.Request)
		if AssertErrors(s.t, s.testCase.Description, result.GQL.Errors, action.ExpectedError) {
			return
		}

		go func() {
			var results []*client.GQLResult
			allActionsAreDone := false
			for !allActionsAreDone || len(results) < len(action.Results) {
				select {
				case s := <-result.Subscription:
					results = append(results, &s)
				case <-time.After(100 * time.Millisecond):
				}
				select {
				case <-s.allActionsDone:
					allActionsAreDone = true
				case <-time.After(100 * time.Millisecond):
				}
			}

			subscriptionAssert <- func() {
				for i, r := range action.Results {
					// This assert should be executed from the main test routine
					// so that failures will be properly handled.
					expectedErrorRaised := assertRequestResults(
						s,
						results[i],
						r,
						action.ExpectedError,
						nil,
						0,
					)

					assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
				}
			}
		}()
	}

	s.subscriptionResultsChans = append(s.subscriptionResultsChans, subscriptionAssert)
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertError(t testing.TB, description string, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		require.NoError(t, err, description)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			// Must be require instead of assert, otherwise will show a fake "error not raised".
			require.ErrorIs(t, err, errors.New(expectedError))
			return false
		}
		return true
	}
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertErrors(
	t testing.TB,
	description string,
	errs []error,
	expectedError string,
) bool {
	if expectedError == "" {
		require.Empty(t, errs, description)
	} else {
		for _, e := range errs {
			// This is always a string at the moment, add support for other types as and when needed
			errorString := e.Error()
			if !strings.Contains(errorString, expectedError) {
				// We use ErrorIs for clearer failures (is a error comparison even if it is just a string)
				require.ErrorIs(t, errors.New(errorString), errors.New(expectedError))
				continue
			}
			return true
		}
	}
	return false
}

func assertRequestResults(
	s *state,
	result *client.GQLResult,
	expectedResults map[string]any,
	expectedError string,
	asserter ResultAsserter,
	nodeID int,
) bool {
	s.currentNodeID = nodeID
	// we skip assertion benchmark because you don't specify expected result for benchmark.
	if AssertErrors(s.t, s.testCase.Description, result.Errors, expectedError) || s.isBench {
		return true
	}

	if expectedResults == nil && result.Data == nil {
		return true
	}

	// Note: if result.Data == nil this panics (the panic seems useful while testing).
	resultantData := result.Data.(map[string]any)
	log.InfoContext(s.ctx, "", corelog.Any("RequestResults", result.Data))

	if asserter != nil {
		asserter.Assert(s.t, resultantData)
		return true
	}

	// merge all keys so we can check for missing values
	keys := make(map[string]struct{})
	for key := range resultantData {
		keys[key] = struct{}{}
	}
	for key := range expectedResults {
		keys[key] = struct{}{}
	}

	stack := &assertStack{}
	for key := range keys {
		stack.pushMap(key)
		expect, ok := expectedResults[key]
		require.True(s.t, ok, "expected key not found: %s", key)

		actual, ok := resultantData[key]
		require.True(s.t, ok, "result key not found: %s", key)

		switch exp := expect.(type) {
		case []map[string]any:
			actualDocs := ConvertToArrayOfMaps(s.t, actual)
			assertRequestResultDocs(
				s,
				nodeID,
				exp,
				actualDocs,
				stack,
			)

		case gomega.OmegaMatcher:
			execGomegaMatcher(exp, s, actual, stack)

		default:
			assertResultsEqual(
				s.t,
				s.clientType,
				expect,
				actual,
				fmt.Sprintf("node: %v, path: %s", nodeID, stack),
			)
		}
		stack.pop()
	}

	return false
}

func assertRequestResultDocs(
	s *state,
	nodeID int,
	expectedResults []map[string]any,
	actualResults []map[string]any,
	stack *assertStack,
) bool {
	// compare results
	require.Equal(s.t, len(expectedResults), len(actualResults),
		s.testCase.Description+" \n(number of results don't match for %s)", stack)

	for actualDocIndex, actualDoc := range actualResults {
		stack.pushArray(actualDocIndex)
		expectedDoc := expectedResults[actualDocIndex]

		require.Equal(
			s.t,
			len(expectedDoc),
			len(actualDoc),
			fmt.Sprintf(
				"%s \n(number of properties don't match for %s)",
				s.testCase.Description,
				stack,
			),
		)

		assertRequestResultDoc(s, nodeID, actualDoc, expectedDoc, stack)

		stack.pop()
	}

	return false
}

func assertRequestResultDoc(
	s *state,
	nodeID int,
	actualDoc map[string]any,
	expectedDoc map[string]any,
	stack *assertStack,
) {
	for field, actualValue := range actualDoc {
		stack.pushMap(field)

		switch expectedValue := expectedDoc[field].(type) {
		case gomega.OmegaMatcher:
			execGomegaMatcher(expectedValue, s, actualValue, stack)

		case DocIndex:
			expectedDocID := s.docIDs[expectedValue.CollectionIndex][expectedValue.Index].String()
			assertResultsEqual(
				s.t,
				s.clientType,
				expectedDocID,
				actualValue,
				fmt.Sprintf("node: %v, path: %s", nodeID, stack),
			)
		case []map[string]any:
			actualValueMap := ConvertToArrayOfMaps(s.t, actualValue)

			assertRequestResultDocs(
				s,
				nodeID,
				expectedValue,
				actualValueMap,
				stack,
			)

		case map[string]any:
			actualMap, ok := actualValue.(map[string]any)
			require.True(s.t, ok, "expected value to be a map %v. Path: %s", actualValue, stack)
			assertRequestResultDoc(s, nodeID, actualMap, expectedValue, stack)

		default:
			assertResultsEqual(
				s.t,
				s.clientType,
				expectedValue,
				actualValue,
				fmt.Sprintf("node: %v, path: %s", nodeID, stack),
			)
		}
		stack.pop()
	}
}

func ConvertToArrayOfMaps(t testing.TB, value any) []map[string]any {
	valueArrayMap, ok := value.([]map[string]any)
	if ok {
		return valueArrayMap
	}
	valueArray, ok := value.([]any)
	require.True(t, ok, "expected value to be an array of maps %v", value)

	valueArrayMap = make([]map[string]any, len(valueArray))
	for i, v := range valueArray {
		valueArrayMap[i], ok = v.(map[string]any)
		require.True(t, ok, "expected value to be an array of maps %v", value)
	}
	return valueArrayMap
}

func assertExpectedErrorRaised(t testing.TB, description string, expectedError string, wasRaised bool) {
	if expectedError != "" && !wasRaised {
		assert.Fail(t, "Expected an error however none was raised.", description)
	}
}

func assertIntrospectionResults(
	s *state,
	action IntrospectionRequest,
) bool {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		result := node.ExecRequest(s.ctx, action.Request)

		if AssertErrors(s.t, s.testCase.Description, result.GQL.Errors, action.ExpectedError) {
			return true
		}
		resultantData := result.GQL.Data.(map[string]any)

		if len(action.ExpectedData) == 0 && len(action.ContainsData) == 0 {
			require.Equal(s.t, action.ExpectedData, resultantData)
		}

		if len(action.ExpectedData) == 0 && len(action.ContainsData) > 0 {
			assertContains(s.t, action.ContainsData, resultantData)
		} else {
			require.Equal(s.t, len(action.ExpectedData), len(resultantData))

			for k, result := range resultantData {
				assert.Equal(s.t, action.ExpectedData[k], result)
			}
		}
	}

	return false
}

// Asserts that the client introspection results conform to our expectations.
func assertClientIntrospectionResults(
	s *state,
	action ClientIntrospectionRequest,
) bool {
	_, nodes := getNodesWithIDs(action.NodeID, s.nodes)
	for _, node := range nodes {
		result := node.ExecRequest(s.ctx, action.Request)

		if AssertErrors(s.t, s.testCase.Description, result.GQL.Errors, action.ExpectedError) {
			return true
		}
		resultantData := result.GQL.Data.(map[string]any)

		if len(resultantData) == 0 {
			return false
		}

		// Iterate through all types, validating each type definition.
		// Inspired from buildClientSchema.ts from graphql-js,
		// which is one way that clients do validate the schema.
		types := resultantData["__schema"].(map[string]any)["types"].([]any)

		for _, typeData := range types {
			typeDef := typeData.(map[string]any)
			kind := typeDef["kind"].(string)

			switch kind {
			case "SCALAR", "INTERFACE", "UNION", "ENUM":
				// No validation for these types in this test
			case "OBJECT":
				fields := typeDef["fields"]
				if fields == nil {
					s.t.Errorf("Fields are missing for OBJECT type %v", typeDef["name"])
				}
			case "INPUT_OBJECT":
				inputFields := typeDef["inputFields"]
				if inputFields == nil {
					s.t.Errorf("InputFields are missing for INPUT_OBJECT type %v", typeDef["name"])
				}
			default:
				// t.Errorf("Unknown type kind: %v", kind)
			}
		}
	}

	return true
}

// Asserts that the `actual` contains the given `contains` value according to the logic
// described on the [RequestTestCase.ContainsData] property.
func assertContains(t testing.TB, contains map[string]any, actual map[string]any) {
	for k, expected := range contains {
		innerActual := actual[k]
		if innerExpected, innerIsMap := expected.(map[string]any); innerIsMap {
			if innerActual == nil {
				assert.Equal(t, innerExpected, innerActual)
			} else if innerActualMap, isMap := innerActual.(map[string]any); isMap {
				// If the inner is another map then we continue down the chain
				assertContains(t, innerExpected, innerActualMap)
			} else {
				// If the types don't match then we use assert.Equal for a clean failure message
				assert.Equal(t, innerExpected, innerActual)
			}
		} else if innerExpected, innerIsArray := expected.([]any); innerIsArray {
			if actualArray, isActualArray := innerActual.([]any); isActualArray {
				// If the inner is an array/slice, then assert that each expected item is present
				// in the actual.  Note how the actual may contain additional items - this should
				// not result in a test failure.
				for _, innerExpectedItem := range innerExpected {
					assert.Contains(t, actualArray, innerExpectedItem)
				}
			} else {
				// If the types don't match then we use assert.Equal for a clean failure message
				assert.Equal(t, expected, innerActual)
			}
		} else {
			assert.Equal(t, expected, innerActual)
		}
	}
}

func assertBackupContent(t testing.TB, expectedContent, filepath string) {
	b, err := os.ReadFile(filepath)
	assert.NoError(t, err)
	assert.Equal(
		t,
		expectedContent,
		string(b),
	)
}

// skipIfMutationTypeUnsupported skips the current test if the given supportedMutationTypes option has value
// and the active mutation type is not contained within that value set.
func skipIfMutationTypeUnsupported(t testing.TB, supportedMutationTypes immutable.Option[[]MutationType]) {
	if supportedMutationTypes.HasValue() {
		var isTypeSupported bool
		for _, supportedMutationType := range supportedMutationTypes.Value() {
			if supportedMutationType == mutationType {
				isTypeSupported = true
				break
			}
		}

		if !isTypeSupported {
			t.Skipf("test does not support given mutation type. Type: %s", mutationType)
		}
	}
}

func skipIfViewCacheTypeUnsupported(t testing.TB, supportedViewTypes immutable.Option[[]ViewType]) {
	if supportedViewTypes.HasValue() {
		var isTypeSupported bool
		for _, supportedViewType := range supportedViewTypes.Value() {
			if supportedViewType == viewType {
				isTypeSupported = true
				break
			}
		}

		if !isTypeSupported {
			t.Skipf("test does not support given view cache type. Type: %s", viewType)
		}
	}
}

// skipIfClientTypeUnsupported returns a new set of client types that match the given supported set.
//
// If supportedClientTypes is none no filtering will take place and the input client set will be returned.
// If the resultant filtered set is empty the test will be skipped.
func skipIfClientTypeUnsupported(
	t testing.TB,
	clients []ClientType,
	supportedClientTypes immutable.Option[[]ClientType],
) []ClientType {
	if !supportedClientTypes.HasValue() {
		return clients
	}

	filteredClients := []ClientType{}
	for _, supportedMutationType := range supportedClientTypes.Value() {
		for _, client := range clients {
			if supportedMutationType == client {
				filteredClients = append(filteredClients, client)
				break
			}
		}
	}

	if len(filteredClients) == 0 {
		t.Skipf("test does not support any given client type. Supported Type: %v", supportedClientTypes.Value())
	}

	return filteredClients
}

func skipIfDocumentACPTypeUnsupported(t testing.TB, supportedACPTypes immutable.Option[[]DocumentACPType]) {
	if supportedACPTypes.HasValue() {
		var isTypeSupported bool
		for _, supportedType := range supportedACPTypes.Value() {
			if supportedType == documentACPType {
				isTypeSupported = true
				break
			}
		}

		if !isTypeSupported {
			t.Skipf("test does not support given acp type. Type: %s", documentACPType)
		}
	}
}

func skipIfDatabaseTypeUnsupported(
	t testing.TB,
	databases []DatabaseType,
	supportedDatabaseTypes immutable.Option[[]DatabaseType],
) []DatabaseType {
	if !supportedDatabaseTypes.HasValue() {
		return databases
	}
	filteredDatabases := []DatabaseType{}
	for _, supportedType := range supportedDatabaseTypes.Value() {
		for _, database := range databases {
			if supportedType == database {
				filteredDatabases = append(filteredDatabases, database)
				break
			}
		}
	}

	if len(filteredDatabases) == 0 {
		t.Skipf("test does not support any given database type. Supported Type: %v", supportedDatabaseTypes.Value())
	}

	return filteredDatabases
}

// skipIfNetworkTest skips the current test if the given actions
// contain network actions and skipNetworkTests is true.
func skipIfNetworkTest(t testing.TB, actions []any) {
	hasNetworkAction := false
	for _, act := range actions {
		switch act.(type) {
		case ConfigureNode:
			hasNetworkAction = true
		}
	}
	if skipNetworkTests && hasNetworkAction {
		t.Skip("test involves network actions")
	}
}

// skipIfBackupTest skips the current test if the given actions
// contain backup actions and skipBackupTests is true.
func skipIfBackupTest(t testing.TB, actions []any) {
	hasBackupAction := false
	for _, act := range actions {
		switch act.(type) {
		case BackupImport:
			hasBackupAction = true
		case BackupExport:
			hasBackupAction = true
		}
	}
	if skipBackupTests && hasBackupAction {
		t.Skip("test involves backup actions")
	}
}

// skipIfVectorEmbeddingTest skips the current test if the given actions
// contain a schema with vector embedding generation and skipVectoEmbeeddingTest is true.
func skipIfVectorEmbeddingTest(t testing.TB, actions []any) {
	hasVectorEmbedding := false
	for _, act := range actions {
		switch a := act.(type) {
		case SchemaUpdate:
			hasVectorEmbedding = strings.Contains(a.Schema, "@embedding")
		}
	}
	if !runVectorEmbeddingTests && hasVectorEmbedding {
		t.Skip("test involves vector embedding generation")
	}
}

func MustParseTime(timeString string) time.Time {
	t, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		panic(err)
	}
	return t
}

func CBORValue(value any) []byte {
	enc, err := cbor.Marshal(value)
	if err != nil {
		panic(err)
	}
	return enc
}

// parseCreateDocs parses and returns documents from a CreateDoc action.
func parseCreateDocs(action CreateDoc, collection client.Collection) ([]*client.Document, error) {
	switch {
	case action.DocMap != nil:
		val, err := client.NewDocFromMap(action.DocMap, collection.Definition())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil

	case client.IsJSONArray([]byte(action.Doc)):
		return client.NewDocsFromJSON([]byte(action.Doc), collection.Definition())

	default:
		val, err := client.NewDocFromJSON([]byte(action.Doc), collection.Definition())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil
	}
}

func performGetNodeIdentityAction(s *state, action GetNodeIdentity) {
	if action.NodeID >= len(s.nodes) {
		s.t.Fatalf("invalid nodeID: %v", action.NodeID)
	}

	actualIdent, err := s.nodes[action.NodeID].GetNodeIdentity(s.ctx)
	require.NoError(s.t, err, s.testCase.Description)

	expectedIdent := getIdentity(s, action.ExpectedIdentity)
	expectedRawIdent := immutable.Some(expectedIdent.IntoRawIdentity().Public())
	require.Equal(s.t, expectedRawIdent, actualIdent, "raw identity at %d mismatch", action.NodeID)
}

// execGomegaMatcher executes the given gomega matcher and asserts the result.
func execGomegaMatcher(exp gomega.OmegaMatcher, s *state, actual any, stack *assertStack) {
	traverseGomegaMatchers(exp, s, func(m TestStateMatcher) { m.SetTestState(s) })

	success, err := exp.Match(actual)
	if err != nil {
		assert.Fail(s.t, "the matcher exited with error", "Error: %s. Path: %s", err, stack)
	}

	if !success {
		assert.Fail(s.t, exp.FailureMessage(actual), "Path: %s", stack)
	}

	traverseGomegaMatchers(exp, s, func(m StatefulMatcher) {
		if !slices.Contains(s.statefulMatchers, m) {
			s.statefulMatchers = append(s.statefulMatchers, m)
		}
	})
}

// traverseGomegaMatchers traverses the given gomega matcher and calls the given function
// for each matcher found with the type T.
func traverseGomegaMatchers[T gomega.OmegaMatcher](exp gomega.OmegaMatcher, s *state, f func(T)) {
	if m, ok := exp.(T); ok {
		f(m)
		return
	}

	switch exp := exp.(type) {
	case *matchers.AndMatcher:
		for _, m := range exp.Matchers {
			traverseGomegaMatchers(m, s, f)
		}
	case *matchers.OrMatcher:
		for _, m := range exp.Matchers {
			traverseGomegaMatchers(m, s, f)
		}
	case *matchers.NotMatcher:
		traverseGomegaMatchers(exp.Matcher, s, f)
	}
}

// resetMatchers resets the state of all stateful matchers.
func resetMatchers(s *state) {
	for _, matcher := range s.statefulMatchers {
		matcher.ResetMatcherState()
	}
}

func performVerifySignatureAction(s *state, action VerifyBlockSignature) {
	_, nodes := getNodesWithIDs(immutable.None[int](), s.nodes)
	for i, node := range nodes {
		ctx := getContextWithIdentity(s.ctx, s, action.Identity, i)
		signerIdentity := getIdentity(s, immutable.Some(action.SignerIdentity))
		err := node.VerifySignature(ctx, action.Cid, signerIdentity.PublicKey)

		if action.ExpectedError != "" {
			require.Error(s.t, err, s.testCase.Description)
			require.Contains(s.t, err.Error(), action.ExpectedError, s.testCase.Description)
		} else {
			require.NoError(s.t, err, s.testCase.Description)
		}
	}
}
