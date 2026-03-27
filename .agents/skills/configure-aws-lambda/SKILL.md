---
name: configure-aws-lambda
description: Step-by-step setup for GitHub Actions OIDC on AWS Lambda projects. Uses built-in templates for aws-lambda.yaml and IAM policies, with config creation/validation before IAM initialization.
---

# AWS Lambda Setup Wizard

This skill configures AWS IAM resources for GitHub Actions OIDC deployment in **six explicit steps**.

Treat the first skill argument as `lambda_directory` (the directory that contains `aws-lambda.yaml`).

Example invocation:

```bash
/configure-aws-lambda aws-lambda-typescript-docker-starter
```

---

## Skill resources

This skill stores reusable templates under:

- `.agents/skills/configure-aws-lambda/resources/aws-lambda.template.yaml`
- `.agents/skills/configure-aws-lambda/resources/github-actions.template.yaml`
- `.agents/skills/configure-aws-lambda/resources/policies/deployment-trust-policy.template.json`
- `.agents/skills/configure-aws-lambda/resources/policies/deployment-inline-policy.template.json`
- `.agents/skills/configure-aws-lambda/resources/policies/execution-trust-policy.template.json`

Scripts must prefer these templates instead of hardcoded inline policy text.

---

## Workflow (must follow in order)

### Step 1 - Verify AWS CLI installation and login status

Run:

```bash
aws --version
aws sts get-caller-identity --query Account --output text
```

Rules:

- If `aws` is missing, stop and instruct installation (`brew install awscli` on macOS).
- If identity lookup fails, stop and ask user to login:
  - `aws login`
  - fallback: `aws configure`
- Do not continue to IAM changes until both checks pass.

### Step 2 - Resolve and confirm all inputs, then sync yaml

Run:

```bash
bash .agents/skills/configure-aws-lambda/scripts/prepare-and-confirm-inputs.sh <lambda_directory>
```

This step must:

- Resolve values from AWS CLI + local git repository:
  - `Region` from `aws configure get region`
  - `AccountId` from `aws sts get-caller-identity --query Account --output text`
  - `GITHUB_OWNER` and `GITHUB_REPO` from `git remote get-url origin`
  - `GITHUB_BRANCH` from `git rev-parse --abbrev-ref HEAD`
- If `aws-lambda.yaml` does not exist, create it from `resources/aws-lambda.template.yaml`.
- If `aws-lambda.yaml` exists, validate it against the template and report missing/extra top-level keys.
- Display **all top-level parameters currently in `aws-lambda.yaml`** for confirmation.
- Display resolved AWS/GitHub values for confirmation.
- Ask for explicit confirmation before writing.
- After confirmation, update `aws-lambda.yaml`.
- Print export-ready GitHub variables for Step 4:
  - `GITHUB_OWNER`
  - `GITHUB_REPO`
  - `GITHUB_BRANCH`

Strict requirements:

- Never hardcode account, region, repo, branch, role names, or function names.
- If any required value cannot be resolved, stop and ask user to fix inputs.

### Step 3 - Initialize OIDC provider

Run:

```bash
bash .agents/skills/configure-aws-lambda/scripts/init-oidc-provider.sh <lambda_directory>
```

Expected behavior:

- Ensure `token.actions.githubusercontent.com` OIDC provider exists (create if missing, reuse if already present).
- Keep idempotent behavior.

### Step 4 - Initialize Deployment role permissions

Export git parameters first, then run:

```bash
GITHUB_OWNER=<owner> GITHUB_REPO=<repo> GITHUB_BRANCH=<branch> \
bash .agents/skills/configure-aws-lambda/scripts/init-deployment-role-permissions.sh <lambda_directory>
```

Expected behavior:

- Create or update deployment role trust policy for GitHub OIDC and Lambda service principal.
- Apply inline deployment policy (ECR + Lambda deploy + `iam:PassRole`) from policy templates.
- IAM updates must be **incremental merge-only**:
  - Never overwrite existing role policy documents wholesale.
  - Preserve existing statements that are unrelated to this Lambda function/project.
  - Only add missing actions/resources/conditions required by current function.
  - This is mandatory when multiple Lambda functions share one IAM role, to prevent cross-function permission loss.

### Step 5 - Initialize Execution role permissions

Run:

```bash
bash .agents/skills/configure-aws-lambda/scripts/init-execution-role-permissions.sh <lambda_directory>
```

Expected behavior:

- Ensure execution role trust allows `lambda.amazonaws.com`.
- Ensure `AWSLambdaBasicExecutionRole` is attached.
- Handle shared-role case:
  - If `ExecutionRoleName == DeploymentRoleName`, do not overwrite trust policy in a way that removes OIDC trust.
  - Preserve OIDC trust and verify Lambda principal is still allowed.
  - For policy updates on shared roles, perform merge-only additions and never remove unrelated existing permissions.

### Step 6 - Configure GitHub Actions workflow

Create workflow file from template:

```bash
FUNCTION_NAME="$(basename "<lambda_directory>")"
mkdir -p .github/workflows
cp .agents/skills/configure-aws-lambda/resources/github-actions.template.yaml .github/workflows/deploy-lambda-${FUNCTION_NAME}.yaml
```

Then update `name`, `on.push.paths`, and `build-function` in `.github/workflows/deploy-lambda-${FUNCTION_NAME}.yaml`:

- Replace workflow `name` with a function-specific name (for example: `Deploy Lambda (${FUNCTION_NAME})`) to avoid ambiguity when multiple Lambda workflows exist.
- Replace `on.push.paths` template default value with `<lambda_directory>/**` so push trigger only runs when this Lambda directory changes.
- Replace `build-function` template default value with `<lambda_directory>` (the first skill argument).
- Example: if invoked as `/configure-aws-lambda aws-lambda-typescript-docker-starter`, then:
  - workflow file: `.github/workflows/deploy-lambda-aws-lambda-typescript-docker-starter.yaml`
  - `name: Deploy Lambda (aws-lambda-typescript-docker-starter)`
  - `paths: [aws-lambda-typescript-docker-starter/**]` (or equivalent YAML list form)
  - `build-function: aws-lambda-typescript-docker-starter`

Expected behavior:

- Workflow exists in current repository under `.github/workflows`.
- Workflow filename must include the Lambda function/project name to avoid conflicts when multiple Lambdas exist.
- Workflow `name` is function-specific, not a shared generic label.
- `on.push.paths` always points to the selected Lambda project directory (`<lambda_directory>/**`), never hardcoded to unrelated values.
- `build-function` always points to the selected Lambda project directory, never hardcoded to unrelated values.
- If target workflow file already exists, show diff and ask for explicit confirmation before overwrite.

---

If you want a full guided flow, run:

1. Step 1 prechecks (`aws --version`, `aws sts get-caller-identity`)
2. `prepare-and-confirm-inputs.sh`
3. `init-oidc-provider.sh`
4. `init-deployment-role-permissions.sh`
5. `init-execution-role-permissions.sh`
6. Configure `.github/workflows/deploy-lambda-<function-name>.yaml` from `github-actions.template.yaml` and replace `name`, `on.push.paths`, and `build-function`

---

## Guardrails

- Always confirm all `aws-lambda.yaml` parameters and resolved GitHub values with the user before writing yaml and before IAM mutations.
- Always enforce account consistency between current AWS identity and `aws-lambda.yaml` after confirmation.
- Keep all IAM operations idempotent.
- For every IAM role update (trust/inline/managed-policy attachment logic), use additive/incremental changes only; never replace whole role policy content when existing statements are present.
- If an IAM API call fails due to permission denial, report the exact failed action and stop.
