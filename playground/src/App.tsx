// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { GraphiQL } from 'graphiql';
import { createGraphiQLFetcher, Fetcher } from '@graphiql/toolkit';
import { GraphiQLPlugin } from '@graphiql/react';
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
const acpClient = import.meta.env.VITE_ACP_CLIENT;

function App() {
  const initRef = useRef(false);
  const clientRef = useRef<any>(null);

  const [isClientReady, setIsClientReady] = useState(false);
  const [isSourceHubAvailable, setIsSourceHubAvailable] = useState(false);

  const checkSourceHubAvailability = useCallback(async (): Promise<boolean> => {
    try {
      const response = await fetch('/api/cosmos/base/tendermint/v1beta1/node_info');
      return response.ok;
    } catch (error) {
      console.log('SourceHub not available:', error);
      return false;
    }
  }, []);

  useEffect(() => {
    // Only initialize client if in WASM mode
    if (mode !== 'wasm') {
      return;
    }

    // Initialize the DefraDB client when the Wasm module signals it's ready
    const initClient = async () => {
      try {
        if (initRef.current) {
          return;
        }
        if (!window.defradb) {
          setTimeout(initClient, 100);
          return;
        }
        // Set ref before async call to prevent race condition
        initRef.current = true;
        let useSourceHub = false;
        if (acpClient === 'sourcehub') {
          useSourceHub = await checkSourceHubAvailability();
        }
        setIsSourceHubAvailable(useSourceHub);
        const db = useSourceHub
          ? await window.defradb.open('sourcehub')
          : await window.defradb.open();
        window.defradbClient = db;
        clientRef.current = db;
        setIsClientReady(true);
        console.log('DefraDB Wasm client initialized with', useSourceHub ? 'SourceHub ACP' : 'Local ACP');
      } catch (error) {
        console.error('Failed to initialize DefraDB Wasm client:', error);
      }
    };
    initClient();
  }, [checkSourceHubAvailability]);

  const wasmFetcher: Fetcher = useCallback(async (graphQLParams: any) => {
    try {
      const query = graphQLParams.query ?? '';
      const variables = graphQLParams.variables ?? {};
      const operationName = graphQLParams.operationName ?? {};
      const args = {
        operationName,
        variables,
      };
      const nodeIdentity = await clientRef.current?.getNodeIdentity();
      // Create context with identity
      const context = {
        identity: nodeIdentity?.PublicKey,
      };
      // All operations go through execRequest
      const result = await clientRef.current?.execRequest(query, args, context);
      return result?.gql;
    } catch (error) {
      console.error('Error executing Wasm request:', error);
      const errorMessage = error instanceof Error ? error.message : String(error);
      return { errors: [{ message: errorMessage }] };
    }
  }, []);

  if (mode === 'wasm') {
    if (!isClientReady) {
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
