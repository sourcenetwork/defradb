import { useEffect } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import { Field, ErrorItem } from '../lib/api'

export type FormData = {
  name: string
  fields: Field[]
}

export type SchemaFormProps = {
  id?: string
  errors?: ErrorItem[]
  isLoading: boolean
  onSubmit: (data: FormData) => void
  values?: FormData
  fieldTypes: string[]
}

const defaultValue: FormData = {
  name: '',
  fields: [{ name: 'name', kind: 'String', internal: false }]
}

export function SchemaForm({ errors, isLoading, onSubmit, values = defaultValue, fieldTypes }: SchemaFormProps) {
  const { control, formState, reset, register, handleSubmit } = useForm<FormData>({ values })
  const { fields, append, remove } = useFieldArray({ control, name: 'fields', keyName: '_id' })

  useEffect(() => {
    if (formState.isSubmitSuccessful) reset(values)
  }, [formState, reset])

  return (
    <form 
      className="graphiql-defradb-form"
      onSubmit={handleSubmit(onSubmit)}
    >
      {errors?.map((error, index) =>
        <div key={index} className="graphiql-defradb-error">
          <span>{error.message}</span>
        </div>
      )}
      <h5 className="graphiql-defradb-input-label">Name</h5>
      <input 
        className="graphiql-defradb-input"
        disabled={isLoading}
        {...register('name', {required: true})} 
      />
      <div className="graphiql-defradb-field-header">
        <h5 className="graphiql-defradb-input-label">Fields</h5>
        <button
          type="button"
          className="graphiql-defradb-button"
          onClick={() => append({ name: '', kind: 'Boolean', internal: false })}
        >
          Add
        </button>
      </div>
      {fields.map((field, index) =>
        <div key={field._id} className="graphiql-defradb-field" style={{ display: field.internal ? 'none' : undefined }}>
          <input
            className="graphiql-defradb-input"
            disabled={isLoading || !!field.id}
            {...register(`fields.${index}.name`)}
          />
          <select
            className="graphiql-defradb-input"
            disabled={isLoading || !!field.id}
            {...register(`fields.${index}.kind`)}
          >
            {fieldTypes.map((value, index) => 
              <option key={index} value={value}>{value}</option>
            )}
          </select>
          <button
            type="button"
            className="graphiql-defradb-button"
            onClick={() => remove(index)}
            disabled={isLoading || !!field.id}
          >
            Remove
          </button>
        </div>
      )}
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