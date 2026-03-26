import assert from 'node:assert/strict'
import test from 'node:test'
import { EventRegistry } from './event-registry'
import { createContext } from '../test-utils'
import { isSnsEvent } from './events'


test('dispatches to the correct handler based on event type', async () => {
  const registry = new EventRegistry()
  const context = createContext()

  registry.registerSqsEvent(async (event) => ({
    processed: 'sqs',
    count: event.Records.length,
  }))

  registry.register(isSnsEvent, async (event) => ({
    processed: 'sns',
    message: event.Records[0].Sns.Message,
  }))

  const sqsEvent = { Records: [{ eventSource: 'aws:sqs', body: 'hello' }] }
  const dispatched = await registry.dispatch(sqsEvent, context)
  assert.deepEqual(dispatched, { matched: true, result: { processed: 'sqs', count: 1 } })
})

test('dispatches to the first matching handler when multiple handlers match', async () => {
  const registry = new EventRegistry()
  const context = createContext()

  const isAnyObject = (event: unknown): event is object =>
    typeof event === 'object' && event !== null

  registry.register(isAnyObject, async () => 'first')
  registry.register(isAnyObject, async () => 'second')

  const dispatched = await registry.dispatch({}, context)
  assert.deepEqual(dispatched, { matched: true, result: 'first' })
})

test('returns matched: false when no handler matches', async () => {
  const registry = new EventRegistry()
  const context = createContext()

  registry.registerSqsEvent(async () => 'sqs')

  const dispatched = await registry.dispatch({ Records: [{ eventSource: 'aws:sns' }] }, context)
  assert.deepEqual(dispatched, { matched: false })
})

test('returns matched: false when registry has no handlers', async () => {
  const registry = new EventRegistry()
  const context = createContext()

  const dispatched = await registry.dispatch({}, context)
  assert.deepEqual(dispatched, { matched: false })
})

test('register returns this for chaining', () => {
  const registry = new EventRegistry()
  const neverMatch = (_: unknown): _ is never => false

  const result = registry
    .register(neverMatch, async () => {})
    .register(neverMatch, async () => {})

  assert.equal(result, registry)
})

test('registerScheduledEvent matches eventbridge schedule event', async () => {
  const registry = new EventRegistry()
  const context = createContext()

  registry.registerScheduledEvent(async (event) => ({
    source: event.source,
    detailType: event['detail-type'],
  }))

  const scheduledEvent = {
    id: 'evt-1',
    version: '0',
    account: '123456789012',
    time: '2026-03-26T00:00:00Z',
    region: 'us-east-1',
    resources: [],
    source: 'aws.events',
    'detail-type': 'Scheduled Event',
    detail: {},
  }

  const dispatched = await registry.dispatch(scheduledEvent, context)
  assert.deepEqual(dispatched, {
    matched: true,
    result: { source: 'aws.events', detailType: 'Scheduled Event' },
  })
})
