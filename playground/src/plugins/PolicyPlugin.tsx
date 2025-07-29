import React from 'react';
import { GraphiQLPlugin } from '@graphiql/react';

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

interface PolicyPluginProps {
  policyRef: React.RefObject<string>;
  clientRef: React.RefObject<any>;
  resultRef: React.RefObject<string>;
  policyIdRef: React.RefObject<string>;
  defaultPolicy: string;
}

export const createPolicyPlugin = (props: PolicyPluginProps): GraphiQLPlugin => ({
  title: 'Add Policy',
  icon: () => <span>🔐</span>,
  content: () => {
    return (
      <div style={{
        padding: '20px',
        backgroundColor: '#202a3b',
        color: '#eaf1fb',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif',
        fontSize: '14px',
        lineHeight: '1.5',
        height: '100%',
        overflowY: 'auto'
      }}>
        <h3 style={{ margin: '0 0 16px 0', fontSize: '18px', fontWeight: '600', color: '#eaf1fb' }}>
          Add Policy
        </h3>
        <p style={{ margin: '0 0 20px 0', color: '#bfc7d5', fontSize: '14px' }}>
          Paste your policy YAML below and click "Add Policy".
        </p>
        <div style={{ marginBottom: '20px' }}>
          <label htmlFor="policy-input" style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Policy YAML
          </label>
          <textarea
            id="policy-input"
            defaultValue={props.defaultPolicy}
            onChange={(e) => {
              props.policyRef.current = e.target.value;
            }}
            style={{
              width: '100%',
              minHeight: '400px',
              fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace',
              fontSize: '13px',
              padding: '16px',
              border: '1px solid #eaf1fb',
              borderRadius: '4px',
              resize: 'vertical',
              backgroundColor: '#2b3546',
              color: '#eaf1fb',
              lineHeight: '1.4',
              outline: 'none',
              boxSizing: 'border-box'
            }}
            placeholder="Enter policy YAML..."
          />
        </div>
        <button
          id="add-policy-button"
          onClick={() => {
            const textarea = document.getElementById('policy-input') as HTMLTextAreaElement;
            const button = document.getElementById('add-policy-button') as HTMLButtonElement;
            const resultDiv = document.getElementById('policy-result');

            if (textarea && button) {
              props.policyRef.current = textarea.value;

              button.disabled = true;
              button.textContent = 'Adding Policy...';
              button.style.backgroundColor = '#2b3546';
              button.style.cursor = 'not-allowed';

              if (resultDiv) {
                resultDiv.innerHTML = '<pre style="margin: 0; color: #bfc7d5; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">Adding policy...</pre>';
                resultDiv.style.display = 'block';
              }

              const handleAddPolicyDirect = async () => {
                if (!props.clientRef.current) {
                  if (resultDiv) {
                    resultDiv.innerHTML = '<pre style="margin: 0; color: #ffb3b3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">Error: Client not initialized</pre>';
                  }
                  return;
                }

                try {
                  const nodeIdentity = await props.clientRef.current.getNodeIdentity();

                  const context = {
                    identity: nodeIdentity.PublicKey
                  };

                  const result = await props.clientRef.current.addDACPolicy(props.policyRef.current, context);
                  const successMessage = `Policy created successfully: ${JSON.stringify(result, null, 2)}`;

                  if (resultDiv) {
                    resultDiv.innerHTML = `<pre style="margin: 0; color: #b3ffb3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">${successMessage}</pre>`;
                  }

                  // Extract policy ID from the result and update state
                  if (result && result.PolicyID) {
                    props.policyIdRef.current = result.PolicyID;
                  } else {
                    console.error("No PolicyID found in result:", result);
                  }
                } catch (error) {
                  const errorMessage = `Error adding policy: ${error instanceof Error ? error.message : String(error)}`;
                  if (resultDiv) {
                    resultDiv.innerHTML = `<pre style="margin: 0; color: #ffb3b3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">${errorMessage}</pre>`;
                  }
                } finally {
                  button.disabled = false;
                  button.textContent = 'Add Policy';
                  button.style.backgroundColor = '#ff5ca7';
                  button.style.cursor = 'pointer';
                }
              };
              handleAddPolicyDirect();
            }
          }}
          style={{
            padding: '10px 20px',
            backgroundColor: '#ff5ca7',
            color: '#eaf1fb',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '14px',
            fontWeight: '500',
            marginBottom: '20px',
            width: '100%',
            transition: 'background-color 0.2s ease'
          }}
        >
          Add Policy
        </button>

        <div
          id="policy-result"
          style={{
            padding: '16px',
            backgroundColor: props.resultRef?.current?.includes('Error') ? '#2d2227' : '#222d26',
            border: `1px solid ${props.resultRef.current.includes('Error') ? '#f5c6cb' : '#c3e6cb'}`,
            borderRadius: '4px',
            marginTop: '10px',
            display: props.resultRef.current ? 'block' : 'none'
          }}
        >
          <pre style={{
            margin: 0,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontSize: '12px',
            fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace',
            color: props.resultRef.current.includes('Error') ? '#ffb3b3' : '#b3ffb3',
            lineHeight: '1.4',
            maxWidth: '100%',
            overflowWrap: 'break-word',
            wordWrap: 'break-word'
          }}>
            {props.resultRef.current}
          </pre>
        </div>
      </div>
    );
  },
}); 