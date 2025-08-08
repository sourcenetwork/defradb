import { useMemo } from 'react';
import { createGraphiQLFetcher, Fetcher } from '@graphiql/toolkit';
import { usePlaygroundStore } from '../store/playgroundStore';
import { policyPlugin } from '../plugins/PolicyPlugin';
import { schemaPlugin } from '../plugins/SchemaPlugin';
import { relationshipPlugin } from '../plugins/RelationshipPlugin';
import { keypairResetPlugin } from '../plugins/KeypairResetPlugin';
import { swaggerPlugin } from '../plugins/SwaggerPlugin';

const mode = import.meta.env.VITE_PLAYGROUND_MODE;

export const useGraphiQLConfig = () => {
  const status = usePlaygroundStore((state) => state.client.status);
  const isInitialized = usePlaygroundStore((state) => state.client.isInitialized);
  const isSourceHubAvailable = usePlaygroundStore((state) => state.client.isSourceHubAvailable);

  return useMemo(() => {
    if (mode === 'wasm') {
      if (!isInitialized || status !== 'ready') {
        return null;
      }
      const wasmFetcher: Fetcher = async (graphQLParams: any) => {
        try {
          const query = graphQLParams.query ?? '';
          const variables = graphQLParams.variables ?? {};
          const operationName = graphQLParams.operationName ?? {};
          const args = {
            operationName,
            variables,
          };
          const nodeIdentity = await window.defradbClient?.getNodeIdentity();
          // Create context with identity
          const context = {
            identity: nodeIdentity?.PublicKey,
          };
          // All operations go through execRequest
          const result = await window.defradbClient?.execRequest(query, args, context);
          return result?.gql;
        } catch (error) {
          console.error('Error executing Wasm request:', error);
          const errorMessage = error instanceof Error ? error.message : String(error);
          return { errors: [{ message: errorMessage }] };
        }
      };
      const plugins = [
        policyPlugin,
        schemaPlugin,
        relationshipPlugin,
        ...(isSourceHubAvailable ? [keypairResetPlugin] : []),
      ];
      return { fetcher: wasmFetcher, plugins };
    } else {
      const baseUrl = import.meta.env.DEV ? 'http://localhost:9181' : '';
      const fetcher = createGraphiQLFetcher({ url: `${baseUrl}/api/v0/graphql` });
      const plugins = [swaggerPlugin];
      return { fetcher, plugins };
    }
  }, [isInitialized, status, isSourceHubAvailable]);
}; 