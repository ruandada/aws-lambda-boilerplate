# aws-lambda-python-docker-starter

Python + Flask AWS Lambda starter for local development, local event simulation, and Docker image deployment.

## Prerequisites

- Python `3.12+`
- Docker

## Quick start (.venv)

```bash
python3 -m venv .venv
source .venv/bin/activate
python -m pip install --upgrade pip
python -m pip install -e ".[dev]"
```

## Local development

Start local Flask server:

```bash
python -m app.cli dev
```

Default port is `3000`. To override:

```bash
PORT=8080 python -m app.cli dev
# or
python -m app.cli dev --port 8080
```

## Local event testing (non-HTTP)

Invoke Lambda handler with local event JSON:

```bash
python -m app.cli test-event test-cases/sqs-event.json
python -m app.cli test-event test-cases/http-v2-event.json
python -m app.cli test-event test-cases/http-v1-event.json
```

`test-event` fails fast for:

- missing file path / unreadable file
- empty event file
- unsupported event shape

## Tests

```bash
pytest
```

## Docker image

Build:

```bash
docker build -t aws-lambda-python-starter .
```

The image uses AWS Lambda Python runtime base and runs `app.entrypoints.lambda_handler.handler` by default.

## Deployment

Deployment guidance (AWS / Docker image / GitHub Actions OIDC) is documented in repository root:

- [/README.md](../README.md)
