// Copyright 2026 Democratized Data Foundation
//
// This file is part of the DefraDB test suite.
//
// The DefraDB test suite is licensed under either:
//
//   (1) GNU Affero General Public License v3
//   (2) Business Source License 1.1
//
// See tests/LICENSE for details.

package tests

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/corekv"
	"github.com/sourcenetwork/corelog"
	"github.com/sourcenetwork/immutable"
	"github.com/sourcenetwork/testo/multiplier"

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	acpTypes "github.com/sourcenetwork/defradb/acp/types"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db"
	"github.com/sourcenetwork/defradb/tests/action"
	changeDetector "github.com/sourcenetwork/defradb/tests/change_detector"
	"github.com/sourcenetwork/defradb/tests/clients"
	"github.com/sourcenetwork/defradb/tests/gen"
	defraMultiplier "github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/predefined"
	"github.com/sourcenetwork/defradb/tests/state"
)

func init() {
	multiplier.Init(multipliersEnvName)
}

const (
	mutationTypeEnvName     = "DEFRA_MUTATION_TYPE"
	viewTypeEnvName         = "DEFRA_VIEW_TYPE"
	skipNetworkTestsEnvName = "DEFRA_SKIP_NETWORK_TESTS"
	vectorEmbeddingEnvName  = "DEFRA_VECTOR_EMBEDDING"
	multipliersEnvName      = "DEFRA_MULTIPLIERS"
)

// ViewType is a type alias for backward compatibility.
type ViewType = state.ViewType

const (
	CachelessViewType    = state.CachelessViewType
	MaterializedViewType = state.MaterializedViewType
)

var (
	log      = corelog.NewLogger("tests.integration")
	viewType state.ViewType
	// skipNetworkTests will skip any tests that involve network actions
	skipNetworkTests = false
	// backupUnsupportedClientTypes lists the client types whose BasicImport/BasicExport are not
	// implemented: the C client (see cbindings/wrapper.go) and the JS client (the Backup API is not
	// suitable for browser environments).
	backupUnsupportedClientTypes = []state.ClientType{state.CClientType, state.JSClientType}
	// runVectorEmbeddingTests will whether tests with vector embedding generation should be executed.
	runVectorEmbeddingTests = false
)

const (
	// subscriptionTimeout is the maximum time to wait for subscription results to be returned.
	subscriptionTimeout = 1 * time.Second
)

const testJSONFile = "/test.json"

func init() {
	// We use environment variables instead of flags `go test ./...` throws for all packages
	// that don't have the flag defined
	if value, ok := os.LookupEnv(mutationTypeEnvName); ok {
		state.ActiveMutationType = state.MutationType(value)
	} else {
		// Default to testing mutations via Collection.Save - it should be simpler and
		// faster. We assume this is desirable when not explicitly testing any particular
		// mutation type.
		state.ActiveMutationType = state.CollectionSaveMutationType
	}

	if value, ok := os.LookupEnv(viewTypeEnvName); ok {
		viewType = state.ViewType(value)
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
	applyMultipliers(t, &testCase)
	collectionNames := getCollectionNames(testCase)
	changeDetector.PreTestChecks(t, collectionNames, testCase.SkipChangeDetector)
	skipIfMutationTypeUnsupported(t, testCase.SupportedMutationTypes)
	skipIfDocumentACPTypeUnsupported(t, testCase.SupportedDocumentACPTypes)
	skipIfNetworkTest(t, testCase.Actions)
	skipIfViewCacheTypeUnsupported(t, testCase.SupportedViewTypes)
	skipIfVectorEmbeddingTest(t, testCase.Actions)

	var clients []state.ClientType
	if httpClient {
		clients = append(clients, state.HTTPClientType)
	}
	if goClient {
		clients = append(clients, state.GoClientType)
	}
	if cliClient {
		clients = append(clients, state.CLIClientType)
	}
	if jsClient {
		clients = append(clients, state.JSClientType)
	}
	if cClient {
		clients = append(clients, state.CClientType)
	}

	var databases []state.DatabaseType
	if badgerInMemory {
		databases = append(databases, BadgerIMType)
	}
	if badgerFile {
		databases = append(databases, BadgerFileType)
	}
	if inMemoryStore {
		databases = append(databases, DefraIMType)
	}
	if levelStore {
		databases = append(databases, LevelStoreType)
	}

	var kmsList []state.KMSType
	if testCase.KMS.Activated {
		kmsList = getKMSTypes()
		for _, excluded := range testCase.KMS.ExcludedTypes {
			kmsList = slices.DeleteFunc(kmsList, func(t state.KMSType) bool { return t == excluded })
		}
	}
	if len(kmsList) == 0 {
		kmsList = []state.KMSType{NoneKMSType}
	}

	// Assert that these are not empty to protect against accidental mis-configurations,
	// otherwise an empty set would silently pass all the tests.
	require.NotEmpty(t, databases)
	require.NotEmpty(t, clients)

	databases = skipIfDatabaseTypeUnsupported(t, databases, testCase.SupportedDatabaseTypes)
	clients = skipIfClientTypeUnsupported(t, clients, testCase.SupportedClientTypes)
	clients = skipIfBackupTest(t, clients, testCase.Actions)

	for _, ct := range clients {
		for _, dbt := range databases {
			for _, kms := range kmsList {
				run := func(st testing.TB) {
					// Some goroutines depend on the context to be cancelled in order to exit.
					// Defer ensures cancel runs on all exit paths, including runtime.Goexit().
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					executeTestCase(
						ctx,
						st,
						collectionNames,
						testCase,
						kms,
						dbt,
						ct,
						action.DocumentACPType,
					)
				}

				if testCase.FlakeRetries > 0 {
					passed := runTestWithRetry(t, testCase.FlakeRetries, run)
					if !passed {
						t.Errorf(
							"test failed after %d retries for client=%v db=%v kms=%v",
							testCase.FlakeRetries, ct, dbt, kms,
						)
					}
				} else {
					run(t)
				}
			}
		}
	}
}

func executeTestCase(
	ctx context.Context,
	t testing.TB,
	collectionNames []string,
	testCase TestCase,
	kms state.KMSType,
	dbt state.DatabaseType,
	clientType state.ClientType,
	documentACPType state.DocumentACPType,
) {
	logAttrs := []slog.Attr{
		corelog.Any("database", dbt),
		corelog.Any("client", clientType),
		corelog.Any("mutationType", state.ActiveMutationType),
		corelog.String("databaseDir", databaseDir),
		corelog.Bool("badgerEncryption", badgerEncryption),
		corelog.Bool("skipNetworkTests", skipNetworkTests),
		corelog.Bool("changeDetector.Enabled", changeDetector.Enabled),
		corelog.Bool("changeDetector.SetupOnly", changeDetector.SetupOnly),
		corelog.String("changeDetector.SourceBranch", changeDetector.SourceBranch),
		corelog.String("changeDetector.TargetBranch", changeDetector.TargetBranch),
		corelog.String("changeDetector.Repository", changeDetector.Repository),
	}

	skipIfUnsupportedLevelDBAction(t, dbt, testCase.Actions)

	if kms != NoneKMSType {
		logAttrs = append(logAttrs, corelog.Any("KMS", kms))
	}

	if value, ok := os.LookupEnv(multipliersEnvName); ok {
		logAttrs = append(logAttrs, corelog.String("multipliers", value))
	}

	log.InfoContext(ctx, t.Name(), logAttrs...)

	startActionIndex, endActionIndex := getActionRange(t, testCase)

	s := state.NewState(
		ctx,
		t,
		testCase.IdentityTypes,
		testCase.EnableSearchableEncryption,
		kms,
		dbt,
		clientType,
		viewType,
		documentACPType,
		collectionNames,
	)
	setStartingNodes(s, testCase)

	// It is very important that the databases are always closed, otherwise resources will leak
	// as tests run.  This is particularly important for file based datastores.
	defer closeNodes(s, Close{})

	// Documents and Collections may already exist in the database if actions have been split
	// by the change detector so we should fetch them here at the start too (if they exist).
	// collections are by node (index), as they are specific to nodes.
	refreshCollections(s, immutable.None[int](), immutable.None[state.Identity]())
	loadCollectionVersions(s)
	refreshDocuments(s, testCase, startActionIndex)

	for i := startActionIndex; i <= endActionIndex; i++ {
		performAction(s, testCase, i, testCase.Actions[i])
	}

	// In the source phase of the change detector, persist the in-memory test
	// state slices that templates depend on so the assert phase can resolve
	// {{.CollectionVersionID<N>}} to the bytes the source side produced. See
	// https://github.com/sourcenetwork/defradb/issues/4752 and the
	// `change_detector` package for details.
	if changeDetector.Enabled && changeDetector.SetupOnly {
		err := changeDetector.WriteTestState(s.T, changeDetector.TestState{
			CollectionVersions:   s.CollectionVersions,
			DocIDs:               docIDsToStrings(s),
			CollectionComposites: collectionCompositesToStrings(s),
		})
		require.NoError(s.T, err)
	}

	// Notify any active subscriptions that all requests have been sent.
	close(s.AllActionsDone)

	for _, resultsChan := range s.SubscriptionResultsChans {
		select {
		case subscriptionAssert := <-resultsChan:
			// We want to assert back in the main thread so failures get recorded properly
			subscriptionAssert()

		// a safety in case the stream hangs - we don't want the tests to run forever.
		case <-time.After(subscriptionTimeout):
			assert.Fail(t, "timeout occurred while waiting for data stream")
		}
	}

	// matchers can be instantiated not as part of the test state, but as a variable for Test... function scope
	// which will outlive all test runs (test instance of type [testUtils.TestCase]) and will be reused
	// by them. So the matchers need to be reset between the test runs.
	resetMatchers(s)
}

func performAction(
	s *state.State,
	testCase TestCase,
	actionIndex int,
	act any,
) {
	if stateful, ok := act.(action.Stateful); ok {
		stateful.SetState(s)
	}

	switch action := act.(type) {
	case action.Action:
		// [action.NewNode] is an action, so node setup runs from here too.
		action.Execute()

	case Restart:
		restartNodes(s, testCase)

	case Close:
		closeNodes(s, action)

	case Start:
		startNodes(s, testCase, action)

	case ConnectPeers:
		connectPeers(s, action)

	case DisconnectPeers:
		disconnectPeers(s, action)

	case AddReplicator:
		addReplicator(s, action)

	case DeleteReplicator:
		deleteReplicator(s, action)

	case AddCollectionSubscription:
		addCollectionSubscription(s, action)

	case DeleteCollectionSubscription:
		deleteCollectionSubscription(s, action)

	case ListP2PCollections:
		listP2PCollections(s, action)

	case AddDocumentSubscription:
		addDocumentSubscription(s, action)

	case DeleteDocumentSubscription:
		deleteDocumentSubscription(s, action)

	case ListP2PDocuments:
		listP2PDocuments(s, action)

	case SetActiveCollectionVersion:
		setActiveCollectionVersion(s, action)

	case ConfigureMigration:
		configureMigration(s, action)

	case AddDACPolicy:
		addDACPolicy(s, action)

	case AddDACActorRelationship:
		addDACActorRelationship(s, action)

	case DeleteDACActorRelationship:
		deleteDACActorRelationship(s, action)

	case ReEnableNAC:
		reEnableNAC(s, action)

	case DisableNAC:
		disableNAC(s, action)

	case AddNACActorRelationship:
		addNACActorRelationship(s, action)

	case DeleteNACActorRelationship:
		deleteNACActorRelationship(s, action)

	case GetNACStatus:
		getNACStatus(s, action)

	case DeleteDoc:
		deleteDoc(s, action)

	case DeleteWithFilter:
		deleteWithFilter(s, action)

	case UpdateWithFilter:
		updateWithFilter(s, action)

	case NewEncryptedIndex:
		newEncryptedIndex(s, action)

	case ListEncryptedIndexes:
		listEncryptedIndexes(s, action)

	case ListAllEncryptedIndexes:
		listAllEncryptedIndexes(s, action)

	case DeleteEncryptedIndex:
		deleteEncryptedIndex(s, action)

	case ExportBackup:
		exportBackup(s, action)

	case ImportBackup:
		importBackup(s, action)

	case IntrospectionRequest:
		assertIntrospectionResults(s, action)

	case ClientIntrospectionRequest:
		assertClientIntrospectionResults(s, action)

	case WaitForSync:
		waitForSync(s, action)

	case WaitForSESync:
		waitForSESync(s, action)

	case SyncDocs:
		syncDocs(s, action)

	case Benchmark:
		benchmarkAction(s, testCase, actionIndex, action)

	case GenerateDocs:
		generateDocs(s, action)

	case AddPredefinedDocs:
		generatePredefinedDocs(s, action)

	case GetNodeIdentity:
		performGetNodeIdentityAction(s, action)

	case VerifyBlockSignature:
		performVerifySignatureAction(s, action)

	case SetupComplete:
		// no-op, just continue.

	default:
		s.T.Fatalf("Unknown action type %T", action)
	}
}

func addGeneratedDocs(s *state.State, docs []gen.GeneratedDoc, nodeID immutable.Option[int]) {
	nameToInd := make(map[string]int)
	for i, name := range s.CollectionNames {
		nameToInd[name] = i
	}
	generatedDocIDs := make(map[string]string)
	for _, doc := range docs {
		collectionID := nameToInd[doc.Col.Name]
		docMap, err := doc.Doc.ToMap()
		if err != nil {
			s.T.Fatalf("Failed to generate docs %s", err)
		}
		// The generator assigns each doc a placeholder DocID, used only to wire up relations
		// between generated docs (see replaceGeneratedDocIDs below). The real DocID is derived
		// from the genesis CID when the doc is saved, so the placeholder is recorded for relation
		// lookup and then dropped from the map to avoid persisting a stale DocID.
		generatedDocID := doc.GeneratedID
		replaceGeneratedDocIDs(docMap, generatedDocIDs)
		delete(docMap, request.DocIDFieldName)

		a := &action.AddDoc{CollectionID: collectionID, DocMap: docMap, NodeID: nodeID}
		a.SetState(s)
		a.Execute()

		if generatedDocID != "" {
			s.DocIDsLock.RLock()
			docIDs := s.DocIDs[collectionID]
			generatedDocIDs[generatedDocID] = docIDs[len(docIDs)-1].String()
			s.DocIDsLock.RUnlock()
		}
	}
}

func replaceGeneratedDocIDs(docMap map[string]any, generatedDocIDs map[string]string) {
	for key, value := range docMap {
		docMap[key] = replaceGeneratedDocID(value, generatedDocIDs)
	}
}

func replaceGeneratedDocID(value any, generatedDocIDs map[string]string) any {
	switch value := value.(type) {
	case string:
		if docID, ok := generatedDocIDs[value]; ok {
			return docID
		}
	case []any:
		for i, item := range value {
			value[i] = replaceGeneratedDocID(item, generatedDocIDs)
		}
	case map[string]any:
		replaceGeneratedDocIDs(value, generatedDocIDs)
	}
	return value
}

func generateDocs(s *state.State, action GenerateDocs) {
	nodeIDs, _ := getNodesWithIDs(action.NodeID, s.Nodes)
	firstNodesID := nodeIDs[0]
	collections := s.Nodes[firstNodesID].Collections
	defs := make([]client.CollectionVersion, 0, len(collections))
	for _, collection := range collections {
		if len(action.ForCollections) == 0 || slices.Contains(action.ForCollections, collection.Name()) {
			defs = append(defs, collection.Version())
		}
	}
	docs, err := gen.AutoGenerate(s.Ctx, defs, action.Options...)
	if err != nil {
		s.T.Fatalf("Failed to generate docs %s", err)
	}
	addGeneratedDocs(s, docs, action.NodeID)
}

func generatePredefinedDocs(s *state.State, action AddPredefinedDocs) {
	nodeIDs, _ := getNodesWithIDs(action.NodeID, s.Nodes)
	firstNodesID := nodeIDs[0]
	collections := s.Nodes[firstNodesID].Collections
	defs := make([]client.CollectionVersion, 0, len(collections))
	for _, col := range collections {
		defs = append(defs, col.Version())
	}
	docs, err := predefined.Add(s.Ctx, defs, action.Docs)
	if err != nil {
		s.T.Fatalf("Failed to generate docs %s", err)
	}
	addGeneratedDocs(s, docs, action.NodeID)
}

func benchmarkAction(
	s *state.State,
	testCase TestCase,
	actionIndex int,
	bench Benchmark,
) {
	if s.DbType == DefraIMType {
		// Benchmarking makes no sense for test in-memory storage
		return
	}
	if len(bench.FocusClients) > 0 {
		isFound := false
		for _, clientType := range bench.FocusClients {
			if s.ClientType == clientType {
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
			performAction(s, testCase, actionIndex, benchCase)
		}
		return time.Since(startTime)
	}

	s.IsBench = true
	defer func() { s.IsBench = false }()

	baseElapsedTime := runBench(bench.BaseCase)
	optimizedElapsedTime := runBench(bench.OptimizedCase)

	factoredBaseTime := int64(float64(baseElapsedTime) / bench.Factor)
	assert.Greater(s.T, factoredBaseTime, optimizedElapsedTime,
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
// WARNING: This will not work with collection definitions ending in `type`, e.g. `user_type`
func getCollectionNames(testCase TestCase) []string {
	nextIndex := 0
	collectionIndexByName := map[string]int{}

	for _, a := range testCase.Actions {
		switch action := a.(type) {
		case *action.AddCollection:
			if action.ExpectedError != "" {
				// If an error is expected then no collections should result from this action
				continue
			}

			nextIndex = getCollectionNamesFromSDL(collectionIndexByName, action.SDL, nextIndex)

		case *action.AddView:
			if action.ExpectedError != "" {
				// If an error is expected then no collections should result from this action
				continue
			}

			nextIndex = getCollectionNamesFromSDL(collectionIndexByName, action.SDL, nextIndex)
		}
	}

	collectionNames := make([]string, len(collectionIndexByName))
	for name, index := range collectionIndexByName {
		collectionNames[index] = name
	}

	return collectionNames
}

func getCollectionNamesFromSDL(result map[string]int, sdl string, nextIndex int) int {
	// WARNING: This will not work with collection definitions ending in `type`, e.g. `user_type`
	splitByType := strings.Split(sdl, "type ")
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
	s *state.State,
	action Close,
) {
	_, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for _, node := range nodes {
		node.Close()
		node.Closed = true
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
func getNodesWithIDs(nodeID immutable.Option[int], nodes []*state.NodeState) ([]int, []*state.NodeState) {
	if !nodeID.HasValue() {
		indexes := make([]int, len(nodes))
		for i := range nodes {
			indexes[i] = i
		}
		return indexes, nodes
	}

	return []int{nodeID.Value()}, []*state.NodeState{nodes[nodeID.Value()]}
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

// applyMultipliers applies the active multipliers to the test actions.
// It converts actions that implement action.Action to action.Actions,
// checks if any multiplier wants to skip based on the actions,
// applies the multipliers, and maps the modified actions back to the original slice.
//
// Note: This implementation assumes multipliers only modify actions in-place and do not
// add or remove actions. If a multiplier changes the action count, the mapping will break.
func applyMultipliers(t testing.TB, testCase *TestCase) {
	actions := make(action.Actions, 0, len(testCase.Actions))
	actionIndices := make([]int, 0, len(testCase.Actions))

	for i, a := range testCase.Actions {
		if act, ok := a.(action.Action); ok {
			actions = append(actions, act)
			actionIndices = append(actionIndices, i)
		}
	}

	if len(actions) == 0 {
		return
	}

	multiplier.Skip(t, actions, testCase.MultiplierIncludes, testCase.MultiplierExcludes)

	activeMultipliers := multiplier.Get()

	// The signed-docs multiplier is incompatible with tests that create the same document
	// independently on more than one node: each node signs with its own key, so the genesis
	// composite block's CID — and therefore the document's DocID — differs per node. That
	// per-signer DocID is intended behaviour, but the cross-node assertions in these tests assume a
	// single shared DocID, so we skip them under signing rather than weaken those assertions.
	if strings.Contains(activeMultipliers, defraMultiplier.SignedDocs) &&
		createsDocsOnMultipleNodes(testCase) {
		t.Skipf("test creates documents on multiple nodes; incompatible with the %q multiplier",
			defraMultiplier.SignedDocs)
	}

	if name, ok := externalNodeMultiplierUnsupported(testCase, activeMultipliers); ok {
		t.Skipf("test supports client types %v, but the %q multiplier runs a node over HTTP",
			testCase.SupportedClientTypes.Value(), name)
	}

	modified := multiplier.Apply(actions)

	for i, idx := range actionIndices {
		testCase.Actions[idx] = modified[i]
	}

	applyTestCaseLevelMultipliers(testCase, activeMultipliers)
}

// externalNodeMultiplierUnsupported reports whether an active multiplier would run
// a node the test cannot drive, and names it.
//
// A node in another process is reached over HTTP whatever the run-wide client type,
// so a test that lists its clients without HTTP cannot run under such a multiplier.
// Listing no clients means any client will do.
func externalNodeMultiplierUnsupported(testCase *TestCase, activeNames string) (string, bool) {
	if !testCase.SupportedClientTypes.HasValue() ||
		slices.Contains(testCase.SupportedClientTypes.Value(), state.HTTPClientType) {
		return "", false
	}

	for name := range strings.SplitSeq(activeNames, ",") {
		name = strings.TrimSpace(name)
		if defraMultiplier.MakesNodeExternal(name) {
			return name, true
		}
	}

	return "", false
}

// applyTestCaseLevelMultipliers mutates TestCase fields based on the given
// comma-separated list of active multiplier names.
//
// Multipliers registered with testo can only modify actions via the
// [multiplier.Multiplier] interface. Some multipliers need to flip
// TestCase-level configuration instead (e.g. [SignedDocs] sets
// [TestCase.EnableSigning]). Each case below is a documented exception that
// cannot be expressed via action rewriting.
//
// The hook only upgrades values; it never downgrades. Tests that already
// configure the flag explicitly are unaffected.
//
// activeNames is passed in rather than read from testo's package-level state
// so this function is directly unit-testable.
// createsDocsOnMultipleNodes reports whether the test independently creates the same document on
// more than one node — an [action.AddDoc] with no explicit NodeID in a multi-node test. Such a
// document is signed (and so gets its DocID) independently on each node. See [applyMultipliers].
func createsDocsOnMultipleNodes(testCase *TestCase) bool {
	nodeCount := 0
	for _, a := range testCase.Actions {
		switch a.(type) {
		case *action.NewNode:
			nodeCount++
		}
	}
	if nodeCount <= 1 {
		return false
	}
	for _, a := range testCase.Actions {
		if addDoc, ok := a.(*action.AddDoc); ok && !addDoc.NodeID.HasValue() {
			return true
		}
	}
	return false
}

func applyTestCaseLevelMultipliers(testCase *TestCase, activeNames string) {
	for _, name := range strings.Split(activeNames, ",") {
		switch strings.TrimSpace(name) {
		case defraMultiplier.SignedDocs:
			testCase.EnableSigning = true
		}
	}
}

// getActionRange returns the index of the first action to be run, and the last.
//
// Not all processes will run all actions - if this is a change detector run they
// will be split.
//
// If a SetupComplete action is provided, the actions will be split there, if not
// they will be split at the first non AddCollection/AddDoc/UpdateDoc action.
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

		// Setup-phase actions, anything else ends the setup phase.
		case *action.AddCollection, *action.AddDoc, *action.UpdateDoc, Restart, *action.CommitTransaction:
			continue

		default:
			firstNonSetupIndex = i
			break ActionLoop
		}
	}

	// setupEnd is the index of the last setup-phase action; the assert phase runs the rest.
	setupEnd := endIndex
	if setupCompleteIndex > -1 {
		setupEnd = setupCompleteIndex
	} else if firstNonSetupIndex > -1 {
		// -1: this index starts the assert phase
		setupEnd = firstNonSetupIndex - 1
	}

	// The phases run as separate processes sharing only committed data, so a
	// transaction that is never committed or spans the split would leave them with
	// different data. Skip such tests (any action can open a transaction).
	if hasUnsplittableTransaction(testCase.Actions, setupEnd) {
		t.Skipf("skipping test with transaction(s) not committed within a single change-detector phase")
	}

	if changeDetector.SetupOnly {
		endIndex = setupEnd
	} else if setupCompleteIndex > -1 || firstNonSetupIndex > -1 {
		startIndex = setupEnd + 1
	} else {
		// if we don't have any non-mutation actions and the change detector is enabled
		// skip this test as we will not gain anything from running (change detector would
		// run an identical profile to a normal test run)
		t.Skipf("no actions to execute")
	}

	return startIndex, endIndex
}

// hasUnsplittableTransaction reports whether any transaction cannot be contained in
// a single change-detector phase: one never committed, or touched on both sides of
// setupEnd (the index of the last setup-phase action).
func hasUnsplittableTransaction(actions []any, setupEnd int) bool {
	type txnUsage struct {
		inSetup   bool
		inAssert  bool
		committed bool
	}
	usage := make(map[int]*txnUsage)

	touch := func(id, actionIndex int) *txnUsage {
		u, ok := usage[id]
		if !ok {
			u = &txnUsage{}
			usage[id] = u
		}
		if actionIndex <= setupEnd {
			u.inSetup = true
		} else {
			u.inAssert = true
		}
		return u
	}

	for i, a := range actions {
		// CommitTransaction has a plain int ID, so it is matched directly here.
		if commit, ok := a.(*action.CommitTransaction); ok {
			touch(commit.TransactionID, i).committed = true
			continue
		}
		if id, ok := actionTransactionID(a); ok {
			touch(id, i)
		}
	}

	for _, u := range usage {
		if !u.committed || (u.inSetup && u.inAssert) {
			return true
		}
	}
	return false
}

// actionTransactionID reads an action's optional TransactionID field by reflection,
// so every action that can run in a transaction is covered.
func actionTransactionID(a any) (int, bool) {
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return 0, false
	}

	field := v.FieldByName("TransactionID")
	if !field.IsValid() {
		return 0, false
	}

	txnID, ok := field.Interface().(immutable.Option[int])
	if !ok || !txnID.HasValue() {
		return 0, false
	}
	return txnID.Value(), true
}

// setStartingNodes adds a set of initial Defra nodes for the test to execute against.
//
// If a node(s) has been explicitly configured via a [action.NewNode] action then no new
// nodes will be added.
func setStartingNodes(
	s *state.State,
	testCase TestCase,
) {
	setupConfig := testCase.nodeSetupConfig()
	for _, a := range testCase.Actions {
		switch cfg := a.(type) {
		case *action.NewNode:
			// Node setup needs a few test-level settings that the action cannot
			// reach on its own.
			cfg.SetupConfig = setupConfig
			s.IsNetworkEnabled = true
		}
	}

	// If nodes have not been explicitly configured via actions, setup a default one.
	if !s.IsNetworkEnabled {
		s.CurrentSetupNodeID = 0
		nodeBuilder := action.DefaultNodeOpts(testCase.nodeSetupConfig())
		nodeBuilder.DB().SetNodeIdentity(state.GetIdentity(s, NodeIdentity(s.CurrentSetupNodeID)))
		st, err := action.SetupNode(
			s,
			acpIdentity.None,
			testCase.nodeSetupConfig(),
			nodeBuilder,
			"",
		)

		require.Nil(s.T, err)
		st.DisableP2P = true
		s.Nodes = append(s.Nodes, st)
	}
}

func startNodes(s *state.State, testCase TestCase, start Start) {
	nodeIDs, nodes := getNodesWithIDs(start.NodeID, s.Nodes)
	// We need to restart the nodes in reverse order, to avoid dial backoff issues.
	for index := len(nodes) - 1; index >= 0; index-- {
		nodeID := nodeIDs[index]

		// databaseDir points a restarting node at its existing store. Restore it
		// with a defer: node setup asserts with require, which ends the goroutine
		// on failure and would otherwise leave the path set for every later test.
		node, err := func() (*state.NodeState, error) {
			originalPath := databaseDir
			defer func() { databaseDir = originalPath }()
			databaseDir = s.Nodes[nodeID].DbPath

			s.CurrentSetupNodeID = nodeID
			p2pOpts := s.Nodes[nodeID].P2POpts
			action.WithListenAddresses(&p2pOpts, s.Nodes[nodeID].CachedAddresses...)
			opts := action.DefaultNodeOpts(testCase.nodeSetupConfig())
			opts.DB().SetNodeIdentity(state.GetIdentity(s, NodeIdentity(s.CurrentSetupNodeID)))
			opts.P2P().SetAll(p2pOpts)
			opts.SetDisableP2P(s.Nodes[nodeID].DisableP2P)
			opts.NodeACP().SetEnabled(start.EnableNAC)
			return action.SetupNode(
				s,
				getIdentityOption(s, start.Identity),
				testCase.nodeSetupConfig(),
				opts,
				s.Nodes[nodeID].Version,
			)
		}()

		expectedErrorRaised := AssertError(s.T, err, start.ExpectedError)
		assertExpectedErrorRaised(s.T, start.ExpectedError, expectedErrorRaised)
		if expectedErrorRaised {
			// If we are testing for failure on start of a node, there will be panics if we don't return
			// when there are errors, so we exit here to assert errors on start.
			return
		}

		require.Equal(s.T, start.ExpectedError, "")
		node.P2P = s.Nodes[nodeID].P2P
		s.Nodes[nodeID] = node
	}

	// If the db was restarted we need to refresh the existing tokens as the audiance value changed,
	// If we don't do this, then any existing tokens will be using the old audiance value upon restart.
	refreshTokens(s)

	// If the db was restarted we need to refresh the collection definitions as the old instances
	// will reference the old (closed) database instances.
	refreshCollections(s, immutable.None[int](), immutable.None[state.Identity]())
}

func restartNodes(
	s *state.State,
	testCase TestCase,
) {
	if s.DbType == BadgerIMType || s.DbType == DefraIMType {
		return
	}
	closeNodes(s, Close{})
	startNodes(s, testCase, Start{})
	reconnectPeers(s)
}

// refreshTokens refreshes all the existing tokens, preserving order.
func refreshTokens(
	s *state.State,
) {
	for identKey, identHolder := range s.Identities {
		identityToUpdate := identHolder.Identity
		if fullIdentityToUpdate, ok := identityToUpdate.(acpIdentity.FullIdentity); ok {
			nodeTokensToUpdate := identHolder.NodeTokens
			for nodeKey := range identHolder.NodeTokens {
				if audience := state.GetNodeAudience(s, nodeKey); audience.HasValue() {
					err := fullIdentityToUpdate.UpdateToken(
						action.AuthTokenExpiration,
						audience,
						immutable.Some(s.RemoteDACAddress),
					)
					require.NoError(s.T, err)
					nodeTokensToUpdate[nodeKey] = fullIdentityToUpdate.BearerToken()
				}
			}
			identHolder.Identity = identityToUpdate
			identHolder.NodeTokens = nodeTokensToUpdate
			s.Identities[identKey] = identHolder
		}
	}
}

// loadCollectionVersions populates s.CollectionVersions so that
// {{.CollectionVersionID<N>}} templates can be resolved against the same
// strings the source side produced under the change detector.
//
// Falls back to seedCollectionVersionsFromState when the sidecar file is
// missing, which happens with source branches that pre-date this mechanism
// or tests without a setup phase.
func loadCollectionVersions(s *state.State) {
	if changeDetector.Enabled && !changeDetector.SetupOnly {
		state, err := changeDetector.ReadTestState(s.T)
		if err == nil {
			s.CollectionVersions = state.CollectionVersions
			return
		}
		if !errors.Is(err, fs.ErrNotExist) {
			require.NoError(s.T, err)
		}
	}
	seedCollectionVersionsFromState(s)
}

// seedCollectionVersionsFromState populates s.CollectionVersions by walking
// the collection version history present in the on-disk database. It is the
// fallback used by loadCollectionVersions when no source-phase sidecar file
// is available — for example when the change detector source branch
// pre-dates the sidecar mechanism, or the test has no setup phase.
//
// Versions are walked from oldest to newest via the PreviousVersion chain so
// the indexing matches the order in which they were created.
func seedCollectionVersionsFromState(s *state.State) {
	if len(s.Nodes) == 0 {
		return
	}

	node := s.Nodes[0]

	identOption := getIdentityForRequestSpecificToNode(s, NodeIdentity(0), 0)
	getOpts := options.GetCollections().SetGetInactive(true)
	if identOption.HasValue() {
		getOpts.SetIdentity(identOption.Value())
	}
	allCols, err := node.GetCollections(s.Ctx, getOpts)
	require.NoError(s.T, err)

	colsByVersionID := make(map[string]client.Collection, len(allCols))
	for _, col := range allCols {
		colsByVersionID[col.Version().VersionID] = col
	}

	// For each active collection (canonical order in node.Collections), walk back
	// the PreviousVersion chain to root, then append from root forward.
	for _, active := range node.Collections {
		if active == nil {
			continue
		}

		var chain []string
		current := active.Version()
		for {
			chain = append(chain, current.VersionID)
			if !current.PreviousVersion.HasValue() {
				break
			}
			prevID := current.PreviousVersion.Value().SourceCollectionID
			prev, ok := colsByVersionID[prevID]
			if !ok {
				break
			}
			current = prev.Version()
		}

		for i := len(chain) - 1; i >= 0; i-- {
			appendCollectionVersionID(s, chain[i])
		}
	}
}

// appendCollectionVersionID appends a collection version ID to s.CollectionVersions
// if it is not already present.
func appendCollectionVersionID(s *state.State, versionID string) {
	if slices.Contains(s.CollectionVersions, versionID) {
		return
	}
	s.CollectionVersions = append(s.CollectionVersions, versionID)
}

// refreshCollections refreshes all the collections of the given names, preserving order.
//
// If a given collection is not present in the database the value at the corresponding
// result-index will be nil.
func refreshCollections(
	s *state.State,
	transactionID immutable.Option[int],
	identity immutable.Option[state.Identity],
) {
	nodeIDs, nodes := getNodesWithIDs(immutable.None[int](), s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]
		nodeIdentity := identity
		if !nodeIdentity.HasValue() {
			// Inject node's identity into the context and options while refreshing so the [GetCollections] call
			// doesn't fail due to lack of authorization(s) if NAC is enabled.
			nodeIdentity = NodeIdentity(nodeID)
		}
		node.Collections = make([]client.Collection, len(s.CollectionNames))
		txn := getTransaction(s, node, transactionID, "")
		ctx := db.InitContext(s.Ctx, txn)
		identOption := getIdentityForRequestSpecificToNode(s, nodeIdentity, nodeID)
		opts := options.GetCollections()
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		allCollections, err := node.GetCollections(ctx, opts)
		require.Nil(s.T, err)

		for i, collectionName := range s.CollectionNames {
			for _, collection := range allCollections {
				if collection.Name() == collectionName {
					if _, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; !ok {
						// If the root is not found here this is likely the first refreshCollections
						// call of the test, we map it by root in case the collection is renamed -
						// we still wish to preserve the original index so test maintainers can reference
						// them in a convenient manner.
						s.CollectionIndexesByCollectionID[collection.Version().CollectionID] = i
					}
					break
				}
			}
		}

		for _, collection := range allCollections {
			if index, ok := s.CollectionIndexesByCollectionID[collection.Version().CollectionID]; ok {
				node.Collections[index] = collection
			}
		}
	}
}

func refreshDocuments(
	s *state.State,
	testCase TestCase,
	startActionIndex int,
) {
	if len(s.Nodes) == 0 {
		// This should only be possible at the moment for P2P testing, for which the
		// change detector is currently disabled.  We'll likely need some fancier logic
		// here if/when we wish to enable it.
		return
	}

	// For now just do the initial setup using the collections on the first node,
	// this may need to become more involved at a later date depending on testing
	// requirements.
	s.DocIDsLock.Lock()
	s.DocIDs = make([][]client.DocID, len(s.Nodes[0].Collections))

	for i := range s.Nodes[0].Collections {
		s.DocIDs[i] = []client.DocID{}
	}
	s.DocIDsLock.Unlock()

	// Post-#4838 a document's public DocID is derived from its genesis CID and is
	// only known once the document has been saved, so it cannot be reconstructed by
	// re-parsing the source-phase AddDoc actions. When the change-detector source
	// phase persisted the authoritative DocIDs, load those and rebuild the commit
	// caches against them. Otherwise fall back to the legacy parse-based
	// reconstruction, which remains correct for source branches that pre-date the
	// genesis-CID DocID model.
	if loadDocIDsFromState(s) {
		s.DocIDsLock.RLock()
		docIDsByCollection := s.DocIDs
		s.DocIDsLock.RUnlock()
		for _, docIDs := range docIDsByCollection {
			for _, docID := range docIDs {
				rebuildDocCommitCIDs(s, 0, docID)
			}
		}
		return
	}

	for i := 0; i < startActionIndex; i++ {
		// We need to add the existing documents in the order in which the test case lists them
		// otherwise they cannot be referenced correctly by other actions.
		switch action := testCase.Actions[i].(type) {
		case *action.AddDoc:
			nodeIDs, _ := getNodesWithIDs(action.NodeID, s.Nodes)
			// Just use the collection from the first relevant node, as all will be the same for this
			// purpose.
			firstNodesID := nodeIDs[0]
			collection := s.Nodes[firstNodesID].Collections[action.CollectionID]

			if action.DocMap != nil {
				substituteRelations(s, action)
			}
			docs, err := parseAddDocs(s.Ctx, action, collection)
			if err != nil {
				// If an err has been returned, ignore it - it may be expected and if not
				// the test will fail later anyway
				continue
			}

			for _, doc := range docs {
				s.DocIDsLock.Lock()
				s.DocIDs[action.CollectionID] = append(s.DocIDs[action.CollectionID], doc.ID())
				s.DocIDsLock.Unlock()

				rebuildDocCommitCIDs(s, firstNodesID, doc.ID())
			}
		}
	}
}

// docIDsToStrings flattens s.DocIDs into a serializable [][]string for the
// change-detector sidecar, preserving the [collection][index] ordering.
func docIDsToStrings(s *state.State) [][]string {
	s.DocIDsLock.RLock()
	defer s.DocIDsLock.RUnlock()

	out := make([][]string, len(s.DocIDs))
	for col, ids := range s.DocIDs {
		out[col] = make([]string, len(ids))
		for i, id := range ids {
			out[col][i] = id.String()
		}
	}
	return out
}

// collectionCompositesToStrings extracts the collection-level composite commit
// CIDs (keyed by CollectionID) recorded on the first node, for the
// change-detector sidecar. Only branchable collections accumulate these.
func collectionCompositesToStrings(s *state.State) map[string][]string {
	if len(s.Nodes) == 0 {
		return nil
	}
	node := s.Nodes[0]

	node.CompositesLock.RLock()
	defer node.CompositesLock.RUnlock()

	out := map[string][]string{}
	for _, collection := range node.Collections {
		if collection == nil {
			continue
		}
		cids := node.Composites[collection.CollectionID()]
		if len(cids) == 0 {
			continue
		}
		strs := make([]string, len(cids))
		for i, c := range cids {
			strs[i] = c.String()
		}
		out[collection.CollectionID()] = strs
	}
	return out
}

// loadDocIDsFromState populates s.DocIDs from the DocIDs the change-detector
// source phase persisted, returning true on success. It returns false (leaving
// s.DocIDs untouched) outside the assert phase or when the source branch did not
// persist DocIDs, so the caller can fall back to legacy reconstruction.
func loadDocIDsFromState(s *state.State) bool {
	if !changeDetector.Enabled || changeDetector.SetupOnly {
		return false
	}

	cdState, err := changeDetector.ReadTestState(s.T)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			require.NoError(s.T, err)
		}
		return false
	}
	if cdState.DocIDs == nil {
		// Source branch pre-dates DocID persistence; fall back to reconstruction.
		return false
	}

	docIDs := make([][]client.DocID, len(cdState.DocIDs))
	for col, ids := range cdState.DocIDs {
		docIDs[col] = make([]client.DocID, len(ids))
		for i, idStr := range ids {
			docID, err := client.NewDocIDFromString(idStr)
			require.NoError(s.T, err)
			docIDs[col][i] = docID
		}
	}

	s.DocIDsLock.Lock()
	s.DocIDs = docIDs
	s.DocIDsLock.Unlock()

	restoreCollectionComposites(s, cdState.CollectionComposites)
	return true
}

// restoreCollectionComposites repopulates the collection-level composite commit
// cache from the change-detector sidecar so {{.CollectionCIDN}} templates resolve
// in the assert phase. Unlike document composites, these cannot be rebuilt from
// the persisted data as they are only observed via update events at creation time.
func restoreCollectionComposites(s *state.State, collectionComposites map[string][]string) {
	if len(collectionComposites) == 0 {
		return
	}

	for _, node := range s.Nodes {
		node.CompositesLock.Lock()
		if node.Composites == nil {
			node.Composites = make(map[string][]cid.Cid)
		}
		for collectionID, cidStrs := range collectionComposites {
			cids := make([]cid.Cid, len(cidStrs))
			for i, cidStr := range cidStrs {
				cids[i] = cid.MustParse(cidStr)
			}
			node.Composites[collectionID] = cids
		}
		node.CompositesLock.Unlock()
	}
}

// rebuildDocCommitCIDs repopulates the composite- and field-level commit caches
// for a single document so that {{.CID...}} and {{.FieldCID...}} references in
// later actions resolve to the CIDs the source phase produced.
func rebuildDocCommitCIDs(s *state.State, nodeIndex int, docID client.DocID) {
	node := s.Nodes[nodeIndex]

	node.CompositesLock.Lock()
	if node.Composites == nil {
		node.Composites = make(map[string][]cid.Cid)
	}
	if node.FieldCIDs == nil {
		node.FieldCIDs = make(map[string]map[string][]cid.Cid)
	}
	if node.FieldCIDs[docID.String()] == nil {
		node.FieldCIDs[docID.String()] = make(map[string][]cid.Cid)
	}
	node.CompositesLock.Unlock()

	// We fetch all commits (composite and field level) for the document in height
	// order so that they can be referenced later in the test if required.
	result := node.ExecRequest(s.Ctx, `query ($docID: [ID!]) {
		_commits(docID: $docID, order: {height: ASC}) {
			cid
			fieldName
		}
	}`, options.ExecRequest().SetVariables(map[string]any{
		"docID": []string{docID.String()},
	}))
	if len(result.GQL.Errors) > 0 {
		s.T.Fatalf("Failed to get existing commits for doc %s: %v", docID, result.GQL.Errors)
	}

	data, ok := result.GQL.Data.(map[string]any)
	if !ok {
		return
	}
	commits, ok := data["_commits"].([]map[string]any)
	if !ok {
		return
	}

	node.CompositesLock.Lock()
	defer node.CompositesLock.Unlock()
	for _, commit := range commits {
		c := cid.MustParse(commit[request.CidFieldName].(string))
		fieldName, _ := commit[request.FieldNameName].(string)
		if fieldName == request.CompositeFieldName {
			node.Composites[docID.String()] = append(node.Composites[docID.String()], c)
			continue
		}
		node.FieldCIDs[docID.String()][fieldName] = append(node.FieldCIDs[docID.String()][fieldName], c)
	}
}

func setActiveCollectionVersion(
	s *state.State,
	act SetActiveCollectionVersion,
) {
	replacedIDs := replaceMap(s, 0, []string{act.VersionID})
	versionID := replacedIDs[act.VersionID]

	nodeIDs, nodes := getNodesWithIDs(act.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.SetActiveCollectionVersion()
		identOption := getIdentityForRequestSpecificToNode(s, act.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		hadTxn := act.TransactionID.HasValue()
		if hadTxn {
			txn, err = s.GetTransaction(node, act.TransactionID)
			require.NoError(s.T, err)
			err = txn.SetActiveCollectionVersion(s.Ctx, versionID, opts)
		} else {
			err = node.SetActiveCollectionVersion(s.Ctx, versionID, opts)
		}

		expectedErrorRaised := AssertError(s.T, err, act.ExpectedError)

		assertExpectedErrorRaised(s.T, act.ExpectedError, expectedErrorRaised)
	}

	if !act.TransactionID.HasValue() {
		refreshCollections(s, immutable.None[int](), immutable.None[state.Identity]())

		// A version switch reindexes in the background; wait so a following query sees a built index.
		for _, node := range s.Nodes {
			action.WaitForNodeIndexesBuilt(s, node)
		}
	}
}

// substituteRelations scans the fields defined in [action.DocMap], if any are of type [DocIndex]
// it will substitute the [DocIndex] for the corresponding document ID found in the state.
//
// If a document at that index is not found it will panic.
func substituteRelations(
	s *state.State,
	action *action.AddDoc,
) {
	for k, v := range action.DocMap {
		index, isIndex := v.(DocIndex)
		if !isIndex {
			continue
		}

		s.DocIDsLock.RLock()
		docID := s.DocIDs[index.CollectionIndex][index.Index]
		s.DocIDsLock.RUnlock()

		action.DocMap[k] = docID.String()
	}
}

// deleteDoc deletes a document using the collection api and caches it in the
// given documents slice.
func deleteDoc(
	s *state.State,
	a DeleteDoc,
) {
	s.DocIDsLock.RLock()
	docID := s.DocIDs[a.CollectionID][a.DocID]
	s.DocIDsLock.RUnlock()

	doNotWaitForUpdate := false

	var collections []client.Collection

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)

	for index, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		txnOption := immutable.None[client.Txn]()
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			doNotWaitForUpdate = true
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]

		collections, err = action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		opts := options.DeleteDocument()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		err = withRetryOnNode(
			node,
			func() error {
				_, err := collection.DeleteDocument(s.Ctx, docID, opts)
				return err
			},
		)
		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}

	if a.ExpectedError == "" && !doNotWaitForUpdate {
		expect := map[string]struct{}{
			docID.String(): {},
		}

		waitForUpdateEvents(s, a.NodeID, a.CollectionID, expect, immutable.None[state.Identity]())
	}
}

// deleteWithFilter deletes the set of matched documents.
func deleteWithFilter(s *state.State, a DeleteWithFilter) {
	var res *client.DeleteResult
	doNotWaitForUpdate := false

	var collections []client.Collection

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)

	for index, node := range nodes {
		var txn client.Txn
		hadTxn := a.TransactionID.HasValue()
		var err error
		txnOption := immutable.None[client.Txn]()
		if hadTxn {
			doNotWaitForUpdate = true
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]
		collections, err = action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		opts := options.DeleteDocumentsWithFilter()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		err = withRetryOnNode(
			node,
			func() error {
				var err error
				res, err = collection.DeleteDocumentsWithFilter(s.Ctx, a.Filter, opts)
				return err
			},
		)

		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}

	if a.ExpectedError == "" && !a.SkipLocalUpdateEvent && !doNotWaitForUpdate {
		expect := make(map[string]struct{}, len(res.DocIDs))
		for _, docID := range res.DocIDs {
			expect[docID] = struct{}{}
		}
		waitForUpdateEvents(s, a.NodeID, a.CollectionID, expect, immutable.None[state.Identity]())
	}
}

// updateWithFilter updates the set of matched documents.
func updateWithFilter(s *state.State, a UpdateWithFilter) {
	var res *client.UpdateResult
	doNotWaitForUpdate := false

	var collections []client.Collection

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)

	for index, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		hadTxn := a.TransactionID.HasValue()
		var err error
		txnOption := immutable.None[client.Txn]()
		if hadTxn {
			doNotWaitForUpdate = true
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]
		collections, err = action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		opts := options.UpdateDocumentsWithFilter()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}
		err = withRetryOnNode(
			node,
			func() error {
				var err error
				res, err = collection.UpdateDocumentsWithFilter(s.Ctx, a.Filter, a.Updater, opts)
				return err
			},
		)

		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}

	if a.ExpectedError == "" && !a.SkipLocalUpdateEvent && !doNotWaitForUpdate {
		waitForUpdateEvents(
			s,
			a.NodeID,
			a.CollectionID,
			getEventsForUpdateWithFilter(s, a, res),
			immutable.None[state.Identity](),
		)
	}
}

func newEncryptedIndex(
	s *state.State,
	a NewEncryptedIndex,
) {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)
	for index, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		hadTxn := a.TransactionID.HasValue()
		txnOption := immutable.None[client.Txn]()
		if hadTxn {
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]

		collections, err := action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		if a.FieldName == "" {
			s.T.Fatalf("fieldName is required for encrypted index")
		}

		indexDesc := client.EncryptedIndexDescription{
			FieldName: a.FieldName,
			Type:      client.EncryptedIndexType(a.Type),
		}

		opts := options.NewEncryptedIndex()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		err = withRetryOnNode(
			node,
			func() error {
				_, err := collection.NewEncryptedIndex(s.Ctx, indexDesc, opts)
				return err
			},
		)
		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}
}

func listEncryptedIndexes(
	s *state.State,
	a ListEncryptedIndexes,
) {
	if len(s.Nodes) == 0 {
		return
	}

	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)
	for index, node := range nodes {
		nodeID := nodeIDs[index]

		opts := options.ListCollectionEncryptedIndexes()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		txnOption := immutable.None[client.Txn]()
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		collections, err := action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		err = withRetryOnNode(
			s.Nodes[nodeID],
			func() error {
				actualIndexes, err := collection.ListEncryptedIndexes(s.Ctx, opts)
				if err != nil {
					return err
				}

				require.ElementsMatch(s.T, a.ExpectedIndexes, actualIndexes,
					"Unexpected encrypted indexes")

				return nil
			},
		)
		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}
}

func listAllEncryptedIndexes(
	s *state.State,
	a ListAllEncryptedIndexes,
) {
	if len(s.Nodes) == 0 {
		return
	}

	nodeIDs, _ := getNodesWithIDs(a.NodeID, s.Nodes)
	for _, nodeID := range nodeIDs {
		opts := options.ListAllEncryptedIndexes()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		err := withRetryOnNode(
			s.Nodes[nodeID],
			func() error {
				allActualIndexes, err := s.Nodes[nodeID].ListAllEncryptedIndexes(s.Ctx, opts)
				if err != nil {
					return err
				}

				for collectionName, expectedIndexes := range a.ExpectedIndexes {
					actualIndexes, exists := allActualIndexes[collectionName]
					require.True(s.T, exists, "Collection %s should exist in actual indexes", collectionName)
					require.ElementsMatch(s.T, expectedIndexes, actualIndexes,
						"Unexpected encrypted indexes for collection %s", collectionName)
					delete(allActualIndexes, collectionName)
				}

				if len(allActualIndexes) > 0 {
					require.Fail(s.T, "Some collection have unexpected indexes", allActualIndexes)
				}

				return nil
			},
		)
		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}
}

func deleteEncryptedIndex(
	s *state.State,
	a DeleteEncryptedIndex,
) {
	nodeIDs, nodes := getNodesWithIDs(a.NodeID, s.Nodes)
	for index, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		txnOption := immutable.None[client.Txn]()
		hadTxn := a.TransactionID.HasValue()
		if hadTxn {
			txn, err = s.GetTransaction(node, a.TransactionID)
			require.NoError(s.T, err)
			txnOption = immutable.Some(txn)
		}

		nodeID := nodeIDs[index]

		collections, err := action.GetCollectionsCanonically(s, node, txnOption, a.Identity)
		if err != nil {
			expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
			assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
			continue
		}
		collection := collections[a.CollectionID]

		if a.FieldName == "" {
			s.T.Fatalf("fieldName is required for deleting encrypted index")
		}

		opts := options.DeleteEncryptedIndex()
		identOption := getIdentityForRequestSpecificToNode(s, a.Identity, nodeID)
		if identOption.HasValue() {
			opts.SetIdentity(identOption.Value())
		}

		err = withRetryOnNode(
			node,
			func() error {
				return collection.DeleteEncryptedIndex(s.Ctx, a.FieldName, opts)
			},
		)
		expectedErrorRaised := AssertError(s.T, err, a.ExpectedError)
		assertExpectedErrorRaised(s.T, a.ExpectedError, expectedErrorRaised)
	}
}

// exportBackup generates a backup using the db api.
func exportBackup(
	s *state.State,
	action ExportBackup,
) {
	if action.Config.Filepath == "" {
		action.Config.Filepath = s.T.TempDir() + testJSONFile
	}

	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for i, node := range nodes {
		opt := options.BasicExport().
			SetFormat(action.Config.Format).
			SetPretty(action.Config.Pretty).
			SetCollections(action.Config.Collections)

		err := withRetryOnNode(
			node,
			func() error { return node.BasicExport(s.Ctx, action.Config.Filepath, opt) },
		)
		expectedErrorRaised = AssertError(s.T, err, action.ExpectedError)

		if !expectedErrorRaised {
			assertBackupContent(s.T, replace(s, nodeIDs[i], action.ExpectedContent), action.Config.Filepath)
		}
	}

	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)
}

// importBackup imports data from a backup using the db api.
func importBackup(
	s *state.State,
	action ImportBackup,
) {
	if action.Filepath == "" {
		action.Filepath = s.T.TempDir() + testJSONFile
	}

	var expectedErrorRaised bool

	nodeIDs, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for i, node := range nodes {
		// we can avoid checking the error here as this would mean the filepath is invalid
		// and we want to make sure that `BasicImport` fails in this case.
		_ = os.WriteFile(action.Filepath, []byte(replace(s, nodeIDs[i], action.ImportContent)), 0664)

		err := withRetryOnNode(
			node,
			func() error { return node.BasicImport(s.Ctx, action.Filepath) },
		)
		expectedErrorRaised = AssertError(s.T, err, action.ExpectedError)
	}

	assertExpectedErrorRaised(s.T, action.ExpectedError, expectedErrorRaised)
}

// withRetryOnNode attempts to perform the given action, retrying up to a DB-defined
// maximum attempt count if a transaction conflict error is returned.
//
// If a P2P-sync commit for the given document is already in progress this
// Save call can fail as the transaction will conflict. We dont want to worry
// about this in our tests so we just retry a few times until it works (or the
// retry limit is breached - important incase this is a different error)
func withRetryOnNode(
	node clients.Client,
	action func() error,
) error {
	for i := 0; i < node.MaxTxnRetries(); i++ {
		err := action()
		// Check the contents of the error instead of the type, because it may have
		// lost its type while passing through the C binding layer.
		if err != nil && strings.Contains(err.Error(), corekv.ErrTxnConflict.Error()) {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		return err
	}
	return nil
}

func getTransaction(
	s *state.State,
	db client.TxnStore,
	transactionSpecifier immutable.Option[int],
	expectedError string,
) client.Txn {
	if !transactionSpecifier.HasValue() {
		return nil
	}

	transactionID := transactionSpecifier.Value()

	if transactionID >= len(s.Txns) {
		// Extend the txn slice so this txn can fit and be accessed by TransactionId
		s.Txns = append(s.Txns, make([]client.Txn, transactionID-len(s.Txns)+1)...)
	}

	if s.Txns[transactionID] == nil {
		// Create a new transaction if one does not already exist.
		txn, err := db.NewTxn(false)
		if AssertError(s.T, err, expectedError) {
			txn.Discard()
			return nil
		}

		s.Txns[transactionID] = txn
	}

	return s.Txns[transactionID]
}

// Asserts as to whether an error has been raised as expected (or not). If an expected
// error has been raised it will return true, returns false in all other cases.
func AssertError(t testing.TB, err error, expectedError string) bool {
	if err == nil {
		return false
	}

	if expectedError == "" {
		require.NoError(t, err)
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
	errs []error,
	expectedError string,
) bool {
	if expectedError == "" {
		require.Empty(t, errs, "Unexpected errors found")
	} else {
		require.NotEmpty(t, errs, "Expected error not found: %s", expectedError)
	}

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
	return false
}

// ConvertToArrayOfMaps converts an interface value to an array of maps.
// This is a wrapper around the action package function for backward compatibility.
func ConvertToArrayOfMaps(t testing.TB, value any) []map[string]any {
	return action.ConvertToArrayOfMaps(t, value)
}

// assertExpectedErrorRaised asserts that an expected error was raised.
// This is a wrapper around the action package function for backward compatibility.
func assertExpectedErrorRaised(t testing.TB, expectedError string, wasRaised bool) {
	action.AssertExpectedErrorRaised(t, expectedError, wasRaised)
}

func assertIntrospectionResults(
	s *state.State,
	action IntrospectionRequest,
) bool {
	_, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for _, node := range nodes {
		result := node.ExecRequest(s.Ctx, action.Request)

		if AssertErrors(s.T, result.GQL.Errors, action.ExpectedError) {
			return true
		}
		resultantData := result.GQL.Data.(map[string]any)

		if len(action.ExpectedData) == 0 && len(action.ContainsData) == 0 {
			require.Equal(s.T, action.ExpectedData, resultantData)
		}

		if len(action.ExpectedData) == 0 && len(action.ContainsData) > 0 {
			assertContains(s.T, action.ContainsData, resultantData)
		} else {
			require.Equal(s.T, len(action.ExpectedData), len(resultantData))

			for k, result := range resultantData {
				assert.Equal(s.T, action.ExpectedData[k], result)
			}
		}
	}

	return false
}

// Asserts that the client introspection results conform to our expectations.
func assertClientIntrospectionResults(
	s *state.State,
	action ClientIntrospectionRequest,
) bool {
	_, nodes := getNodesWithIDs(action.NodeID, s.Nodes)
	for _, node := range nodes {
		result := node.ExecRequest(s.Ctx, action.Request)

		if AssertErrors(s.T, result.GQL.Errors, action.ExpectedError) {
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
					s.T.Errorf("Fields are missing for OBJECT type %v", typeDef["name"])
				}
			case "INPUT_OBJECT":
				inputFields := typeDef["inputFields"]
				if inputFields == nil {
					s.T.Errorf("InputFields are missing for INPUT_OBJECT type %v", typeDef["name"])
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
func skipIfMutationTypeUnsupported(t testing.TB, supportedMutationTypes immutable.Option[[]state.MutationType]) {
	if supportedMutationTypes.HasValue() {
		var isTypeSupported bool
		for _, supportedMutationType := range supportedMutationTypes.Value() {
			if supportedMutationType == state.ActiveMutationType {
				isTypeSupported = true
				break
			}
		}

		if !isTypeSupported {
			t.Skipf("test does not support given mutation type. Type: %s", state.ActiveMutationType)
		}
	}
}

func skipIfViewCacheTypeUnsupported(t testing.TB, supportedViewTypes immutable.Option[[]ViewType]) {
	if supportedViewTypes.HasValue() {
		var isTypeSupported bool
		if slices.Contains(supportedViewTypes.Value(), viewType) {
			isTypeSupported = true
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
	clients []state.ClientType,
	supportedClientTypes immutable.Option[[]state.ClientType],
) []state.ClientType {
	if !supportedClientTypes.HasValue() {
		return clients
	}

	filteredClients := []state.ClientType{}
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

func skipIfDocumentACPTypeUnsupported(t testing.TB, supportedACPTypes immutable.Option[[]state.DocumentACPType]) {
	if supportedACPTypes.HasValue() {
		var isTypeSupported bool
		for _, supportedType := range supportedACPTypes.Value() {
			if supportedType == action.DocumentACPType {
				isTypeSupported = true
				break
			}
		}

		if !isTypeSupported {
			t.Skipf("test does not support given acp type. Type: %s", action.DocumentACPType)
		}
	}
}

func skipIfDatabaseTypeUnsupported(
	t testing.TB,
	databases []state.DatabaseType,
	supportedDatabaseTypes immutable.Option[[]state.DatabaseType],
) []state.DatabaseType {
	if !supportedDatabaseTypes.HasValue() {
		return databases
	}
	filteredDatabases := []state.DatabaseType{}
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
		case *action.NewNode:
			hasNetworkAction = true
		}
	}
	if skipNetworkTests && hasNetworkAction {
		t.Skip("test involves network actions")
	}
}

// skipIfBackupTest removes any client type that doesn't support the Backup API from clients, if the
// given actions contain backup actions. Skips the test entirely if no client type remains.
func skipIfBackupTest(t testing.TB, clients []state.ClientType, actions []any) []state.ClientType {
	hasBackupAction := false
	for _, act := range actions {
		switch act.(type) {
		case ImportBackup:
			hasBackupAction = true
		case ExportBackup:
			hasBackupAction = true
		}
	}
	if !hasBackupAction {
		return clients
	}

	filteredClients := make([]state.ClientType, 0, len(clients))
	for _, ct := range clients {
		if !slices.Contains(backupUnsupportedClientTypes, ct) {
			filteredClients = append(filteredClients, ct)
		}
	}
	if len(filteredClients) == 0 {
		t.Skip("test involves backup actions, but no selected client type supports them")
	}
	return filteredClients
}

// skipIfVectorEmbeddingTest skips the current test if the given actions
// contain a collection definition with vector embedding generation and skipVectoEmbeeddingTest is true.
func skipIfVectorEmbeddingTest(t testing.TB, actions []any) {
	hasVectorEmbedding := false
	for _, act := range actions {
		switch a := act.(type) {
		case *action.AddCollection:
			hasVectorEmbedding = strings.Contains(a.SDL, "@embedding")
		}
	}
	if !runVectorEmbeddingTests && hasVectorEmbedding {
		t.Skip("test involves vector embedding generation")
	}
}

// skipIfUnsupportedLevelDBAction skips the test if it contains an action that leveldb does
// not support.
func skipIfUnsupportedLevelDBAction(t testing.TB, dbt state.DatabaseType, actions []any) {
	if dbt != LevelStoreType {
		return
	}

	for _, act := range actions {
		switch a := act.(type) {
		case *action.Truncate:
			if a.TransactionID.HasValue() {
				// https://github.com/sourcenetwork/defradb/issues/4983
				t.Skip("explicit transactions for truncate with leveldb are not supported")
			}
		case *action.RefreshViews, *action.AddView:
			// These actions are skipped due to:
			// https://github.com/sourcenetwork/defradb/issues/4959
			t.Skip("RefreshViews does not yet support the leveldb store")
		case *action.Parallel:
			for _, inner := range a.Children {
				switch inner.(type) {
				case *action.Truncate, *action.RefreshViews, *action.AddView:
					t.Skip("write actions that acquire write locks are unsupported with leveldb in" +
						" the test framework.  Another action may lock the store by opening a transaction" +
						" after one of these acquires the write lock, but before it itself locks the leveldb" +
						" store - causing a deadlock.")
				}
			}
		}
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

// parseAddDocs parses and returns documents from an AddDoc action.
func parseAddDocs(
	ctx context.Context,
	action *action.AddDoc,
	collection client.Collection,
) ([]*client.Document, error) {
	switch {
	case action.DocMap != nil:
		val, err := client.NewDocFromMap(ctx, action.DocMap, collection.Version())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil

	case client.IsJSONArray([]byte(action.Doc)):
		return client.NewDocsFromJSON(ctx, []byte(action.Doc), collection.Version())

	default:
		val, err := client.NewDocFromJSON(ctx, []byte(action.Doc), collection.Version())
		if err != nil {
			return nil, err
		}
		return []*client.Document{val}, nil
	}
}

func performGetNodeIdentityAction(s *state.State, action GetNodeIdentity) {
	if action.NodeID >= len(s.Nodes) {
		s.T.Fatalf("invalid nodeID: %v", action.NodeID)
	}

	actualIdent, err := s.Nodes[action.NodeID].GetNodeIdentity(s.Ctx)
	require.NoError(s.T, err)

	expectedIdent := state.GetIdentity(s, action.ExpectedIdentity)
	expectedRawIdent := expectedIdent.ToPublicRawIdentity()
	expectedRawIdentOpt := immutable.Some(expectedRawIdent)
	require.Equal(s.T, expectedRawIdentOpt, actualIdent, "raw identity at %d mismatch", action.NodeID)
}

// resetMatchers resets the state of all stateful matchers.
func resetMatchers(s *state.State) {
	for _, matcher := range s.StatefulMatchers {
		matcher.ResetMatcherState()
	}
}

func performVerifySignatureAction(s *state.State, action VerifyBlockSignature) {
	_, nodes := getNodesWithIDs(immutable.None[int](), s.Nodes)
	for i, node := range nodes {
		// Check if a transaction is attached to this action. If so, we will be using it.
		var txn client.Txn
		var err error
		hadTxn := action.TransactionID.HasValue()
		if hadTxn {
			txn, err = s.GetTransaction(node, action.TransactionID)
			require.NoError(s.T, err)
		}

		actorIdentity := getIdentityForRequestSpecificToNode(s, action.Identity, i)
		opt := options.WithIdentity(options.VerifySignature(), actorIdentity)
		signerIdentity := state.GetIdentity(s, immutable.Some(action.SignerIdentity))
		cid := replace(s, i, action.Cid)

		if hadTxn {
			err = txn.VerifySignature(s.Ctx, cid, signerIdentity.PublicKey(), opt)
		} else {
			err = node.VerifySignature(s.Ctx, cid, signerIdentity.PublicKey(), opt)
		}

		if action.ExpectedError != "" {
			require.Error(s.T, err)
			require.Contains(s.T, err.Error(), action.ExpectedError)
		} else {
			require.NoError(s.T, err)
		}
	}
}

func FormatExpectedErrorWithPermission(permission acpTypes.NodeResourcePermission) string {
	return fmt.Sprintf("%s. Permission: %s", client.ErrNotAuthorizedToPerformOperation, permission.String())
}
