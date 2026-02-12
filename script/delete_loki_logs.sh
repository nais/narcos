#!/bin/bash

# Start port-forwarding to Loki compactor
# Sets the global variable PORT_FORWARD_PID
start_port_forward() {
  kubectl port-forward -n nais-system loki-compactor-0 3100 >/dev/null 2>&1 &
  PORT_FORWARD_PID=$!
  sleep 3
}

# Stop port-forwarding
stop_port_forward() {
  if [[ -n "$PORT_FORWARD_PID" ]]; then
    kill $PORT_FORWARD_PID 2>/dev/null
  fi
}

# List all delete requests from Loki
# Usage: list_delete_requests
list_delete_requests() {
  echo "Fetching delete requests from Loki..."

  start_port_forward

  curl -s "localhost:3100/loki/api/v1/delete" | jq

  stop_port_forward
}

# Delete logs from Loki using structured parameters
# Usage: delete_logs <namespace> <app_name> <days_since> ["<additional_filters>"] ["<regex_pattern>"]
delete_logs() {
  local NAMESPACE="$1"
  local APP_NAME="$2"
  local DAYS_SINCE="$3"
  local ADDITIONAL_FILTERS="$4"
  local REGEX_PATTERN="$5"

  if [[ -z "$NAMESPACE" || -z "$APP_NAME" || -z "$DAYS_SINCE" ]]; then
    echo "Usage: $0 delete <namespace> <app_name> <days_since> [\"<additional_filters>\"] [\"<regex_pattern>\"]"
    echo "Example with regex: $0 delete myns myapp 7 '' 'some string: 1234'"
    return 1
  fi

  # Calculate the start timestamp (DAYS_SINCE ago, in seconds since epoch)
  if date -v-"${DAYS_SINCE}"d +%s >/dev/null 2>&1; then
    # macOS
    START_TS=$(date -v-"${DAYS_SINCE}"d +%s)
  else
    # Linux
    START_TS=$(date --date="${DAYS_SINCE} days ago" +%s)
  fi

  # Compose the query
  local QUERY="{service_namespace=\"${NAMESPACE}\", service_name = \"${APP_NAME}\"}"
  if [[ -n "$ADDITIONAL_FILTERS" ]]; then
    QUERY="${QUERY} | ${ADDITIONAL_FILTERS}"
  fi
  if [[ -n "$REGEX_PATTERN" ]]; then
    QUERY="${QUERY} |~ \"(?i)${REGEX_PATTERN}\""
  fi

  # URI-encode the query
  ENCODED_QUERY=$(jq -rn --arg str "$QUERY" '$str|@uri')

  # Show the query and confirm
  echo "The following query will be used for deletion:"
  echo "Query: $QUERY"
  echo "Start timestamp (epoch): $START_TS"
  echo "cURL command:"
  echo "curl -X POST \"localhost:3100/loki/api/v1/delete?query=${ENCODED_QUERY}&start=${START_TS}\""
  read -p "Proceed with deletion? [y/N]: " CONFIRM
  if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Aborted."
    return 1
  fi

  # Start port-forwarding Loki compactor in the background
  start_port_forward

  # Call Loki's delete API
  curl -X POST "localhost:3100/loki/api/v1/delete?query=${ENCODED_QUERY}&start=${START_TS}"

  # Read back the newly added delete request before terminating the port-forward
  echo -e "\n\nFetching delete requests from Loki:"
  curl -s "localhost:3100/loki/api/v1/delete" | jq

  stop_port_forward
}

# Show usage information
show_usage() {
  echo "Usage: $0 <command> [options]"
  echo ""
	echo "We're running one Loki per cluster, so remember to run against the correct instance"
	echo ""
  echo "Commands:"
  echo "  delete <namespace> <app_name> <days_since> [\"<additional_filters>\"] [\"<regex_pattern>\"]"
  echo "         Delete logs matching the specified criteria"
  echo "         Example: $0 delete myns myapp 7"
  echo "         Example with regex: $0 delete myns myapp 7 '' 'some string: 1234'"
  echo ""
  echo "  list   List all delete requests from Loki"
  echo "         Example: $0 list"
}

# Main script logic
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  COMMAND="$1"
  shift

  case "$COMMAND" in
    delete)
      delete_logs "$@"
      ;;
    list)
      list_delete_requests
      ;;
    *)
      show_usage
      exit 1
      ;;
  esac
fi
