// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"fmt"
	"strings"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
)

const (
	errFailedToGetHeads                          string = "failed to get document heads"
	errFailedToCreateCollectionQuery             string = "failed to create collection prefix query"
	errFailedToGetCollection                     string = "failed to get collection"
	errFailedToGetAllCollections                 string = "failed to get all collections"
	errDocVerification                           string = "the document verification failed"
	errAddingP2PCollection                       string = "cannot add collection ID"
	errRemovingP2PCollection                     string = "cannot remove collection ID"
	errAddCollectionWithPatch                    string = "adding collections via patch is not supported"
	errRemoveReferencedCollection                string = "cannot remove a collection while another field references it"
	errCollectionIDDoesntMatch                   string = "CollectionID does not match existing"
	errCollectionRootDoesntMatch                 string = "CollectionRoot does not match existing"
	errCannotSetVersionID                        string = "setting the VersionID is not supported"
	errRelationalFieldMissingIDField             string = "missing id field for relation object field"
	errRelatedFieldKindMismatch                  string = "invalid Kind of the related field"
	errRelationalFieldIDInvalidType              string = "relational id field of invalid kind"
	errDuplicateField                            string = "duplicate field"
	errCannotMutateField                         string = "mutating an existing field is not supported"
	errCannotMoveField                           string = "moving fields is not currently supported"
	errCannotDeleteField                         string = "deleting an existing field is not supported"
	errFieldKindNotFound                         string = "no type found for given name"
	errFieldKindDoesNotMatchFieldDefinition      string = "field Kind does not match field definition"
	errDocumentAlreadyExists                     string = "a document with the given ID already exists"
	errDocumentDeleted                           string = "a document with the given ID has been deleted"
	errIndexMissingFields                        string = "index missing fields"
	errNonZeroIndexIDProvided                    string = "non-zero index ID provided"
	errIndexFieldMissingName                     string = "index field missing name"
	errIndexWithNameAlreadyExists                string = "index with name already exists"
	errInvalidStoredIndex                        string = "invalid stored index"
	errInvalidStoredIndexKey                     string = "invalid stored index key"
	errNonExistingFieldForIndex                  string = "making a new index on a non-existing property"
	errFailedToStoreIndexedField                 string = "failed to store indexed field"
	errFailedToReadStoredIndexDesc               string = "failed to read stored index description"
	errCanNotDeleteIndexedField                  string = "can not delete indexed field"
	errCanNotCreateNewIndexWithPatch             string = "making new indexes via patch is not supported"
	errCanNotDropIndexWithPatch                  string = "dropping indexes via patch is not supported"
	errIndexWithNameDoesNotExists                string = "index with name doesn't exists"
	errCorruptedIndex                            string = "corrupted index. Please delete and recreate the index"
	errInvalidFieldValue                         string = "invalid field value"
	errUnsupportedIndexFieldType                 string = "unsupported index field type"
	errCannotIndexAccumulatedCRDTField           string = "indexing accumulated CRDT fields is not yet supported"
	errIndexDescriptionHasNoFields               string = "index description has no fields"
	errCreateFile                                string = "failed to create file"
	errRemoveFile                                string = "failed to remove file"
	errOpenFile                                  string = "failed to open file"
	errCloseFile                                 string = "failed to close file"
	errFailedtoCloseQueryReqAllIDs               string = "failed to close query requesting all docIDs"
	errFailedToReadByte                          string = "failed to read byte"
	errFailedToWriteString                       string = "failed to write string"
	errJSONDecode                                string = "failed to decode JSON"
	errDocFromMap                                string = "failed to create a new doc from map"
	errDocAdd                                    string = "failed to add a new doc to collection"
	errDocUpdate                                 string = "failed to update doc to collection"
	errEmptyFilter                               string = "filter cannot be empty"
	errUnsupportedFilterType                     string = "unsupported filter type"
	errExpectedJSONObject                        string = "expected JSON object"
	errExpectedJSONArray                         string = "expected JSON array"
	errIndexDoesNotMatchName                     string = "the index used does not match the given name"
	errCanNotIndexNonUniqueFields                string = "can not index a doc's field(s) that violates unique index"
	errInvalidViewQuery                          string = "the query provided is not valid as a View"
	errCollectionAlreadyExists                   string = "collection already exists"
	errMultipleActiveCollectionVersions          string = "multiple versions of same collection cannot be active"
	errCollectionSourcesCannotBeAddedRemoved     string = "collection sources cannot be added or removed"
	errCollectionSourceIDMutated                 string = "collection source ID cannot be mutated"
	errCollectionSourceWrongCollection           string = "collection source must belong to host collection"
	errCollectionIndexesCannotBeMutated          string = "collection indexes cannot be mutated"
	errCollectionEncryptedIndexesCannotBeMutated string = "collection encrypted indexes cannot be mutated"
	errCollectionPolicyCannotBeMutated           string = "collection policy cannot be mutated"
	errCollectionIDCannotBeMutated               string = "collection ID cannot be mutated"
	errCollectionVersionIDCannotBeMutated        string = "collection version ID cannot be mutated"
	errCollectionIDCannotBeEmpty                 string = "collection ID cannot be empty"
	errCannotDeleteOldVersion                    string = "cannot delete a version that is used by a newer version, " +
		"first delete the new version"
	errCanNotHavePolicyWithoutACP          string = "can not specify policy on collection, without acp"
	errRelationMissingField                string = "relation missing field"
	errMultipleRelationPrimaries           string = "relation can only have a single field set as primary"
	errP2PColHasPolicy                     string = "p2p collection specified has a policy on it"
	errReplicatorColHasPolicy              string = "replicator collection specified has a policy on it"
	errNoTransactionInContext              string = "no transaction in context"
	errReplicatorExists                    string = "replicator already exists"
	errReplicatorDocID                     string = "failed to get docID for replicator"
	errCanNotEncryptBuiltinField           string = "can not encrypt build-in field"
	errSelfReferenceWithoutSelf            string = "must specify 'Self' kind for self referencing relations"
	errColNotMaterialized                  string = "non-materialized collections are not supported"
	errColMutatingIsBranchable             string = "mutating IsBranchable is not supported"
	errMaterializedViewAndACPNotSupported  string = "materialized views do not support ACP"
	errInvalidDefaultFieldValue            string = "default field value is invalid"
	errDocIDNotFound                       string = "docID not found"
	errCollectionRootNotFound              string = "collection root not found"
	errGetEmbeddingFunc                    string = "failed to get embedding function"
	errGetEmbeddingField                   string = "failed getting vector embedding field"
	errFieldNotFound                       string = "field not found"
	errGetDocForEmbedding                  string = "failed to get previous document for embedding generation"
	errMissingSignature                    string = "block is missing required signature"
	errNoIdentityInContext                 string = "no identity found in context"
	errMissingPermission                   string = "missing permission"
	errCollectionNameMutated               string = "collection name cannot be mutated"
	errUnsupportedTxnType                  string = "unsupported transaction type"
	errEncryptedIndexUnknownField          string = "encrypted index on non-existent field"
	errEncryptedIndexAlreadyExists         string = "encrypted index already exists on this field"
	errEncryptedIndexDoesNotExist          string = "encrypted index does not exist on this field"
	errNACIsAlreadyDisabled                string = "node acp is already disabled"
	errNACIsAlreadyEnabled                 string = "node acp is already enabled"
	errNACIsNotConfigured                  string = "node acp is not configured"
	errRelationNameEmpty                   string = "relation name cannot be empty"
	errInvalidCID                          string = "invalid CID"
	errUnknownCID                          string = "unknown CID, collection ids cannot be manually defined"
	errMigrationBetweenNonAdjacentVersions string = "cannot migrate between non-adjacent collection versions"
	errLensRuntimeNotSupported             string = "the selected lens runtime is not supported by this build"
	errLensCIDNotFound                     string = "lens CID not found"
	errOneToOneMustBeUnique                string = "one-to-one relation must have a unique index"

	errCreateMergeTxn         string = "failed to create merge transaction"
	errGetShortIDForMerge     string = "failed to get short collection ID for merge"
	errGetMergeTargetHeads    string = "failed to get merge target heads"
	errLoadComposites         string = "failed to load composites for merge"
	errMergeComposites        string = "failed to merge composites"
	errSyncIndexedDoc         string = "failed to sync indexed document after merge"
	errLoadBlockForMerge      string = "failed to load block for merge"
	errDecodeBlockForMerge    string = "failed to decode block for merge"
	errLoadParentComposite    string = "failed to load parent composite for merge"
	errLoadMergeTargetBlock   string = "failed to load merge target block"
	errDecodeMergeTargetBlock string = "failed to decode merge target block"
	errGenerateMergeLink      string = "failed to generate link for merge composite"
	errProcessBlockMerge      string = "failed to process block during merge"
	errProcessEncryptedBlock  string = "failed to process encrypted block"
	errInitCRDTForMerge       string = "failed to init CRDT for merge"
	errProcessCRDTBlock       string = "failed to process CRDT block"
	errLoadChildBlock         string = "failed to load child block for merge"
	errDecodeChildBlock       string = "failed to decode child block for merge"
	errProcessChildBlock      string = "failed to process child block for merge"
	errLoadEncryptionBlock    string = "failed to load encryption block"
	errGetHeadsForMerge       string = "failed to get heads for merge target"
	errLoadBlockFromStore     string = "failed to get block from blockstore"
	errDecodeBlockFromStore   string = "failed to decode block from bytes"
	errParseDocIDMerge        string = "failed to parse doc ID during merge"
	errGetShortFieldIDMerge   string = "failed to get short field ID during merge"
	errGetDocStatus           string = "failed to get document status"
	errGetShortIDForDoc       string = "failed to get short collection ID for document"

	errDeleteNACState             string = "failed to delete NAC state"
	errCommitNACTransaction       string = "failed to commit NAC transaction"
	errParseDatastoreKey          string = "failed to parse datastore key"
	errStoreViewCacheItem         string = "failed to store view cache item"
	errStoreDocMarker             string = "failed to store document marker"
	errSetEmbeddingField          string = "failed to set embedding field"
	errStoreIndexKey              string = "failed to store index key"
	errDeleteIndexedDoc           string = "failed to delete indexed document"
	errDeleteIndexKey             string = "failed to delete index key"
	errUpdateIndex                string = "failed to update index"
	errTruncateDatastoreKey       string = "failed to delete key during truncate"
	errTruncateHeadstoreKey       string = "failed to delete headstore key during truncate"
	errTruncateDeleteBlocks       string = "failed to delete blocks during truncate"
	errDeleteViewCacheItem        string = "failed to delete view cache item"
	errParseViewCacheKey          string = "failed to parse view cache key"
	errStoreNACState              string = "failed to store NAC state"
	errMarshalNACState            string = "failed to marshal NAC state"
	errCheckNACState              string = "failed to check NAC state"
	errCheckDBInitialized         string = "failed to check if database is initialized"
	errCheckCIDExists             string = "failed to check if CID exists in blockstore"
	errCheckIndexKeyExists        string = "failed to check if index key exists"
	errCheckUniqueIndexConstraint string = "failed to check unique index constraint"
	errCheckCollectionDocs        string = "failed to check if collection has documents"
	errCreateTruncateIterator     string = "failed to create iterator for truncate"
	errDumpDBState                string = "failed to iterate datastore during dump"
	errGetAllDocIDs               string = "failed to get all document IDs"
	errCreateDeleteIndexIterator  string = "failed to create iterator for index deletion"
	errCreateViewCacheIterator    string = "failed to create view cache iterator"
	errTxnDiscarded               string = "this transaction has been discarded. Create a new one"
)

var (
	ErrFailedToGetCollection                     = errors.New(errFailedToGetCollection)
	ErrSubscriptionsNotAllowed                   = errors.New("server does not accept subscriptions")
	ErrEmptyFilter                               = errors.New(errEmptyFilter)
	ErrUnsupportedFilterType                     = errors.New(errUnsupportedFilterType)
	ErrCollectionAlreadyExists                   = errors.New(errCollectionAlreadyExists)
	ErrCollectionNameEmpty                       = errors.New("collection name can't be empty")
	ErrCollectionRootEmpty                       = errors.New("collection root can't be empty")
	ErrCollectionVersionIDEmpty                  = errors.New("collection version ID can't be empty")
	ErrKeyEmpty                                  = errors.New("key cannot be empty")
	ErrCannotSetVersionID                        = errors.New(errCannotSetVersionID)
	ErrIndexMissingFields                        = errors.New(errIndexMissingFields)
	ErrIndexFieldMissingName                     = errors.New(errIndexFieldMissingName)
	ErrCorruptedIndex                            = errors.New(errCorruptedIndex)
	ErrExpectedJSONObject                        = errors.New(errExpectedJSONObject)
	ErrExpectedJSONArray                         = errors.New(errExpectedJSONArray)
	ErrInvalidViewQuery                          = errors.New(errInvalidViewQuery)
	ErrCanNotIndexNonUniqueFields                = errors.New(errCanNotIndexNonUniqueFields)
	ErrMultipleActiveCollectionVersions          = errors.New(errMultipleActiveCollectionVersions)
	ErrCollectionSourcesCannotBeAddedRemoved     = errors.New(errCollectionSourcesCannotBeAddedRemoved)
	ErrCollectionSourceIDMutated                 = errors.New(errCollectionSourceIDMutated)
	ErrCollectionSourceWrongCollection           = errors.New(errCollectionSourceWrongCollection)
	ErrCollectionIndexesCannotBeMutated          = errors.New(errCollectionIndexesCannotBeMutated)
	ErrCollectionEncryptedIndexesCannotBeMutated = errors.New(errCollectionEncryptedIndexesCannotBeMutated)
	ErrCollectionCollectionIDCannotBeMutated     = errors.New(errCollectionIDCannotBeMutated)
	ErrCollectionVersionIDCannotBeMutated        = errors.New(errCollectionVersionIDCannotBeMutated)
	ErrCollectionIDCannotBeEmpty                 = errors.New(errCollectionIDCannotBeEmpty)
	ErrCannotDeleteOldVersion                    = errors.New(errCannotDeleteOldVersion)
	ErrCanNotHavePolicyWithoutACP                = errors.New(errCanNotHavePolicyWithoutACP)
	ErrRemoveReferencedCollection                = errors.New(errRemoveReferencedCollection)
	ErrRelationMissingField                      = errors.New(errRelationMissingField)
	ErrMultipleRelationPrimaries                 = errors.New(errMultipleRelationPrimaries)
	ErrP2PColHasPolicy                           = errors.New(errP2PColHasPolicy)
	ErrNoTransactionInContext                    = errors.New(errNoTransactionInContext)
	ErrReplicatorColHasPolicy                    = errors.New(errReplicatorColHasPolicy)
	ErrCanNotEncryptBuiltinField                 = errors.New(errCanNotEncryptBuiltinField)
	ErrSelfReferenceWithoutSelf                  = errors.New(errSelfReferenceWithoutSelf)
	ErrColNotMaterialized                        = errors.New(errColNotMaterialized)
	ErrMaterializedViewAndACPNotSupported        = errors.New(errMaterializedViewAndACPNotSupported)
	ErrDocIDNotFound                             = errors.New(errDocIDNotFound)
	ErrCollectionRootNotFound                    = errors.New(errCollectionRootNotFound)
	ErrColMutatingIsBranchable                   = errors.New(errColMutatingIsBranchable)
	ErrGetEmbeddingField                         = errors.New(errGetEmbeddingField)
	ErrFieldNotFound                             = errors.New(errFieldNotFound)
	ErrGetDocForEmbedding                        = errors.New(errGetDocForEmbedding)
	ErrGetEmbeddingFunc                          = errors.New(errGetEmbeddingFunc)
	ErrMissingSignature                          = errors.New(errMissingSignature)
	ErrMissingPermission                         = errors.New(errMissingPermission)
	ErrNoIdentityInContext                       = errors.New(errNoIdentityInContext)
	ErrCollectionNameMutated                     = errors.New(errCollectionNameMutated)
	ErrUnsupportedTxnType                        = errors.New(errUnsupportedTxnType)
	ErrNACIsAlreadyDisabled                      = errors.New(errNACIsAlreadyDisabled)
	ErrNACIsAlreadyEnabled                       = errors.New(errNACIsAlreadyEnabled)
	ErrNACIsNotConfigured                        = errors.New(errNACIsNotConfigured)
	ErrNACRelationshipOperationRequiresIdentity  = errors.New("node acp relationship operation requires identity")
	ErrRelationNameEmpty                         = errors.New(errRelationNameEmpty)
	ErrInvalidCID                                = errors.New(errInvalidCID)
	ErrUnknownCID                                = errors.New(errUnknownCID)
	ErrNoP2P                                     = errors.New("no p2p system configured")
	ErrBadDocsResultType                         = errors.New("bad docs result type")
	ErrMigrationBetweenNonAdjacentVersions       = errors.New(errMigrationBetweenNonAdjacentVersions)
	ErrLensRuntimeNotSupported                   = errors.New(errLensRuntimeNotSupported)
	ErrLensCIDNotFound                           = errors.New(errLensCIDNotFound)
	ErrDocumentAlreadyExists                     = errors.New(errDocumentAlreadyExists)
	ErrIndexWithNameAlreadyExists                = errors.New(errIndexWithNameAlreadyExists)
	ErrIndexWithNameDoesNotExists                = errors.New(errIndexWithNameDoesNotExists)
	ErrEncryptedIndexAlreadyExists               = errors.New(errEncryptedIndexAlreadyExists)
	ErrEncryptedIndexDoesNotExist                = errors.New(errEncryptedIndexDoesNotExist)
	ErrReplicatorExists                          = errors.New(errReplicatorExists)
	ErrTxnDiscarded                              = errors.New(errTxnDiscarded)
)

// NewErrFailedToGetHeads returns a new error indicating that the heads of a document
// could not be obtained.
func NewErrFailedToGetHeads(inner error) error {
	return errors.Wrap(errFailedToGetHeads, inner)
}

// NewErrFailedToCreateCollectionQuery returns a new error indicating that the query
// to create a collection failed.
func NewErrFailedToCreateCollectionQuery(inner error) error {
	return errors.Wrap(errFailedToCreateCollectionQuery, inner)
}

// NewErrInvalidStoredIndex returns a new error indicating that the stored
// index in the database is invalid.
func NewErrInvalidStoredIndex(inner error) error {
	return errors.Wrap(errInvalidStoredIndex, inner)
}

// NewErrInvalidStoredIndexKey returns a new error indicating that the stored
// index in the database is invalid.
func NewErrInvalidStoredIndexKey(key string) error {
	return errors.New(errInvalidStoredIndexKey, errors.NewKV("Key", key))
}

// NewErrNonExistingFieldForIndex returns a new error indicating the attempt to make a new index
// on a non-existing field.
func NewErrNonExistingFieldForIndex(field string) error {
	return errors.New(errNonExistingFieldForIndex, errors.NewKV("Field", field))
}

// NewErrFailedToStoreIndexedField returns a new error indicating that the indexed field
// could not be stored.
func NewErrFailedToStoreIndexedField(key string, inner error) error {
	return errors.Wrap(errFailedToStoreIndexedField, inner, errors.NewKV("Key", key))
}

// NewErrFailedToReadStoredIndexDesc returns a new error indicating that the stored index
// description could not be read.
func NewErrFailedToReadStoredIndexDesc(inner error) error {
	return errors.Wrap(errFailedToReadStoredIndexDesc, inner)
}

// NewCanNotDeleteIndexedField returns a new error a failed attempt to delete an indexed field
func NewCanNotDeleteIndexedField(inner error) error {
	return errors.Wrap(errCanNotDeleteIndexedField, inner)
}

// NewErrNonZeroIndexIDProvided returns a new error indicating that a non-zero index ID was
// provided.
func NewErrNonZeroIndexIDProvided(indexID uint32) error {
	return errors.New(errNonZeroIndexIDProvided, errors.NewKV("ID", indexID))
}

// NewErrFailedToGetCollection returns a new error indicating that the collection could not
// be obtained.
func NewErrFailedToGetCollection(name string, inner error) error {
	return errors.Wrap(errFailedToGetCollection, inner, errors.NewKV("Name", name))
}

// NewErrFailedToGetAllCollections returns a new error indicating that the collection list could not
// be obtained.
func NewErrFailedToGetAllCollections(inner error) error {
	return errors.Wrap(errFailedToGetAllCollections, inner)
}

// NewErrDocVerification returns a new error indicating that the document verification failed.
//
// This occurs when a documents contents fail the verification during an Add()
// call against the supplied Document ID (docID).
func NewErrDocVerification(expected string, actual string) error {
	return errors.New(
		errDocVerification,
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

// NewErrAddingP2PCollection returns a new error indicating that adding a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrAddingP2PCollection(inner error) error {
	return errors.Wrap(errAddingP2PCollection, inner)
}

// NewErrRemovingP2PCollection returns a new error indicating that removing a collection ID to the
// persisted list of P2P collection IDs was not successful.
func NewErrRemovingP2PCollection(inner error) error {
	return errors.Wrap(errRemovingP2PCollection, inner)
}

func NewErrAddCollectionWithPatch(name string) error {
	return errors.New(
		errAddCollectionWithPatch,
		errors.NewKV("Name", name),
	)
}

func NewErrRemoveReferencedCollection(inner error, removed []string) error {
	return errors.Wrap(
		errRemoveReferencedCollection,
		inner,
		errors.NewKV("Removed", strings.Join(removed, ",")),
	)
}

// NewErrRemoveReferencedCollectionFromField errors when a patch removes a collection
// that is still being referenced by a field on another collection in the post-patch
// state. It identifies which removed collection is still in use and the host
// collection/field doing the referencing.
func NewErrRemoveReferencedCollectionFromField(removedName, hostCollection, hostField string) error {
	return errors.New(
		errRemoveReferencedCollection,
		errors.NewKV("Removed", removedName),
		errors.NewKV("ReferencedBy", hostCollection),
		errors.NewKV("Field", hostField),
	)
}

func NewErrCollectionIDDoesntMatch(name string, existingID, proposedID string) error {
	return errors.New(
		errCollectionIDDoesntMatch,
		errors.NewKV("Name", name),
		errors.NewKV("ExistingID", existingID),
		errors.NewKV("ProposedID", proposedID),
	)
}

func NewErrCollectionRootDoesntMatch(name, existingRoot, proposedRoot string) error {
	return errors.New(
		errCollectionRootDoesntMatch,
		errors.NewKV("Name", name),
		errors.NewKV("ExistingRoot", existingRoot),
		errors.NewKV("ProposedRoot", proposedRoot),
	)
}

func NewErrRelationalFieldMissingIDField(name string, expectedName string) error {
	return errors.New(
		errRelationalFieldMissingIDField,
		errors.NewKV("Field", name),
		errors.NewKV("ExpectedIDFieldName", expectedName),
	)
}

func NewErrRelatedFieldKindMismatch(relationName string, expected client.FieldKind, actual client.FieldKind) error {
	return errors.New(
		errRelatedFieldKindMismatch,
		errors.NewKV("RelationName", relationName),
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

func NewErrRelationalFieldIDInvalidType(name string, expected, actual client.FieldKind) error {
	return errors.New(
		errRelationalFieldIDInvalidType,
		errors.NewKV("Field", name),
		errors.NewKV("Expected", expected),
		errors.NewKV("Actual", actual),
	)
}

func NewErrRelationNameEmpty(name string) error {
	return errors.New(
		errRelationNameEmpty,
		errors.NewKV("Field", name),
	)
}

func NewErrFieldKindNotFound(name string, kind string) error {
	return errors.New(
		errFieldKindNotFound,
		errors.NewKV("Field", name),
		errors.NewKV("Kind", kind),
	)
}

func NewErrFieldKindDoesNotMatchFieldDefinition(kind string, definition string) error {
	return errors.New(
		errFieldKindDoesNotMatchFieldDefinition,
		errors.NewKV("Kind", kind),
		errors.NewKV("Definition", definition),
	)
}

func NewErrDuplicateField(name string) error {
	return errors.New(errDuplicateField, errors.NewKV("Name", name))
}

func NewErrCannotMutateField(name string) error {
	return errors.New(
		errCannotMutateField,
		errors.NewKV("ProposedName", name),
	)
}

func NewErrCannotMoveField(name string, proposedIndex, existingIndex int) error {
	return errors.New(
		errCannotMoveField,
		errors.NewKV("Name", name),
		errors.NewKV("ProposedIndex", proposedIndex),
		errors.NewKV("ExistingIndex", existingIndex),
	)
}

func NewErrCanNotEncryptBuiltinField(name string) error {
	return errors.New(errCanNotEncryptBuiltinField, errors.NewKV("Name", name))
}

func NewErrCannotDeleteField(name string) error {
	return errors.New(
		errCannotDeleteField,
		errors.NewKV("Name", name),
	)
}

func NewErrDocumentAlreadyExists(docID string) error {
	return errors.New(
		errDocumentAlreadyExists,
		errors.NewKV("DocID", docID),
	)
}

func NewErrDocumentDeleted(docID string) error {
	return errors.New(
		errDocumentDeleted,
		errors.NewKV("DocID", docID),
	)
}

// NewErrIndexWithNameAlreadyExists returns a new error indicating that an index with the
// given name already exists.
func NewErrIndexWithNameAlreadyExists(indexName string) error {
	return errors.New(
		errIndexWithNameAlreadyExists,
		errors.NewKV("Name", indexName),
	)
}

// NewErrIndexWithNameDoesNotExists returns a new error indicating that an index with the
// given name does not exist.
func NewErrIndexWithNameDoesNotExists(indexName string) error {
	return errors.New(
		errIndexWithNameDoesNotExists,
		errors.NewKV("Name", indexName),
	)
}

// NewErrCorruptedIndex returns a new error indicating that an index with the
// given name has been corrupted.
func NewErrCorruptedIndex(indexName string) error {
	return errors.New(
		errCorruptedIndex,
		errors.NewKV("Name", indexName),
	)
}

// NewErrCannotCreateNewIndexWithPatch returns a new error indicating that making a new index
// via patch is not supported.
func NewErrCannotCreateNewIndexWithPatch(proposedName string) error {
	return errors.New(
		errCanNotCreateNewIndexWithPatch,
		errors.NewKV("ProposedName", proposedName),
	)
}

// NewErrCannotDropIndexWithPatch returns a new error indicating that an index cannot be dropped
// with a patch.
func NewErrCannotDropIndexWithPatch(indexName string) error {
	return errors.New(
		errCanNotDropIndexWithPatch,
		errors.NewKV("Name", indexName),
	)
}

// NewErrInvalidFieldValue returns a new error indicating that the given value is invalid for the
// given field kind.
func NewErrInvalidFieldValue(kind client.FieldKind, value any) error {
	return errors.New(
		errInvalidFieldValue,
		errors.NewKV("Kind", kind),
		errors.NewKV("Value", value),
	)
}

// NewErrUnsupportedIndexFieldType returns a new error indicating that the given field kind is not
// supported for indexing.
func NewErrUnsupportedIndexFieldType(kind client.FieldKind) error {
	return errors.New(
		errUnsupportedIndexFieldType,
		errors.NewKV("Kind", kind),
	)
}

// NewErrCannotIndexAccumulatedCRDTField returns a new error indicating that the given field
// cannot be indexed because it uses an accumulated CRDT type.
func NewErrCannotIndexAccumulatedCRDTField(fieldName, crdtType string) error {
	return errors.New(
		errCannotIndexAccumulatedCRDTField,
		errors.NewKV("Field", fieldName),
		errors.NewKV("CRDTType", crdtType),
	)
}

// NewErrIndexDescHasNoFields returns a new error indicating that the given index
// description has no fields.
func NewErrIndexDescHasNoFields(desc client.IndexDescription) error {
	return errors.New(
		errIndexDescriptionHasNoFields,
		errors.NewKV("Description", desc),
	)
}

// NewErrCreateFile returns a new error indicating there was a failure in creating a file.
func NewErrCreateFile(inner error, filepath string) error {
	return errors.Wrap(errCreateFile, inner, errors.NewKV("Filepath", filepath))
}

// NewErrOpenFile returns a new error indicating there was a failure in opening a file.
func NewErrOpenFile(inner error, filepath string) error {
	return errors.Wrap(errOpenFile, inner, errors.NewKV("Filepath", filepath))
}

// NewErrCloseFile returns a new error indicating there was a failure in closing a file.
func NewErrCloseFile(closeErr, other error) error {
	if other != nil {
		return errors.Wrap(errCloseFile, closeErr, errors.NewKV("Other error", other))
	}
	return errors.Wrap(errCloseFile, closeErr)
}

// NewErrRemoveFile returns a new error indicating there was a failure in removing a file.
func NewErrRemoveFile(removeErr, other error, filepath string) error {
	if other != nil {
		return errors.Wrap(
			errRemoveFile,
			removeErr,
			errors.NewKV("Other error", other),
			errors.NewKV("Filepath", filepath),
		)
	}
	return errors.Wrap(
		errRemoveFile,
		removeErr,
		errors.NewKV("Filepath", filepath),
	)
}

// NewErrFailedToReadByte returns a new error indicating there was a failure in read a byte
// from the Reader
func NewErrFailedToReadByte(inner error) error {
	return errors.Wrap(errFailedToReadByte, inner)
}

// NewErrFailedToWriteString returns a new error indicating there was a failure in writing
// a string to the Writer
func NewErrFailedToWriteString(inner error) error {
	return errors.Wrap(errFailedToWriteString, inner)
}

// NewErrJSONDecode returns a new error indicating there was a failure in decoding some JSON
// from the JSON decoder
func NewErrJSONDecode(inner error) error {
	return errors.Wrap(errJSONDecode, inner)
}

// NewErrDocFromMap returns a new error indicating there was a failure to create
// a new doc from a map
func NewErrDocFromMap(inner error) error {
	return errors.Wrap(errDocFromMap, inner)
}

// NewErrDocAdd returns a new error indicating there was a failure to add
// a new doc to a collection
func NewErrDocAdd(inner error) error {
	return errors.Wrap(errDocAdd, inner)
}

// NewErrDocUpdate returns a new error indicating there was a failure to update
// a doc to a collection
func NewErrDocUpdate(inner error) error {
	return errors.Wrap(errDocUpdate, inner)
}

func NewErrIndexDoesNotMatchName(index, name string) error {
	return errors.New(
		errIndexDoesNotMatchName,
		errors.NewKV("Index", index),
		errors.NewKV("Name", name),
	)
}

func NewErrCanNotIndexNonUniqueFields(docID string, fieldValues ...errors.KV) error {
	kvPairs := make([]errors.KV, 0, len(fieldValues)+1)
	kvPairs = append(kvPairs, errors.NewKV("DocID", docID))
	kvPairs = append(kvPairs, fieldValues...)

	return errors.New(errCanNotIndexNonUniqueFields, kvPairs...)
}

func NewErrInvalidViewQueryCastFailed(query string) error {
	return errors.New(
		errInvalidViewQuery,
		errors.NewKV("Query", query),
		errors.NewKV("Reason", "Internal error, cast failed"),
	)
}

func NewErrInvalidViewQueryMissingQuery() error {
	return errors.New(
		errInvalidViewQuery,
		errors.NewKV("Reason", "No query provided"),
	)
}

func NewErrCollectionAlreadyExists(name string) error {
	return errors.New(
		errCollectionAlreadyExists,
		errors.NewKV("Name", name),
	)
}

func NewErrCollectionIDAlreadyExists(id string) error {
	return errors.New(
		errCollectionAlreadyExists,
		errors.NewKV("ID", id),
	)
}

func NewErrMultipleActiveCollectionVersions(name string, root string) error {
	return errors.New(
		errMultipleActiveCollectionVersions,
		errors.NewKV("Name", name),
		errors.NewKV("Root", root),
	)
}

func NewErrCollectionSourcesCannotBeAddedRemoved(colID string) error {
	return errors.New(
		errCollectionSourcesCannotBeAddedRemoved,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionSourceIDMutated(colID string, newSrcID string, oldSrcID string) error {
	return errors.New(
		errCollectionSourceIDMutated,
		errors.NewKV("CollectionID", colID),
		errors.NewKV("NewCollectionSourceID", newSrcID),
		errors.NewKV("OldCollectionSourceID", oldSrcID),
	)
}

func NewErrCollectionIndexesCannotBeMutated(colID string) error {
	return errors.New(
		errCollectionIndexesCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionEncryptedIndexesCannotBeMutated(colID string) error {
	return errors.New(
		errCollectionEncryptedIndexesCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionPolicyCannotBeMutated(colID string) error {
	return errors.New(
		errCollectionPolicyCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCollectionIDCannotBeMutated(collectionVersionID string) error {
	return errors.New(
		errCollectionIDCannotBeMutated,
		errors.NewKV("CollectionVersionID", collectionVersionID),
	)
}

func NewErrCollectionVersionIDCannotBeMutated(colID string) error {
	return errors.New(
		errCollectionVersionIDCannotBeMutated,
		errors.NewKV("CollectionID", colID),
	)
}

func NewErrCannotDeleteOldVersion(old, new string) error {
	return errors.New(
		errCannotDeleteOldVersion,
		errors.NewKV("TargetCollectionID", old),
		errors.NewKV("UsedByCollectionID", new),
	)
}

func NewErrRelationMissingField(objectName, relationName string) error {
	return errors.New(
		errRelationMissingField,
		errors.NewKV("Object", objectName),
		errors.NewKV("RelationName", relationName),
	)
}

func NewErrReplicatorExists(collection string, peerID peer.ID) error {
	return errors.New(
		errReplicatorExists,
		errors.NewKV("Collection", collection),
		errors.NewKV("PeerID", peerID.String()),
	)
}

func NewErrReplicatorDocID(inner error, kv ...errors.KV) error {
	return errors.Wrap(errReplicatorDocID, inner, kv...)
}

func NewErrSelfReferenceWithoutSelf(fieldName string) error {
	return errors.New(
		errSelfReferenceWithoutSelf,
		errors.NewKV("Field", fieldName),
	)
}

func NewErrColNotMaterialized(collection string) error {
	return errors.New(
		errColNotMaterialized,
		errors.NewKV("Collection", collection),
	)
}

func NewErrColMutatingIsBranchable(collection string) error {
	return errors.New(
		errColMutatingIsBranchable,
		errors.NewKV("Collection", collection),
	)
}

func NewErrMaterializedViewAndACPNotSupported(collection string) error {
	return errors.New(
		errMaterializedViewAndACPNotSupported,
		errors.NewKV("Collection", collection),
	)
}

func NewErrDefaultFieldValueInvalid(collection string, inner error) error {
	return errors.New(
		errInvalidDefaultFieldValue,
		errors.NewKV("Collection", collection),
		errors.NewKV("Inner", inner),
	)
}

func NewErrDocIDNotFound(docID string) error {
	return errors.New(errDocIDNotFound, errors.NewKV("DocID", docID))
}

func NewErrCollectionRootNotFound(collectionRoot string) error {
	return errors.New(errCollectionRootNotFound, errors.NewKV("CollectionRoot", collectionRoot))
}

func NewErrGetEmbeddingField(inner error) error {
	return errors.Wrap(errGetEmbeddingField, inner)
}

func NewErrGetEmbeddingFunc(inner error) error {
	return errors.Wrap(errGetEmbeddingFunc, inner)
}

func NewErrEmbeddingFieldNotFound(field string) error {
	return errors.New(errFieldNotFound, errors.NewKV("Embedding field", field))
}

func NewErrGetDocForEmbedding(inner error) error {
	return errors.Wrap(errGetDocForEmbedding, inner)
}

func NewErrCollectionNameMutated(newName string, oldName string) error {
	return errors.New(
		errCollectionNameMutated,
		errors.NewKV("NewName", newName),
		errors.NewKV("OldName", oldName),
	)
}

func NewErrUnsupportedTxnType(actual any) error {
	return errors.New(errUnsupportedTxnType, errors.NewKV("Actual", fmt.Sprintf("%T", actual)))
}

func NewErrEncryptedIndexOnNonExistentField(fieldName string) error {
	return errors.New(
		errEncryptedIndexUnknownField,
		errors.NewKV("Field", fieldName),
	)
}

func NewErrEncryptedIndexAlreadyExists(fieldName string) error {
	return errors.New(
		errEncryptedIndexAlreadyExists,
		errors.NewKV("Field", fieldName),
	)
}

func NewErrEncryptedIndexDoesNotExist(fieldName string) error {
	return errors.New(
		errEncryptedIndexDoesNotExist,
		errors.NewKV("Field", fieldName),
	)
}

func NewErrInvalidCID(name string, value string, inner error) error {
	return errors.New(
		inner.Error(),
		errors.NewKV(name, value),
	)
}

func NewErrUnknownCID(name string, value string) error {
	return errors.New(
		errUnknownCID,
		errors.NewKV(name, value),
	)
}

func NewErrCollectionSourceWrongCollection(hostCollectionID string, sourceCollectionID string) error {
	return errors.New(
		errCollectionSourceWrongCollection,
		errors.NewKV("HostCollectionID", hostCollectionID),
		errors.NewKV("SourceCollectionID", sourceCollectionID),
	)
}

func NewErrMigrationBetweenNonAdjacentVersions(sourceVersion string, destinationVersion string) error {
	return errors.New(
		errMigrationBetweenNonAdjacentVersions,
		errors.NewKV("SourceVersionID", sourceVersion),
		errors.NewKV("DestinationVersionID", destinationVersion),
	)
}

func NewErrLensRuntimeNotSupported(lens LensRuntimeType) error {
	return errors.New(errLensRuntimeNotSupported, errors.NewKV("Lens", lens))
}

func NewErrLensCIDNotFound(cid string) error {
	return errors.New(errLensCIDNotFound, errors.NewKV("CID", cid))
}

func NewErrCheckCollectionDocs(inner error) error {
	return errors.Wrap(errCheckCollectionDocs, inner)
}

func NewErrCreateTruncateIterator(inner error) error {
	return errors.Wrap(errCreateTruncateIterator, inner)
}

func NewErrDumpDBState(inner error) error {
	return errors.Wrap(errDumpDBState, inner)
}

func NewErrGetAllDocIDs(inner error) error {
	return errors.Wrap(errGetAllDocIDs, inner)
}

func NewErrCreateDeleteIndexIterator(inner error) error {
	return errors.Wrap(errCreateDeleteIndexIterator, inner)
}

func NewErrCreateViewCacheIterator(inner error) error {
	return errors.Wrap(errCreateViewCacheIterator, inner)
}

// NewErrOneToOneRelationMustBeUnique returns an error indicating that a one-to-one
// relation field cannot have a non-unique index.
func NewErrOneToOneRelationMustBeUnique(objectName, fieldName string) error {
	return errors.New(
		errOneToOneMustBeUnique,
		errors.NewKV("Object", objectName),
		errors.NewKV("Field", fieldName),
	)
}

// NewErrUnsupportedFilterType returns a new error indicating that the given filter type is not supported.
func NewErrUnsupportedFilterType(actualType string) error {
	return errors.New(errUnsupportedFilterType, errors.NewKV("ActualType", actualType))
}

func NewErrCreateMergeTxn(inner error, docID string, cid string) error {
	return errors.Wrap(errCreateMergeTxn, inner,
		errors.NewKV("DocID", docID), errors.NewKV("CID", cid))
}

func NewErrGetShortIDForMerge(inner error, collectionID string) error {
	return errors.Wrap(errGetShortIDForMerge, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrGetMergeTargetHeads(inner error, docID string, key string) error {
	return errors.Wrap(errGetMergeTargetHeads, inner, errors.NewKV("DocID", docID), errors.NewKV("Key", key))
}

func NewErrLoadComposites(inner error, cid string, docID string) error {
	return errors.Wrap(errLoadComposites, inner,
		errors.NewKV("CID", cid), errors.NewKV("DocID", docID))
}

func NewErrMergeComposites(inner error, docID string) error {
	return errors.Wrap(errMergeComposites, inner, errors.NewKV("DocID", docID))
}

func NewErrSyncIndexedDoc(inner error, docID string) error {
	return errors.Wrap(errSyncIndexedDoc, inner, errors.NewKV("DocID", docID))
}

func NewErrLoadBlockForMerge(inner error, cid string) error {
	return errors.Wrap(errLoadBlockForMerge, inner, errors.NewKV("CID", cid))
}

func NewErrDecodeBlockForMerge(inner error, cid string) error {
	return errors.Wrap(errDecodeBlockForMerge, inner, errors.NewKV("CID", cid))
}

func NewErrLoadParentComposite(inner error, cid string) error {
	return errors.Wrap(errLoadParentComposite, inner, errors.NewKV("CID", cid))
}

func NewErrLoadMergeTargetBlock(inner error, link string) error {
	return errors.Wrap(errLoadMergeTargetBlock, inner, errors.NewKV("Link", link))
}

func NewErrDecodeMergeTargetBlock(inner error, link string) error {
	return errors.Wrap(errDecodeMergeTargetBlock, inner, errors.NewKV("Link", link))
}

func NewErrGenerateMergeLink(inner error) error {
	return errors.Wrap(errGenerateMergeLink, inner)
}

func NewErrProcessBlockMerge(inner error, cid string) error {
	return errors.Wrap(errProcessBlockMerge, inner, errors.NewKV("CID", cid))
}

func NewErrProcessEncryptedBlock(inner error, cid string) error {
	return errors.Wrap(errProcessEncryptedBlock, inner, errors.NewKV("CID", cid))
}

func NewErrInitCRDTForMerge(inner error, cid string) error {
	return errors.Wrap(errInitCRDTForMerge, inner, errors.NewKV("CID", cid))
}

func NewErrProcessCRDTBlock(inner error, cid string) error {
	return errors.Wrap(errProcessCRDTBlock, inner, errors.NewKV("CID", cid))
}

func NewErrLoadChildBlock(inner error, cid string) error {
	return errors.Wrap(errLoadChildBlock, inner, errors.NewKV("CID", cid))
}

func NewErrDecodeChildBlock(inner error, cid string) error {
	return errors.Wrap(errDecodeChildBlock, inner, errors.NewKV("CID", cid))
}

func NewErrProcessChildBlock(inner error, cid string) error {
	return errors.Wrap(errProcessChildBlock, inner, errors.NewKV("CID", cid))
}

func NewErrLoadEncryptionBlock(inner error, cid string) error {
	return errors.Wrap(errLoadEncryptionBlock, inner, errors.NewKV("CID", cid))
}

func NewErrGetHeadsForMerge(inner error, key string) error {
	return errors.Wrap(errGetHeadsForMerge, inner, errors.NewKV("Key", key))
}

func NewErrLoadBlockFromStore(inner error, cid string) error {
	return errors.Wrap(errLoadBlockFromStore, inner, errors.NewKV("CID", cid))
}

func NewErrDecodeBlockFromStore(inner error, cid string) error {
	return errors.Wrap(errDecodeBlockFromStore, inner, errors.NewKV("CID", cid))
}

func NewErrParseDocIDMerge(inner error, rawDocID string) error {
	return errors.Wrap(errParseDocIDMerge, inner, errors.NewKV("RawDocID", rawDocID))
}

func NewErrGetShortFieldIDMerge(inner error, fieldID string, field string) error {
	return errors.Wrap(errGetShortFieldIDMerge, inner,
		errors.NewKV("FieldID", fieldID), errors.NewKV("Field", field))
}

func NewErrGetDocStatus(inner error, docID string) error {
	return errors.Wrap(errGetDocStatus, inner, errors.NewKV("DocID", docID))
}

func NewErrGetShortIDForDoc(inner error, collectionID string) error {
	return errors.Wrap(errGetShortIDForDoc, inner, errors.NewKV("CollectionID", collectionID))
}

func NewErrStoreViewCacheItem(inner error) error {
	return errors.Wrap(errStoreViewCacheItem, inner)
}

func NewErrStoreDocMarker(inner error, docID string) error {
	return errors.Wrap(errStoreDocMarker, inner, errors.NewKV("DocID", docID))
}

func NewErrSetEmbeddingField(inner error, fieldName string) error {
	return errors.Wrap(errSetEmbeddingField, inner, errors.NewKV("Field", fieldName))
}

func NewErrStoreIndexKey(inner error) error {
	return errors.Wrap(errStoreIndexKey, inner)
}

func NewErrStoreNACState(inner error) error {
	return errors.Wrap(errStoreNACState, inner)
}

func NewErrMarshalNACState(inner error) error {
	return errors.Wrap(errMarshalNACState, inner)
}

func NewErrCheckNACState(inner error) error {
	return errors.Wrap(errCheckNACState, inner)
}

func NewErrCheckDBInitialized(inner error) error {
	return errors.Wrap(errCheckDBInitialized, inner)
}

func NewErrCheckCIDExists(inner error, cidType string, cidValue string) error {
	return errors.Wrap(errCheckCIDExists, inner,
		errors.NewKV("Type", cidType), errors.NewKV("CID", cidValue))
}

func NewErrCheckIndexKeyExists(inner error, indexName string) error {
	return errors.Wrap(errCheckIndexKeyExists, inner, errors.NewKV("IndexName", indexName))
}

func NewErrCheckUniqueIndexConstraint(inner error) error {
	return errors.Wrap(errCheckUniqueIndexConstraint, inner)
}

func NewErrDeleteNACState(inner error) error {
	return errors.Wrap(errDeleteNACState, inner)
}

func NewErrCommitNACTransaction(inner error) error {
	return errors.Wrap(errCommitNACTransaction, inner)
}

func NewErrParseDatastoreKey(inner error) error {
	return errors.Wrap(errParseDatastoreKey, inner)
}

func NewErrDeleteIndexedDoc(inner error, indexName string) error {
	return errors.Wrap(errDeleteIndexedDoc, inner, errors.NewKV("IndexName", indexName))
}

func NewErrDeleteIndexKey(inner error) error {
	return errors.Wrap(errDeleteIndexKey, inner)
}

func NewErrUpdateIndex(inner error, indexName string) error {
	return errors.Wrap(errUpdateIndex, inner, errors.NewKV("IndexName", indexName))
}

func NewErrTruncateDatastoreKey(inner error, key string) error {
	return errors.Wrap(errTruncateDatastoreKey, inner, errors.NewKV("Key", key))
}

func NewErrTruncateHeadstoreKey(inner error, key string) error {
	return errors.Wrap(errTruncateHeadstoreKey, inner, errors.NewKV("Key", key))
}

func NewErrTruncateDeleteBlocks(inner error, cid string) error {
	return errors.Wrap(errTruncateDeleteBlocks, inner, errors.NewKV("CID", cid))
}

func NewErrDeleteViewCacheItem(inner error) error {
	return errors.Wrap(errDeleteViewCacheItem, inner)
}

func NewErrParseViewCacheKey(inner error) error {
	return errors.Wrap(errParseViewCacheKey, inner)
}
