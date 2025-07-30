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

import React, { useState, useCallback } from 'react';
import { GraphiQLPlugin } from '@graphiql/react';
import styles from './PluginStyles.module.css';
import { PolicyIcon } from '../icons';

export const DEFAULT_POLICY = `name: Test Policy
description: A test policy for playground

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + collaborator
      update:
        expr: owner + collaborator
      delete:
        expr: owner + collaborator

    relations:
      owner:
        types:
          - actor
      collaborator:
        types:
          - actor`;

type ResultType = 'success' | 'error' | 'info';

interface PolicyResult {
  message: string;
  type: ResultType;
}

interface PolicyPluginProps {
  policyRef: React.RefObject<string>;
  clientRef: React.RefObject<any>;
  resultRef: React.RefObject<string>;
  policyIdRef: React.RefObject<string>;
  defaultPolicy: string;
}

export const createPolicyPlugin = (props: PolicyPluginProps): GraphiQLPlugin => ({
  title: 'Add Policy',
  icon: PolicyIcon,
  content: () => {
    const [policyText, setPolicyText] = useState(props.defaultPolicy);
    const [isLoading, setIsLoading] = useState(false);
    const [result, setResult] = useState<PolicyResult | null>(null);

    const handlePolicyChange = useCallback((value: string) => {
      setPolicyText(value);
      props.policyRef.current = value;
    }, [props]);

    const handleAddPolicy = useCallback(async () => {
      if (!props.clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error',
        });
        return;
      }

      setIsLoading(true);
      setResult({
        message: 'Adding policy...',
        type: 'info',
      });

      try {
        const nodeIdentity = await props.clientRef.current.getNodeIdentity();
        const context = {
          identity: nodeIdentity.PublicKey,
        };

        const response = await props.clientRef.current.addDACPolicy(policyText, context);
        const successMessage = `Policy created successfully: ${JSON.stringify(response, null, 2)}`;

        setResult({
          message: successMessage,
          type: 'success',
        });
        props.resultRef.current = successMessage;

        if (response?.PolicyID) {
          props.policyIdRef.current = response.PolicyID;
        } else {
          console.error('No PolicyID found in result:', response);
        }
      } catch (error) {
        const errorMessage = `Error adding policy: ${error instanceof Error ? error.message : String(error)}`;
        setResult({
          message: errorMessage,
          type: 'error',
        });
        props.resultRef.current = errorMessage;
      } finally {
        setIsLoading(false);
      }
    }, [policyText, props]);

    return (
      <main className={styles.pluginContainer}>
        <header>
          <h3 className={styles.pluginTitle}>Add Policy</h3>
          <p id="policy-description" className={styles.pluginDescription}>
            Paste your policy YAML below and click "Add Policy".
          </p>
        </header>
        
        <form 
          onSubmit={(e) => {
            e.preventDefault();
            if (!isLoading && policyText.trim()) {
              handleAddPolicy();
            }
          }}
          noValidate
        >
          <fieldset className={styles.formGroup}>
            <label htmlFor="policy-input" className={styles.formLabel}>
              Policy YAML
            </label>
            <textarea
              id="policy-input"
              name="policy"
              value={policyText}
              onChange={(e) => handlePolicyChange(e.target.value)}
              className={`${styles.textarea} ${styles.large}`}
              placeholder="Enter policy YAML..."
              aria-describedby="policy-description"
              required
              minLength={10}
              rows={20}
              spellCheck={false}
            />
          </fieldset>
          
          <button
            type="submit"
            disabled={isLoading || !policyText.trim()}
            className={`${styles.button} ${styles.primary} ${styles.fullWidth}`}
            aria-describedby={result ? 'policy-result' : undefined}
            aria-busy={isLoading}
          >
            {isLoading ? 'Adding Policy...' : 'Add Policy'}
          </button>
        </form>

        {result && (
          <section
            id="policy-result"
            className={`${styles.resultContainer} ${styles[result.type]}`}
            role={result.type === 'error' ? 'alert' : 'status'}
            aria-live={result.type === 'error' ? 'assertive' : 'polite'}
            aria-label={`Policy operation result: ${result.type}`}
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