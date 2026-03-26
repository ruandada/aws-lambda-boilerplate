import express, { type Express } from 'express'

// Placeholder for future async bootstrap logic (db, config, secrets, etc.).
async function setup(): Promise<void> {}

export async function createApp(): Promise<Express> {
  // Keep setup before app creation so initialization order is explicit.
  await setup()

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
