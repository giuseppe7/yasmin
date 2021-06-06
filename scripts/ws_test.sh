#!/bin/bash
# See https://api.slack.com/apis/connections/socket-implement

# Check for necessary other CLI applications.
if ! command -v jq &> /dev/null; then
    echo "jq could not be found."
    exit 1
fi

# Check for environment variables and necessary applications.
if [[ -z "${YASMIN_APP_TOKEN}" ]]; then
  printf 'Ensure YASMIN environment variables are set before continuing.\n'
  exit 1
fi

printf '\nInvoking the api.test endpoint...\n'
RESULT=$( 
  curl -s -X POST "https://slack.com/api/apps.connections.open" \
    -H "Content-type: application/x-www-form-urlencoded" \
    -H "Authorization: Bearer ${YASMIN_APP_TOKEN}"
)
echo ${RESULT} | jq
