#!/bin/bash
BINARY_DIR=../bin
BINARY_NAME=aws-iot-authorizer
PACKAGE_NAME=$BINARY_DIR/$BINARY_NAME.zip
AUTHORIZER_LAMBDA_FUNCTION_NAME="IotDeviceAuthenticationHandler"
AUTHORIZER_LAMBDA_ROLE="arn:aws:iam::037420171134:role/authorizer-lambda"

# Create the lambda for the custom authorizer.
aws lambda create-function --region "us-west-2" --runtime "go1.x" \
  --function-name $AUTHORIZER_LAMBDA_FUNCTION_NAME \
  --zip-file fileb://$PACKAGE_NAME \
  --role $AUTHORIZER_LAMBDA_ROLE \
  --handler $BINARY_NAME
