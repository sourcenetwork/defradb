import { useState } from 'react'
import { useSchemaContext } from '@graphiql/react'
import { useQueryClient } from '@tanstack/react-query'
import { SchemaForm, FormData } from './SchemaForm'
import { loadSchema, ErrorItem } from '../lib/api'

export type SchemaLoadFormProps = {
  fieldTypes: string[]
}

export function SchemaLoadForm({ fieldTypes }: SchemaLoadFormProps) {
  const queryClient = useQueryClient()
  const schemaContext = useSchemaContext({ nonNull: true })

  const [errors, setErrors] = useState<ErrorItem[]>()
  const [isLoading, setIsLoading] = useState(false)

  const onSubmit = async (data: FormData) => {
    setErrors(undefined)
    setIsLoading(true)

    try {
      const res = await loadSchema(data.name, data.fields)
      if (res.errors) {
        setErrors(res.errors)
      } else {
        schemaContext.introspect()
        queryClient.invalidateQueries(['schemas'])
      }
    } catch(err: any) {
      setErrors([{ message: err.message }])
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <SchemaForm
      errors={errors}
      isLoading={isLoading}
      onSubmit={onSubmit}
      fieldTypes={fieldTypes}
    />
  ) 
}