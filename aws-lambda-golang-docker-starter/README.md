# aws-lambda-golang-docker-starter

Go + Cobra AWS Lambda starter for local development, local event simulation, and Docker image deployment.

## Prerequisites

- Go `1.24+`
- Docker

## Quick start

```bash
go mod tidy
go run ./cmd dev
```

Default port is `3000`. To override:

```bash
PORT=8080 go run ./cmd dev
# or
go run ./cmd dev --port 8080
```

## CLI commands (Cobra)

```bash
# start local HTTP server
go run ./cmd dev

# invoke Lambda handler with event JSON
go run ./cmd test-event ./test-cases/sqs-event.json
go run ./cmd test-event ./test-cases/http-v2-event.json
go run ./cmd test-event ./test-cases/http-v1-event.json

# start Lambda runtime (same as default root command)
go run ./cmd lambda
go run ./cmd
```

`test-event` command fails fast for:

- missing file path / unreadable file
- empty event file
- unsupported event shape

## Docker image

Build:

```bash
docker build -t aws-lambda-golang-starter .
```

The image uses AWS Lambda custom runtime base (`provided:al2023`) and runs `bootstrap` by default.

## Deployment

Deployment guidance (AWS / Docker image / GitHub Actions OIDC) is documented in repository root:

- [/README.md](../README.md)
