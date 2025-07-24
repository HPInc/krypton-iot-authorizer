#!/bin/bash
AUTHORIZER_NAME="KryptonDeviceAuthenticationAuthorizer"
AUTHORIZER_LAMBDA_FUNCTION_NAME="IotDeviceAuthenticationHandler"
AUTHORIZER_ARN="arn:aws:iot:us-west-2:037420171134:authorizer/KryptonDeviceAuthenticationAuthorizer"

# The ARN (Amazon Resource Name) of the authorizer lambda function.
AUTHORIZER_FUNCTION_ARN="arn:aws:lambda:us-west-2:037420171134:function:$AUTHORIZER_LAMBDA_FUNCTION_NAME"

# Register the authorizer.
aws iot create-authorizer --region "us-west-2" --authorizer-name $AUTHORIZER_NAME \
  --authorizer-function-arn $AUTHORIZER_FUNCTION_ARN \
  --token-key-name Authorization --status ACTIVE \
  --signing-disabled

aws lambda add-permission --region "us-west-2" --function-name $AUTHORIZER_LAMBDA_FUNCTION_NAME \
  --principal iot.amazonaws.com --source-arn $AUTHORIZER_ARN \
  --statement-id Id-456 --action "lambda:InvokeFunction"
