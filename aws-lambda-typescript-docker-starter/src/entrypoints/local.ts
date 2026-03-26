import { createExpressApp } from '../app'
import { setup } from '../registry'

const DEFAULT_PORT = 3000
const parsedPort = Number(process.env.PORT)
const port = Number.isFinite(parsedPort) && parsedPort > 0 ? parsedPort : DEFAULT_PORT

export async function startLocalServer(): Promise<void> {
  await setup()

  const app = createExpressApp()
  app.listen(port, () => {
    // Keep logging concise for local debugging.
    console.log(`Local server is running on http://localhost:${port}`)
  })
}

startLocalServer()
