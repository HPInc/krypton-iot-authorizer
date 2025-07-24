// package github.com/HPInc/krypton-iot-authorizer
// Author: Mahesh Unnikrishnan
// Component: Krypton AWS IoT Authorizer Lambda
// (C) HP Development Company, LP
package main

import "net/url"

// Check if the list of protocols sent by the broker for the event contains
// the specified protocol.
func containsProtocol(protocols *[]string, matchProtocol string) bool {
	for _, protocol := range *protocols {
		if protocol == matchProtocol {
			return true
		}
	}
	return false
}

// Extract the specified query parameter from the parsed URL map. If the parameter
// does not exist, return an empty string.
// Eg: For MQTT connect requests, the username field contains a list of parameters,
// including the device/app access token.
func getQueryParameter(values url.Values, parameterName string) string {
	value, ok := values[parameterName]
	if !ok {
		return ""
	}
	return value[0]
}
