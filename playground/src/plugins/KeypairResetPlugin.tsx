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
import { KeyRound } from 'lucide-react';
import { usePlaygroundStore } from '../store/playgroundStore';
import styles from './PluginStyles.module.css';

export const keypairResetPlugin: GraphiQLPlugin = {
  title: 'Keypair Reset',
  icon: () => <KeyRound size={16} />,
  content: () => <KeypairResetComponent />,
};

const KeypairResetComponent = () => {
  const isResetting = usePlaygroundStore((state) => state.keypair.isLoading);
  const result = usePlaygroundStore((state) => state.keypair.result);
  const resetKeypair = usePlaygroundStore((state) => state.resetKeypair);

  return (
    <main className={styles.pluginContainer}>
      <header>
        <h3 className={styles.pluginTitle}>Keypair Reset</h3>
        <p id="keypair-description" className={styles.pluginDescription}>
          Optionally, reset the keypair used for SourceHub ACP operations and reload the page.
          This is useful to get a fresh keypair after resetting the SourceHub state.
        </p>
      </header>

      <section>
        <button
          type="button"
          onClick={resetKeypair}
          disabled={isResetting}
          className={`${styles.button} ${styles.primary} ${styles.fullWidth}`}
          aria-describedby={result ? 'keypair-result keypair-description' : 'keypair-description'}
          aria-busy={isResetting}
        >
          {isResetting ? 'Resetting Keypair...' : 'Reset Keypair'}
        </button>
      </section>

      {result && (
        <section
          id="keypair-result"
          className={`${styles.resultContainer} ${styles[result.type]}`}
          role={result.type === 'error' ? 'alert' : 'status'}
          aria-live={result.type === 'error' ? 'assertive' : 'polite'}
          aria-label={`Keypair reset result: ${result.type}`}
        >
          <pre className={`${styles.resultText} ${styles[result.type]}`}>
            {result.message}
          </pre>
        </section>
      )}
    </main>
  );
};