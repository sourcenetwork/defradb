import React from 'react';
import { GraphiQLPlugin } from '@graphiql/react';

export const DEFAULT_SCHEMA = `type Users @policy(
  id: "policy_id",
  resource: "users"
) {
  name: String
  age: Int
}`;

interface SchemaPluginProps {
  schemaRef: React.RefObject<string>;
  clientRef: React.RefObject<any>;
  policyIdRef: React.RefObject<string>;
  defaultSchema: string;
}

export const createSchemaPlugin = (props: SchemaPluginProps): GraphiQLPlugin => ({
  title: 'Add Schema',
  icon: () => <span>📋</span>,
  content: () => {
    const currentSchema = props.defaultSchema.replace('policy_id', props.policyIdRef.current);

    React.useEffect(() => {
      const textarea = document.getElementById('schema-input') as HTMLTextAreaElement;
      if (textarea) {
        const updatedSchema = props.defaultSchema.replace('policy_id', props.policyIdRef.current);
        textarea.value = updatedSchema;
        props.schemaRef.current = updatedSchema;
      }
    });

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
          Add Schema
        </h3>
        <p style={{ margin: '0 0 20px 0', color: '#bfc7d5', fontSize: '14px' }}>
          Edit your schema below and click "Add Schema". The policy ID will be automatically populated if a policy was previously created.
        </p>
        <div style={{ marginBottom: '20px' }}>
          <label htmlFor="schema-input" style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Schema GraphQL
          </label>
          <textarea
            id="schema-input"
            defaultValue={currentSchema}
            onChange={(e) => {
              props.schemaRef.current = e.target.value;
            }}
            style={{
              width: '100%',
              minHeight: '200px',
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
            placeholder="Enter schema GraphQL..."
          />
        </div>
        <button
          id="add-schema-button"
          onClick={() => {
            const textarea = document.getElementById('schema-input') as HTMLTextAreaElement;
            const button = document.getElementById('add-schema-button') as HTMLButtonElement;
            const resultDiv = document.getElementById('schema-result');

            if (textarea && button) {
              props.schemaRef.current = textarea.value;

              button.disabled = true;
              button.textContent = 'Adding Schema...';
              button.style.backgroundColor = '#2b3546';
              button.style.cursor = 'not-allowed';

              if (resultDiv) {
                resultDiv.innerHTML = '<pre style="margin: 0; color: #bfc7d5; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">Adding schema...</pre>';
                resultDiv.style.display = 'block';
              }

              const handleAddSchemaDirect = async () => {
                if (!props.clientRef.current) {
                  if (resultDiv) {
                    resultDiv.innerHTML = '<pre style="margin: 0; color: #ffb3b3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">Error: Client not initialized</pre>';
                  }
                  return;
                }

                try {
                  const result = await props.clientRef.current.addSchema(props.schemaRef.current);
                  const successMessage = `Schema added successfully: ${JSON.stringify(result, null, 2)}`;

                  if (resultDiv) {
                    resultDiv.innerHTML = `<pre style="margin: 0; color: #b3ffb3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">${successMessage}</pre>`;
                  }
                } catch (error) {
                  const errorMessage = `Error adding schema: ${error instanceof Error ? error.message : String(error)}`;
                  if (resultDiv) {
                    resultDiv.innerHTML = `<pre style="margin: 0; color: #ffb3b3; white-space: pre-wrap; word-break: break-word; max-width: 100%; overflow-wrap: break-word; word-wrap: break-word;">${errorMessage}</pre>`;
                  }
                } finally {
                  button.disabled = false;
                  button.textContent = 'Add Schema';
                  button.style.backgroundColor = '#ff5ca7';
                  button.style.cursor = 'pointer';
                }
              };
              handleAddSchemaDirect();
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
          Add Schema
        </button>

        <div
          id="schema-result"
          style={{
            padding: '16px',
            backgroundColor: '#222d26',
            border: '1px solid #c3e6cb',
            borderRadius: '4px',
            marginTop: '10px',
            display: 'none'
          }}
        >
          <pre style={{
            margin: 0,
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontSize: '12px',
            fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace',
            color: '#b3ffb3',
            lineHeight: '1.4',
            maxWidth: '100%',
            overflowWrap: 'break-word',
            wordWrap: 'break-word'
          }}>
          </pre>
        </div>
      </div>
    );
  },
}); 