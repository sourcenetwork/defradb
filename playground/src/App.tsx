// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import { useEffect } from 'react';
import { GraphiQL } from 'graphiql';
import { usePlaygroundStore } from './store/playgroundStore';
import { useGraphiQLConfig } from './hooks/useGraphiQLConfig';
import { DefraDBLogo } from './assets/icons/DefraDBLogo';
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
  const initializeClient = usePlaygroundStore((state) => state.initializeClient);
  const config = useGraphiQLConfig();

  useEffect(() => {
    // Only initialize client if in WASM mode
    if (mode !== 'wasm') {
      return;
    }
    initializeClient();
  }, []);

  if (config === null) {
    return null;
  }

  return (
    <div className="defradb-playground">
      <GraphiQL
        fetcher={config.fetcher}
        plugins={config.plugins}
      >
        <GraphiQL.Logo>
          <DefraDBLogo />
          DefraDB Playground
        </GraphiQL.Logo>
      </GraphiQL>
    </div>
  );
}

export default App;
