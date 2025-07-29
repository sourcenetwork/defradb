import React from 'react';
import { GraphiQLPlugin } from '@graphiql/react';

interface KeypairPluginProps {
  clientRef: React.RefObject<any>;
}

export const createKeypairPlugin = (props: KeypairPluginProps): GraphiQLPlugin => ({
  title: 'Keypair Reset',
  icon: () => <span>🔑</span>,
  content: () => {
    const [isResetting, setIsDeleting] = React.useState(false);
    const [result, setResult] = React.useState<string>('');

    const handleResetKeypair = async () => {
      if (!props.clientRef.current) {
        setResult('Error: Client not initialized');
        return;
      }
      setIsDeleting(true);
      setResult('Deleting keypair...');
      try {
        const acpDeleteKeypair = (window as any).acp_DeleteKeypair;
        if (acpDeleteKeypair) {
          const error = await acpDeleteKeypair();
          if (error) {
            setResult(`Error deleting keypair: ${error.message || error}`);
          } else {
            setResult('Keypair reset successfully! Refreshihg the page...');
          }
        } else {
          setResult('Error: acp_DeleteKeypair function not found');
        }
      } catch (error) {
        setResult(`Error deleting keypair: ${error instanceof Error ? error.message : String(error)}`);
      } finally {
        setIsDeleting(false);
      }
      window.location.reload();
    };

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
          Keypair Reset
        </h3>
        <p style={{ margin: '0 0 20px 0', color: '#bfc7d5', fontSize: '14px' }}>
          Optionally, reset the keypair used for SourceHub ACP operations and reload the page. This is useful to get a fresh keypair after resetting the SourceHub state.
        </p>
        
        <button
          onClick={handleResetKeypair}
          disabled={isResetting}
          style={{
            padding: '10px 20px',
            backgroundColor: isResetting ? '#2b3546' : '#ff5ca7',
            color: '#eaf1fb',
            border: 'none',
            borderRadius: '4px',
            cursor: isResetting ? 'not-allowed' : 'pointer',
            fontSize: '14px',
            fontWeight: '500',
            marginBottom: '20px',
            width: '100%',
            transition: 'background-color 0.2s ease'
          }}
        >
          {isResetting ? 'Resetting Keypair...' : 'Reset Keypair'}
        </button>

        {result && (
          <div
            style={{
              padding: '16px',
              backgroundColor: result.includes('Error') ? '#2d2227' : '#222d26',
              border: `1px solid ${result.includes('Error') ? '#f5c6cb' : '#c3e6cb'}`,
              borderRadius: '4px',
              marginTop: '10px'
            }}
          >
            <pre style={{
              margin: 0,
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              fontSize: '12px',
              fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace',
              color: result.includes('Error') ? '#ffb3b3' : '#b3ffb3',
              lineHeight: '1.4',
              maxWidth: '100%',
              overflowWrap: 'break-word',
              wordWrap: 'break-word'
            }}>
              {result}
            </pre>
          </div>
        )}
      </div>
    );
  },
}); 