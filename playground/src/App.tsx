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
