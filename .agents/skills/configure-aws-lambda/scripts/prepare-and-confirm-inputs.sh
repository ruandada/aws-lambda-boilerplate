#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

main() {
  local lambda_dir="${1:-}"
  [[ -n "${lambda_dir}" ]] || die "usage: $0 <lambda_directory>"

  need_cmd git
  need_cmd python3
  ensure_aws_cli

  local config_file
  config_file="$(config_file_path_for_dir "${lambda_dir}")"
  local lambda_parent_dir
  lambda_parent_dir="$(dirname "${config_file}")"
  [[ -d "${lambda_parent_dir}" ]] || die "lambda directory not found: ${lambda_parent_dir}"
  local template_file
  template_file="${SCRIPT_DIR}/../resources/aws-lambda.template.yaml"
  [[ -f "${template_file}" ]] || die "template file not found: ${template_file}"

  section "Step 1 - AWS prechecks"
  local caller_account
  caller_account="$(ensure_aws_login)"
  local aws_region
  aws_region="$(get_aws_region)"
  ok "AWS CLI and login are ready"
  info "AccountId (AWS): ${caller_account}"
  info "Region (AWS): ${aws_region}"

  section "Step 2 - Resolve GitHub metadata"
  local origin_url
  origin_url="$(git remote get-url origin 2>/dev/null || true)"
  [[ -n "${origin_url}" ]] || die "git remote origin is missing"

  local parsed
  parsed="$(parse_github_owner_repo_from_origin "${origin_url}")"
  local github_owner github_repo
  github_owner="$(awk '{print $1}' <<< "${parsed}")"
  github_repo="$(awk '{print $2}' <<< "${parsed}")"
  local github_branch
  github_branch="$(current_git_branch)"

  info "GITHUB_OWNER: ${github_owner}"
  info "GITHUB_REPO: ${github_repo}"
  info "GITHUB_BRANCH: ${github_branch}"

  section "Step 2 - Ensure and validate aws-lambda.yaml"
  if [[ ! -f "${config_file}" ]]; then
    info "aws-lambda.yaml does not exist. Creating from template."
    local default_function_name
    default_function_name="$(basename "${lambda_dir}")"
    render_template_file \
      "${template_file}" \
      "${config_file}" \
      "FUNCTION_NAME=${default_function_name}" \
      "IMAGE_REPOSITORY=${default_function_name}" \
      "REGION=${aws_region}" \
      "ACCOUNT_ID=${caller_account}" \
      "DEPLOYMENT_ROLE_NAME=github-actions-oidc-deployment-role" \
      "EXECUTION_ROLE_NAME=lambda-execution-role" \
      "ARCHITECTURE=x86_64"
    ok "Created ${config_file} from template."
  fi

  local validation_rc=0
  python3 - <<'PY' "${config_file}" "${template_file}" || validation_rc=$?
import sys
from pathlib import Path

path = Path(sys.argv[1])
template = Path(sys.argv[2])
text = path.read_text().splitlines()
template_text = template.read_text().splitlines()
pairs = []
template_keys = []

for raw in template_text:
    line = raw.strip()
    if not line or line.startswith("#") or ":" not in line:
        continue
    key = line.split(":", 1)[0].strip()
    if key:
        template_keys.append(key)

actual_keys = []
for raw in text:
    line = raw.strip()
    if not line or line.startswith("#"):
        continue
    if ":" not in line:
        continue
    key, val = line.split(":", 1)
    key = key.strip()
    val = val.strip()
    if not key:
        continue
    pairs.append((key, val))
    actual_keys.append(key)

missing = [k for k in template_keys if k not in actual_keys]
extra = [k for k in actual_keys if k not in template_keys]

print(f"Config file: {path}")
for k, v in pairs:
    print(f"  - {k}: {v}")

if missing:
    print()
    print("Template validation: MISSING required keys")
    for k in missing:
        print(f"  - {k}")
if extra:
    print()
    print("Template validation: EXTRA keys (allowed)")
    for k in extra:
        print(f"  - {k}")

if missing:
    raise SystemExit(2)
PY

  [[ ${validation_rc} -eq 0 ]] || die "aws-lambda.yaml is not compliant with template. Add missing keys and rerun."

  echo
  echo "Resolved values to sync:"
  echo "  - Region: ${aws_region}"
  echo "  - AccountId: ${caller_account}"
  echo "  - GITHUB_OWNER: ${github_owner}"
  echo "  - GITHUB_REPO: ${github_repo}"
  echo "  - GITHUB_BRANCH: ${github_branch}"
  echo
  echo "Will update in aws-lambda.yaml:"
  echo "  - Region -> ${aws_region}"
  echo "  - AccountId -> ${caller_account}"
  echo

  section "Next command exports"
  echo "export GITHUB_OWNER='${github_owner}'"
  echo "export GITHUB_REPO='${github_repo}'"
  echo "export GITHUB_BRANCH='${github_branch}'"
}

main "$@"
