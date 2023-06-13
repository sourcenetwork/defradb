import { useState } from 'react'
import { GraphiQLPlugin, useSchemaContext } from '@graphiql/react'
import { ErrorItem, loadSchema } from './api'
import './plugin.css'

function Errors({ errors }: { errors: ErrorItem[] }) {
  return (
    <ul className="graphiql-defradb-errors">
      {errors.map((error, index) => 
        <li key={index} >{error.message}</li>
      )}
    </ul>
  )
}

const schemaAddPlaceholder = 
`type User {
  name: String 
  age: Int 
  verified: Boolean 
  points: Float
}`

function SchemaAddForm() {
  const schemaContext = useSchemaContext({ nonNull: true })
  const [value, setValue] = useState<string>('')
  const [errors, setErrors] = useState<ErrorItem[]>()
  const [isLoading, setIsLoading] = useState(false)

  const onSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setErrors(undefined)
    setIsLoading(true)

    try {
      const res = await loadSchema(value)
      if (res.errors) {
        setErrors(res.errors)
      } else {
        setValue('')
        schemaContext.introspect()
      }
    } catch(err) {
      setErrors([{ message: 'something went wrong' }])
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={onSubmit}>
      <h4 className="graphiql-defradb-subheader">Add Schema</h4>
      {errors && <Errors errors={errors} />}
      <textarea
        value={value}
        onChange={e => setValue(e.target.value)}
        className="graphiql-defradb-textarea"
        placeholder={schemaAddPlaceholder}
        disabled={isLoading}
      />
      <button 
        type="submit"
        className="graphiql-defradb-button"
        disabled={isLoading}
      >
        Submit
      </button>
    </form>
  )
}

export const plugin: GraphiQLPlugin = {
  title: 'DefraDB',
  icon: () =>{
    return (<div>DB</div>)
  },
  content: () => {
    return (
      <div>
        <h2 className="graphiql-defradb-header">DefraDB</h2>
        <SchemaAddForm />
      </div>
    )
  },
}