import { createGraphiQLFetcher } from '@graphiql/toolkit'
import { GraphiQL } from 'graphiql'
import 'graphiql/graphiql.css'

const fetcher = createGraphiQLFetcher({ url: 'http://localhost:9181/api/v0/graphql' })

function App() {
  return (
    <GraphiQL fetcher={fetcher}>
    </GraphiQL>
  )
}

export default App
