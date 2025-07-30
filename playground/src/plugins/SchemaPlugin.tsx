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

import React, { useState, useCallback, useEffect } from 'react';
import { GraphiQLPlugin } from '@graphiql/react';
import styles from './PluginStyles.module.css';
import { SchemaIcon } from '../icons';

export const DEFAULT_SCHEMA = `type Users @policy(
  id: "policy_id",
  resource: "users"
) {
  name: String
  age: Int
}`;

type ResultType = 'success' | 'error' | 'info';

interface SchemaResult {
  message: string;
  type: ResultType;
}

interface SchemaPluginProps {
  schemaRef: React.RefObject<string>;
  clientRef: React.RefObject<any>;
  policyIdRef: React.RefObject<string>;
  defaultSchema: string;
}

export const createSchemaPlugin = (props: SchemaPluginProps): GraphiQLPlugin => ({
  title: 'Add Schema',
  icon: SchemaIcon,
  content: () => {
    const [schemaText, setSchemaText] = useState(() => 
      props.defaultSchema.replace('policy_id', props.policyIdRef.current || 'policy_id')
    );
    const [isLoading, setIsLoading] = useState(false);
    const [result, setResult] = useState<SchemaResult | null>(null);

    useEffect(() => {
      if (props.policyIdRef.current) {
        const updatedSchema = props.defaultSchema.replace('policy_id', props.policyIdRef.current);
        setSchemaText(updatedSchema);
        props.schemaRef.current = updatedSchema;
      }
    }, [props.policyIdRef.current, props.defaultSchema]);

    const handleSchemaChange = useCallback((value: string) => {
      setSchemaText(value);
      props.schemaRef.current = value;
    }, [props]);

    const handleAddSchema = useCallback(async () => {
      if (!props.clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error'
        });
        return;
      }

      setIsLoading(true);
      setResult({
        message: 'Adding schema...',
        type: 'info'
      });

      try {
        const response = await props.clientRef.current.addSchema(schemaText);
        const successMessage = `Schema added successfully: ${JSON.stringify(response, null, 2)}`;

        setResult({
          message: successMessage,
          type: 'success'
        });
      } catch (error) {
        const errorMessage = `Error adding schema: ${error instanceof Error ? error.message : String(error)}`;
        setResult({
          message: errorMessage,
          type: 'error'
        });
      } finally {
        setIsLoading(false);
      }
    }, [schemaText, props.clientRef]);

    return (
      <main className={styles.pluginContainer}>
        <header>
          <h3 className={styles.pluginTitle}>Add Schema</h3>
          <p id="schema-description" className={styles.pluginDescription}>
            Edit your schema below and click "Add Schema". The policy ID will be automatically populated if a policy was previously created.
          </p>
        </header>
        
        <form 
          onSubmit={(e) => {
            e.preventDefault();
            if (!isLoading && schemaText.trim()) {
              handleAddSchema();
            }
          }}
          noValidate
        >
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
  },
}); 