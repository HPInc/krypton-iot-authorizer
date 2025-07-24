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
	sharedSubscriptionFormat = "$share/krypton/"

	//////////////////// Subscribe topics /////////////////////////////////////
	//                Topics scheduler needs to subscribe to.
	// - topic on which devices publish task responses.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topicfilter/SHARED_SUBSCRIPTION/v1/@cloud/task_responses
	// arn:aws:iot:us-west-2:11111122222:topicfilter/$share/krypton/v1/@cloud/task_responses
	taskResponsesTopicSubscribeFormat = "arn:aws:iot:%s:%s:topicfilter/%sv1/@cloud/task_responses"

	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/SHARED_SUBSCRIPTION/v1/@cloud/task_responses
	// arn:aws:iot:us-west-2:11111122222:topic/$share/krypton/v1/@cloud/task_responses
	taskResponsesTopicReceiveFormat = "arn:aws:iot:%s:%s:topic/%sv1/@cloud/task_responses"

	// - topic on which devices publish messages intended for their management service
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topicfilter/SHARED_SUBSCRIPTION/v1/@cloud
	// arn:aws:iot:us-west-2:11111122222:topicfilter/v1/@cloud
	// arn:aws:iot:us-west-2:11111122222:topicfilter/$share/krypton/v1/@cloud
	serviceMessageTopicSubscribeFormat = "arn:aws:iot:%s:%s:topicfilter/%sv1/@cloud"

	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/SHARED_SUBSCRIPTION/v1/@cloud
	// arn:aws:iot:us-west-2:11111122222:topic/v1/@cloud
	// arn:aws:iot:us-west-2:11111122222:topic/$share/krypton/v1/@cloud
	serviceMessageTopicReceiveFormat = "arn:aws:iot:%s:%s:topic/%sv1/@cloud"
	///////////////////////////////////////////////////////////////////////////

	//////////////////// Publish topics ///////////////////////////////////////
	//            Topics to which the scheduler publishes messages.
	// - topic to which the scheduler publishes task responses.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/*/tasks
	// arn:aws:iot:us-west-2:11111122222:topic/v1/*/tasks
	deviceTasksTopic = "arn:aws:iot:%s:%s:topic/v1/*/tasks"

	// - topic to which scheduler publishes broadcast messages from the
	// corresponding device managment service.
	// arn:aws:iot:AWS_REGION:AWS_ACCOUNT_ID:topic/v1/@devices/*
	// arn:aws:iot:us-west-2:11111122222:topic/v1/@devices/*
	deviceBroadcastMessageTopic = "arn:aws:iot:%s:%s:topic/v1/@devices/*"
	///////////////////////////////////////////////////////////////////////////
)

func createIotPolicyDocumentForScheduler(awsRegion string, awsAccount string,
	clientID string) []*events.IAMPolicyDocument {
	policyDoc := events.IAMPolicyDocument{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action: connectAction,
				Effect: "Allow",
				Resource: []string{fmt.Sprintf(clientResourceFormat, awsRegion,
					awsAccount, clientID)},
			},
			{
				Action: subscribeAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(taskResponsesTopicSubscribeFormat, awsRegion, awsAccount, ""),
					fmt.Sprintf(taskResponsesTopicSubscribeFormat, awsRegion, awsAccount,
						sharedSubscriptionFormat),
					fmt.Sprintf(serviceMessageTopicSubscribeFormat, awsRegion, awsAccount, ""),
					fmt.Sprintf(serviceMessageTopicSubscribeFormat, awsRegion, awsAccount,
						sharedSubscriptionFormat)},
			},
			{
				Action: receiveAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(taskResponsesTopicReceiveFormat, awsRegion, awsAccount, ""),
					fmt.Sprintf(taskResponsesTopicReceiveFormat, awsRegion, awsAccount,
						sharedSubscriptionFormat),
					fmt.Sprintf(serviceMessageTopicReceiveFormat, awsRegion, awsAccount, ""),
					fmt.Sprintf(serviceMessageTopicReceiveFormat, awsRegion, awsAccount,
						sharedSubscriptionFormat)},
			},
			{
				Action: publishAction,
				Effect: "Allow",
				Resource: []string{
					fmt.Sprintf(deviceTasksTopic, awsRegion, awsAccount),
					fmt.Sprintf(deviceBroadcastMessageTopic, awsRegion, awsAccount)},
			},
		},
	}
	return []*events.IAMPolicyDocument{&policyDoc}
}
