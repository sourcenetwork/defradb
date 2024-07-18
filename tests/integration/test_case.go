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

// SchemaUpdate is an action that will update the database schema.
//
// WARNING: getCollectionNames will not work with schemas ending in `type`, e.g. `user_type`
// and should be updated if such a name is desired.
type SchemaUpdate struct {
	// NodeID may hold the ID (index) of a node to apply this update to.
	//
	// If a value is not provided the update will be applied to all nodes.
	NodeID immutable.Option[int]

	// The schema update.
	Schema string

	// Optionally, the expected results.
	//
	// Each item will be compared individually, if ID, RootID, SchemaVersionID or Fields on the
	// expected item are default they will not be compared with the actual.
	//
	// Assertions on Indexes and Sources will not distinguish between nil and empty (in order
	// to allow their ommission in most cases).
	ExpectedResults []client.CollectionDescription

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

type SchemaPatch struct {
	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	Patch string

	// If SetAsDefaultVersion has a value, and that value is false then the schema version
	// resulting from this patch will not be made default.
	SetAsDefaultVersion immutable.Option[bool]

	Lens immutable.Option[model.Lens]

	ExpectedError string
}

type PatchCollection struct {
	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	// The Patch to apply to the collection description.
	Patch string

	ExpectedError string
}

// GetSchema is an action that fetches schema using the provided options.
type GetSchema struct {
	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

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

	ExpectedResults []client.SchemaDescription

	ExpectedError string
}

// GetCollections is an action that fetches collections using the provided options.
//
// ID, RootID and SchemaVersionID will only be asserted on if an expected value is provided.
type GetCollections struct {
	// NodeID may hold the ID (index) of a node to apply this patch to.
	//
	// If a value is not provided the patch will be applied to all nodes.
	NodeID immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

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

	// Any error expected from the action. Optional.
	ExpectedError string
}

// SetActiveSchemaVersion is an action that will set the active schema version to the
// given value.
type SetActiveSchemaVersion struct {
	// NodeID may hold the ID (index) of a node to set the default schema version on.
	//
	// If a value is not provided the default will be set on all nodes.
	NodeID immutable.Option[int]

	SchemaVersionID string
	ExpectedError   string
}

// CreateView is an action that will create a new View.
type CreateView struct {
	// NodeID may hold the ID (index) of a node to create this View on.
	//
	// If a value is not provided the view will be created on all nodes.
	NodeID immutable.Option[int]

	// The query that this View is to be based off of. Required.
	Query string

	// The SDL containing all types used by the view output.
	SDL string

	// An optional Lens transform to add to the view.
	Transform immutable.Option[model.Lens]

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// CreateDoc will attempt to create the given document in the given collection
// using the set [MutationType].
type CreateDoc struct {
	// NodeID may hold the ID (index) of a node to apply this create to.
	//
	// If a value is not provided the document will be created in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided the created document(s) will be public.
	//
	// If an Identity is provided and the collection has a policy, then the
	// created document(s) will be owned by this Identity.
	Identity immutable.Option[int]

	// Specifies whether the document should be encrypted.
	IsDocEncrypted bool

	// Individual fields of the document to encrypt.
	EncryptedFields []string

	// The collection in which this document should be created.
	CollectionID int

	// The document to create, in JSON string format.
	//
	// If [DocMap] is provided this value will be ignored.
	Doc string

	// The document to create, in map format.
	//
	// If this is provided [Doc] will be ignored.
	DocMap map[string]any

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
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
	// NodeID may hold the ID (index) of a node to apply this create to.
	//
	// If a value is not provided the document will be created in all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only delete public document(s).
	//
	// If an Identity is provided and the collection has a policy, then
	// can also delete private document(s) that are owned by this Identity.
	Identity immutable.Option[int]

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
}

// UpdateDoc will attempt to update the given document using the set [MutationType].
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
	Identity immutable.Option[int]

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
	// NodeID may hold the ID (index) of a node to create the secondary index on.
	//
	// If a value is not provided the index will be created in all nodes.
	NodeID immutable.Option[int]

	// The collection for which this index should be created.
	CollectionID int

	// The name of the index to create. If not provided, one will be generated.
	IndexName string

	// The name of the field to index. Used only for single field indexes.
	FieldName string

	// The fields to index. Used only for composite indexes.
	Fields []IndexedField

	// If Unique is true, the index will be created as a unique index.
	Unique bool

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// DropIndex will attempt to drop the given secondary index from the given collection
// using the collection api.
type DropIndex struct {
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

	// The index name of the secondary index within the collection.
	// If it is provided, `IndexID` is ignored.
	IndexName string

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// GetIndex will attempt to get the given secondary index from the given collection
// using the collection api.
type GetIndexes struct {
	// NodeID may hold the ID (index) of a node to create the secondary index on.
	//
	// If a value is not provided the indexes will be retrieved from the first nodes.
	NodeID immutable.Option[int]

	// The collection for which this indexes should be retrieved.
	CollectionID int

	// The expected indexes to be returned.
	ExpectedIndexes []client.IndexDescription

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// ResultAsserter is an interface that can be implemented to provide custom result
// assertions.
type ResultAsserter interface {
	// Assert will be called with the test and the result of the request.
	Assert(t testing.TB, result []map[string]any)
}

// ResultAsserterFunc is a function that can be used to implement the ResultAsserter
type ResultAsserterFunc func(testing.TB, []map[string]any) (bool, string)

func (f ResultAsserterFunc) Assert(t testing.TB, result []map[string]any) {
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
	// Reps is the number of times to run the benchmark.
	Reps int
	// FocusClients is the list of clients to run the benchmark on.
	FocusClients []ClientType
	// Factor is the factor by which the optimized case should be better than the base case.
	Factor float64
}

// Request represents a standard Defra (GQL) request.
type Request struct {
	// NodeID may hold the ID (index) of a node to execute this request on.
	//
	// If a value is not provided the request will be executed against all nodes,
	// in which case the expected results must all match across all nodes.
	NodeID immutable.Option[int]

	// The identity of this request. Optional.
	//
	// If an Identity is not provided then can only operate over public document(s).
	//
	// If an Identity is provided and the collection has a policy, then can
	// operate over private document(s) that are owned by this Identity.
	Identity immutable.Option[int]

	// Used to identify the transaction for this to run against. Optional.
	TransactionID immutable.Option[int]

	// The request to execute.
	Request string

	// The expected (data) results of the issued request.
	Results []map[string]any

	// Asserter is an optional custom result asserter.
	Asserter ResultAsserter

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
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
	// NodeID may hold the ID (index) of a node to execute the generation on.
	//
	// If a value is not provided the docs generation will be executed against all nodes,
	NodeID immutable.Option[int]

	// The list of documents to replicate.
	Docs predefined.DocsList
}

// TransactionCommit represents a commit request for a transaction of the given id.
type TransactionCommit struct {
	// Used to identify the transaction to commit.
	TransactionID int

	// Any error expected from the action. Optional.
	//
	// String can be a partial, and the test will pass if an error is returned that
	// contains this string.
	ExpectedError string
}

// SubscriptionRequest represents a subscription request.
//
// The subscription will remain active until shortly after all actions have been processed.
// The results of the subscription will then be asserted upon.
type SubscriptionRequest struct {
	// NodeID is the node ID (index) of the node in which to subscribe to.
	NodeID immutable.Option[int]

	// The subscription request to submit.
	Request string

	// The expected (data) results yielded through the subscription across its lifetime.
	Results []map[string]any

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

// BackupExport will attempt to export data from the datastore using the db api.
type BackupExport struct {
	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the indexes will be retrieved from the first nodes.
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

// BackupExport will attempt to export data from the datastore using the db api.
type BackupImport struct {
	// NodeID may hold the ID (index) of a node to generate the backup from.
	//
	// If a value is not provided the indexes will be retrieved from the first nodes.
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
