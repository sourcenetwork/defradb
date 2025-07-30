// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React, { useEffect, useCallback } from 'react';
import { GraphiQL } from 'graphiql';
import { createGraphiQLFetcher, Fetcher } from '@graphiql/toolkit';
import { GraphiQLPlugin } from '@graphiql/react';
import { usePlaygroundStore } from './store/playgroundStore';
import { policyPlugin } from './plugins/PolicyPlugin';
import { schemaPlugin } from './plugins/SchemaPlugin';
import { relationshipPlugin } from './plugins/RelationshipPlugin';
import { keypairResetPlugin } from './plugins/KeypairResetPlugin';
import 'swagger-ui-react/swagger-ui.css';
import 'graphiql/graphiql.css';

// Declare global types for DefraDB WASM integration
declare global {
  interface Window {
    defradb: {
      open(_acpType?: string): Promise<any>;
    };
    defradbClient: any;
  }
}

const mode = import.meta.env.VITE_PLAYGROUND_MODE;

function App() {
  const status = usePlaygroundStore((state) => state.client.status);
  const isInitialized = usePlaygroundStore((state) => state.client.isInitialized);
  const isSourceHubAvailable = usePlaygroundStore((state) => state.client.isSourceHubAvailable);
  const initializeClient = usePlaygroundStore((state) => state.initializeClient);

  useEffect(() => {
    // Only initialize client if in WASM mode
    if (mode !== 'wasm') {
      return;
    }
    initializeClient();
  }, []);

  const wasmFetcher: Fetcher = useCallback(async (graphQLParams: any) => {
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
  }, []);

  if (mode === 'wasm') {
    if (!isInitialized || status !== 'ready') {
      return null;
    }

    return (
      <div className="defradb-playground">
        <GraphiQL
          fetcher={wasmFetcher}
          plugins={[
            ...(isSourceHubAvailable ? [keypairResetPlugin] : []),
            policyPlugin,
            schemaPlugin,
            relationshipPlugin,
          ]}
        />
      </div>
    );
  } else {
    const baseUrl = import.meta.env.DEV ? 'http://localhost:9181' : '';
    const SwaggerUI = React.lazy(() => import('swagger-ui-react'));
    const fetcher = createGraphiQLFetcher({ url: `${baseUrl}/api/v0/graphql` });
    const plugin: GraphiQLPlugin = {
      title: 'DefraDB API',
      icon: () => (<div>API</div>),
      content: () => (
        <React.Suspense>
          <SwaggerUI url={`${baseUrl}/openapi.json`} />
        </React.Suspense>
      ),
    };
    return (<GraphiQL fetcher={fetcher} plugins={[plugin]} />);
  }
}

export default App;
