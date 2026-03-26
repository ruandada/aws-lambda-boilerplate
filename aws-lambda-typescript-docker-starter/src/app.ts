import serverlessExpress from '@codegenie/serverless-express'
import type { ConfigureResult } from '@codegenie/serverless-express/src/configure'
import { Handler } from 'aws-lambda'
import express, { type Express } from 'express'
import EventEmitter from 'node:events'

export function createExpressApp(): Express {
  const app = express()

  app.get('/', (_req, res) => {
    res.status(200).json({
      message: 'Hello World',
    })
  })

  app.get('/api/greet/:name', (req, res) => {
    const { name } = req.params
    const from = typeof req.query.from === 'string' ? req.query.from : 'starter'

    res.status(200).json({
      message: `Hello, ${name}!`,
      from,
    })
  })

  return app
}

type ServerlessInstance = Handler<unknown, unknown> & ConfigureResult<any, any>

/**
 * Creates a serverless express instance.
 * @returns A serverless express instance.
 */
export function createServerlessExpressApp(): ServerlessInstance {
  return serverlessExpress({
    app: createExpressApp(),
    framework: {
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
    },
  })
}
