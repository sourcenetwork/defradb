import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { useSchemaContext } from '@graphiql/react'
import { useQueryClient } from '@tanstack/react-query'
import { loadSchema, ErrorItem } from '../lib/api'

export type FormData = {
  schema: string
}

const defaultValues: FormData = {
  schema: '',
}

export function SchemaLoadForm() {
  const queryClient = useQueryClient()
  const schemaContext = useSchemaContext({ nonNull: true })

  const { formState, reset, register, handleSubmit } = useForm<FormData>({ defaultValues })

  const [errors, setErrors] = useState<ErrorItem[]>()
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    if (formState.isSubmitSuccessful) reset(defaultValues)
  }, [formState, reset])

  const onSubmit = async (data: FormData) => {
    setErrors(undefined)
    setIsLoading(true)

    try {
      const res = await loadSchema(data.schema)
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
    <form
      className="graphiql-defradb-load-form"
      onSubmit={handleSubmit(onSubmit)}
    >
      {errors?.map((error, index) =>
        <div key={index} className="graphiql-defradb-error">
          <span>{error.message}</span>
        </div>
      )}
      <textarea 
        className="graphiql-defradb-textarea"
        disabled={isLoading}
        {...register('schema', {required: true})} 
      />
      <button 
        type="submit"
        className="graphiql-button"
        disabled={isLoading}
      >
        Submit
      </button>
    </form>
  ) 
}