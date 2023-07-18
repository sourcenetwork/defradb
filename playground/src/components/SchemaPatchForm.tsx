// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

import { useState } from 'react'
import { useForm, useFieldArray } from 'react-hook-form'
import { useSchemaContext } from '@graphiql/react'
import { useQueryClient } from '@tanstack/react-query'
import { patchSchema, Field, ErrorItem } from '../lib/api'

export type FormData = {
  name: string
  fields: Field[]
}

export type SchemaPatchFormProps = {
  values?: FormData
  fieldTypes: string[]
}

export function SchemaPatchForm({ values, fieldTypes }: SchemaPatchFormProps) {
  const queryClient = useQueryClient()
  const schemaContext = useSchemaContext({ nonNull: true })

  const [errors, setErrors] = useState<ErrorItem[]>()
  const [isLoading, setIsLoading] = useState(false)

  const { control, register, handleSubmit } = useForm<FormData>({ values })
  const { fields, append, remove } = useFieldArray({ control, name: 'fields', keyName: '_id' })

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
    <form 
      className="graphiql-defradb-form"
      onSubmit={handleSubmit(onSubmit)}
    >
      {errors?.map((error, index) =>
        <div key={index} className="graphiql-defradb-error">
          <span>{error.message}</span>
        </div>
      )}
      <div className="graphiql-defradb-field-header">
        <h5 className="graphiql-defradb-input-label">Fields</h5>
        <button
          type="button"
          className="graphiql-button"
          onClick={() => append({ name: '', kind: 'String', internal: false })}
        >
          Add
        </button>
      </div>
      {fields.map((field, index) =>
        <div 
          key={field._id} 
          className="graphiql-defradb-field" 
          style={{ display: field.internal ? 'none' : undefined }}
        >
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
          {!field.id &&
            <button
              type="button"
              className="graphiql-button"
              onClick={() => remove(index)}
              disabled={isLoading || !!field.id}
            >
              Remove
            </button>
          }
        </div>
      )}
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