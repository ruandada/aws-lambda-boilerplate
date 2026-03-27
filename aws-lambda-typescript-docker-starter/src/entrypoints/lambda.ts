import type { Context } from 'aws-lambda'
import { createEventRegistry } from '../registry'
import { EventRegistry } from '../events/event-registry'

let registryInstancePromise: Promise<EventRegistry> | null = null

export const handler = async (event: unknown, context: Context): Promise<unknown> => {
  if (!registryInstancePromise) {
    registryInstancePromise = createEventRegistry()
  }

  const registry = await registryInstancePromise
  const dispatchOutput = await registry.dispatch(event, context)

  if (!dispatchOutput.matched) {
    throw new Error('Unsupported event type.')
  }

  return dispatchOutput.result
}
