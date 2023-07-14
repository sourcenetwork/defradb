import { compare } from 'fast-json-patch'

export type Extensions = {
  status: number
  httpError: string
  stack?: string 
}

export type ErrorItem = {
  message: string
  extensions?: Extensions 
}

export type Field = {
  id?: string
  name: string
  kind: string
  internal: boolean
}

export type Collection = {
  id: string
  name: string
}

export type CollectionWithFields = Collection & {
  fields: Field[]
}

export type Response<T> = {
  data: T
  errors?: ErrorItem[]
}

export type ListSchemaResponse = Response<{
  collections?: CollectionWithFields[]
}>

export type LoadSchemaResponse = Response<{
  result?: string
  collections?: Collection[]
}>

export type PatchSchemaResponse = Response<{
  result?: string
}>

const baseUrl = import.meta.env.DEV ? 'http://localhost:9181/api/v0' : '/api/v0'

export async function listSchema(): Promise<ListSchemaResponse> {
  return fetch(baseUrl + '/schema').then(res => res.json())
}

export async function loadSchema(schema: string): Promise<LoadSchemaResponse> {
  return fetch(baseUrl + '/schema', { method: 'POST', body: schema }).then(res => res.json())
}

export async function patchSchema(nameA: string, fieldsA: Field[], nameB: string, fieldsB: Field[]): Promise<PatchSchemaResponse> {
  const schemaA = { Name: nameA, Fields: fieldsA.map(field => ({ Name: field.name, Kind: field.kind })) }
  const schemaB = { Name: nameB, Fields: fieldsB.map(field => ({ Name: field.name, Kind: field.kind })) }

  const collectionA = { [nameA]: { Name: nameA, Schema: schemaA } }
  const collectionB = { [nameB]: { Name: nameB, Schema: schemaB } }
  
  const body = JSON.stringify(compare(collectionA, collectionB))
  return fetch(baseUrl + '/schema', { method: 'PATCH', body }).then(res => res.json())
}
