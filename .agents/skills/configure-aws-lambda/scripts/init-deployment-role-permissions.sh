#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

main() {
  local lambda_dir="${1:-}"
  [[ -n "${lambda_dir}" ]] || die "usage: $0 <lambda_directory>"

  need_cmd git
  need_cmd mktemp
  need_cmd python3
  ensure_aws_cli

  local config_file
  config_file="$(config_file_for_dir "${lambda_dir}")"
  load_lambda_config "${config_file}"

  GITHUB_OWNER="${GITHUB_OWNER:-}"
  GITHUB_REPO="${GITHUB_REPO:-}"
  GITHUB_BRANCH="${GITHUB_BRANCH:-}"
  [[ -n "${GITHUB_OWNER}" ]] || die "GITHUB_OWNER is required."
  [[ -n "${GITHUB_REPO}" ]] || die "GITHUB_REPO is required."
  [[ -n "${GITHUB_BRANCH}" ]] || die "GITHUB_BRANCH is required."

  section "Initialize Deployment role permissions"
  local caller_account
  caller_account="$(ensure_aws_login)"
  ok "Current AWS account: ${caller_account}"
  ensure_account_matches_config "${caller_account}"

  local provider_arn
  provider_arn="$(ensure_oidc_provider)"

  local resources_dir
  resources_dir="${SCRIPT_DIR}/../resources/policies"

  local trust_file
  trust_file="$(mktemp)"
  local current_trust_file
  current_trust_file="$(mktemp)"
  local merged_trust_file
  merged_trust_file="$(mktemp)"
  render_template_file \
    "${resources_dir}/deployment-trust-policy.template.json" \
    "${trust_file}" \
    "OIDC_PROVIDER_ARN=${provider_arn}" \
    "GITHUB_OWNER=${GITHUB_OWNER}" \
    "GITHUB_REPO=${GITHUB_REPO}" \
    "GITHUB_BRANCH=${GITHUB_BRANCH}"

  if aws iam get-role --role-name "${DEPLOYMENT_ROLE_NAME}" >/dev/null 2>&1; then
    aws iam get-role \
      --role-name "${DEPLOYMENT_ROLE_NAME}" \
      --query 'Role.AssumeRolePolicyDocument' \
      --output json > "${current_trust_file}"
    merge_policy_documents "${current_trust_file}" "${trust_file}" "${merged_trust_file}"
    aws iam update-assume-role-policy \
      --role-name "${DEPLOYMENT_ROLE_NAME}" \
      --policy-document "file://${merged_trust_file}" >/dev/null
    ok "Incrementally merged trust policy for role: ${DEPLOYMENT_ROLE_NAME}"
  else
    aws iam create-role \
      --role-name "${DEPLOYMENT_ROLE_NAME}" \
      --assume-role-policy-document "file://${trust_file}" >/dev/null
    ok "Created role: ${DEPLOYMENT_ROLE_NAME}"
  fi

  local policy_file
  policy_file="$(mktemp)"
  local current_inline_policy_file
  current_inline_policy_file="$(mktemp)"
  local merged_inline_policy_file
  merged_inline_policy_file="$(mktemp)"
  render_template_file \
    "${resources_dir}/deployment-inline-policy.template.json" \
    "${policy_file}" \
    "REGION=${REGION}" \
    "ACCOUNT_ID=${ACCOUNT_ID}" \
    "IMAGE_REPOSITORY=${IMAGE_REPOSITORY}" \
    "FUNCTION_NAME=${FUNCTION_NAME}" \
    "EXECUTION_ROLE_ARN=${EXECUTION_ROLE_ARN}" \
    "DEPLOYMENT_ROLE_NAME=${DEPLOYMENT_ROLE_NAME}"

  if aws iam get-role-policy \
    --role-name "${DEPLOYMENT_ROLE_NAME}" \
    --policy-name "${INLINE_POLICY_NAME}" >/dev/null 2>&1; then
    aws iam get-role-policy \
      --role-name "${DEPLOYMENT_ROLE_NAME}" \
      --policy-name "${INLINE_POLICY_NAME}" \
      --query 'PolicyDocument' \
      --output json > "${current_inline_policy_file}"
    merge_policy_documents "${current_inline_policy_file}" "${policy_file}" "${merged_inline_policy_file}"
  else
    cp "${policy_file}" "${merged_inline_policy_file}"
  fi

  aws iam put-role-policy \
    --role-name "${DEPLOYMENT_ROLE_NAME}" \
    --policy-name "${INLINE_POLICY_NAME}" \
    --policy-document "file://${merged_inline_policy_file}" >/dev/null
  ok "Inline policy incrementally merged: ${INLINE_POLICY_NAME}"

  local role_arn
  role_arn="$(aws iam get-role --role-name "${DEPLOYMENT_ROLE_NAME}" --query 'Role.Arn' --output text)"

  rm -f \
    "${trust_file}" \
    "${current_trust_file}" \
    "${merged_trust_file}" \
    "${policy_file}" \
    "${current_inline_policy_file}" \
    "${merged_inline_policy_file}"

  echo
  echo "Deployment role ready:"
  echo "  ${role_arn}"
}

main "$@"
