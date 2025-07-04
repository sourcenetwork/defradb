// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React, { useState, useEffect, useRef } from 'react';
import { GraphiQL } from 'graphiql';
import { createGraphiQLFetcher, Fetcher } from '@graphiql/toolkit';
import { GraphiQLPlugin } from '@graphiql/react';
import 'swagger-ui-react/swagger-ui.css';
import 'graphiql/graphiql.css';

const mode = import.meta.env.VITE_PLAYGROUND_MODE;

function App() {
  if (mode === 'wasm') {
    const [client, setClient] = useState<any>(null);
    const initRef = useRef(false);

    useEffect(() => {
      // Initialize the DefraDB client when the Wasm module signals it's ready.
      const initClient = async () => {
        try {
          if (initRef.current) {
            return;
          }
          // @ts-expect-error - window.defradb is set in cmd/defradb/main_js.go.
          if (!window.defradb) {
            setTimeout(initClient, 100);
          } else {
            // @ts-expect-error - window.defradb.open() creates a db client.
            const db = await window.defradb.open();
            // @ts-expect-error - expose window.defradbClient globally.
            window.defradbClient = db;
            initRef.current = true;
            console.log('DefraDB Wasm client initialized.');
            setClient(db);
          }
        } catch (error) {
          console.error('Failed to initialize DefraDB Wasm client:', error);
        }
      };
      initClient();
    }, []);

    if (!client) {
      return <></>;
    }

    const wasmFetcher: Fetcher = async (graphQLParams: any) => {
      try {
        const query = graphQLParams.query || '';
        const variables = graphQLParams.variables || {};
        const operationName = graphQLParams.operationName || {};
        const args = {
          operationName,
          variables,
      };
        // All operations go through execRequest.
        const result = await client.execRequest(query, args);
        return result.gql;
      } catch (error) {
        console.error('Error executing Wasm request:', error);
        const errorMessage = error instanceof Error ? error.message : String(error);
        return { errors: [{ message: errorMessage }] };
      }
    }
    return <GraphiQL fetcher={wasmFetcher} />;
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
