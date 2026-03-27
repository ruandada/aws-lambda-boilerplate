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

  section "Initialize Execution role permissions"
  local caller_account
  caller_account="$(ensure_aws_login)"
  ok "Current AWS account: ${caller_account}"
  ensure_account_matches_config "${caller_account}"

  if [[ "${EXECUTION_ROLE_NAME}" == "${DEPLOYMENT_ROLE_NAME}" ]]; then
    info "ExecutionRoleName equals DeploymentRoleName (${EXECUTION_ROLE_NAME})."
    info "Skipping trust overwrite to avoid removing OIDC trust from the shared role."

    local has_lambda_principal
    has_lambda_principal="$(aws iam get-role \
      --role-name "${EXECUTION_ROLE_NAME}" \
      --query "contains(to_string(Role.AssumeRolePolicyDocument), 'lambda.amazonaws.com')" \
      --output text)"

    [[ "${has_lambda_principal}" == "True" ]] || die "Shared role trust policy does not include lambda.amazonaws.com. Run deployment-role init first."
  else
    local exec_trust_file
    exec_trust_file="$(mktemp)"
    local current_exec_trust_file
    current_exec_trust_file="$(mktemp)"
    local merged_exec_trust_file
    merged_exec_trust_file="$(mktemp)"

    local resources_dir
    resources_dir="${SCRIPT_DIR}/../resources/policies"
    render_template_file \
      "${resources_dir}/execution-trust-policy.template.json" \
      "${exec_trust_file}"

    if aws iam get-role --role-name "${EXECUTION_ROLE_NAME}" >/dev/null 2>&1; then
      aws iam get-role \
        --role-name "${EXECUTION_ROLE_NAME}" \
        --query 'Role.AssumeRolePolicyDocument' \
        --output json > "${current_exec_trust_file}"
      merge_policy_documents "${current_exec_trust_file}" "${exec_trust_file}" "${merged_exec_trust_file}"
      aws iam update-assume-role-policy \
        --role-name "${EXECUTION_ROLE_NAME}" \
        --policy-document "file://${merged_exec_trust_file}" >/dev/null
      ok "Incrementally merged execution role trust policy: ${EXECUTION_ROLE_NAME}"
    else
      aws iam create-role \
        --role-name "${EXECUTION_ROLE_NAME}" \
        --assume-role-policy-document "file://${exec_trust_file}" >/dev/null
      ok "Created execution role: ${EXECUTION_ROLE_NAME}"
    fi

    rm -f "${exec_trust_file}" "${current_exec_trust_file}" "${merged_exec_trust_file}"
  fi

  ensure_execution_basic_policy "${EXECUTION_ROLE_NAME}"

  echo
  echo "Execution role ready:"
  echo "  ${EXECUTION_ROLE_NAME}"
}

main "$@"
