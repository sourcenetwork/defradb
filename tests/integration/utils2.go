// Copyright 2022 Democratized Data Foundation
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
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bxcodec/faker/support/slice"
	"github.com/fxamacker/cbor/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/datastore"
	badgerds "github.com/sourcenetwork/defradb/datastore/badger/v4"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/request/graphql"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/gen"
	"github.com/sourcenetwork/defradb/tests/predefined"
)

const mutationTypeEnvName = "DEFRA_MUTATION_TYPE"

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

var (
	log          = corelog.NewLogger("tests.integration")
	mutationType MutationType
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
		// The http / cli client will return an error instead of panicing at the moment.
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
	t *testing.T,
	testCase TestCase,
) {
	flattenActions(&testCase)
	collectionNames := getCollectionNames(testCase)
	changeDetector.PreTestChecks(t, collectionNames)
	skipIfMutationTypeUnsupported(t, testCase.SupportedMutationTypes)

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

	var databases []DatabaseType
	if badgerInMemory {
		databases = append(databases, badgerIMType)
	}
	if badgerFile {
		databases = append(databases, badgerFileType)
	}
	if inMemoryStore {
		databases = append(databases, defraIMType)
	}

	// Assert that these are not empty to protect against accidental mis-configurations,
	// otherwise an empty set would silently pass all the tests.
	require.NotEmpty(t, databases)
	require.NotEmpty(t, clients)

	ctx := context.Background()
	for _, ct := range clients {
		for _, dbt := range databases {
			executeTestCase(ctx, t, collectionNames, testCase, dbt, ct)
		}
	}
}

func executeTestCase(
	ctx context.Context,
	t *testing.T,
	collectionNames []string,
	testCase TestCase,
	dbt DatabaseType,
	clientType ClientType,
) {
	log.InfoContext(
		ctx,
		testCase.Description,
		"database", dbt,
		"client", clientType,
		"mutationType", mutationType,
		"databaseDir", databaseDir,
		"changeDetector.Enabled", changeDetector.Enabled,
		"changeDetector.SetupOnly", changeDetector.SetupOnly,
		"changeDetector.SourceBranch", changeDetector.SourceBranch,
		"changeDetector.TargetBranch", changeDetector.TargetBranch,
		"changeDetector.Repository", changeDetector.Repository,
	)

	startActionIndex, endActionIndex := getActionRange(t, testCase)

	s := newState(ctx, t, testCase, dbt, clientType, collectionNames)
	setStartingNodes(s)

	// It is very important that the databases are always closed, otherwise resources will leak
	// as tests run.  This is particularly important for file based datastores.
	defer closeNodes(s)

	// Documents and Collections may already exist in the database if actions have been split
	// by the change detector so we should fetch them here at the start too (if they exist).
	// collections are by node (index), as they are specific to nodes.
	refreshCollections(s)
	refreshDocuments(s, startActionIndex)
	refreshIndexes(s)

	for i := startActionIndex; i <= endActionIndex; i++ {
		performAction(s, i, testCase.Actions[i])
	}

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
		restartNodes(s, actionIndex)

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

	case GetSchema:
		getSchema(s, action)

	case GetCollections:
		getCollections(s, action)

	case SetActiveSchemaVersion:
		setActiveSchemaVersion(s, action)

	case CreateView:
		createView(s, action)

	case ConfigureMigration:
		configureMigration(s, action)

	case CreateDoc:
		createDoc(s, action)

	case DeleteDoc:
		deleteDoc(s, action)

	case UpdateDoc:
		updateDoc(s, action)

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

	case Benchmark:
		benchmarkAction(s, actionIndex, action)

	case GenerateDocs:
		generateDocs(s, action)

	case CreatePredefinedDocs:
		generatePredefinedDocs(s, action)

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
		createDoc(s, CreateDoc{CollectionID: nameToInd[doc.Col.Description.Name.Value()], Doc: docJSON, NodeID: nodeID})
	}
}

func generateDocs(s *state, action GenerateDocs) {
	collections := getNodeCollections(action.NodeID, s.collections)
	defs := make([]client.CollectionDefinition, 0, len(collections[0]))
	for _, col := range collections[0] {
		if len(action.ForCollections) == 0 || slice.Contains(action.ForCollections, col.Name().Value()) {
			defs = append(defs, col.Definition())
		}
	}
	docs, err := gen.AutoGenerate(defs, action.Options...)
	if err != nil {
		s.t.Fatalf("Failed to generate docs %s", err)
	}
	createGenerateDocs(s, docs, action.NodeID)
}

func generatePredefinedDocs(s *state, action CreatePredefinedDocs) {
	collections := getNodeCollections(action.NodeID, s.collections)
	defs := make([]client.CollectionDefinition, 0, len(collections[0]))
	for _, col := range collections[0] {
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
	if s.dbt == defraIMType {
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
	// Skip the first, as that preceeds `type ` if `type ` is present,
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
) {
	for _, node := range s.nodes {
		node.Close()
	}
}

// getNodes gets the set of applicable nodes for the given nodeID.
//
// If nodeID has a value it will return that node only, otherwise all nodes will be returned.
func getNodes(nodeID immutable.Option[int], nodes []clients.Client) []clients.Client {
	if !nodeID.HasValue() {
		return nodes
	}

	return []clients.Client{nodes[nodeID.Value()]}
}

// getNodeCollections gets the set of applicable collections for the given nodeID.
//
// If nodeID has a value it will return collections for that node only, otherwise all collections across all
// nodes will be returned.
func getNodeCollections(nodeID immutable.Option[int], collections [][]client.Collection) [][]client.Collection {
	if !nodeID.HasValue() {
		return collections
	}

	return [][]client.Collection{collections[nodeID.Value()]}
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
func getActionRange(t *testing.T, testCase TestCase) (int, int) {
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
			// run an idential profile to a normal test run)
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
	hasExplicitNode := false
	for _, action := range s.testCase.Actions {
		switch action.(type) {
		case ConfigureNode:
			hasExplicitNode = true
		}
	}

	// If nodes have not been explicitly configured via actions, setup a default one.
	if !hasExplicitNode {
		db, path, err := setupDatabase(s)
		require.Nil(s.t, err)

		c, err := setupClient(s, &net.Node{DB: db})
		require.Nil(s.t, err)

		s.nodes = append(s.nodes, c)
		s.dbPaths = append(s.dbPaths, path)
	}
}

func restartNodes(
	s *state,
	actionIndex int,
) {
	if s.dbt == badgerIMType || s.dbt == defraIMType {
		return
	}
	closeNodes(s)

	// We need to restart the nodes in reverse order, to avoid dial backoff issues.
	for i := len(s.nodes) - 1; i >= 0; i-- {
		originalPath := databaseDir
		databaseDir = s.dbPaths[i]
		db, _, err := setupDatabase(s)
		require.Nil(s.t, err)
		databaseDir = originalPath

		if len(s.nodeConfigs) == 0 {
			// If there are no explicit node configuration actions the node will be
			// basic (i.e. no P2P stuff) and can be yielded now.
			c, err := setupClient(s, &net.Node{DB: db})
			require.NoError(s.t, err)
			s.nodes[i] = c
			continue
		}

		key := s.nodePrivateKeys[i]
		cfg := s.nodeConfigs[i]
		// We need to make sure the node is configured with its old address, otherwise
		// a new one may be selected and reconnnection to it will fail.
		var addresses []string
		for _, addr := range s.nodeAddresses[i].Addrs {
			addresses = append(addresses, addr.String())
		}
		cfg.Net.P2PAddresses = addresses

		var n *net.Node
		n, err = net.NewNode(
			s.ctx,
			db,
			net.WithListenAddresses(cfg.Net.P2PAddresses...),
			net.WithEnablePubSub(cfg.Net.PubSubEnabled),
			net.WithEnableRelay(cfg.Net.RelayEnabled),
			net.WithPrivateKey(key),
		)
		require.NoError(s.t, err)

		if err := n.Start(); err != nil {
			n.Close()
			require.NoError(s.t, err)
		}

		c, err := setupClient(s, n)
		require.NoError(s.t, err)
		s.nodes[i] = c
	}

	// The index of the action after the last wait action before the current restart action.
	// We wish to resume the wait clock from this point onwards.
	waitGroupStartIndex := 0
actionLoop:
	for i := actionIndex; i >= 0; i-- {
		switch s.testCase.Actions[i].(type) {
		case WaitForSync:
			// +1 as we do not wish to resume from the wait itself, but the next action
			// following it. This may be the current restart action.
			waitGroupStartIndex = i + 1
			break actionLoop
		}
	}

	for _, tc := range s.testCase.Actions {
		switch action := tc.(type) {
		case ConnectPeers:
			// Give the nodes a chance to connect to each other and learn about each other's subscribed topics.
			time.Sleep(100 * time.Millisecond)
			setupPeerWaitSync(
				s, waitGroupStartIndex, action, s.nodes[action.SourceNodeID], s.nodes[action.TargetNodeID],
			)
		case ConfigureReplicator:
			// Give the nodes a chance to connect to each other and learn about each other's subscribed topics.
			time.Sleep(100 * time.Millisecond)
			setupReplicatorWaitSync(
				s, waitGroupStartIndex, action, s.nodes[action.SourceNodeID], s.nodes[action.TargetNodeID],
			)
		}
	}

	// If the db was restarted we need to refresh the collection definitions as the old instances
	// will reference the old (closed) database instances.
	refreshCollections(s)
	refreshIndexes(s)
}

// refreshCollections refreshes all the collections of the given names, preserving order.
//
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func refreshCollections(
	s *state,
) {
	s.collections = make([][]client.Collection, len(s.nodes))

	for nodeID, node := range s.nodes {
		s.collections[nodeID] = make([]client.Collection, len(s.collectionNames))
		allCollections, err := node.GetAllCollections(s.ctx, false)
		require.Nil(s.t, err)

		for i, collectionName := range s.collectionNames {
			for _, collection := range allCollections {
				if collection.Name().Value() == collectionName {
					s.collections[nodeID][i] = collection
					break
				}
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

	cfg := action()
	db, path, err := setupDatabase(s) //disable change dector, or allow it?
	require.NoError(s.t, err)

	privateKey, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
	require.NoError(s.t, err)

	var n *net.Node
	log.InfoContext(s.ctx, "Starting P2P node", "P2P address", cfg.Net.P2PAddresses)
	n, err = net.NewNode(
		s.ctx,
		db,
		net.WithListenAddresses(cfg.Net.P2PAddresses...),
		net.WithEnablePubSub(cfg.Net.PubSubEnabled),
		net.WithEnableRelay(cfg.Net.RelayEnabled),
		net.WithPrivateKey(privateKey),
	)
	require.NoError(s.t, err)

	log.InfoContext(s.ctx, "Starting P2P node", "P2P address", n.PeerInfo())
	if err := n.Start(); err != nil {
		n.Close()
		require.NoError(s.t, err)
	}

	s.nodeAddresses = append(s.nodeAddresses, n.PeerInfo())
	s.nodeConfigs = append(s.nodeConfigs, cfg)
	s.nodePrivateKeys = append(s.nodePrivateKeys, privateKey)

	c, err := setupClient(s, n)
	require.NoError(s.t, err)

	s.nodes = append(s.nodes, c)
	s.dbPaths = append(s.dbPaths, path)
}

func refreshDocuments(
	s *state,
	startActionIndex int,
) {
	if len(s.collections) == 0 {
		// This should only be possible at the moment for P2P testing, for which the
		// change detector is currently disabled.  We'll likely need some fancier logic
		// here if/when we wish to enable it.
		return
	}

	// For now just do the initial setup using the collections on the first node,
	// this may need to become more involved at a later date depending on testing
	// requirements.
	s.documents = make([][]*client.Document, len(s.collections[0]))

	for i := range s.collections[0] {
		s.documents[i] = []*client.Document{}
	}

	for i := 0; i < startActionIndex; i++ {
		switch action := s.testCase.Actions[i].(type) {
		case CreateDoc:
			// Just use the collection from the first relevant node, as all will be the same for this
			// purpose.
			collection := getNodeCollections(action.NodeID, s.collections)[0][action.CollectionID]

			// We need to add the existing documents in the order in which the test case lists them
			// otherwise they cannot be referenced correctly by other actions.
			doc, err := client.NewDocFromJSON([]byte(action.Doc), collection.Schema())
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			// The document may have been mutated by other actions, so to be sure we have the latest
			// version without having to worry about the individual update mechanics we fetch it.
			doc, err = collection.Get(s.ctx, doc.ID(), false)
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			s.documents[action.CollectionID] = append(s.documents[action.CollectionID], doc)
		}
	}
}

func refreshIndexes(
	s *state,
) {
	if len(s.collections) == 0 {
		return
	}

	s.indexes = make([][][]client.IndexDescription, len(s.collections))

	for i, nodeCols := range s.collections {
		s.indexes[i] = make([][]client.IndexDescription, len(nodeCols))

		for j, col := range nodeCols {
			if col == nil {
				continue
			}
			colIndexes, err := col.GetIndexes(s.ctx)
			if err != nil {
				continue
			}

			s.indexes[i][j] = colIndexes
		}
	}
}

func getIndexes(
	s *state,
	action GetIndexes,
) {
	if len(s.collections) == 0 {
		return
	}

	var expectedErrorRaised bool
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		err := withRetry(
			actionNodes,
			nodeID,
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
	t *testing.T,
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
		assertIndexesEqual(expectedMap[key], actualMap[key], t, testDescription)
	}
}

func assertIndexesEqual(expectedIndex, actualIndex client.IndexDescription,
	t *testing.T,
	testDescription string,
) {
	assert.Equal(t, expectedIndex.Name, actualIndex.Name, testDescription)
	assert.Equal(t, expectedIndex.ID, actualIndex.ID, testDescription)

	toNames := func(fields []client.IndexedFieldDescription) []string {
		names := make([]string, len(fields))
		for i, field := range fields {
			names[i] = field.Name
		}
		return names
	}

	require.ElementsMatch(t, toNames(expectedIndex.Fields), toNames(actualIndex.Fields), testDescription)

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
		assert.Equal(t, expectedMap[key], actualMap[key], testDescription)
	}
}

// updateSchema updates the schema using the given details.
func updateSchema(
	s *state,
	action SchemaUpdate,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		_, err := node.AddSchema(s.ctx, action.Schema)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}

	// If the schema was updated we need to refresh the collection definitions.
	refreshCollections(s)
	refreshIndexes(s)
}

func patchSchema(
	s *state,
	action SchemaPatch,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
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
	refreshIndexes(s)
}

func getSchema(
	s *state,
	action GetSchema,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		var results []client.SchemaDescription
		var err error
		switch {
		case action.VersionID.HasValue():
			result, e := node.GetSchemaByVersionID(s.ctx, action.VersionID.Value())
			err = e
			results = []client.SchemaDescription{result}
		case action.Root.HasValue():
			results, err = node.GetSchemasByRoot(s.ctx, action.Root.Value())
		case action.Name.HasValue():
			results, err = node.GetSchemasByName(s.ctx, action.Name.Value())
		default:
			results, err = node.GetAllSchemas(s.ctx)
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
	for _, node := range getNodes(action.NodeID, s.nodes) {
		db := getStore(s, node, action.TransactionID, "")
		results, err := db.GetAllCollections(s.ctx, action.GetInactive)

		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

		if !expectedErrorRaised {
			require.Equal(s.t, len(action.ExpectedResults), len(results))

			for i, expected := range action.ExpectedResults {
				actual := results[i].Description()
				if expected.ID != 0 {
					require.Equal(s.t, expected.ID, actual.ID)
				}
				if expected.RootID != 0 {
					require.Equal(s.t, expected.RootID, actual.RootID)
				}
				if expected.SchemaVersionID != "" {
					require.Equal(s.t, expected.SchemaVersionID, actual.SchemaVersionID)
				}

				require.Equal(s.t, expected.Name, actual.Name)

				if expected.Indexes != nil || len(actual.Indexes) != 0 {
					// Dont bother asserting this if the expected is nil and the actual is nil/empty.
					// This is to say each test action from having to bother declaring an empty slice (if there are no indexes)
					require.Equal(s.t, expected.Indexes, actual.Indexes)
				}

				if expected.Sources != nil || len(actual.Sources) != 0 {
					// Dont bother asserting this if the expected is nil and the actual is nil/empty.
					// This is to say each test action from having to bother declaring an empty slice (if there are no sources)
					require.Equal(s.t, expected.Sources, actual.Sources)
				}
			}
		}
	}
}

func setActiveSchemaVersion(
	s *state,
	action SetActiveSchemaVersion,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		err := node.SetActiveSchemaVersion(s.ctx, action.SchemaVersionID)
		expectedErrorRaised := AssertError(s.t, s.testCase.Description, err, action.ExpectedError)

		assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
	}

	refreshCollections(s)
	refreshIndexes(s)
}

func createView(
	s *state,
	action CreateView,
) {
	for _, node := range getNodes(action.NodeID, s.nodes) {
		_, err := node.AddView(s.ctx, action.Query, action.SDL)
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
	var mutation func(*state, CreateDoc, client.P2P, []client.Collection) (*client.Document, error)

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
	var doc *client.Document
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		err := withRetry(
			actionNodes,
			nodeID,
			func() error {
				var err error
				doc, err = mutation(s, action, actionNodes[nodeID], collections)
				return err
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)

	if action.CollectionID >= len(s.documents) {
		// Expand the slice if required, so that the document can be accessed by collection index
		s.documents = append(s.documents, make([][]*client.Document, action.CollectionID-len(s.documents)+1)...)
	}
	s.documents[action.CollectionID] = append(s.documents[action.CollectionID], doc)
}

func createDocViaColSave(
	s *state,
	action CreateDoc,
	node client.P2P,
	collections []client.Collection,
) (*client.Document, error) {
	var err error
	doc, err := client.NewDocFromJSON([]byte(action.Doc), collections[action.CollectionID].Schema())
	if err != nil {
		return nil, err
	}

	return doc, collections[action.CollectionID].Save(s.ctx, doc)
}

func createDocViaColCreate(
	s *state,
	action CreateDoc,
	node client.P2P,
	collections []client.Collection,
) (*client.Document, error) {
	var err error
	doc, err := client.NewDocFromJSON([]byte(action.Doc), collections[action.CollectionID].Schema())
	if err != nil {
		return nil, err
	}

	return doc, collections[action.CollectionID].Create(s.ctx, doc)
}

func createDocViaGQL(
	s *state,
	action CreateDoc,
	node client.P2P,
	collections []client.Collection,
) (*client.Document, error) {
	collection := collections[action.CollectionID]

	input, err := jsonToGQL(action.Doc)
	require.NoError(s.t, err)

	request := fmt.Sprintf(
		`mutation {
			create_%s(input: %s) {
				_docID
			}
		}`,
		collection.Name().Value(),
		input,
	)

	db := getStore(s, node, immutable.None[int](), action.ExpectedError)

	result := db.ExecRequest(s.ctx, request)
	if len(result.GQL.Errors) > 0 {
		return nil, result.GQL.Errors[0]
	}

	resultantDocs, ok := result.GQL.Data.([]map[string]any)
	if !ok || len(resultantDocs) == 0 {
		return nil, nil
	}

	docIDString := resultantDocs[0]["_docID"].(string)
	docID, err := client.NewDocIDFromString(docIDString)
	require.NoError(s.t, err)

	doc, err := collection.Get(s.ctx, docID, false)
	require.NoError(s.t, err)

	return doc, nil
}

// deleteDoc deletes a document using the collection api and caches it in the
// given documents slice.
func deleteDoc(
	s *state,
	action DeleteDoc,
) {
	doc := s.documents[action.CollectionID][action.DocID]

	var expectedErrorRaised bool
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		err := withRetry(
			actionNodes,
			nodeID,
			func() error {
				_, err := collections[action.CollectionID].DeleteWithDocID(s.ctx, doc.ID())
				return err
			},
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// updateDoc updates a document using the chosen [mutationType].
func updateDoc(
	s *state,
	action UpdateDoc,
) {
	var mutation func(*state, UpdateDoc, client.P2P, []client.Collection) error

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
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		err := withRetry(
			actionNodes,
			nodeID,
			func() error { return mutation(s, action, actionNodes[nodeID], collections) },
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

func updateDocViaColSave(
	s *state,
	action UpdateDoc,
	node client.P2P,
	collections []client.Collection,
) error {
	cachedDoc := s.documents[action.CollectionID][action.DocID]

	doc, err := collections[action.CollectionID].Get(s.ctx, cachedDoc.ID(), true)
	if err != nil {
		return err
	}

	err = doc.SetWithJSON([]byte(action.Doc))
	if err != nil {
		return err
	}

	s.documents[action.CollectionID][action.DocID] = doc

	return collections[action.CollectionID].Save(s.ctx, doc)
}

func updateDocViaColUpdate(
	s *state,
	action UpdateDoc,
	node client.P2P,
	collections []client.Collection,
) error {
	cachedDoc := s.documents[action.CollectionID][action.DocID]

	doc, err := collections[action.CollectionID].Get(s.ctx, cachedDoc.ID(), true)
	if err != nil {
		return err
	}

	err = doc.SetWithJSON([]byte(action.Doc))
	if err != nil {
		return err
	}

	s.documents[action.CollectionID][action.DocID] = doc

	return collections[action.CollectionID].Update(s.ctx, doc)
}

func updateDocViaGQL(
	s *state,
	action UpdateDoc,
	node client.P2P,
	collections []client.Collection,
) error {
	doc := s.documents[action.CollectionID][action.DocID]
	collection := collections[action.CollectionID]

	input, err := jsonToGQL(action.Doc)
	require.NoError(s.t, err)

	request := fmt.Sprintf(
		`mutation {
			update_%s(docID: "%s", input: %s) {
				_docID
			}
		}`,
		collection.Name().Value(),
		doc.ID().String(),
		input,
	)

	db := getStore(s, node, immutable.None[int](), action.ExpectedError)

	result := db.ExecRequest(s.ctx, request)
	if len(result.GQL.Errors) > 0 {
		return result.GQL.Errors[0]
	}
	return nil
}

// createIndex creates a secondary index using the collection api.
func createIndex(
	s *state,
	action CreateIndex,
) {
	if action.CollectionID >= len(s.indexes) {
		// Expand the slice if required, so that the index can be accessed by collection index
		s.indexes = append(s.indexes,
			make([][][]client.IndexDescription, action.CollectionID-len(s.indexes)+1)...)
	}
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		indexDesc := client.IndexDescription{
			Name: action.IndexName,
		}
		if action.FieldName != "" {
			indexDesc.Fields = []client.IndexedFieldDescription{
				{
					Name: action.FieldName,
				},
			}
		} else if len(action.FieldsNames) > 0 {
			for i := range action.FieldsNames {
				dir := client.Ascending
				if len(action.Directions) > i {
					dir = action.Directions[i]
				}
				indexDesc.Fields = append(indexDesc.Fields, client.IndexedFieldDescription{
					Name:      action.FieldsNames[i],
					Direction: dir,
				})
			}
		}
		indexDesc.Unique = action.Unique
		err := withRetry(
			actionNodes,
			nodeID,
			func() error {
				desc, err := collections[action.CollectionID].CreateIndex(s.ctx, indexDesc)
				if err != nil {
					return err
				}
				s.indexes[nodeID][action.CollectionID] =
					append(s.indexes[nodeID][action.CollectionID], desc)
				return nil
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
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, collections := range getNodeCollections(action.NodeID, s.collections) {
		indexName := action.IndexName
		if indexName == "" {
			indexName = s.indexes[nodeID][action.CollectionID][action.IndexID].Name
		}

		err := withRetry(
			actionNodes,
			nodeID,
			func() error {
				return collections[action.CollectionID].DropIndex(s.ctx, indexName)
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
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, node := range actionNodes {
		err := withRetry(
			actionNodes,
			nodeID,
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
	actionNodes := getNodes(action.NodeID, s.nodes)
	for nodeID, node := range actionNodes {
		err := withRetry(
			actionNodes,
			nodeID,
			func() error { return node.BasicImport(s.ctx, action.Filepath) },
		)
		expectedErrorRaised = AssertError(s.t, s.testCase.Description, err, action.ExpectedError)
	}
	assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
}

// withRetry attempts to perform the given action, retrying up to a DB-defined
// maximum attempt count if a transaction conflict error is returned.
//
// If a P2P-sync commit for the given document is already in progress this
// Save call can fail as the transaction will conflict. We dont want to worry
// about this in our tests so we just retry a few times until it works (or the
// retry limit is breached - important incase this is a different error)
func withRetry(
	nodes []clients.Client,
	nodeID int,
	action func() error,
) error {
	for i := 0; i < nodes[nodeID].MaxTxnRetries(); i++ {
		err := action()
		if err != nil && errors.Is(err, badgerds.ErrTxnConflict) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return nil
}

func getStore(
	s *state,
	db client.DB,
	transactionSpecifier immutable.Option[int],
	expectedError string,
) client.Store {
	if !transactionSpecifier.HasValue() {
		return db
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

	return db.WithTxn(s.txns[transactionID])
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
	for nodeID, node := range getNodes(action.NodeID, s.nodes) {
		db := getStore(s, node, action.TransactionID, action.ExpectedError)
		result := db.ExecRequest(s.ctx, action.Request)

		anyOfByFieldKey := map[docFieldKey][]any{}
		expectedErrorRaised = assertRequestResults(
			s,
			&result.GQL,
			action.Results,
			action.ExpectedError,
			action.Asserter,
			nodeID,
			anyOfByFieldKey,
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

	for _, node := range getNodes(action.NodeID, s.nodes) {
		result := node.ExecRequest(s.ctx, action.Request)
		if AssertErrors(s.t, s.testCase.Description, result.GQL.Errors, action.ExpectedError) {
			return
		}

		go func() {
			data := []map[string]any{}
			errs := []error{}

			allActionsAreDone := false
			expectedDataRecieved := len(action.Results) == 0
			stream := result.Pub.Stream()
			for {
				select {
				case s := <-stream:
					sResult, _ := s.(client.GQLResult)
					sData, _ := sResult.Data.([]map[string]any)
					errs = append(errs, sResult.Errors...)
					data = append(data, sData...)

					if len(data) >= len(action.Results) {
						expectedDataRecieved = true
					}

				case <-s.allActionsDone:
					allActionsAreDone = true
				}

				if expectedDataRecieved && allActionsAreDone {
					finalResult := &client.GQLResult{
						Data:   data,
						Errors: errs,
					}

					subscriptionAssert <- func() {
						// This assert should be executed from the main test routine
						// so that failures will be properly handled.
						expectedErrorRaised := assertRequestResults(
							s,
							finalResult,
							action.Results,
							action.ExpectedError,
							nil,
							// anyof is not yet supported by subscription requests
							0,
							map[docFieldKey][]any{},
						)

						assertExpectedErrorRaised(s.t, s.testCase.Description, action.ExpectedError, expectedErrorRaised)
					}

					return
				}
			}
		}()
	}

	s.subscriptionResultsChans = append(s.subscriptionResultsChans, subscriptionAssert)
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertError(t *testing.T, description string, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		require.NoError(t, err, description)
		return false
	} else {
		if !strings.Contains(err.Error(), expectedError) {
			assert.ErrorIs(t, err, errors.New(expectedError))
			return false
		}
		return true
	}
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertErrors(
	t *testing.T,
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

// docFieldKey is an internal key type that wraps docIndex and fieldName
type docFieldKey struct {
	docIndex  int
	fieldName string
}

func assertRequestResults(
	s *state,
	result *client.GQLResult,
	expectedResults []map[string]any,
	expectedError string,
	asserter ResultAsserter,
	nodeID int,
	anyOfByField map[docFieldKey][]any,
) bool {
	// we skip assertion benchmark because you don't specify expected result for benchmark.
	if AssertErrors(s.t, s.testCase.Description, result.Errors, expectedError) || s.isBench {
		return true
	}

	if expectedResults == nil && result.Data == nil {
		return true
	}

	// Note: if result.Data == nil this panics (the panic seems useful while testing).
	resultantData := result.Data.([]map[string]any)

	if asserter != nil {
		asserter.Assert(s.t, resultantData)
		return true
	}

	log.InfoContext(s.ctx, "", "RequestResults", result.Data)

	// compare results
	require.Equal(s.t, len(expectedResults), len(resultantData),
		s.testCase.Description+" \n(number of results don't match)")

	for docIndex, result := range resultantData {
		expectedResult := expectedResults[docIndex]
		for field, actualValue := range result {
			expectedValue := expectedResult[field]

			switch r := expectedValue.(type) {
			case AnyOf:
				assertResultsAnyOf(s.t, s.clientType, r, actualValue)

				dfk := docFieldKey{docIndex, field}
				valueSet := anyOfByField[dfk]
				valueSet = append(valueSet, actualValue)
				anyOfByField[dfk] = valueSet
			default:
				assertResultsEqual(
					s.t,
					s.clientType,
					expectedValue,
					actualValue,
					fmt.Sprintf("node: %v, doc: %v", nodeID, docIndex),
				)
			}
		}
	}

	return false
}

func assertExpectedErrorRaised(t *testing.T, description string, expectedError string, wasRaised bool) {
	if expectedError != "" && !wasRaised {
		assert.Fail(t, "Expected an error however none was raised.", description)
	}
}

func assertIntrospectionResults(
	s *state,
	action IntrospectionRequest,
) bool {
	for _, node := range getNodes(action.NodeID, s.nodes) {
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
	for _, node := range getNodes(action.NodeID, s.nodes) {
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
func assertContains(t *testing.T, contains map[string]any, actual map[string]any) {
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

func assertBackupContent(t *testing.T, expectedContent, filepath string) {
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
func skipIfMutationTypeUnsupported(t *testing.T, supportedMutationTypes immutable.Option[[]MutationType]) {
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

func ParseSDL(gqlSDL string) (map[string]client.CollectionDefinition, error) {
	parser, err := graphql.NewParser()
	if err != nil {
		return nil, err
	}
	cols, err := parser.ParseSDL(context.Background(), gqlSDL)
	if err != nil {
		return nil, err
	}
	result := make(map[string]client.CollectionDefinition)
	for _, col := range cols {
		result[col.Description.Name.Value()] = col
	}
	return result, nil
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
