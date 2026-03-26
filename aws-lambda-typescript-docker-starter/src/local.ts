import { createApp } from './app'

const DEFAULT_PORT = 3000
const parsedPort = Number(process.env.PORT)
const port = Number.isFinite(parsedPort) && parsedPort > 0 ? parsedPort : DEFAULT_PORT

async function startLocalServer(): Promise<void> {
  const app = await createApp()
  app.listen(port, () => {
    // Keep logging concise for local debugging.
    console.log(`Local server is running on http://localhost:${port}`)
  })
}

startLocalServer()
