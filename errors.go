// package github.com/HPInc/krypton-iot-authorizer
// Author: Mahesh Unnikrishnan
// Component: Krypton AWS IoT Authorizer Lambda
// (C) HP Development Company, LP
package main

import (
	"errors"
	"net/http"
)

var (
	ErrNoLambdaContext              = errors.New("failed to retrieve context from lambda function")
	ErrMissingAssets                = errors.New("required assets are missing to create a public key")
	ErrInvalidToken                 = errors.New("invalid token provided")
	ErrInvalidTokenHeaderKid        = errors.New("invalid token signing kid specified")
	ErrInvalidTokenHeaderSigningAlg = errors.New("invalid token signing algorithm specified")
	ErrInvalidIssuerClaim           = errors.New("specified token contains an invalid issuer claim")
	ErrInvalidAudienceClaim         = errors.New("specified token contains an invalid audience claim")
	ErrOverflowDetected             = errors.New("integer overflow detected while parsing exponent from the JWKS")
	ErrUnauthorized                 = errors.New(http.StatusText(http.StatusUnauthorized))
	ErrBadRequest                   = errors.New(http.StatusText(http.StatusBadRequest))
)
