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
          // @ts-expect-error - defradbClient is set by Go
          if (!window.defradb) {
            setTimeout(initClient, 100);
          } else {
            // @ts-expect-error - defradb is a global object created by Wasm
            const db = await window.defradb.open();
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
        const lowerQuery = query.toLowerCase();
        if (lowerQuery.includes('addschema')) {
          const schemaMatch = query.match(/schema:\s*"""([\s\S]*)"""/);
          if (schemaMatch && schemaMatch[1]) {
            const schema = schemaMatch[1];
            const result = await client.addSchema(schema);
            return { data: { addSchema: result } };
          }
        }
        if (lowerQuery.includes('patchschema')) {
          const patchMatch = query.match(/patch:\s*"""([\s\S]*?)"""/);
          const migrationMatch = query.match(/migration:\s*"""([\s\S]*?)"""/);
          const setActiveMatch = query.match(/setasdefaultversion:\s*(true|false)/i);
          const patch = patchMatch ? patchMatch[1] : '';
          const migration = migrationMatch ? JSON.parse(migrationMatch[1]) : undefined;
          const setAsDefaultVersion = setActiveMatch ? setActiveMatch[1].toLowerCase() === 'true' : false;
          const result = await client.patchSchema(patch, migration, setAsDefaultVersion);
          return { data: { patchSchema: result } };
        }
        if (lowerQuery.includes('addview')) {
          const queryMatch = query.match(/query:\s*"""([\s\S]*?)"""/);
          const sdlMatch = query.match(/sdl:\s*"""([\s\S]*?)"""/);
          const transformMatch = query.match(/transform:\s*"""([\s\S]*?)"""/);
          const gqlQuery = queryMatch ? queryMatch[1] : '';
          const sdl = sdlMatch ? sdlMatch[1] : '';
          const transform = transformMatch ? JSON.parse(transformMatch[1]) : undefined;
          const result = await client.addView(gqlQuery, sdl, transform);
          await client.refreshViews({});
          return { data: { addView: result } };
        }
        // All other operations (queries, other mutations) go through execRequest.
        const result = await client.execRequest(query, JSON.stringify(variables));
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
