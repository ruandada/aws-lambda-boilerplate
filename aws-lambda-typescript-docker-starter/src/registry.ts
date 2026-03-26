import { createServerlessExpressApp } from './app'
import { EventRegistry } from './entrypoints/event-registry'

// Placeholder for future async bootstrap logic (db, config, secrets, etc.).
export async function setup(): Promise<void> {}

export async function createEventRegistry(): Promise<EventRegistry> {
  // Keep setup before app creation so initialization order is explicit.
  await setup()

  const registry = new EventRegistry()

  // By default, register a handler for HTTP events.
  const app = createServerlessExpressApp()

  registry.registerHttpEvent(async (event, context) => {
    return app(event, context, () => {})
  })

  // You can register other non-HTTP events here.
  registry.registerSqsEvent(async (event) => {
    console.log('SQS event received.')
  })

  return registry
}
