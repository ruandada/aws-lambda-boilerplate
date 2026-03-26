import assert from 'node:assert/strict'
import test from 'node:test'
import { handler } from './lambda'
import { createApiGatewayV2Event, createContext } from '../test-utils'

test('handler returns hello world for API Gateway v2 event', async () => {
  const event = createApiGatewayV2Event('/')
  const context = createContext()
  const result = await handler(event, context)

  assert.ok(result && typeof result === 'object')
  const response = result as { statusCode?: unknown; body?: unknown }
  assert.equal(response.statusCode, 200)
  assert.equal(response.body, JSON.stringify({ message: 'Hello World' }))
})

test('handler rejects unsupported event', async () => {
  const event = { foo: 'bar' }
  const context = createContext()

  await assert.rejects(async () => handler(event, context), /Unsupported event type/)
})
