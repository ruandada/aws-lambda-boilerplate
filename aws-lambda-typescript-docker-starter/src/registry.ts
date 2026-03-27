import type { ALBEvent, APIGatewayProxyEvent, APIGatewayProxyEventV2 } from 'aws-lambda'
import { createServerlessExpressApp } from './app'
import { EventRegistry } from './events/event-registry'

// Placeholder for future async bootstrap logic (db, config, secrets, etc.).
export async function setup(): Promise<void> {}

/**
 * Backfill fields that @codegenie/serverless-express accesses unconditionally
 * (e.g. `Object.entries(event.headers)`) so hand-crafted test events don't crash.
 */
function normalizeHttpEvent(event: APIGatewayProxyEvent | APIGatewayProxyEventV2 | ALBEvent): void {
  const e = event as Record<string, any>
  if (!e.headers || typeof e.headers !== 'object') {
    e.headers = {}
  }
  if (!e.requestContext || typeof e.requestContext !== 'object') {
    e.requestContext = {}
  }
}

export async function createEventRegistry(): Promise<EventRegistry> {
  // Keep setup before app creation so initialization order is explicit.
  await setup()

  const registry = new EventRegistry()

  // By default, register a handler for HTTP events.
  const app = createServerlessExpressApp()

  registry.registerHttpEvent(async (event, context) => {
    normalizeHttpEvent(event)
    return app(event, context, () => {})
  })

  // You can register other non-HTTP events here.
  registry.registerSqsEvent(async () => {
    console.log('SQS event received.')
  })

  return registry
}
