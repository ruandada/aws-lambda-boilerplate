#!/usr/bin/env bash

set -euo pipefail

OIDC_URL="https://token.actions.githubusercontent.com"
BASIC_EXECUTION_POLICY_ARN="arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"

DRY_RUN=0
LAMBDA_DIR="."

usage() {
  cat <<'USAGE'
Usage:
  ./configure-aws-lambda.sh [--dry-run] [lambda_directory]

Configure GitHub Actions OIDC IAM permissions and workflow files for an AWS Lambda
Docker-image project. Omit lambda_directory when the Lambda project is the repo root.
USAGE
}

die() { echo "error: $*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"; }

header() {
  echo
  echo "AWS Lambda GitHub Actions setup"
  echo "================================"
}

section() {
  echo
  echo "$1"
  printf '%*s\n' "${#1}" '' | tr ' ' '-'
}

kv() {
  printf '  %-18s %s\n' "$1:" "$2"
}

ok() {
  printf '  [ok] %s\n' "$*"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1 ;;
    -h|--help) usage; exit 0 ;;
    -*) die "unknown option: $1" ;;
    *) LAMBDA_DIR="$1" ;;
  esac
  shift
done

need aws
need git
need python3
need mktemp

REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "${REPO_ROOT}"

normalize_path() {
  local path="${1%/}"
  [[ -n "${path}" ]] || path="."
  [[ "${path}" == "." ]] && echo "." && return
  echo "${path#./}"
}

LAMBDA_DIR="$(normalize_path "${LAMBDA_DIR}")"
[[ -d "${LAMBDA_DIR}" ]] || die "lambda directory not found: ${LAMBDA_DIR}"

FUNCTION_NAME="$(basename "${LAMBDA_DIR}")"
if [[ "${LAMBDA_DIR}" == "." ]]; then
  FUNCTION_NAME="$(basename "${REPO_ROOT}")"
fi

CONFIG_FILE="${LAMBDA_DIR}/aws-lambda.yaml"
WORKFLOW_DIR=".github/workflows"
WORKFLOW_FILE="${WORKFLOW_DIR}/deploy-lambda-${FUNCTION_NAME}.yaml"

aws_account() {
  aws sts get-caller-identity --query Account --output text 2>/dev/null \
    || die "AWS login required. Run aws login or aws configure."
}

aws_region() {
  local region
  region="$(aws configure get region 2>/dev/null || true)"
  [[ -n "${region}" ]] || die "AWS region is not configured. Run aws configure."
  echo "${region}"
}

github_owner_repo() {
  local url owner repo
  url="$(git remote get-url origin 2>/dev/null || true)"
  [[ -n "${url}" ]] || die "git remote origin is missing"

  if [[ "${url}" =~ ^git@github\.com:([^/]+)/([^/]+)(\.git)?$ ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]%.git}"
  elif [[ "${url}" =~ ^https://github\.com/([^/]+)/([^/]+)(\.git)?$ ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]%.git}"
  else
    die "unsupported GitHub origin URL: ${url}"
  fi

  echo "${owner} ${repo}"
}

upsert_config() {
  local file="$1"
  local function_name="$2"
  local region="$3"
  local account_id="$4"

  python3 - "$file" "$function_name" "$region" "$account_id" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
function_name, region, account_id = sys.argv[2:5]

defaults = {
    "$schema": "https://raw.githubusercontent.com/ruandada/aws-lambda-boilerplate/refs/heads/main/definition/aws-lambda.schema.json",
    "FunctionName": function_name,
    "ImageRepository": function_name,
    "Region": region,
    "AccountId": account_id,
    "DeploymentRoleName": "github-actions-oidc-deployment-role",
    "ExecutionRoleName": "lambda-execution-role",
    "Architecture": "x86_64",
}

values = {}
order = []
if path.exists():
    for raw in path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or ":" not in line:
            continue
        key, value = line.split(":", 1)
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        values[key] = value
        order.append(key)

values["Region"] = region
values["AccountId"] = account_id
for key, value in defaults.items():
    values.setdefault(key, value)

keys = [key for key in defaults if key in values]
keys.extend(key for key in order if key not in keys and key in values)

path.parent.mkdir(parents=True, exist_ok=True)
path.write_text("\n".join(f'{key}: "{values[key]}"' for key in keys) + "\n")
PY
}

yaml_get() {
  local file="$1" key="$2"
  python3 - "$file" "$key" <<'PY'
from pathlib import Path
import sys

for raw in Path(sys.argv[1]).read_text().splitlines():
    line = raw.strip()
    if not line or line.startswith("#") or ":" not in line:
        continue
    key, value = line.split(":", 1)
    if key.strip() == sys.argv[2]:
        print(value.strip().strip('"').strip("'"))
        raise SystemExit
raise SystemExit(1)
PY
}

render_json() {
  local output="$1"
  shift
  python3 - "$output" "$@" <<'PY'
import json
import sys
from pathlib import Path

output = Path(sys.argv[1])
values = dict(item.split("=", 1) for item in sys.argv[2:])
doc = values["DOC"]

if doc != "deployment-policy":
    for key in (
        "REGION",
        "ACCOUNT_ID",
        "IMAGE_REPOSITORY",
        "FUNCTION_NAME",
        "EXECUTION_ROLE_ARN",
        "DEPLOYMENT_ROLE_NAME",
    ):
        values.setdefault(key, "")

if doc != "deployment-trust":
    for key in (
        "SID_SUFFIX",
        "OIDC_PROVIDER_ARN",
        "GITHUB_OWNER",
        "GITHUB_REPO",
        "GITHUB_BRANCH",
    ):
        values.setdefault(key, "")

deployment_trust = {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole",
        },
        {
            "Sid": "GitHubActionsOIDCTrust" + values["SID_SUFFIX"],
            "Effect": "Allow",
            "Principal": {"Federated": values["OIDC_PROVIDER_ARN"]},
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
                },
                "StringLike": {
                    "token.actions.githubusercontent.com:sub": [
                        f'repo:{values["GITHUB_OWNER"]}/{values["GITHUB_REPO"]}:ref:refs/heads/{values["GITHUB_BRANCH"]}'
                    ]
                },
            },
        },
    ],
}

execution_trust = {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole",
        }
    ],
}

deployment_policy = {
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "EcrAuth",
            "Effect": "Allow",
            "Action": ["ecr:GetAuthorizationToken"],
            "Resource": "*",
        },
        {
            "Sid": "EcrPushPullAndRepo",
            "Effect": "Allow",
            "Action": [
                "ecr:BatchCheckLayerAvailability",
                "ecr:BatchGetImage",
                "ecr:CompleteLayerUpload",
                "ecr:CreateRepository",
                "ecr:DescribeImages",
                "ecr:DescribeRepositories",
                "ecr:GetRepositoryPolicy",
                "ecr:InitiateLayerUpload",
                "ecr:ListImages",
                "ecr:PutImage",
                "ecr:SetRepositoryPolicy",
                "ecr:UploadLayerPart",
            ],
            "Resource": [
                f'arn:aws:ecr:{values["REGION"]}:{values["ACCOUNT_ID"]}:repository/{values["IMAGE_REPOSITORY"]}',
                "*",
            ],
        },
        {
            "Sid": "LambdaDeploy",
            "Effect": "Allow",
            "Action": [
                "lambda:AddPermission",
                "lambda:CreateFunction",
                "lambda:CreateFunctionUrlConfig",
                "lambda:GetFunction",
                "lambda:GetFunctionConfiguration",
                "lambda:GetFunctionUrlConfig",
                "lambda:UpdateFunctionCode",
                "lambda:UpdateFunctionConfiguration",
                "lambda:UpdateFunctionUrlConfig",
            ],
            "Resource": f'arn:aws:lambda:{values["REGION"]}:{values["ACCOUNT_ID"]}:function:{values["FUNCTION_NAME"]}*',
        },
        {
            "Sid": "PassRolesForLambda",
            "Effect": "Allow",
            "Action": ["iam:PassRole"],
            "Resource": [
                values["EXECUTION_ROLE_ARN"],
                f'arn:aws:iam::{values["ACCOUNT_ID"]}:role/{values["DEPLOYMENT_ROLE_NAME"]}',
            ],
        },
    ],
}

docs = {
    "deployment-trust": deployment_trust,
    "execution-trust": execution_trust,
    "deployment-policy": deployment_policy,
}
output.write_text(json.dumps(docs[values["DOC"]], indent=2) + "\n")
PY
}

merge_policy() {
  local current="$1" desired="$2" output="$3"
  python3 - "$current" "$desired" "$output" <<'PY'
import json
import sys
from pathlib import Path

current, desired, output = map(Path, sys.argv[1:4])
base = json.loads(current.read_text()) if current.exists() and current.stat().st_size else {"Version": "2012-10-17", "Statement": []}
want = json.loads(desired.read_text())

def as_list(value):
    if value is None:
        return []
    return value if isinstance(value, list) else [value]

def key(stmt):
    sid = stmt.get("Sid")
    if sid:
        return "sid:" + sid
    shape = {
        "Effect": stmt.get("Effect"),
        "Action": sorted(as_list(stmt.get("Action"))),
        "Principal": stmt.get("Principal"),
        "Condition": stmt.get("Condition"),
    }
    return json.dumps(shape, sort_keys=True, separators=(",", ":"))

statements = as_list(base.get("Statement"))
index = {key(stmt): i for i, stmt in enumerate(statements)}
for stmt in as_list(want.get("Statement")):
    stmt_key = key(stmt)
    if stmt_key in index:
        statements[index[stmt_key]] = stmt
    else:
        statements.append(stmt)

base["Version"] = base.get("Version", want.get("Version", "2012-10-17"))
base["Statement"] = statements
output.write_text(json.dumps(base, indent=2) + "\n")
PY
}

ensure_oidc_provider() {
  local provider_arn
  provider_arn="$(aws iam list-open-id-connect-providers --query "OpenIDConnectProviderList[?contains(Arn, 'token.actions.githubusercontent.com')].Arn | [0]" --output text)"
  if [[ -z "${provider_arn}" || "${provider_arn}" == "None" ]]; then
    provider_arn="$(aws iam create-open-id-connect-provider \
      --url "${OIDC_URL}" \
      --client-id-list sts.amazonaws.com \
      --query OpenIDConnectProviderArn \
      --output text)"
  fi
  echo "${provider_arn}"
}

ensure_role_trust() {
  local role_name="$1" desired="$2"
  local current merged
  current="$(mktemp)"
  merged="$(mktemp)"

  if aws iam get-role --role-name "${role_name}" >/dev/null 2>&1; then
    aws iam get-role --role-name "${role_name}" --query 'Role.AssumeRolePolicyDocument' --output json > "${current}"
    merge_policy "${current}" "${desired}" "${merged}"
    aws iam update-assume-role-policy --role-name "${role_name}" --policy-document "file://${merged}" >/dev/null
  else
    aws iam create-role --role-name "${role_name}" --assume-role-policy-document "file://${desired}" >/dev/null
  fi

  rm -f "${current}" "${merged}"
}

ensure_inline_policy() {
  local role_name="$1" policy_name="$2" desired="$3"
  local current merged
  current="$(mktemp)"
  merged="$(mktemp)"

  if aws iam get-role-policy --role-name "${role_name}" --policy-name "${policy_name}" >/dev/null 2>&1; then
    aws iam get-role-policy --role-name "${role_name}" --policy-name "${policy_name}" --query 'PolicyDocument' --output json > "${current}"
    merge_policy "${current}" "${desired}" "${merged}"
  else
    cp "${desired}" "${merged}"
  fi

  aws iam put-role-policy --role-name "${role_name}" --policy-name "${policy_name}" --policy-document "file://${merged}" >/dev/null
  rm -f "${current}" "${merged}"
}

ensure_basic_execution_policy() {
  local role_name="$1"
  local attached
  attached="$(aws iam list-attached-role-policies \
    --role-name "${role_name}" \
    --query "AttachedPolicies[?PolicyArn=='${BASIC_EXECUTION_POLICY_ARN}'] | length(@)" \
    --output text)"

  if [[ "${attached}" == "0" ]]; then
    aws iam attach-role-policy --role-name "${role_name}" --policy-arn "${BASIC_EXECUTION_POLICY_ARN}" >/dev/null
  fi
}

write_workflow() {
  local file="$1" lambda_dir="$2" function_name="$3"
  mkdir -p "$(dirname "$file")"
  python3 - "$file" "$lambda_dir" "$function_name" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
lambda_dir, function_name = sys.argv[2:4]
path_filter = "" if lambda_dir == "." else f"""    paths:
      - {lambda_dir}/**
"""

path.write_text(f"""name: Deploy Lambda ({function_name})

on:
  push:
    branches: [main]
{path_filter}  workflow_dispatch:

permissions:
  id-token: write
  contents: read

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v6

      - name: Deploy AWS Lambda
        id: deploy
        uses: ruandada/deploy-aws-lambda@v1
        with:
          build-function: {lambda_dir}

      - name: Print outputs
        run: |
          echo "latest=${{{{ steps.deploy.outputs.image-uri-latest }}}}"
          echo "sha=${{{{ steps.deploy.outputs.image-uri-sha }}}}"
          echo "tag=${{{{ steps.deploy.outputs.image-tag-sha }}}}"
          echo "url=${{{{ steps.deploy.outputs.lambda-url }}}}"
""")
PY
}

header
section "Target"
kv "Mode" "$([[ "${DRY_RUN}" == "1" ]] && echo "dry run" || echo "apply")"
kv "Project directory" "${LAMBDA_DIR}"
kv "Function name" "${FUNCTION_NAME}"

section "Resolve context"
ACCOUNT_ID="$(aws_account)"
REGION="$(aws_region)"
read -r GITHUB_OWNER GITHUB_REPO < <(github_owner_repo)
GITHUB_BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo main)"
kv "AWS account" "${ACCOUNT_ID}"
kv "AWS region" "${REGION}"
kv "GitHub repo" "${GITHUB_OWNER}/${GITHUB_REPO}"
kv "GitHub branch" "${GITHUB_BRANCH}"

section "Write local files"
upsert_config "${CONFIG_FILE}" "${FUNCTION_NAME}" "${REGION}" "${ACCOUNT_ID}"
ok "Updated ${CONFIG_FILE}"

FUNCTION_NAME="$(yaml_get "${CONFIG_FILE}" FunctionName)"
IMAGE_REPOSITORY="$(yaml_get "${CONFIG_FILE}" ImageRepository)"
DEPLOYMENT_ROLE_NAME="$(yaml_get "${CONFIG_FILE}" DeploymentRoleName)"
EXECUTION_ROLE_NAME="$(yaml_get "${CONFIG_FILE}" ExecutionRoleName)"
EXECUTION_ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${EXECUTION_ROLE_NAME}"
INLINE_POLICY_NAME="${FUNCTION_NAME}-deployment-inline-policy"

write_workflow "${WORKFLOW_FILE}" "${LAMBDA_DIR}" "${FUNCTION_NAME}"
ok "Updated ${WORKFLOW_FILE}"

if [[ "${DRY_RUN}" == "1" ]]; then
  section "Summary"
  kv "Status" "dry run complete"
  kv "AWS IAM changes" "skipped"
  kv "Config file" "${CONFIG_FILE}"
  kv "Workflow file" "${WORKFLOW_FILE}"
  exit 0
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

section "Configure AWS IAM"
PROVIDER_ARN="$(ensure_oidc_provider)"
SID_SUFFIX="$(printf '%s-%s' "${GITHUB_OWNER}" "${GITHUB_REPO}" | tr -cd '[:alnum:]')"
ok "OIDC provider ready"

deployment_trust="${tmpdir}/deployment-trust.json"
execution_trust="${tmpdir}/execution-trust.json"
deployment_policy="${tmpdir}/deployment-policy.json"

render_json "${deployment_trust}" \
  DOC=deployment-trust \
  OIDC_PROVIDER_ARN="${PROVIDER_ARN}" \
  GITHUB_OWNER="${GITHUB_OWNER}" \
  GITHUB_REPO="${GITHUB_REPO}" \
  GITHUB_BRANCH="${GITHUB_BRANCH}" \
  SID_SUFFIX="${SID_SUFFIX}"

render_json "${execution_trust}" DOC=execution-trust

render_json "${deployment_policy}" \
  DOC=deployment-policy \
  REGION="${REGION}" \
  ACCOUNT_ID="${ACCOUNT_ID}" \
  IMAGE_REPOSITORY="${IMAGE_REPOSITORY}" \
  FUNCTION_NAME="${FUNCTION_NAME}" \
  EXECUTION_ROLE_ARN="${EXECUTION_ROLE_ARN}" \
  DEPLOYMENT_ROLE_NAME="${DEPLOYMENT_ROLE_NAME}"

ensure_role_trust "${DEPLOYMENT_ROLE_NAME}" "${deployment_trust}"
ok "Deployment role trust ready"
ensure_role_trust "${EXECUTION_ROLE_NAME}" "${execution_trust}"
ok "Execution role trust ready"
ensure_inline_policy "${DEPLOYMENT_ROLE_NAME}" "${INLINE_POLICY_NAME}" "${deployment_policy}"
ok "Deployment inline policy ready"
ensure_basic_execution_policy "${EXECUTION_ROLE_NAME}"
ok "Execution basic logging policy ready"

section "Summary"
kv "Status" "configured"
kv "Function name" "${FUNCTION_NAME}"
kv "Config file" "${CONFIG_FILE}"
kv "Workflow file" "${WORKFLOW_FILE}"
kv "Deployment role" "${DEPLOYMENT_ROLE_NAME}"
kv "Execution role" "${EXECUTION_ROLE_NAME}"
