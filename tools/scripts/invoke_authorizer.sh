#!/bin/bash
# Sample test script to invoke a custom IoT authorizer in AWS IoT Core. The script requires
# you to provide the access token in a file named 'token.txt'. Ensure this file is not checked
# into version control for security reasons and delete it after use.
# Usage: ./invoke_authorizer.sh
# Dependencies: AWS CLI configured with appropriate permissions.
AUTHORIZER_NAME="KryptonDeviceAuthenticationAuthorizer"

TOKEN=$(cat token.txt)
# Enable the following line to debug and see the token.
# echo $TOKEN

if [ -z "$TOKEN" ]; then
  echo "Error: token.txt is empty or not found. Please provide a valid token."
  exit 1
fi

aws iot test-invoke-authorizer --region "us-west-2" --authorizer-name $AUTHORIZER_NAME \
  --http-context '{"headers":{"Authorization": "Bearer"}}'
