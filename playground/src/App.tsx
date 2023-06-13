import { createGraphiQLFetcher } from '@graphiql/toolkit'
import { GraphiQL } from 'graphiql'
import { plugin } from './plugin'
import 'graphiql/graphiql.css'

const fetcher = createGraphiQLFetcher({ url: 'http://localhost:9181/api/v0/graphql' })

function App() {
  return (
    <GraphiQL fetcher={fetcher} plugins={[plugin]}>
    </GraphiQL>
  )
}

export default App
