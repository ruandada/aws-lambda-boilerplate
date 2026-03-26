import serverlessExpress from '@codegenie/serverless-express'
import type { Context, Handler } from 'aws-lambda'
import { EventEmitter } from 'node:events'
import { createApp } from '../app'
import type { ConfigureResult } from '@codegenie/serverless-express/src/configure'
import type Framework from '@codegenie/serverless-express/src/frameworks'

type ServerlessInstance = Handler<unknown, unknown> & ConfigureResult<any, any>

let serverlessExpressInstancePromise: Promise<ServerlessInstance> | null = null

function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isHttpEvent(event: unknown): boolean {
  if (!isObject(event)) {
    return false
  }

  if (typeof event.httpMethod === 'string') {
    return true
  }

  if (event.version === '2.0' && isObject(event.requestContext)) {
    const http = event.requestContext.http
    return isObject(http) && typeof http.method === 'string'
  }

  return false
}

export const handler = async (event: unknown, context: Context): Promise<unknown> => {
  if (!isHttpEvent(event)) {
    throw new Error('Unsupported event type. This handler accepts HTTP-class events only.')
  }

  if (!serverlessExpressInstancePromise) {
    // The `ServerlessRequest.socket` constructed by `serverless-express` may not be an
    // EventEmitter in Node 24, which causes `on-finished`/`ee-first` to throw
    // `ee.on is not a function` when attempting to listen on `socket`.
    const patchedFramework: Framework = {
      sendRequest: ({ app, request, response }: any) => {
        const socket = request?.socket
        if (socket && typeof socket.on !== 'function') {
          const emitter = new EventEmitter()
          socket.on = emitter.on.bind(emitter)
          socket.once = emitter.once.bind(emitter)
          socket.removeListener = emitter.removeListener.bind(emitter)
          socket.emit = emitter.emit.bind(emitter)
        }

        // Express 5 adapter ultimately calls `app.handle(req, res)`
        app.handle(request, response)
      },
    }

    serverlessExpressInstancePromise = createApp().then((app) =>
      serverlessExpress({ app, framework: patchedFramework }),
    )
  }

  const serverlessExpressInstance = await serverlessExpressInstancePromise
  return serverlessExpressInstance(event, context, () => {})
}
