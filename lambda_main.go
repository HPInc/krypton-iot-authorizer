// package github.com/HPInc/krypton-iot-authorizer
// Author: Mahesh Unnikrishnan
// Component: Krypton AWS IoT Authorizer Lambda
// (C) HP Development Company, LP
package main

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const (
	ENV_DSTS_JWKS_URL   = "DSTS_JWKS_URL"
	dstsIssuerName      = "HP Device Token Service"
	schedulerAppID      = "bebc5cbf-acc0-431f-8c4e-c582dc2489e2"
	authorizerUserAgent = "krypton-iot-authorizer"

	// Request headers
	headerIotCustomAuthorizer = "x-amz-customauthorizer-name"
	headerUserAgent           = "User-Agent"
	headerAuthorization       = "Authorization"
	bearerTokenPrefix         = "Bearer "

	// Query parameters
	paramDeviceToken = "device_token"
	paramClientId    = "client_id"

	// Token types - asserted as values of the 'typ' claim.
	TokenTypeDeviceAccessToken = "device"
	TokenTypeAppAccessToken    = "app"
)

var (
	// Device STS JWKs endpoint URL.
	dstsJwksUrl string
)

type DstsTokenClaims struct {
	// Standard JWT claims such as 'aud', 'exp', 'jti', 'iat', 'iss', 'nbf',
	// 'sub'
	// 'sub' claim is set to the unique ID assigned to the device after enrollment.
	jwt.RegisteredClaims

	// Type of token. Possible values are:
	//  - device: device access tokens
	//  - app: app access token
	TokenType string `json:"typ"`

	// The ID of the tenant to which the device belongs.
	TenantID string `json:"tid"`

	// The device management service responsible for managing this device.
	ManagementService string `json:"ms"`
}

func IotDeviceAuthenticationHandler(ctx context.Context,
	event events.IoTCoreCustomAuthorizerRequest) (events.IoTCoreCustomAuthorizerResponse, error) {
	var clientID, deviceAccessToken string

	lambdaCtx, exists := lambdacontext.FromContext(ctx)
	if !exists {
		iotLogger.Error("No context information found in lambda context!")
		return failedAuthResponse(), ErrNoLambdaContext
	}

	// Parse the lambda context to determine the region and AWS account.
	// Example:
	// "arn:aws:lambda:us-west-2:719809574944:function:krypton-iot-authorizer-lambda"
	result := strings.Split(lambdaCtx.InvokedFunctionArn, ":")
	awsRegion := result[3]
	awsAccount := result[4]

	// Log information about the event received.
	eventJson, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		iotLogger.Error("Error unmarshaling IoT device authorization event",
			zap.Error(err),
		)
		return failedAuthResponse(), err
	}

	iotLogger.Debug("Krypton IoT authorizer lambda invoked!",
		zap.String("AWS region:", awsRegion),
		zap.String("AWS account:", awsAccount),
		zap.ByteString("Event info:", eventJson),
	)

	// Check to see if the request specified the device access token in the
	// authorization header.
	if containsProtocol(&event.Protocols, protocolHttp) {
		if event.ProtocolData != nil {
			if event.ProtocolData.HTTP != nil && event.ProtocolData.HTTP.Headers != nil {
				authzHeader, ok := event.ProtocolData.HTTP.Headers[headerAuthorization]
				if ok {
					if !strings.HasPrefix(authzHeader, bearerTokenPrefix) {
						iotLogger.Error("No bearer token specified in authorization header")
						return events.IoTCoreCustomAuthorizerResponse{}, ErrUnauthorized
					}
					deviceAccessToken = strings.TrimPrefix(authzHeader, bearerTokenPrefix)

					// Parse the client ID from the query string, if specified.
					if event.ProtocolData.HTTP.QueryString != "" {
						values, err := url.ParseQuery(event.ProtocolData.HTTP.QueryString)
						if err != nil {
							iotLogger.Error("Failed to parse the query string specified in the HTTP request!",
								zap.Error(err),
							)
							return events.IoTCoreCustomAuthorizerResponse{}, ErrBadRequest
						}
						clientID = getQueryParameter(values, paramClientId)
					}
				}
			}
		}
	}

	if deviceAccessToken == "" {
		if containsProtocol(&event.Protocols, protocolMqtt) {
			if (event.ProtocolData != nil) && (event.ProtocolData.MQTT != nil) {
				if event.ProtocolData.MQTT.Username == "" {
					iotLogger.Error("MQTT username field was not specified!")
					return events.IoTCoreCustomAuthorizerResponse{}, ErrBadRequest
				}

				values, err := url.ParseQuery(event.ProtocolData.MQTT.Username)
				if err != nil {
					iotLogger.Error("Failed to parse the username specified in the MQTT context!",
						zap.Error(err),
					)
					return events.IoTCoreCustomAuthorizerResponse{}, ErrBadRequest
				}

				deviceAccessToken = getQueryParameter(values, paramDeviceToken)
				clientID = event.ProtocolData.MQTT.ClientID
			}
		}
	}

	if deviceAccessToken == "" {
		iotLogger.Error("Device access token was not specified in HTTP header or MQTT context!")
		return events.IoTCoreCustomAuthorizerResponse{}, ErrUnauthorized
	}

	// Validate the provided DSTS access token - it may be either an app access
	// token or a device access token.
	claims, err := validateDstsAccessToken(deviceAccessToken)
	if err != nil {
		iotLogger.Error("Failed to validate the specified access token!",
			zap.Error(err),
		)
		return failedAuthResponse(), ErrUnauthorized
	}

	switch claims.TokenType {
	case TokenTypeDeviceAccessToken:
		if claims.Subject != clientID {
			iotLogger.Error("Client ID does not match the device ID (sub) of the device access token!")
			return failedAuthResponse(), ErrUnauthorized
		}
		return successDeviceAuthResponse(awsRegion, awsAccount, claims.Subject), nil

	case TokenTypeAppAccessToken:
		// Ensure that the token was issued only to the well known scheduler App ID
		// that is registered with the DSTS.
		if claims.Subject != schedulerAppID {
			iotLogger.Error("Only the scheduler app is authorized to connect to the IoT broker!")
			return failedAuthResponse(), ErrUnauthorized
		}

		// Ensure the client ID requested in the message has the scheduler App ID
		// as its prefix. The actual client ID may contain a short unique string
		// so that multiple scheduler pods can connect to the broker with causing
		// each other to be disconnected due to IoT core's client ID uniqueness
		// requirements.
		if !strings.HasPrefix(clientID, claims.Subject) {
			iotLogger.Error("Client ID does not start with the app ID (sub) of the app access token!")
			return failedAuthResponse(), ErrUnauthorized
		}
		return successAppAuthResponse(awsRegion, awsAccount, clientID), nil

	default:
		iotLogger.Error("Invalid token type specified in the access token!",
			zap.String("Token type:", claims.TokenType),
		)
		return failedAuthResponse(), ErrUnauthorized
	}
}

func successDeviceAuthResponse(awsRegion string, awsAccount string,
	deviceID string) events.IoTCoreCustomAuthorizerResponse {
	// Construct a successful response.
	response := events.IoTCoreCustomAuthorizerResponse{
		IsAuthenticated: true,
		PrincipalID:     strings.Replace(deviceID, "-", "", -1),
		PolicyDocuments: createIotPolicyDocumentForDevice(awsRegion,
			awsAccount, deviceID),
		RefreshAfterInSeconds:    defaultRefreshAfterSeconds,
		DisconnectAfterInSeconds: defaultDisconnectAfterSeconds,
	}
	iotLogger.Debug("Device token validated successfully. Sending IoT policy document!",
		zap.Any("Policy document:", response),
	)
	return response
}

func successAppAuthResponse(awsRegion string, awsAccount string,
	clientID string) events.IoTCoreCustomAuthorizerResponse {
	// Construct a successful response.
	response := events.IoTCoreCustomAuthorizerResponse{
		IsAuthenticated: true,
		PrincipalID:     strings.Replace(clientID, "-", "", -1),
		PolicyDocuments: createIotPolicyDocumentForScheduler(awsRegion,
			awsAccount, clientID),
		RefreshAfterInSeconds:    defaultRefreshAfterSeconds,
		DisconnectAfterInSeconds: defaultDisconnectAfterSeconds,
	}
	iotLogger.Debug("App token validated successfully. Sending IoT policy document!",
		zap.Any("Policy document:", response),
	)
	return response
}

func failedAuthResponse() events.IoTCoreCustomAuthorizerResponse {
	// Construct a successful response.
	response := events.IoTCoreCustomAuthorizerResponse{
		IsAuthenticated:          false,
		PrincipalID:              "",
		PolicyDocuments:          nil,
		RefreshAfterInSeconds:    300,
		DisconnectAfterInSeconds: 300,
	}
	iotLogger.Info("Device authentication failed. Sending failure response to IoT core!",
		zap.Any("Failure response:", response),
	)
	return response
}

func validateDstsAccessToken(accessToken string) (*DstsTokenClaims, error) {
	var claims DstsTokenClaims

	token, err := jwt.ParseWithClaims(accessToken, &claims, getSigningKey)
	if err != nil {
		return nil, err
	} else if !token.Valid {
		return nil, ErrInvalidToken
	}

	if token.Header["alg"] == nil {
		return nil, ErrInvalidTokenHeaderSigningAlg
	}

	if !strings.HasPrefix(claims.Issuer, dstsIssuerName) {
		return nil, ErrInvalidIssuerClaim
	}

	return &claims, nil
}

func main() {
	initLogger()
	defer shutdownLogger()

	dstsJwksUrl = os.Getenv(ENV_DSTS_JWKS_URL)
	if dstsJwksUrl == "" {
		iotLogger.Panic("Required DSTS JWKS URL environment variable is not specified!")
		return
	}

	// Get the token signing key from the DSTS.
	err := getJWKSSigningKey(dstsJwksUrl)
	if err != nil {
		iotLogger.Error("Failed to get the JWKS signing key!",
			zap.Error(err),
		)
		return
	}

	lambda.Start(IotDeviceAuthenticationHandler)
}
