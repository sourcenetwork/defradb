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

import { GraphiQLPlugin } from '@graphiql/react';
import { Users } from 'lucide-react';
import { usePlaygroundStore } from '../store/playgroundStore';
import styles from './PluginStyles.module.css';

export const relationshipPlugin: GraphiQLPlugin = {
  title: 'Actor Relationships',
  icon: () => <Users size={16} />,
  content: () => <RelationshipComponent />,
};

const RelationshipComponent = () => {
  const relationship = usePlaygroundStore((state) => state.relationship.data);
  const isLoading = usePlaygroundStore((state) => state.relationship.isLoading);
  const result = usePlaygroundStore((state) => state.relationship.result);
  const setRelationshipData = usePlaygroundStore((state) => state.setRelationshipData);
  const addRelationship = usePlaygroundStore((state) => state.addRelationship);
  const deleteRelationship = usePlaygroundStore((state) => state.deleteRelationship);
  const verifyAccess = usePlaygroundStore((state) => state.verifyAccess);

  const handleInputChange = (field: keyof typeof relationship, value: string) => {
    setRelationshipData({ [field]: value });
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
              onChange={(e) => handleInputChange('collectionName', e.target.value)}
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
              onChange={(e) => handleInputChange('docID', e.target.value)}
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
              onChange={(e) => handleInputChange('relation', e.target.value)}
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
              onChange={(e) => handleInputChange('targetActor', e.target.value)}
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
            onClick={addRelationship}
            disabled={isLoading.add}
            className={`${styles.button} ${styles.success}`}
            aria-busy={isLoading.add}
          >
            {isLoading.add ? 'Adding...' : 'Add Relationship'}
          </button>

          <button
            type="button"
            onClick={deleteRelationship}
            disabled={isLoading.delete}
            className={`${styles.button} ${styles.danger}`}
            aria-busy={isLoading.delete}
          >
            {isLoading.delete ? 'Deleting...' : 'Delete Relationship'}
          </button>
        </section>

        <section className={styles.formGroup}>
          <button
            type="button"
            onClick={verifyAccess}
            disabled={isLoading.verify}
            className={`${styles.button} ${styles.info} ${styles.fullWidth}`}
            aria-busy={isLoading.verify}
          >
            {isLoading.verify ? 'Verifying...' : 'Verify Access'}
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
}; 