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
	"testing"
	"time"

	"github.com/lens-vm/lens/host-go/config/model"
	"github.com/sourcenetwork/immutable"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/net"
	"github.com/sourcenetwork/defradb/tests/gen"
	"github.com/sourcenetwork/defradb/tests/predefined"
)

// TestCase contains the details of the test case to execute.
type TestCase struct {
	// Test description, optional.
	Description string

	// Actions contains the set of actions and their expected results that
	// this test should execute.  They will execute in the order that they
	// are provided.
	Actions []any

	// If provided a value, SupportedMutationTypes will cause this test to be skipped
	// if the active mutation type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between mutation types, or we need to temporarily document a bug.
	SupportedMutationTypes immutable.Option[[]MutationType]

	// If provided a value, SupportedClientTypes will limit the client types under test to those
	// within this set.  If no active clients pass this filter the test will be skipped.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between client types, or we need to temporarily document a bug.
	SupportedClientTypes immutable.Option[[]ClientType]

	// If provided a value, SupportedACPTypes will cause this test to be skipped
	// if the active acp type is not within the given set.
	//
	// This is to only be used in the very rare cases where we really do want behavioural
	// differences between acp types, or we need to temporarily document a bug.
	SupportedACPTypes immutable.Option[[]ACPType]

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
	SupportedDatabaseTypes immutable.Option[[]DatabaseType]

	// Configuration for KMS to be used in the test
	KMS KMS
}

// KMS contains the configuration for KMS to be used in the test
type KMS struct {
	// ExcludedTypes specifies the KMS types that should be excluded from the test.
	// If none are specified all types will be used.
	ExcludedTypes []KMSType
	// Activated indicates if the KMS should be used in the test
	Activated bool
}

// SetupComplete is a flag to explicitly notify the change detector at which point
// setup is complete so that it may split actions across database code-versions.
//
// If a SetupComplete action is not provided the change detector will split before
// the first item that is neither a SchemaUpdate, CreateDoc or UpdateDoc action.
type SetupComplete struct{}

// ConfigureNode allows the explicit configuration of new Defra nodes.
//
// If no nodes are explicitly configured, a default one will be setup.  There is no
// upper limit to the number that can be configured.
//
// Nodes may be explicitly referenced by index by other actions using `NodeID` properties.
// If the action has a `NodeID` property and it is not specified, the action will be
// effected on all nodes.
type ConfigureNode func() []net.NodeOpt

// Restart is an action that will close and then start all nodes.
type Restart struct{}

// Close is an action that will close a node.
type Close struct {
	// NodeID may hold the ID (index) of a node to close.
	//
	// If a value is not provided the close will be applied to all nodes.
	NodeID immutable.Option[int]
}

// Start is an action that will start a node that has been previously closed.
type Start struct {
	// NodeID may hold the ID (index) of a node to start.
	//
	// If a value is not provided the start will be applied to all nodes.
	NodeID immutable.Option[int]
}

// SchemaUpdate is an action that will update the database schema.
//
// WARNING: getCollectionNames will not work with schemas ending in `type`, e.g. `user_type`
// and should be updated if such a name is desired.
type SchemaUpdate struct {

	// The schema update.
	Schema string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Optionally, the expected results.
	//
	// Each item will be compared individually, if ID, RootID, SchemaVersionID or Fields on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	ExpectedResults []client.CollectionDescription

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]
}

type SchemaPatch struct {
	Patch string

	ExpectedError string

	Lens immutable.Option[model.Lens]

	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	// If SetAsDefaultVersion has a value, and that value is false then the schema version
	// resulting from this patch will not be made default.
	SetAsDefaultVersion immutable.Option[bool]
}

type PatchCollection struct {

	// The Patch to apply to the collection description.
	Patch string

	ExpectedError string
	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]
}

// GetSchema is an action that fetches schema using the provided options.
type GetSchema struct {

	// The VersionID of the schema version to fetch.
	//
	// This option will be prioritized over all other options.
	VersionID immutable.Option[string]

	// The Root of the schema versions to fetch.
	//
	// This option will be prioritized over Name.
	Root immutable.Option[string]

	// The Name of the schema versions to fetch.
	Name immutable.Option[string]

	ExpectedError string

	ExpectedResults []client.SchemaDescription

	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]
}

// GetCollections is an action that fetches collections using the provided options.
//
// ID, RootID and SchemaVersionID will only be asserted on if an expected value is provided.
type GetCollections struct {

	// Any error expected from the action. Optional.
	ExpectedError string

	// The expected results.
	//
	// Each item will be compared individually, if ID, RootID or SchemaVersionID on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	ExpectedResults []client.CollectionDescription

	// An optional set of fetch options for the collections.
	FilterOptions client.CollectionFetchOptions

	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]
}

// SetActiveSchemaVersion is an action that will set the active schema version to the
// given value.
type SetActiveSchemaVersion struct {
	SchemaVersionID string
	ExpectedError   string
	// NodeID may hold the ID (index) of a node to set the default schema version on.
	//
	// If a value is not provided the default will be set on all nodes.
	NodeID immutable.Option[int]
}

// CreateView is an action that will create a new View.
type CreateView struct {

	// The query that this View is to be based off of. Required.
	Query string

	// The SDL containing all types used by the view output.
	SDL string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// An optional Lens transform to add to the view.
	Transform immutable.Option[model.Lens]

	// NodeID may hold the ID (index) of a node to create this View on.
	//
	// If a value is not provided the view will be created on all nodes.
	NodeID immutable.Option[int]
}

// RefreshViews action will execute a call to `store.RefreshViews` using the provided options.
type RefreshViews struct {

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// The set of fetch options for the views.
	FilterOptions client.CollectionFetchOptions

	// NodeID may hold the ID (index) of a node to create this View on.
	//
	// If a value is not provided the view will be created on all nodes.
	NodeID immutable.Option[int]
}

// CreateDoc will attempt to create the given document in the given collection
// using the set [MutationType].
type CreateDoc struct {

	// The document to create, in map format.
	//
	// If this is provided [Doc] will be ignored.
	DocMap map[string]any

	// The identity of this request. Optional.
	//
	// If an Identity is not provided the created document(s) will be public.
	//
	// If an Identity is provided and the collection has a policy, then the
	// created document(s) will be owned by this Identity.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[identity]

	// The document to create, in JSON string format.
	//
	// If [DocMap] is provided this value will be ignored.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// Individual fields of the document to encrypt.
	EncryptedFields []string

	// NodeID may hold the ID (index) of a node to apply this create to.
	//
	// If a value is not provided the document will be created in all nodes.
	NodeID immutable.Option[int]

	// The collection in which this document should be created.
	CollectionID int

	// Specifies whether the document should be encrypted.
	IsDocEncrypted bool
}

// DocIndex represents a relation field value, it allows relation fields to be set without worrying
// about the specific document id.
//
// The test harness will substitute this struct for the document at the given index before
// performing the host action.
//
// The targeted document must have been defined in an action prior to the action that this index
// is hosted upon.
type DocIndex struct {
	// CollectionIndex is the index of the collection holding the document to target.
	CollectionIndex int

	// Index is the index within the target collection at which the document exists.
	//
	// This is dependent on the order in which test [CreateDoc] actions were defined.
	Index int
}

// NewDocIndex creates a new [DocIndex] instance allowing relation fields to be set without worrying
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

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only delete public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also delete private document(s) that are owned by this Identity.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[identity]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID may hold the ID (index) of a node to apply this create to.
	//
	// If a value is not provided the document will be created in all nodes.
	NodeID immutable.Option[int]

	// The collection in which this document should be deleted.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int
}

// UpdateDoc will attempt to update the given document using the set [MutationType].
type UpdateDoc struct {

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only update public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also update private document(s) that are owned by this Identity.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[identity]

	// The document update, in JSON string format. Will only update the properties
	// provided.
	Doc string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The collection in which this document exists.
	CollectionID int

	// The index-identifier of the document within the collection.  This is based on
	// the order in which it was created, not the ordering of the document within the
	// database.
	DocID int

	// Skip waiting for an update event on the local event bus.
	//
	// This should only be used for tests that do not correctly
	// publish an update event to the local event bus.
	SkipLocalUpdateEvent bool
}

// UpdateWithFilter will update the set of documents that match the given filter.
type UpdateWithFilter struct {

	// The filter to match documents against.
	Filter any

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only update public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also update private document(s) that are owned by this Identity.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[identity]

	// The update to apply to matched documents.
	Updater string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The collection in which this document exists.
	CollectionID int

	// Skip waiting for an update event on the local event bus.
	//
	// This should only be used for tests that do not correctly
	// publish an update event to the local event bus.
	SkipLocalUpdateEvent bool
}

// IndexField describes a field to be indexed.
type IndexedField struct {
	// Name contains the name of the field.
	Name string
	// Descending indicates whether the field is indexed in descending order.
	Descending bool
}

// CreateIndex will attempt to create the given secondary index for the given collection
// using the collection api.
type CreateIndex struct {

	// The name of the index to create. If not provided, one will be generated.
	IndexName string

	// The name of the field to index. Used only for single field indexes.
	FieldName string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// The fields to index. Used only for composite indexes.
	Fields []IndexedField

	// NodeID may hold the ID (index) of a node to create the secondary index on.
	//
	// If a value is not provided the index will be created in all nodes.
	NodeID immutable.Option[int]

	// The collection for which this index should be created.
	CollectionID int

	// If Unique is true, the index will be created as a unique index.
	Unique bool
}

// DropIndex will attempt to drop the given secondary index from the given collection
// using the collection api.
type DropIndex struct {

	// The index name of the secondary index within the collection.
	// If it is provided, `IndexID` is ignored.
	IndexName string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID may hold the ID (index) of a node to delete the secondary index from.
	//
	// If a value is not provided the index will be deleted from all nodes.
	NodeID immutable.Option[int]

	// The collection from which the index should be deleted.
	CollectionID int

	// The index-identifier of the secondary index within the collection.
	// This is based on the order in which it was created, not the ordering of
	// the indexes within the database.
	IndexID int
}

// GetIndex will attempt to get the given secondary index from the given collection
// using the collection api.
type GetIndexes struct {

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// The expected indexes to be returned.
	ExpectedIndexes []client.IndexDescription

	// NodeID may hold the ID (index) of a node to create the secondary index on.
	//
	// If a value is not provided the indexes will be retrieved from the first nodes.
	NodeID immutable.Option[int]

	// The collection for which this indexes should be retrieved.
	CollectionID int
}

// ResultAsserter is an interface that can be implemented to provide custom result
// assertions.
type ResultAsserter interface {
	// Assert will be called with the test and the result of the request.
	Assert(t testing.TB, result map[string]any)
}

// ResultAsserterFunc is a function that can be used to implement the ResultAsserter
type ResultAsserterFunc func(testing.TB, map[string]any) (bool, string)

func (f ResultAsserterFunc) Assert(t testing.TB, result map[string]any) {
	f(t, result)
}

// Benchmark is an action that will run another test action for benchmark test.
// It will run benchmarks for a base case and optimized case and assert that
// the optimized case performs better by at least the given factor.
type Benchmark struct {
	// BaseCase is a test action which is the base case to benchmark.
	BaseCase any
	// OptimizedCase is a test action which is the optimized case to benchmark.
	OptimizedCase any
	// FocusClients is the list of clients to run the benchmark on.
	FocusClients []ClientType
	// Reps is the number of times to run the benchmark.
	Reps int
	// Factor is the factor by which the optimized case should be better than the base case.
	Factor float64
}

// Request represents a standard Defra (GQL) request.
type Request struct {

	// Variables sets the variables option for the request.
	Variables immutable.Option[map[string]any]

	// Asserter is an optional custom result asserter.
	Asserter ResultAsserter

	// The expected (data) results of the issued request.
	Results map[string]any

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only operate over public document(s).
	//
	// If an Identity is provided and the collection has a policy, then can
	// operate over private document(s) that are owned by this Identity.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	Identity immutable.Option[identity]

	// OperationName sets the operation name option for the request.
	OperationName immutable.Option[string]

	// The request to execute.
	Request string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID may hold the ID (index) of a node to execute this request on.
	//
	// If a value is not provided the request will be executed against all nodes,
	// in which case the expected results must all match across all nodes.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]
}

// GenerateDocs is an action that will trigger generation of documents.
type GenerateDocs struct {

	// Options to be passed to the auto doc generator.
	Options []gen.Option

	// The list of collection names to generate docs for.
	// If not provided, docs will be generated for all collections.
	ForCollections []string
	// NodeID may hold the ID (index) of a node to execute the generation on.
	//
	// If a value is not provided the docs generation will be executed against all nodes,
	NodeID immutable.Option[int]
}

// CreatePredefinedDocs is an action that will trigger creation of predefined documents.
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
type CreatePredefinedDocs struct {

	// The list of documents to replicate.
	Docs predefined.DocsList
	// NodeID may hold the ID (index) of a node to execute the generation on.
	//
	// If a value is not provided the docs generation will be executed against all nodes,
	NodeID immutable.Option[int]
}

// TransactionCommit represents a commit request for a transaction of the given id.
type TransactionCommit struct {

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// Used to identify the transaction to commit.
	TransactionID int
}

// SubscriptionRequest represents a subscription request.
//
// The subscription will remain active until shortly after all actions have been processed.
// The results of the subscription will then be asserted upon.
type SubscriptionRequest struct {

	// The subscription request to submit.
	Request string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// The expected (data) results yielded through the subscription across its lifetime.
	Results []map[string]any

	// NodeID is the node ID (index) of the node in which to subscribe to.
	NodeID immutable.Option[int]
}

type IntrospectionRequest struct {

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

	// The introspection request to use when fetching schema state.
	//
	// Available properties can be found in the GQL spec:
	// https://spec.graphql.org/October2021/#sec-Introspection
	Request string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID is the node ID (index) of the node in which to introspect.
	NodeID immutable.Option[int]
}

// ClientIntrospectionRequest represents a GraphQL client introspection request.
// The GraphQL clients usually use this to fetch the schema state with a default introspection
// query they provide.
type ClientIntrospectionRequest struct {

	// The introspection request to use when fetching schema state.
	Request string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID is the node ID (index) of the node in which to introspect.
	NodeID immutable.Option[int]
}

// BackupExport will attempt to export data from the datastore using the db api.
type BackupExport struct {

	// Content expected to be found in the backup file.
	ExpectedContent string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string

	// The backup configuration.
	Config client.BackupConfig

	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the backup export will be done for all the nodes.
	// todo: https://github.com/sourcenetwork/defradb/issues/3067
	NodeID immutable.Option[int]
}

// BackupExport will attempt to export data from the datastore using the db api.
type BackupImport struct {

	// The backup file path.
	Filepath string

	// The backup file content.
	ImportContent string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the backup import will be done for all the nodes.
	// todo: https://github.com/sourcenetwork/defradb/issues/3067
	NodeID immutable.Option[int]
}

// GetNodeIdentity is an action that calls the [DB.GetNodeIdentity] method and asserts the result.
// It checks if a node at the given index has an identity matching another identity under the same index.
type GetNodeIdentity struct {

	// ExpectedIdentity holds the identity that is expected to be found.
	//
	// Use `UserIdentity` to create a user identity and `NodeIdentity` to create a node identity.
	// Default value is `NoIdentity()`.
	ExpectedIdentity immutable.Option[identity]
	// NodeID holds the ID (index) of a node to get the identity from.
	NodeID int
}

// Wait is an action that will wait for the given duration.
type Wait struct {
	// Duration is the duration to wait.
	Duration time.Duration
}
