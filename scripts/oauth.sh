#!/bin/bash
# See https://api.slack.com/docs/sign-in-with-slack for documentation.
# TODO: Move from script to simple go cli?

# Check for necessary other CLI applications.
if ! command -v jq &> /dev/null; then
    echo "jq could not be found."
    exit 1
fi

# Check for environment variables and necessary applications.
if [[ -z "${CLIENT_ID}" || -z "{$CLIENT_SECRET}" ]]; then
  printf 'Ensure environment variables CLIENT_ID and CLIENT_SECRET are set before continuing.\n'
  exit 1
fi
USER_SCOPE='identity.basic,identity.team'

printf '\n'
printf '1. Authorization Request. Opening authorization URL to grant access and receive grant URL with code...\n'
LOCATION=$(
  curl -is "https://slack.com/oauth/v2/authorize?user_scope=${USER_SCOPE}&client_id=${CLIENT_ID}&tracked=1" \
    | grep -Fi location \
    | sed -e 's/location: //g'
)
# TODO: Really, I want to launch Chrome?
open -n -a "Google Chrome" --args "${LOCATION}"

# Parse the code from the redirect response.
printf '\n'
printf '2. Authorization Grant. Enter in the CODE received:\n'
read CODE
RESULT=$(
  curl -s "https://slack.com/api/oauth.v2.access?client_id=${CLIENT_ID}&client_secret=${CLIENT_SECRET}&code=${CODE}" 
) 
echo ${RESULT} | jq

# Review the results of the call to viewing oneself via the API.
SUCCESS=$( echo ${RESULT} | jq '.ok')
if [ ${SUCCESS} != "true" ] ; then
  printf '\nOAUTH flow unsuccessful.\n'
  exit 1
fi

# Retain the triad.
USER_ID=$( echo ${RESULT} | jq -r '.authed_user.id' )
TEAM_ID=$( echo ${RESULT} | jq -r '.team.id' )
ACCESS_TOKEN=$( echo ${RESULT} | jq -r '.authed_user.access_token' )
printf '\nOAUTH flow successful.\n'

# Request the caller's identity via the API with the obtained token.
printf '\n'
printf '3. Protected Resource. Invoking the API with obtained token...\n'
RESULT=$( curl -s "https://slack.com/api/users.identity" -H "Authorization: Bearer ${ACCESS_TOKEN}" )
echo ${RESULT} | jq

# Request the the api.test for good measure.
printf '\n'
printf '4. Protected Resource. Invoking the api.test endpoint...\n'
RESULT=$( curl -s "https://slack.com/api/api.test" --request POST -H "Content-type: application/json" -H "Authorization: Bearer ${YASMIN_ACCESS_TOKEN}" )
echo ${RESULT} | jq

# All set, use these below.
printf '\n'
printf '4. Done, configure your environment.\n'
echo "export YASMIN_USER_ID=${USER_ID}"
echo "export YASMIN_TEAM_ID=${TEAM_ID}"
echo "export YASMIN_ACCESS_TOKEN=${ACCESS_TOKEN}"