#!/usr/bin/env bash
set -euo pipefail

CONFIG_FILE="${XDG_CONFIG_HOME:-${HOME}/.config}/kube-tenant-map"

# Define some logging utilities
if tput setaf 1 &> /dev/null; then
  tput sgr0
  RED=$(tput setaf 1)
  YELLOW=$(tput setaf 3)
  WHITE=$(tput setaf 7)
  BOLD=$(tput bold)
  RESET=$(tput sgr0)
fi

function log::_level_color {
  local log_level
  log_level=$1

  case $log_level in
    INFO) echo "${BOLD}${WHITE} " ;;
    WARN) echo "${BOLD}${YELLOW} " ;;
    ERROR) echo "${BOLD}${RED}" ;;
  esac
}

function log::_write_log {
  local timestamp log_level color
  log_level=$1
  shift

  timestamp=$(date +'%Y.%m.%d %H:%M:%S')
  color=$(log::_level_color "${log_level}")
  >&2 printf '[%s%s%s|%s]: %s\n' "$color" "$log_level" "$RESET" "$timestamp" "${*}"
}

function log::error {
  log::_write_log "ERROR" "$@"
}

function log::warn {
  log::_write_log "WARN" "$@"
}

function log::info {
  log::_write_log "INFO" "$@"
}


# Check dependencies
for dep in kubectx fzf narc nais jq; do
  if ! command -v "$dep" &>/dev/null; then
    log::error "Error: required dependency '$dep' not found in PATH" >&2
    exit 1
  fi
done

# Check config file
if [[ ! -f "$CONFIG_FILE" ]]; then
  log::error "Error: config file not found: $CONFIG_FILE" >&2
  log::error "" >&2
  log::error "Expected format (one entry per line):" >&2
  log::error "  glob-pattern=tenant" >&2
  log::error "" >&2
  log::error "Example:" >&2
  log::error "  nav-*=NAV" >&2
  log::error "  myorg-*=myorg.io" >&2
  exit 1
fi

# Require at least one argument (the command to run)
if [[ $# -eq 0 ]]; then
  log::info "Usage: $0 <command> [args...]" >&2
  exit 1
fi

# Read config: build arrays of patterns and tenants
declare -a PATTERNS=()
declare -a TENANT_FOR_PATTERN=()

while IFS='=' read -r pattern tenant || [[ -n "$pattern" ]]; do
  # Skip blank lines and comments
  [[ -z "$pattern" || "$pattern" == \#* ]] && continue
  PATTERNS+=("$pattern")
  TENANT_FOR_PATTERN+=("$tenant")
done < "$CONFIG_FILE"

# Function: extract clustername from a context name
extract_clustername() {
  local ctx="$1"
  local i
  for i in "${!PATTERNS[@]}"; do
    # shellcheck disable=SC2053
    if [[ "$ctx" == ${PATTERNS[$i]} ]]; then
      local pattern="${PATTERNS[$i]}"
      # Strip literal prefix (everything before the '*')
      local prefix="${pattern%%\**}"
      local remainder="${ctx#"$prefix"}"
      # Strip trailing -v<digits> suffix
      if [[ "$remainder" =~ ^(.*)-v[0-9]+$ ]]; then
        remainder="${BASH_REMATCH[1]}"
      fi
      echo "$remainder"
      return 0
    fi
  done
  return 1
}

# Function: resolve tenant for a context name
resolve_tenant() {
  local ctx="$1"
  local i
  for i in "${!PATTERNS[@]}"; do
    # shellcheck disable=SC2053
    if [[ "$ctx" == ${PATTERNS[$i]} ]]; then
      echo "${TENANT_FOR_PATTERN[$i]}"
      return 0
    fi
  done
  return 1
}

# Get all contexts, filter to those matching a pattern
declare -a MATCHING_CONTEXTS=()
while IFS= read -r ctx; do
  if resolve_tenant "$ctx" &>/dev/null; then
    MATCHING_CONTEXTS+=("$ctx")
  fi
done < <(kubectx)

if [[ ${#MATCHING_CONTEXTS[@]} -eq 0 ]]; then
  log::error "No contexts matched any pattern in $CONFIG_FILE" >&2
  exit 1
fi

# Let user pick contexts via fzf --multi
mapfile -t SELECTED < <(printf '%s\n' "${MATCHING_CONTEXTS[@]}" | fzf --multi --prompt="Select contexts: ")

if [[ ${#SELECTED[@]} -eq 0 ]]; then
  log::info "No contexts selected, exiting." >&2
  exit 0
fi

# Group selected contexts by tenant
declare -A TENANT_CONTEXTS  # tenant -> newline-separated contexts

for ctx in "${SELECTED[@]}"; do
  tenant=$(resolve_tenant "$ctx")
  if [[ -z "${TENANT_CONTEXTS[$tenant]+set}" ]]; then
    TENANT_CONTEXTS[$tenant]="$ctx"
  else
    TENANT_CONTEXTS[$tenant]+=$'\n'"$ctx"
  fi
done

# Track failures: "context:exitcode"
declare -a FAILURES=()

# For each tenant: connect naisdevice, then run command per context
for tenant in "${!TENANT_CONTEXTS[@]}"; do
  log::info "==> Tenant: $tenant"

  # Set naisdevice tenant
  if ! narc tenant set "$tenant"; then
    log::error "Error: 'narc tenant set $tenant' failed, skipping tenant" >&2
    # Record all contexts for this tenant as failures
    while IFS= read -r ctx; do
      FAILURES+=("${ctx}:narc-failed")
    done <<< "${TENANT_CONTEXTS[$tenant]}"
    continue
  fi

  # Connect naisdevice
  if ! nais device connect; then
    log::error "Error: 'nais device connect' failed for tenant $tenant, skipping" >&2
    while IFS= read -r ctx; do
      FAILURES+=("${ctx}:connect-failed")
    done <<< "${TENANT_CONTEXTS[$tenant]}"
    continue
  fi

  # Poll for Connected status (up to 60s)
  log::info "Waiting for naisdevice to connect..."
  connected=false
  for _ in $(seq 1 12); do
    if nais device status 2>/dev/null | grep -q "Connected"; then
      connected=true
      break
    fi
    sleep 5
  done

  if [[ "$connected" != true ]]; then
    log::error "Error: timed out waiting for naisdevice to connect for tenant $tenant, skipping" >&2
    while IFS= read -r ctx; do
      FAILURES+=("${ctx}:connect-timeout")
    done <<< "${TENANT_CONTEXTS[$tenant]}"
    continue
  fi

  log::info "Connected to naisdevice for tenant $tenant"

  # Run command for each context in this tenant
  while IFS= read -r ctx; do
    log::info "--> Context: $ctx"

    # Check gateway connectivity for this context
    clustername=$(extract_clustername "$ctx")
    gateway_connected=false
    for _ in $(seq 1 6); do
      if nais device gateway list --output json 2>/dev/null \
          | jq -e --arg name "$clustername" '.[] | select(.name == $name or .name == ("nais-device-gw-k8s-" + $name)) | .connected' \
          | grep -q true; then
        gateway_connected=true
        break
      fi
      log::info "Waiting for gateway to be connected"
      sleep 5
    done

    if [[ "$gateway_connected" != true ]]; then
      log::error "Error: gateway '$clustername' not connected for context $ctx after 30s, skipping" >&2
      FAILURES+=("${ctx}:gateway-timeout")
      continue
    fi

    log::info "Gateway for cluster ${clustername} is connected"
    sleep 2

    export NAIS_TENANT="$tenant"
    export KUBE_CONTEXT="$ctx"
    exit_code=0
    "$@" || exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
      log::error "Error: command failed for context $ctx (exit code $exit_code)" >&2
      FAILURES+=("${ctx}:${exit_code}")
    fi
  done <<< "${TENANT_CONTEXTS[$tenant]}"
done

# Summary
echo ""
if [[ ${#FAILURES[@]} -eq 0 ]]; then
  log::info "All contexts completed successfully."
else
  log::warn "Failures summary:"
  for entry in "${FAILURES[@]}"; do
    ctx="${entry%%:*}"
    code="${entry##*:}"
    log::warn "  $ctx  (exit: $code)"
  done
  exit 1
fi
