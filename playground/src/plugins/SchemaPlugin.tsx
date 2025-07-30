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
import { GraphiQLPlugin } from '@graphiql/react';
import { Database } from 'lucide-react';
import { usePlaygroundStore } from '../store/playgroundStore';
import styles from './PluginStyles.module.css';

export const schemaPlugin: GraphiQLPlugin = {
  title: 'Add Schema',
  icon: () => <Database size={16} />,
  content: () => <SchemaComponent />,
};

const SchemaComponent = () => {
  const schemaText = usePlaygroundStore((state) => state.schema.text);
  const isLoading = usePlaygroundStore((state) => state.schema.isLoading);
  const result = usePlaygroundStore((state) => state.schema.result);
  const setSchemaText = usePlaygroundStore((state) => state.setSchemaText);
  const addSchema = usePlaygroundStore((state) => state.addSchema);

  const handleSchemaChange = (value: string) => {
    setSchemaText(value);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isLoading && schemaText.trim()) {
      addSchema();
    }
  };

  return (
    <main className={styles.pluginContainer}>
      <header>
        <h3 className={styles.pluginTitle}>Add Schema</h3>
        <p id="schema-description" className={styles.pluginDescription}>
          Edit your schema below and click "Add Schema". The policy ID will be automatically populated if a policy was previously created.
        </p>
      </header>

      <form onSubmit={handleSubmit} noValidate>
        <fieldset className={styles.formGroup}>
          <label htmlFor="schema-input" className={styles.formLabel}>
            Schema GraphQL
          </label>
          <textarea
            id="schema-input"
            name="schema"
            value={schemaText}
            onChange={(e) => handleSchemaChange(e.target.value)}
            className={styles.textarea}
            placeholder="Enter schema GraphQL..."
            aria-describedby="schema-description"
            required
            minLength={5}
            rows={12}
            spellCheck={false}
          />
        </fieldset>

        <button
          type="submit"
          disabled={isLoading || !schemaText.trim()}
          className={`${styles.button} ${styles.primary} ${styles.fullWidth}`}
          aria-describedby={result ? 'schema-result' : undefined}
          aria-busy={isLoading}
        >
          {isLoading ? 'Adding Schema...' : 'Add Schema'}
        </button>
      </form>

      {result && (
        <section
          id="schema-result"
          className={`${styles.resultContainer} ${styles[result.type]}`}
          role={result.type === 'error' ? 'alert' : 'status'}
          aria-live={result.type === 'error' ? 'assertive' : 'polite'}
          aria-label={`Schema operation result: ${result.type}`}
        >
          <pre className={`${styles.resultText} ${styles[result.type]}`}>
            {result.message}
          </pre>
        </section>
      )}
    </main>
  );
}; 