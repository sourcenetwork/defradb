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
	"time"

	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/crypto"
	"github.com/sourcenetwork/defradb/tests/action"
	"github.com/sourcenetwork/defradb/tests/gen"
	"github.com/sourcenetwork/defradb/tests/multiplier"
	"github.com/sourcenetwork/defradb/tests/predefined"
	"github.com/sourcenetwork/defradb/tests/state"
)

// TestCase contains the details of the test case to execute.
type TestCase struct {
	// Actions contains the set of actions and their expected results that
	// this test should execute.  They will execute in the order that they
	// are provided.
	Actions []any

	// If provided a value, SupportedMutationTypes will cause this test to be skipped
	// if the active mutation type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between mutation types, or we need to temporarily document a bug.
	SupportedMutationTypes immutable.Option[[]state.MutationType]

	// If provided a value, SupportedClientTypes will limit the client types under test to those
	// within this set.  If no active clients pass this filter the test will be skipped.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between client types, or we need to temporarily document a bug.
	SupportedClientTypes immutable.Option[[]state.ClientType]

	// If provided a value, SupportedDocumentACPTypes will cause this test to be skipped
	// if the active acp type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between acp types, or we need to temporarily document a bug.
	SupportedDocumentACPTypes immutable.Option[[]state.DocumentACPType]

	// If provided a value, SupportedACPTypes will cause this test to be skipped
	// if the active view type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between view types, or we need to temporarily document a bug.
	SupportedViewTypes immutable.Option[[]ViewType]

	// If provided a value, SupportedDatabaseTypes will cause this test to be skipped
	// if the active database type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between database types, or we need to temporarily document a bug.
	SupportedDatabaseTypes immutable.Option[[]state.DatabaseType]

	// Configuration for KMS to be used in the test
	KMS KMS

	// EnableSigning indicates if signing should be enabled for the test.
	// Use [IdentityTypes] to customize the key type that is used for identity and signing.
	EnableSigning bool

	// EnableSearchableEncryption indicates if searchable encryption should be enabled for the test.
	// When enabled, a searchable encryption key will be generated and passed to the database.
	EnableSearchableEncryption bool

	// IdentityTypes is a map of identity to key type.
	// Use it to customize the key type that is used for identity and signing.
	IdentityTypes map[state.Identity]crypto.KeyType

	// The test will be skipped if the current active set of multipliers
	// does not contain all of the given multiplier names.
	MultiplierIncludes []multiplier.Name

	// The test will be skipped if the current active set of multipliers
	// contains any of the given multiplier names.
	MultiplierExcludes []multiplier.Name

	// FlakeRetries specifies the number of times a flaky test should be retried
	// if it fails. If a test succeeds on any attempt, it is considered passed.
	// A value of 0 (default) means no retries - the test runs once as normal.
	// A value of N means the test will be attempted up to N+1 times total
	// (1 initial + N retries).
	FlakeRetries uint
}

// KMS contains the configuration for KMS to be used in the test
type KMS struct {
	// Activated indicates if the KMS should be used in the test
	Activated bool
	// ExcludedTypes specifies the KMS types that should be excluded from the test.
	// If none are specified all types will be used.
	ExcludedTypes []state.KMSType
}

// SetupComplete is a flag to explicitly notify the change detector at which point
// setup is complete so that it may split actions across database code-versions.
//
// If a SetupComplete action is not provided the change detector will split before
// the first item that is neither an AddCollection, AddDoc or UpdateDoc action.
type SetupComplete struct{}

// ConfigureNode allows the explicit configuration of new Defra nodes.
//
// If no nodes are explicitly configured, a default one will be setup.  There is no
// upper limit to the number that can be configured.
//
// Nodes may be explicitly referenced by index by other actions using `NodeID` properties.
// If the action has a `NodeID` property and it is not specified, the action will be
// effected on all nodes.
type ConfigureNode func() options.NodeP2POptions

// Restart is an action that will close and then start all nodes.
type Restart struct{}

// Close is an action that will close a node.
type Close struct {
	// NodeID may hold the ID (index) of a node to close.
	//
	// If a value is not provided the close will be applied to all nodes.
	NodeID immutable.Option[int]
}

// Todo: https://github.com/sourcenetwork/defradb/issues/3872
// Start should be improved and a bit more smart to not restart a node (also won't require Close() in that case),
// if the first action item is `Start` with flag config options. Likely should fix tests where [EnableNAC]
// is true to start node with nac from the start once this is fixed.
// Start is an action that will start a node that has been previously closed.
type Start struct {
	// NodeID may hold the ID (index) of a node to start.
	//
	// If a value is not provided the start will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of the user starting the node.
	//
	// If this identity if used in combination with enabling node acp, then this
	// Identity becomes the node acp owner for that node.
	//
	// To disable/purge/re-enable node acp after a successful start, must use the
	// respective client commands instead.
	Identity immutable.Option[state.Identity]

	// EnableNAC is true when the node is being started with an attempt to setup and enable
	// the node access control for that node.
	//
	// Must have a valid identity to enable (if enabling for the first time).
	//
	// False by default.
	EnableNAC bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// SetActiveCollectionVersion is an action that will set the active collection version to the
// given value.
type SetActiveCollectionVersion struct {
	// NodeID may hold the ID (index) of a node to set the collection version on.
	//
	// If a value is not provided the version will be set on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// VersionID to set as active collection version.
	VersionID string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// DocIndex represents a relation field value, it allows relation fields to be set without worrying
// about the specific document id.
//
// The test harness will substitute this struct for the document at the given index before
// performing the host action.
//
// The targeted document must have been defined in an action prior to the action that this index
// is hosted upon.
// This is a type alias for backward compatibility.
type DocIndex = action.DocIndex

// NewDocIndex creates a new [DocIndex] instance allowing relation fields to be referenced without worrying
// about the specific document id.
func NewDocIndex(collectionIndex int, index int) DocIndex {
	return DocIndex{
		CollectionIndex: collectionIndex,
		Index:           index,
	}
}

// DeleteDoc will attempt to delete the given document in the given collection
// using the collection api.
type DeleteDoc struct {
	// NodeID may hold the ID (index) of a node to apply this delete to.
	//
	// If a value is not provided the document will be deleted in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only delete public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also delete private document(s) that are owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection in which this document should be deleted.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// UpdateDoc will attempt to update the given document using the set [state.MutationType].
type UpdateDoc struct {
	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only update public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also update private document(s) that are owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection in which this document exists.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int

	// The document update, in JSON string format. Will only update the properties
	// provided.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Skip waiting for an update event on the local event bus.
	//
	// This should only be used for tests that do not correctly
	// publish an update event to the local event bus.
	SkipLocalUpdateEvent bool

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// UpdateWithFilter will update the set of documents that match the given filter.
type UpdateWithFilter struct {
	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only update public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also update private document(s) that are owned by this Identity.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection in which this document exists.
	CollectionID int

	// The filter to match documents against.
	Filter any

	// The update to apply to matched documents.
	Updater string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Skip waiting for an update event on the local event bus.
	//
	// This should only be used for tests that do not correctly
	// publish an update event to the local event bus.
	SkipLocalUpdateEvent bool

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// NewEncryptedIndex will attempt to make a new encrypted index on the given collection
// using the collection api.
type NewEncryptedIndex struct {
	// NodeID may hold the ID (index) of a node to make the new encrypted index on.
	//
	// If a value is not provided the index will be made on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection on which this index should be made.
	CollectionID int

	// The name of the field to index. Used only for single field indexes.
	FieldName string

	// The type of new index to make.
	Type string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// ListEncryptedIndexes will attempt to list encrypted index from the given collection
// using the collection api.
type ListEncryptedIndexes struct {
	// NodeID may hold the ID (index) of a node to list the encrypted index on.
	//
	// If a value is not provided the encrypted indexes will be retrieved from the first nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which this encrypted indexes should be retrieved.
	CollectionID int

	// The expected encrypted indexes to be returned.
	ExpectedIndexes []client.EncryptedIndexDescription

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// ListAllEncryptedIndexes will attempt to list encrypted index from all collections.
type ListAllEncryptedIndexes struct {
	// NodeID may hold the ID (index) of a node to list the encrypted index on.
	//
	// If a value is not provided the encrypted indexes will be retrieved from the first nodes.
	NodeID immutable.Option[int]

	// Identity is the identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The expected encrypted indexes by collection names to be returned.
	ExpectedIndexes map[client.CollectionName][]client.EncryptedIndexDescription

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// DeleteEncryptedIndex will attempt to delete the given encrypted index for the given collection
// using the collection api.
type DeleteEncryptedIndex struct {
	// NodeID may hold the ID (index) of a node to drop the encrypted index on.
	//
	// If a value is not provided the index will be dropped on all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection for which this index should be dropped.
	CollectionID int

	// The name of the field whose encrypted index should be dropped.
	FieldName string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// TransactionID to use for the action. Optional.
	TransactionID immutable.Option[int]
}

// ResultAsserter is an interface that can be implemented to provide custom result
// assertions. This is a type alias for backward compatibility.
type ResultAsserter = action.ResultAsserter

// ResultAsserterFunc is a function that can be used to implement the ResultAsserter.
// This is a type alias for backward compatibility.
type ResultAsserterFunc = action.ResultAsserterFunc

// Benchmark is an action that will run another test action for benchmark test.
// It will run benchmarks for a base case and optimized case and assert that
// the optimized case performs better by at least the given factor.
type Benchmark struct {
	// BaseCase is a test action which is the base case to benchmark.
	BaseCase any
	// OptimizedCase is a test action which is the optimized case to benchmark.
	OptimizedCase any
	// Reps is the number of times to run the benchmark.
	Reps int
	// FocusClients is the list of clients to run the benchmark on.
	FocusClients []state.ClientType
	// Factor is the factor by which the optimized case should be better than the base case.
	Factor float64
}

// GenerateDocs is an action that will trigger generation of documents.
type GenerateDocs struct {
	// NodeID may hold the ID (index) of a node to execute the generation on.
	//
	// If a value is not provided the docs generation will be executed against all nodes,
	NodeID immutable.Option[int]

	// Options to be passed to the auto doc generator.
	Options []gen.Option

	// The list of collection names to generate docs for.
	// If not provided, docs will be generated for all collections.
	ForCollections []string
}

// AddPredefinedDocs is an action that will trigger creation of predefined documents.
// Predefined docs allows specifying a database state with complex schemas that can be used by
// multiple tests while allowing each test to select a subset of the schemas (collection and
// collection's fields) to work with.
// Example:
//
//	 gen.DocsList{
//		ColName: "User",
//		Docs: []map[string]any{
//		  {
//			"name":     "Shahzad",
//			"devices": []map[string]any{
//			  {
//				"model": "iPhone Xs",
//			  }},
//		  }},
//	 }
//
// For more information refer to tests/predefined/README.md
type AddPredefinedDocs struct {
	// NodeID may hold the ID (index) of a node to execute the generation on.
	//
	// If a value is not provided the docs generation will be executed against all nodes,
	NodeID immutable.Option[int]

	// The list of documents to replicate.
	Docs predefined.DocsList
}

// CommitTransaction represents a commit request for a transaction of the given id.
type CommitTransaction struct {
	// Used to identify the transaction to commit.
	TransactionID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

type IntrospectionRequest struct {
	// NodeID is the node ID (index) of the node in which to introspect.
	NodeID immutable.Option[int]

	// The introspection request to use when fetching schema state.
	//
	// Available properties can be found in the GQL spec:
	// https://spec.graphql.org/October2021/#sec-Introspection
	Request string

	// The full data expected to be returned from the introspection request.
	ExpectedData map[string]any

	// If [ExpectedData] is nil and this is populated, the test framework will assert
	// that the value given exists in the actual results.
	//
	// If this contains nested maps it only requires the last (i.e. non-map) value to
	// be present along the given path.  If an array/slice is present in this chain,
	// it will assert that the items in the expected-array have exact matches in the
	// corresponding result-array (inner maps are not traversed beyond the array,
	// the full array-item must match exactly).
	ContainsData map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// ClientIntrospectionRequest represents a GraphQL client introspection request.
// The GraphQL clients usually use this to fetch the schema state with a default introspection
// query they provide.
type ClientIntrospectionRequest struct {
	// NodeID is the node ID (index) of the node in which to introspect.
	NodeID immutable.Option[int]

	// The introspection request to use when fetching schema state.
	Request string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// ExportBackup will attempt to export data from the datastore using the db api.
type ExportBackup struct {
	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the backup export will be done for all the nodes.
	// todo: https://github.com/sourcenetwork/defradb/issues/3067
	NodeID immutable.Option[int]

	// The backup configuration.
	Config client.BackupConfig

	// Content expected to be found in the backup file.
	ExpectedContent string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// ImportBackup will attempt to import data into the datastore using the db API.
type ImportBackup struct {
	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the backup import will be done for all the nodes.
	// todo: https://github.com/sourcenetwork/defradb/issues/3067
	NodeID immutable.Option[int]

	// The backup file path.
	Filepath string

	// The backup file content.
	ImportContent string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// GetNodeIdentity is an action that calls the [DB.GetNodeIdentity] method and asserts the result.
// It checks if a node at the given index has an identity matching another identity under the same index.
type GetNodeIdentity struct {
	// NodeID holds the ID (index) of a node to get the identity from.
	NodeID int

	// ExpectedIdentity holds the identity that is expected to be found.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	ExpectedIdentity immutable.Option[state.Identity]
}

// Wait is an action that will wait for the given duration.
type Wait struct {
	// Duration is the duration to wait.
	Duration time.Duration
}

// VerifyBlockSignature is an action that will verify the signature of the given block.
type VerifyBlockSignature struct {
	// The cid of the block to verify the signature of.
	Cid string

	// The identity of this request. Optional.
	//
	// Use `ClientIdentity` to create a client identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The identity of the author of the block to verify the signature of.
	SignerIdentity state.Identity

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Used to identify the transaction for this to be executed in. Optional.
	TransactionID immutable.Option[int]
}

// SyncDocs will synchronize documents from the network via P2P.
type SyncDocs struct {
	// NodeID holds the ID (index) of a node to execute the sync on.
	NodeID int

	// The identity of this request. Optional.
	//
	// If node acp is enabled, identity will be used to check if this operation can be performed.
	Identity immutable.Option[state.Identity]

	// The collection containing the documents to sync.
	CollectionID int

	// The indices of documents to sync (references to previously added documents).
	// Uses the same DocIndex pattern as other test actions - these will be resolved
	// to actual document IDs at runtime by the test framework.
	DocIDs []int

	// The source nodes to sync documents from.
	// This is used by testing framework to determine from which nodes the expected doc heads can
	// be looked up for WaitForSync action.
	// There must an item for each document in DocIDs.
	SourceNodes []int

	// Any error expected from the action.
	ExpectedError string
}
