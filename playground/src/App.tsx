// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import React from 'react'
import { GraphiQL } from 'graphiql'
import { createGraphiQLFetcher } from '@graphiql/toolkit'
import { GraphiQLPlugin } from '@graphiql/react'
import 'swagger-ui-react/swagger-ui.css'
import 'graphiql/graphiql.css'

const baseUrl = import.meta.env.DEV ? 'http://localhost:9181' : ''
const SwaggerUI = React.lazy(() => import('swagger-ui-react'))
const fetcher = createGraphiQLFetcher({ url: `${baseUrl}/api/v0/graphql` })

const plugin: GraphiQLPlugin = {
  title: 'DefraDB API',
  icon: () => (<div>API</div>),
  content: () => (
    <React.Suspense>
      <SwaggerUI url={`${baseUrl}/openapi.json`} />
    </React.Suspense>
  ),
}

function App() {
  return (<GraphiQL fetcher={fetcher} plugins={[plugin]} />)
}

export default App
