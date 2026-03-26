#!/usr/bin/env bash

set -euo pipefail

OIDC_URL="https://token.actions.githubusercontent.com"
OIDC_THUMBPRINT="6938fd4d98bab03faadb97b34396831e3780aea1"

die() { echo "[ERROR] $*" >&2; exit 1; }
info() { echo "[INFO]  $*"; }
ok() { echo "[OK]    $*"; }
section() { echo; echo "========== $* =========="; }

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

yaml_get_required() {
  local file="$1"
  local key="$2"
  local value
  value="$(awk -F':' -v k="${key}" '
    $1 ~ "^[[:space:]]*"k"[[:space:]]*$" {
      sub(/^[[:space:]]+/, "", $2)
      sub(/[[:space:]]+$/, "", $2)
      gsub(/^'\''|'\''$/, "", $2)
      gsub(/^"|"$/, "", $2)
      print $2
      exit
    }
  ' "${file}")"
  [[ -n "${value}" ]] || die "missing required key ${key} in ${file}"
  echo "${value}"
}

ensure_aws_cli() {
  command -v aws >/dev/null 2>&1 || die "AWS CLI is not installed. Install first: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html"
}

ensure_aws_login() {
  set +e
  local caller_account
  caller_account="$(aws sts get-caller-identity --query Account --output text 2>/dev/null)"
  local rc=$?
  set -e

  if [[ $rc -eq 0 && -n "${caller_account}" && "${caller_account}" != "None" ]]; then
    echo "${caller_account}"
    return
  fi

  die "AWS login required. Run: aws login (or aws configure)."
}

repo_root() {
  git rev-parse --show-toplevel 2>/dev/null || die "must run inside a git repository"
}

config_file_path_for_dir() {
  local lambda_dir="$1"
  local root
  root="$(repo_root)"
  echo "${root}/${lambda_dir}/aws-lambda.yaml"
}

get_aws_region() {
  local region
  region="$(aws configure get region 2>/dev/null || true)"
  [[ -n "${region}" ]] || die "AWS region is empty. Configure it first: aws login (or aws configure)."
  echo "${region}"
}

parse_github_owner_repo_from_origin() {
  local origin_url="$1"
  local owner repo

  if [[ "${origin_url}" =~ ^git@github\.com:([^/]+)/([^/]+)$ ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]%.git}"
  elif [[ "${origin_url}" =~ ^https://github\.com/([^/]+)/([^/]+)$ ]]; then
    owner="${BASH_REMATCH[1]}"
    repo="${BASH_REMATCH[2]%.git}"
  else
    die "unsupported origin URL format: ${origin_url}"
  fi

  [[ -n "${owner}" && -n "${repo}" ]] || die "failed to parse GitHub owner/repo from origin URL"
  echo "${owner} ${repo}"
}

current_git_branch() {
  local branch
  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
  [[ -n "${branch}" ]] || die "cannot read current git branch"
  echo "${branch}"
}

config_file_for_dir() {
  local lambda_dir="$1"
  local cfg
  cfg="$(config_file_path_for_dir "${lambda_dir}")"
  [[ -f "${cfg}" ]] || die "config file not found: ${cfg}"
  echo "${cfg}"
}

load_lambda_config() {
  local config_file="$1"
  FUNCTION_NAME="$(yaml_get_required "${config_file}" "FunctionName")"
  IMAGE_REPOSITORY="$(yaml_get_required "${config_file}" "ImageRepository")"
  REGION="$(yaml_get_required "${config_file}" "Region")"
  ACCOUNT_ID="$(yaml_get_required "${config_file}" "AccountId")"
  DEPLOYMENT_ROLE_NAME="$(yaml_get_required "${config_file}" "DeploymentRoleName")"
  EXECUTION_ROLE_NAME="$(yaml_get_required "${config_file}" "ExecutionRoleName")"
  EXECUTION_ROLE_ARN="arn:aws:iam::${ACCOUNT_ID}:role/${EXECUTION_ROLE_NAME}"
  INLINE_POLICY_NAME="${DEPLOYMENT_ROLE_NAME}-inline-policy"
}

ensure_account_matches_config() {
  local caller_account="$1"
  [[ "${caller_account}" == "${ACCOUNT_ID}" ]] || die "aws-lambda account (${ACCOUNT_ID}) does not match current AWS account (${caller_account})"
}

ensure_oidc_provider() {
  local provider_arn
  provider_arn="$(aws iam list-open-id-connect-providers --query "OpenIDConnectProviderList[?contains(Arn, 'token.actions.githubusercontent.com')].Arn | [0]" --output text)"

  if [[ "${provider_arn}" == "None" || -z "${provider_arn}" ]]; then
    provider_arn="$(aws iam create-open-id-connect-provider \
      --url "${OIDC_URL}" \
      --client-id-list sts.amazonaws.com \
      --thumbprint-list "${OIDC_THUMBPRINT}" \
      --query OpenIDConnectProviderArn \
      --output text)"
    ok "Created OIDC provider: ${provider_arn}" >&2
  else
    ok "OIDC provider already exists: ${provider_arn}" >&2
  fi

  echo "${provider_arn}"
}

ensure_execution_basic_policy() {
  local role_name="$1"
  local basic_policy_arn="arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  local attached

  attached="$(aws iam list-attached-role-policies \
    --role-name "${role_name}" \
    --query "AttachedPolicies[?PolicyArn=='${basic_policy_arn}'] | length(@)" \
    --output text)"

  if [[ "${attached}" == "0" ]]; then
    aws iam attach-role-policy \
      --role-name "${role_name}" \
      --policy-arn "${basic_policy_arn}" >/dev/null
    ok "Attached managed policy: AWSLambdaBasicExecutionRole"
  else
    ok "Managed policy already attached: AWSLambdaBasicExecutionRole"
  fi
}

render_template_file() {
  local template_file="$1"
  local output_file="$2"
  shift 2

  python3 - <<'PY' "${template_file}" "${output_file}" "$@"
import re
import sys
from pathlib import Path

template = Path(sys.argv[1]).read_text()
out_file = Path(sys.argv[2])
pairs = sys.argv[3:]

values = {}
for pair in pairs:
    if "=" not in pair:
        continue
    key, value = pair.split("=", 1)
    values[key] = value

def repl(match):
    key = match.group(1)
    if key not in values:
        raise SystemExit(f"missing template variable: {key}")
    return values[key]

rendered = re.sub(r"\{\{([A-Z0-9_]+)\}\}", repl, template)
out_file.write_text(rendered)
PY
}
