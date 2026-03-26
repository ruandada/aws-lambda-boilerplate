#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/common.sh
source "${SCRIPT_DIR}/lib/common.sh"

main() {
  local lambda_dir="${1:-}"
  [[ -n "${lambda_dir}" ]] || die "usage: $0 <lambda_directory>"

  need_cmd git
  ensure_aws_cli

  local config_file
  config_file="$(config_file_for_dir "${lambda_dir}")"
  load_lambda_config "${config_file}"

  section "Initialize OIDC provider"
  local caller_account
  caller_account="$(ensure_aws_login)"
  ok "Current AWS account: ${caller_account}"
  ensure_account_matches_config "${caller_account}"

  local provider_arn
  provider_arn="$(ensure_oidc_provider)"

  echo
  echo "OIDC provider ready:"
  echo "  ${provider_arn}"
}

main "$@"
