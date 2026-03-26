# AWS Lambda Boilerplate

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-orange)](https://aws.amazon.com/lambda/)
[![GitHub Actions](https://img.shields.io/badge/CI-GitHub_Actions-blue)](https://github.com/features/actions)

Production-oriented boilerplate for building and deploying AWS Lambda functions with TypeScript, Docker image packaging, and GitHub Actions OIDC.

This repository includes:

- A ready-to-copy Lambda starter project (`aws-lambda-typescript-docker-starter`)
- A local agent skill (`/configure-aws-lambda`) to bootstrap AWS IAM + GitHub Actions deployment
- Reusable templates for `aws-lambda.yaml`, workflow, and IAM policies

---

## Table of Contents

- [Why this boilerplate](#why-this-boilerplate)
- [Project Structure](#project-structure)
- [Get Started](#get-started)
- [Deployment Flow](#deployment-flow)
- [Configuration Reference](#configuration-reference)
- [Contributing](#contributing)
- [License](#license)

## Why this boilerplate

- Fast local development with TypeScript + Express
- Lambda container image deployment workflow
- GitHub Actions OIDC-based deployment (no long-lived AWS keys in GitHub secrets)
- Consistent project convention through `aws-lambda.yaml`
- Easy to scale to multiple Lambda services in one repository

## Project Structure

```text
.
├── aws-lambda-typescript-docker-starter/   # Starter Lambda project
├── definition/aws-lambda.schema.json        # Schema for aws-lambda.yaml
└── .agents/skills/configure-aws-lambda/     # Agent skill, scripts, templates, IAM policies
```

## Get Started

### Prerequisites

- AWS CLI installed and logged in
- Docker (for Lambda image build/deploy workflow)
- Node.js (latest LTS recommended) and `pnpm`
- A GitHub repository connected as your `origin` remote

### 1) Copy a boilerplate directory

From the repository root, copy the starter to your own Lambda directory:

```bash
cp -R aws-lambda-typescript-docker-starter my-lambda-service
```

You can choose any target folder name. It will become your Lambda function/project directory.

### 2) Build your Lambda code in the new directory

```bash
cd my-lambda-service
pnpm install
pnpm run dev
```

At this stage, implement your business logic and verify local behavior.

### 3) Run the local agent skill to configure AWS + GitHub Actions

From repository root, run:

```bash
/configure-aws-lambda my-lambda-service
```

The skill guides you through:

- Validating AWS CLI/login status
- Preparing and validating `aws-lambda.yaml`
- Initializing OIDC provider for GitHub Actions
- Creating/updating deployment and execution IAM roles
- Generating workflow file under `.github/workflows/`

### 4) Commit and push code to deploy

After configuration and development are done:

```bash
git add .
git commit -m "feat: add my lambda service"
git push origin main
```

The generated GitHub Actions workflow will deploy your AWS Lambda function automatically.

## Deployment Flow

1. Push to `main` (or trigger workflow manually with `workflow_dispatch`)
2. GitHub Actions requests AWS credentials via OIDC
3. Workflow builds and deploys Lambda container image
4. Deployment outputs include image URI and Lambda endpoint information

## Configuration Reference

Each Lambda directory uses an `aws-lambda.yaml` file, including:

- `FunctionName`
- `ImageRepository`
- `Region`
- `AccountId`
- `DeploymentRoleName`
- `ExecutionRoleName`
- `Architecture`

Schema reference:

- `definition/aws-lambda.schema.json`

## Contributing

Issues and pull requests are welcome.

Recommended contribution flow:

1. Fork the repository
2. Create a feature branch
3. Keep changes focused and documented
4. Open a pull request with context and test notes

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE).
