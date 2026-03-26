import { readFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { handler } from './lambda'
import { createContext } from '../test-utils'

function printUsage(): void {
  console.error('Usage: pnpm test-event <event-json-file>')
}

async function loadEventFromFile(filePath: string): Promise<unknown> {
  const absolutePath = resolve(process.cwd(), filePath)
  const content = await readFile(absolutePath, 'utf8')
  return JSON.parse(content) as unknown
}

function printResult(result: unknown): void {
  if (typeof result === 'undefined') {
    console.log('Handler completed with no return value.')
    return
  }

  try {
    console.log(JSON.stringify(result, null, 2))
  } catch {
    console.log(String(result))
  }
}

export async function runFromCli(argv: string[]): Promise<number> {
  const eventFilePath = argv[0]
  if (!eventFilePath) {
    printUsage()
    return 1
  }

  try {
    const event = await loadEventFromFile(eventFilePath)
    const context = createContext()
    const result = await handler(event, context)
    printResult(result)
    return 0
  } catch (error) {
    console.error('Failed to execute test event.')
    console.error(error)
    return 1
  }
}

runFromCli(process.argv.slice(2)).then((exitCode) => {
  if (exitCode !== 0) {
    process.exitCode = exitCode
  }
})
