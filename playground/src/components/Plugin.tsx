import { useQuery } from '@tanstack/react-query'
import { SchemaLoadForm } from './SchemaLoadForm'
import { SchemaPatchForm } from './SchemaPatchForm'
import { listSchema } from '../lib/api'

export function Plugin() {
  const { data } = useQuery({ queryKey: ['schemas'], queryFn: listSchema })
  return (
    <div>
      <h2 className="graphiql-defradb-header">DefraDB</h2>
      <div className="graphiql-defradb-plugin">
        <div>
          <h3 className="graphiql-defradb-subheader">Create</h3>
          <SchemaLoadForm />
        </div>
        { data?.data.collections?.map((schema, index) => 
          <div key={index}>
            <h3 className="graphiql-defradb-subheader">{schema.name}</h3>
            <SchemaPatchForm values={schema} />
          </div>
        )}
      </div>
    </div>
  )
}