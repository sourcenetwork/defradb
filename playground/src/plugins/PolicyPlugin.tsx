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
import { FileText } from 'lucide-react';
import { usePlaygroundStore } from '../store/playgroundStore';
import styles from './PluginStyles.module.css';

export const policyPlugin: GraphiQLPlugin = {
  title: 'Add Policy',
  icon: () => <FileText size={16} />,
  content: () => <PolicyComponent />,
};

const PolicyComponent = () => {
  const policyText = usePlaygroundStore((state) => state.policy.text);
  const isLoading = usePlaygroundStore((state) => state.policy.isLoading);
  const result = usePlaygroundStore((state) => state.policy.result);
  const setPolicyText = usePlaygroundStore((state) => state.setPolicyText);
  const addPolicy = usePlaygroundStore((state) => state.addPolicy);

  const handlePolicyChange = (value: string) => {
    setPolicyText(value);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!isLoading && policyText.trim()) {
      addPolicy();
    }
  };

  return (
    <main className={styles.pluginContainer}>
      <header>
        <h3 className={styles.pluginTitle}>Add Policy</h3>
        <p id="policy-description" className={styles.pluginDescription}>
          Paste your policy YAML below and click "Add Policy".
        </p>
      </header>

      <form onSubmit={handleSubmit} noValidate>
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
}; 