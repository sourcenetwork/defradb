export type Extensions = {
  status: number
  httpError: string
  stack?: string 
}

export type ErrorItem = {
  message: string
  extensions?: Extensions 
}

export type CollectionResponse = {
  name: string
  id: string
}

export type LoadSchemaResponse = {
  errors?: ErrorItem[]
  result?: string
  collections?: CollectionResponse[]
}

export async function loadSchema(body: string): Promise<LoadSchemaResponse> {
  return fetch('http://localhost:9181/api/v0/schema/load', { method: 'POST', body }).then(res => res.json())
}