/* Copyright 2025 Democratized Data Foundation
 *
 * Use of this software is governed by the Business Source License
 * included in the file licenses/BSL.txt.
 *
 * As of the Change Date specified in that file, in accordance with
 * the Business Source License, use of this software will be governed
 * by the Apache License, Version 2.0, included in the file
 * licenses/APL.txt.
 */

import React from 'react';
import styles from './PluginStyles.module.css';
import { RelationshipIcon } from '../icons';

export const DEFAULT_RELATIONSHIP = {
  collectionName: 'Users',
  docID: 'document_id',
  relation: 'collaborator',
  targetActor: 'did:key:alice',
};

type ResultType = 'success' | 'error' | 'info';

interface RelationshipResult {
  message: string;
  type: ResultType;
}

export interface RelationshipPluginProps {
  clientRef: React.RefObject<any>;
  policyIdRef: React.RefObject<string>;
  relationshipRef: React.RefObject<typeof DEFAULT_RELATIONSHIP>;
  resultRef: React.RefObject<string>;
}

export const createRelationshipPlugin = ({
  clientRef,
  policyIdRef,
  relationshipRef,
  resultRef,
}: RelationshipPluginProps) => ({
  title: 'Actor Relationships',
  icon: RelationshipIcon,
  content: () => {
    const [isVerifyLoading, setIsVerifyLoading] = React.useState(false);
    const [isAddLoading, setIsAddLoading] = React.useState(false);
    const [isDeleteLoading, setIsDeleteLoading] = React.useState(false);
    const [result, setResult] = React.useState<RelationshipResult | null>(null);
    const [relationship, setRelationship] = React.useState(DEFAULT_RELATIONSHIP);

    React.useEffect(() => {
      relationshipRef.current = relationship;
    }, [relationship, relationshipRef]);

    const handleAddRelationship = async () => {
      if (!clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error',
        });
        return;
      }

      setIsAddLoading(true);
      setResult({
        message: 'Adding relationship...',
        type: 'info',
      });

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };

        const response = await clientRef.current.addDACActorRelationship(
          relationship.collectionName,
          relationship.docID,
          relationship.relation,
          relationship.targetActor,
          context,
        );

        const successMessage = `Relationship added successfully: ${JSON.stringify(response, null, 2)}`;
        setResult({
          message: successMessage,
          type: 'success',
        });
        resultRef.current = successMessage;
      } catch (error) {
        const errorMessage = `Error adding relationship: ${error instanceof Error ? error.message : String(error)}`;
        setResult({
          message: errorMessage,
          type: 'error',
        });
        resultRef.current = errorMessage;
      } finally {
        setIsAddLoading(false);
      }
    };

    const handleDeleteRelationship = async () => {
      if (!clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error',
        });
        return;
      }

      setIsDeleteLoading(true);
      setResult({
        message: 'Deleting relationship...',
        type: 'info',
      });

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };

        const response = await clientRef.current.deleteDACActorRelationship(
          relationship.collectionName,
          relationship.docID,
          relationship.relation,
          relationship.targetActor,
          context,
        );

        const successMessage = `Relationship deleted successfully: ${JSON.stringify(response, null, 2)}`;
        setResult({
          message: successMessage,
          type: 'success',
        });
        resultRef.current = successMessage;
      } catch (error) {
        const errorMessage = `Error deleting relationship: ${error instanceof Error ? error.message : String(error)}`;
        setResult({
          message: errorMessage,
          type: 'error',
        });
        resultRef.current = errorMessage;
      } finally {
        setIsDeleteLoading(false);
      }
    };

    const handleVerifyAccess = async () => {
      if (!clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error',
        });
        return;
      }

      setIsVerifyLoading(true);
      setResult({
        message: 'Verifying access...',
        type: 'info',
      });

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };

        if (!policyIdRef.current) {
          const errorMessage = 'Error: Policy ID not available. Please add a policy first.';
          setResult({
            message: errorMessage,
            type: 'error',
          });
          resultRef.current = errorMessage;
          return;
        }

        const response = await clientRef.current.verifyDACAccess(
          'read',
          relationship.targetActor,
          policyIdRef.current,
          relationship.collectionName.toLowerCase(),
          relationship.docID,
          context,
        );

        const successMessage = `Access verification result: ${JSON.stringify(response, null, 2)}`;
        setResult({
          message: successMessage,
          type: 'success',
        });
        resultRef.current = successMessage;
      } catch (error) {
        const errorMessage = `Error verifying access: ${error instanceof Error ? error.message : String(error)}`;
        setResult({
          message: errorMessage,
          type: 'error',
        });
        resultRef.current = errorMessage;
      } finally {
        setIsVerifyLoading(false);
      }
    };

    return (
      <main className={styles.pluginContainer}>
        <header>
          <h3 className={styles.pluginTitle}>Actor Relationships</h3>
          <p id="relationship-description" className={styles.pluginDescription}>
            Add or delete actor relationships for testing ACP functionality.
          </p>
        </header>

        <form noValidate>
          <fieldset>
            <legend className="sr-only">Relationship Configuration</legend>
            
            <div className={styles.formGroup}>
              <label htmlFor="collection-name" className={styles.formLabel}>
                Collection Name
              </label>
              <input
                id="collection-name"
                name="collectionName"
                type="text"
                value={relationship.collectionName}
                onChange={(e) => setRelationship(prev => ({ ...prev, collectionName: e.target.value }))}
                className={styles.input}
                placeholder="e.g. Users"
                required
                minLength={1}
                aria-describedby="relationship-description"
              />
            </div>

            <div className={styles.formGroup}>
              <label htmlFor="document-id" className={styles.formLabel}>
                Document ID
              </label>
              <input
                id="document-id"
                name="docID"
                type="text"
                value={relationship.docID}
                onChange={(e) => setRelationship(prev => ({ ...prev, docID: e.target.value }))}
                className={styles.input}
                placeholder="e.g. bae-12345678-1234-1234-1234-123456789abc"
                required
                pattern="^[a-zA-Z0-9\-_]+$"
                title="Document ID should contain only alphanumeric characters, hyphens, and underscores"
              />
            </div>

            <div className={styles.formGroup}>
              <label htmlFor="relation" className={styles.formLabel}>
                Relation
              </label>
              <input
                id="relation"
                name="relation"
                type="text"
                value={relationship.relation}
                onChange={(e) => setRelationship(prev => ({ ...prev, relation: e.target.value }))}
                className={styles.input}
                placeholder="e.g. owner | reader | collaborator"
                required
                minLength={1}
              />
            </div>

            <div className={styles.formGroup}>
              <label htmlFor="target-actor" className={styles.formLabel}>
                Target Actor (DID)
              </label>
              <input
                id="target-actor"
                name="targetActor"
                type="text"
                value={relationship.targetActor}
                onChange={(e) => setRelationship(prev => ({ ...prev, targetActor: e.target.value }))}
                className={styles.input}
                placeholder="e.g. did:key:alice or * for all actors"
                required
                minLength={1}
              />
            </div>
          </fieldset>

          <section className={styles.buttonGroup} role="group" aria-label="Relationship actions">
            <button
              type="button"
              onClick={handleAddRelationship}
              disabled={isAddLoading}
              className={`${styles.button} ${styles.success}`}
              aria-busy={isAddLoading}
            >
              {isAddLoading ? 'Adding...' : 'Add Relationship'}
            </button>

            <button
              type="button"
              onClick={handleDeleteRelationship}
              disabled={isDeleteLoading}
              className={`${styles.button} ${styles.danger}`}
              aria-busy={isDeleteLoading}
            >
              {isDeleteLoading ? 'Deleting...' : 'Delete Relationship'}
            </button>
          </section>

          <section className={styles.formGroup}>
            <button
              type="button"
              onClick={handleVerifyAccess}
              disabled={isVerifyLoading}
              className={`${styles.button} ${styles.info} ${styles.fullWidth}`}
              aria-busy={isVerifyLoading}
            >
              {isVerifyLoading ? 'Verifying...' : 'Verify Access'}
            </button>
          </section>
        </form>

        {result && (
          <section
            className={`${styles.resultContainer} ${styles[result.type]}`}
            role={result.type === 'error' ? 'alert' : 'status'}
            aria-live={result.type === 'error' ? 'assertive' : 'polite'}
            aria-label={`Relationship operation result: ${result.type}`}
          >
            <pre className={`${styles.resultText} ${styles[result.type]}`}>
              {result.message}
            </pre>
          </section>
        )}
      </main>
    );
  },
}); 