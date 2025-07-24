#!/bin/bash
BINARY_DIR=../bin
BINARY_NAME=aws-iot-authorizer
PACKAGE_NAME=$BINARY_DIR/$BINARY_NAME.zip
AUTHORIZER_LAMBDA_FUNCTION_NAME="IotDeviceAuthenticationHandler"

aws lambda update-function-code --region "us-west-2" --function-name $AUTHORIZER_LAMBDA_FUNCTION_NAME \
  --zip-file fileb://$PACKAGE_NAME
