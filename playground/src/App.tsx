// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import { GraphiQL } from 'graphiql'
import { createGraphiQLFetcher } from '@graphiql/toolkit'
import { GraphiQLPlugin } from '@graphiql/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Plugin } from './components/Plugin'
import 'graphiql/graphiql.css'

const client = new QueryClient()
const fetcher = createGraphiQLFetcher({ url: 'http://localhost:9181/api/v0/graphql' })

const plugin: GraphiQLPlugin = {
  title: 'DefraDB',
  icon: () => (<div>DB</div>),
  content: () => (<Plugin />),
}

function App() {
  return (
    <QueryClientProvider client={client}>
      <GraphiQL fetcher={fetcher} plugins={[plugin]} />
    </QueryClientProvider>
  )
}

export default App
