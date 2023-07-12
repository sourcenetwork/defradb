import { useQuery } from '@tanstack/react-query'
import { SchemaLoadForm } from './SchemaLoadForm'
import { SchemaPatchForm } from './SchemaPatchForm'
import { listSchema } from '../lib/api'

const defaultFieldTypes = [
  'ID',
  'Boolean',
  '[Boolean]',
  '[Boolean!]',
  'Int',
  '[Int]',
  '[Int!]',
  'DateTime',
  'Float',
  '[Float]',
  '[Float!]',
  'String',
  '[String]',
  '[String!]',
]

export function Plugin() {
  const { data } = useQuery({ queryKey: ['schemas'], queryFn: listSchema })

  const collections = data?.data?.collections ?? []
  const schemaFieldTypes = collections.map(col => [`${col.name}`, `[${col.name}]`]).flat()
  const fieldTypes = [...defaultFieldTypes, ...schemaFieldTypes]

  return (
    <div>
      <h2 className="graphiql-defradb-header">DefraDB</h2>
      <div className="graphiql-defradb-plugin">
        <div>
          <h3 className="graphiql-defradb-subheader">Create</h3>
          <SchemaLoadForm fieldTypes={fieldTypes} />
        </div>
        { collections?.map((schema, index) => 
          <div key={index}>
            <h3 className="graphiql-defradb-subheader">{schema.name}</h3>
            <SchemaPatchForm fieldTypes={fieldTypes} values={schema} />
          </div>
        )}
      </div>
    </div>
  )
}