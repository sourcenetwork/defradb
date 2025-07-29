import React from 'react';

export const DEFAULT_RELATIONSHIP = {
  collectionName: 'Users',
  docID: 'document_id',
  relation: 'collaborator',
  targetActor: 'did:key:alice'
};

export interface RelationshipPluginProps {
  clientRef: React.RefObject<any>;
  policyIdRef: React.RefObject<string>;
  relationshipRef: React.RefObject<typeof DEFAULT_RELATIONSHIP>;
  resultRef: React.RefObject<string>;
}

export const createRelationshipPlugin = ({
  clientRef,
  policyIdRef,
  relationshipRef,
  resultRef,
}: RelationshipPluginProps) => ({
  title: 'Actor Relationships',
  icon: () => <span>👥</span>,
  content: () => {
    const [isVerifyLoading, setIsVerifyLoading] = React.useState(false);
    const [isAddLoading, setIsAddLoading] = React.useState(false);
    const [isDeleteLoading, setIsDeleteLoading] = React.useState(false);

    const [relationship, setRelationship] = React.useState(DEFAULT_RELATIONSHIP);

    React.useEffect(() => {
      relationshipRef.current = relationship;
    }, [relationship, relationshipRef]);

    const handleAddRelationship = async () => {
      if (!clientRef.current) {
        resultRef.current = 'Error: Client not initialized';
        return;
      }

      setIsAddLoading(true);
      resultRef.current = 'Adding relationship...';

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();

        const context = {
          identity: nodeIdentity.PublicKey
        };

        const result = await clientRef.current.addDACActorRelationship(
          relationship.collectionName,
          relationship.docID,
          relationship.relation,
          relationship.targetActor,
          context
        );

        resultRef.current = `Relationship added successfully: ${JSON.stringify(result, null, 2)}`;
      } catch (error) {
        resultRef.current = `Error adding relationship: ${error instanceof Error ? error.message : String(error)}`;
      } finally {
        setIsAddLoading(false);
      }
    };

    const handleDeleteRelationship = async () => {
      if (!clientRef.current) {
        resultRef.current = 'Error: Client not initialized';
        return;
      }

      setIsDeleteLoading(true);
      resultRef.current = 'Deleting relationship...';

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();

        const context = {
          identity: nodeIdentity.PublicKey
        };

        const result = await clientRef.current.deleteDACActorRelationship(
          relationship.collectionName,
          relationship.docID,
          relationship.relation,
          relationship.targetActor,
          context
        );

        resultRef.current = `Relationship deleted successfully: ${JSON.stringify(result, null, 2)}`;
      } catch (error) {
        resultRef.current = `Error deleting relationship: ${error instanceof Error ? error.message : String(error)}`;
      } finally {
        setIsDeleteLoading(false);
      }
    };

    const handleVerifyAccess = async () => {
      if (!clientRef.current) {
        resultRef.current = 'Error: Client not initialized';
        return;
      }

      setIsVerifyLoading(true);
      resultRef.current = 'Verifying access...';

      try {
        const nodeIdentity = await clientRef.current.getNodeIdentity();

        const context = {
          identity: nodeIdentity.PublicKey
        };

        if (!policyIdRef.current) {
          resultRef.current = 'Error: Policy ID not available. Please add a policy first.';
          return;
        }

        const result = await clientRef.current.verifyDACAccess(
          "read",
          relationship.targetActor,
          policyIdRef.current,
          relationship.collectionName.toLowerCase(),
          relationship.docID,
          context
        );

        resultRef.current = `Access verification result: ${JSON.stringify(result, null, 2)}`;
      } catch (error) {
        resultRef.current = `Error verifying access: ${error instanceof Error ? error.message : String(error)}`;
      } finally {
        setIsVerifyLoading(false);
      }
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
          Actor Relationships
        </h3>
        <p style={{ margin: '0 0 20px 0', color: '#bfc7d5', fontSize: '14px' }}>
          Add or delete actor relationships for testing ACP functionality.
        </p>

        <div style={{ marginBottom: '20px' }}>
          <label style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Collection Name
          </label>
          <input
            type="text"
            value={relationship.collectionName}
            onChange={(e) => setRelationship(prev => ({ ...prev, collectionName: e.target.value }))}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #eaf1fb',
              borderRadius: '4px',
              backgroundColor: '#2b3546',
              color: '#eaf1fb',
              fontSize: '14px',
              outline: 'none',
              boxSizing: 'border-box'
            }}
            placeholder="e.g. Users"
          />
        </div>

        <div style={{ marginBottom: '20px' }}>
          <label style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Document ID
          </label>
          <input
            type="text"
            value={relationship.docID}
            onChange={(e) => setRelationship(prev => ({ ...prev, docID: e.target.value }))}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #eaf1fb',
              borderRadius: '4px',
              backgroundColor: '#2b3546',
              color: '#eaf1fb',
              fontSize: '14px',
              outline: 'none',
              boxSizing: 'border-box'
            }}
            placeholder="e.g. bae-12345678-1234-1234-1234-123456789abc"
          />
        </div>

        <div style={{ marginBottom: '20px' }}>
          <label style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Relation
          </label>
          <input
            type="text"
            value={relationship.relation}
            onChange={(e) => setRelationship(prev => ({ ...prev, relation: e.target.value }))}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #eaf1fb',
              borderRadius: '4px',
              backgroundColor: '#2b3546',
              color: '#eaf1fb',
              fontSize: '14px',
              outline: 'none',
              boxSizing: 'border-box'
            }}
            placeholder="e.g. collaborator | reader | editor"
          />
        </div>

        <div style={{ marginBottom: '20px' }}>
          <label style={{
            display: 'block',
            marginBottom: '8px',
            fontWeight: '500',
            color: '#eaf1fb',
            fontSize: '14px'
          }}>
            Target Actor (DID)
          </label>
          <input
            type="text"
            value={relationship.targetActor}
            onChange={(e) => setRelationship(prev => ({ ...prev, targetActor: e.target.value }))}
            style={{
              width: '100%',
              padding: '8px 12px',
              border: '1px solid #eaf1fb',
              borderRadius: '4px',
              backgroundColor: '#2b3546',
              color: '#eaf1fb',
              fontSize: '14px',
              outline: 'none',
              boxSizing: 'border-box'
            }}
            placeholder="e.g. did:key:alice or * for all actors"
          />
        </div>

        <div style={{ display: 'flex', gap: '10px', marginBottom: '20px' }}>
          <button
            onClick={handleAddRelationship}
            disabled={isAddLoading}
            style={{
              flex: 1,
              padding: '10px 20px',
              backgroundColor: isAddLoading ? '#2b3546' : '#4CAF50',
              color: '#eaf1fb',
              border: 'none',
              borderRadius: '4px',
              cursor: isAddLoading ? 'not-allowed' : 'pointer',
              fontSize: '14px',
              fontWeight: '500',
              transition: 'background-color 0.2s ease'
            }}
          >
            {isAddLoading ? 'Adding...' : 'Add Relationship'}
          </button>

          <button
            onClick={handleDeleteRelationship}
            disabled={isDeleteLoading}
            style={{
              flex: 1,
              padding: '10px 20px',
              backgroundColor: isDeleteLoading ? '#2b3546' : '#f44336',
              color: '#eaf1fb',
              border: 'none',
              borderRadius: '4px',
              cursor: isDeleteLoading ? 'not-allowed' : 'pointer',
              fontSize: '14px',
              fontWeight: '500',
              transition: 'background-color 0.2s ease'
            }}
          >
            {isDeleteLoading ? 'Deleting...' : 'Delete Relationship'}
          </button>
        </div>

        <div style={{ marginBottom: '20px' }}>
          <button
            onClick={handleVerifyAccess}
            disabled={isVerifyLoading}
            style={{
              width: '100%',
              padding: '10px 20px',
              backgroundColor: isVerifyLoading ? '#2b3546' : '#2196F3',
              color: '#eaf1fb',
              border: 'none',
              borderRadius: '4px',
              cursor: isVerifyLoading ? 'not-allowed' : 'pointer',
              fontSize: '14px',
              fontWeight: '500',
              transition: 'background-color 0.2s ease'
            }}
          >
            {isVerifyLoading ? 'Verifying...' : 'Verify Access'}
          </button>
        </div>

        {resultRef.current && (
          <div style={{
            padding: '16px',
            backgroundColor: resultRef.current.includes('Error') ? '#2d2227' : '#222d26',
            border: `1px solid ${resultRef.current.includes('Error') ? '#f5c6cb' : '#c3e6cb'}`,
            borderRadius: '4px',
            marginTop: '10px'
          }}>
            <pre style={{
              margin: 0,
              whiteSpace: 'pre-wrap',
              wordBreak: 'break-word',
              fontSize: '12px',
              fontFamily: 'Consolas, "Liberation Mono", Menlo, Courier, monospace',
              color: resultRef.current.includes('Error') ? '#ffb3b3' : '#b3ffb3',
              lineHeight: '1.4',
              maxWidth: '100%',
              overflowWrap: 'break-word',
              wordWrap: 'break-word'
            }}>
              {resultRef.current}
            </pre>
          </div>
        )}
      </div>
    );
  },
}); 