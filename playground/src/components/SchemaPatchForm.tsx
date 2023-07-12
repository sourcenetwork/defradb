import { useState } from 'react'
import { useSchemaContext } from '@graphiql/react'
import { useQueryClient } from '@tanstack/react-query'
import { SchemaForm, FormData } from './SchemaForm'
import { patchSchema, ErrorItem } from '../lib/api'

export type SchemaPatchFormProps = {
  values?: FormData
  fieldTypes: string[]
}

export function SchemaPatchForm({ values, fieldTypes }: SchemaPatchFormProps) {
  const queryClient = useQueryClient()
  const schemaContext = useSchemaContext({ nonNull: true })

  const [errors, setErrors] = useState<ErrorItem[]>()
  const [isLoading, setIsLoading] = useState(false)

  const onSubmit = async (data: FormData) => {
    setErrors(undefined)
    setIsLoading(true)

    try {
      const res = await patchSchema(values!.name, values!.fields, data.name, data.fields)
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
      values={values}
      fieldTypes={fieldTypes}
    />
  ) 
}