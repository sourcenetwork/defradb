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
import styles from './PluginStyles.module.css';
import { KeypairIcon } from '../icons';

type ResultType = 'success' | 'error' | 'info';

interface KeypairResult {
  message: string;
  type: ResultType;
}

interface KeypairPluginProps {
  clientRef: React.RefObject<any>;
}

export const createKeypairPlugin = (props: KeypairPluginProps): GraphiQLPlugin => ({
  title: 'Keypair Reset',
  icon: KeypairIcon,
  content: () => {
    const [isResetting, setIsResetting] = React.useState(false);
    const [result, setResult] = React.useState<KeypairResult | null>(null);

    const handleResetKeypair = async () => {
      if (!props.clientRef.current) {
        setResult({
          message: 'Error: Client not initialized',
          type: 'error'
        });
        return;
      }

      setIsResetting(true);
      setResult({
        message: 'Deleting keypair...',
        type: 'info'
      });

      try {
        const acpDeleteKeypair = (window as any).acp_DeleteKeypair;
        if (acpDeleteKeypair) {
          const error = await acpDeleteKeypair();
          if (error) {
            setResult({
              message: `Error deleting keypair: ${error.message || error}`,
              type: 'error'
            });
          } else {
            setResult({
              message: 'Keypair reset successfully! Reloading the page...',
              type: 'success'
            });
            setTimeout(() => {
              window.location.reload();
            }, 2000);
          }
        } else {
          setResult({
            message: 'Error: acp_DeleteKeypair function not found',
            type: 'error'
          });
        }
      } catch (error) {
        setResult({
          message: `Error deleting keypair: ${error instanceof Error ? error.message : String(error)}`,
          type: 'error'
        });
      } finally {
        setIsResetting(false);
      }
    };

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
            onClick={handleResetKeypair}
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
  },
}); 