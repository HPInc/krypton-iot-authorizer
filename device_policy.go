// package github.com/HPInc/krypton-iot-authorizer
// Author: Mahesh Unnikrishnan
// Component: Krypton AWS IoT Authorizer Lambda
// (C) HP Development Company, LP
package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

const (
	// ARN of the client allowed to connect to the hub.
	clientResourceFormat = "arn:aws:iot:%s:%s:client/%s"

	//////////////////// Subscribe topics /////////////////////////////////////
	//                Topics devices need to subscribe to.
	// - topic on which tasks intended for the device to execute are published.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topicfilter/v1/DEVICE_ID/tasks
	// arn:aws:iot:us-west-2:11111122222:topicfilter/v1/d4a8cd9a-be0e-4e71-b1b5-91d0226dad0d/tasks
	deviceTasksTopicSubscribeFormat = "arn:aws:iot:%s:%s:topicfilter/v1/%s/tasks"

	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/DEVICE_ID/tasks
	// arn:aws:iot:us-west-2:11111122222:topic/v1/d4a8cd9a-be0e-4e71-b1b5-91d0226dad0d/tasks
	deviceTasksTopicReceiveFormat = "arn:aws:iot:%s:%s:topic/v1/%s/tasks"

	// - topic on which the management service broadcasts messages intended for
	// all devices managed by it.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topicfilter/v1/@devices/MANAGEMENT_SERVICE
	// arn:aws:iot:us-west-2:11111122222:topicfilter/v1/@devices/hpcem
	deviceServiceBroadcastTopicSubscribeFormat = "arn:aws:iot:%s:%s:topicfilter/v1/@devices/+"

	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/@devices/MANAGEMENT_SERVICE
	// arn:aws:iot:us-west-2:11111122222:topic/v1/@devices/hpcem
	deviceServiceBroadcastTopicReceiveFormat = "arn:aws:iot:%s:%s:topic/v1/@devices/+"
	///////////////////////////////////////////////////////////////////////////

	//////////////////// Publish topics ///////////////////////////////////////
	//                Topics to which devices publish messages.
	// - topic to which devices publish task responses.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/@cloud/task_responses
	// arn:aws:iot:us-west-2:11111122222:topic/v1/@cloud/task_responses
	cloudTaskResponsesTopic = "arn:aws:iot:%s:%s:topic/v1/@cloud/task_responses"

	// - topic to which devices publish messages intended for their
	// device managment service.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/@cloud
	// arn:aws:iot:us-west-2:11111122222:topic/v1/@cloud
	cloudServiceMessageTopic = "arn:aws:iot:%s:%s:topic/v1/@cloud"
	///////////////////////////////////////////////////////////////////////////
)

var (
	// AWS IoT Core policy actions.
	connectAction   = []string{"iot:Connect"}
	publishAction   = []string{"iot:Publish"}
	receiveAction   = []string{"iot:Receive"}
	subscribeAction = []string{"iot:Subscribe"}
)

func createIotPolicyDocumentForDevice(awsRegion string, awsAccount string,
	deviceID string) []*events.IAMPolicyDocument {
	policyDoc := events.IAMPolicyDocument{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action: connectAction,
				Effect: "Allow",
				Resource: []string{fmt.Sprintf(clientResourceFormat, awsRegion,
					awsAccount, deviceID)},
			},
			{
				Action: subscribeAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(deviceTasksTopicSubscribeFormat, awsRegion, awsAccount, deviceID),
					fmt.Sprintf(deviceServiceBroadcastTopicSubscribeFormat, awsRegion,
						awsAccount)},
			},
			{
				Action: receiveAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(deviceTasksTopicReceiveFormat, awsRegion, awsAccount, deviceID),
					fmt.Sprintf(deviceServiceBroadcastTopicReceiveFormat, awsRegion,
						awsAccount)},
			},
			{
				Action: publishAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(cloudTaskResponsesTopic, awsRegion, awsAccount),
					fmt.Sprintf(cloudServiceMessageTopic, awsRegion, awsAccount)},
			},
		},
	}
	return []*events.IAMPolicyDocument{&policyDoc}
}
