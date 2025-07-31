import React from 'react';
import { GraphiQLPlugin } from '@graphiql/react';

const SwaggerUI = React.lazy(() => import('swagger-ui-react'));

const baseUrl = import.meta.env.DEV ? 'http://localhost:9181' : '';

export const swaggerPlugin: GraphiQLPlugin = {
  title: 'DefraDB API',
  icon: () => (<div>API</div>),
  content: () => (
    <React.Suspense>
      <SwaggerUI url={`${baseUrl}/openapi.json`} />
    </React.Suspense>
  ),
}; 