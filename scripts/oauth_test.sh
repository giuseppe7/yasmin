#!/bin/bash
# See https://api.slack.com/docs/sign-in-with-slack for documentation.

# Check for necessary other CLI applications.
if ! command -v jq &> /dev/null; then
    echo "jq could not be found."
    exit 1
fi

# Check for environment variables and necessary applications.
if [[ -z "${YASMIN_USER_ID}" || -z "{$YASMIN_TEAM_ID}" || -z "${YASMIN_ACCESS_TOKEN}" ]]; then
  printf 'Ensure YASMIN environment variables are set before continuing.\n'
  exit 1
fi

printf '\nInvoking the api.test endpoint...\n'
RESULT=$( curl -s "https://slack.com/api/api.test" --request POST -H "Content-type: application/json" -H "Authorization: Bearer ${YASMIN_ACCESS_TOKEN}" )
echo ${RESULT} | jq

printf '\nInvoking the users.identity endpoint...\n'
RESULT=$( curl -s "https://slack.com/api/users.identity" -H "Content-type: application/json" -H "Authorization: Bearer ${YASMIN_ACCESS_TOKEN}" )
echo ${RESULT} | jq

printf '\nInvoking the conversations.list endpoint...\n'
RESULT=$( curl -s "https://slack.com/api/conversations.list?limit=50" -H "Content-type: application/json" -H "Authorization: Bearer ${YASMIN_ACCESS_TOKEN}" )
echo ${RESULT} | jq