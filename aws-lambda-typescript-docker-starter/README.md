# aws-lambda-typescript-docker-starter

TypeScript + Express AWS Lambda starter for local development and quick iteration.

## Prerequisites

- Node.js `v24.14.0`
- pnpm (or Corepack-enabled Node)

## Local development

Install dependencies:

```bash
pnpm install
```

Start dev server:

```bash
pnpm run dev
```

Default port is `3000`. To override:

```bash
PORT=8080 pnpm run dev
```

Useful checks:

```bash
pnpm run build
pnpm run typecheck
```

## Local event testing (non-HTTP)

You can invoke Lambda locally with an event JSON file:

```bash
pnpm test-event test-cases/sqs-event.json
```

If no handler matches the event shape, the command exits with an error.

## Deployment

Deployment guidance (AWS / Docker image / GitHub Actions OIDC) is documented in the repository root README:

- [/README.md](../README.md)
