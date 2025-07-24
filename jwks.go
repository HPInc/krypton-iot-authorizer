package main

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

const (
	// ktyRSA is the key type (kty) in the JWT header for RSA.
	ktyRSA       = "RSA"
	rsaPublicKey = "RSA PUBLIC KEY"

	httpRequestTimeout = time.Second * 3
)

var signingKeyTable map[string]*rsa.PublicKey

// jsonWebKey represents a JSON Web Key inside a JWKS.
type jsonWebKey struct {
	Curve    string `json:"crv"`
	Exponent string `json:"e"`
	K        string `json:"k"`
	ID       string `json:"kid"`
	Modulus  string `json:"n"`
	Type     string `json:"kty"`
	Use      string `json:"use"`
	X        string `json:"x"`
	Y        string `json:"y"`
}

// rawJWKS represents a JWKS in JSON format.
type rawJWKS struct {
	Keys []*jsonWebKey `json:"keys"`
}

func getSigningKey(token *jwt.Token) (interface{}, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, ErrInvalidTokenHeaderKid
	}

	// Check if a signing key corresponding to the kid was found in the
	// signing key table.
	pubKey, ok := signingKeyTable[kid]
	if !ok {
		// Key with this kid was not found - fetch the JWKS keys from the
		// DSTS to check if this is a new signing key.
		err := getJWKSSigningKey(dstsJwksUrl)
		if err != nil {
			iotLogger.Error("Failed to get JWKS signing keys from DSTS!")
			return nil, err
		}

		pubKey, ok = signingKeyTable[kid]
		if !ok {
			return nil, fmt.Errorf("no public key to validate kid: %s", kid)
		}
	}

	return pubKey, nil
}

func getJWKSSigningKey(url string) error {
	var rawKS rawJWKS

	// Fetch the
	jwksBytes, err := getKeysFromServer(url)
	if err != nil {
		iotLogger.Error("Error fetching keys.",
			zap.String("url:", url),
			zap.Error(err))
		return err
	}

	err = json.Unmarshal(jwksBytes, &rawKS)
	if err != nil {
		iotLogger.Error("Failed to JSON unmarshal JWKS response!",
			zap.Error(err),
		)
		return err
	}

	signingKeyTable = make(map[string]*rsa.PublicKey, len(rawKS.Keys))
	for _, key := range rawKS.Keys {
		switch keyType := key.Type; keyType {
		case ktyRSA:
			publicKey, err := parseRSASigningKey(key)
			if err != nil {
				iotLogger.Error("Error parsing signing key",
					zap.String("type:", key.Type),
					zap.String("kid:", key.ID),
					zap.Error(err))
				return err
			}

			signingKeyTable[key.ID] = publicKey

		default:
			continue
		}
	}

	return nil
}

// Retrieve the token signing keys in JWKS format from the DSTS JWKS endpoint.
func getKeysFromServer(keysUrl string) (keys []byte, err error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), httpRequestTimeout)
	defer cancelFunc()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, keysUrl, nil)
	if err != nil {
		iotLogger.Error("Failed to create HTTP request to the DSTS JWKS endpoint!",
			zap.String("JWKS URL:", keysUrl),
			zap.Error(err),
		)
		return nil, err
	}

	req.Header.Set(headerUserAgent, authorizerUserAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		iotLogger.Error("Failed to retrieve JWKS signing keys from DSTS!",
			zap.String("JWKS URL:", keysUrl),
			zap.Error(err),
		)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		iotLogger.Error("HTTP request to get JWKS signing keys failed!",
			zap.String("JWKS URL:", keysUrl),
			zap.Int("status", resp.StatusCode))
		return nil, err
	}
	return io.ReadAll(resp.Body)
}

// parseRSASigningKey parses a jsonWebKey and turns it into an RSA public key.
func parseRSASigningKey(j *jsonWebKey) (publicKey *rsa.PublicKey, err error) {
	if j.Exponent == "" || j.Modulus == "" {
		return nil, ErrMissingAssets
	}

	// Decode the exponent from Base64.
	// According to RFC 7518, this is a Base64 URL unsigned integer.
	// https://tools.ietf.org/html/rfc7518#section-6.3
	exponent, err := base64urlTrailingPadding(j.Exponent)
	if err != nil {
		return nil, err
	}

	modulus, err := base64urlTrailingPadding(j.Modulus)
	if err != nil {
		return nil, err
	}

	// Turn the exponent into an integer.
	// According to RFC 7517, these numbers are in big-endian format.
	// https://tools.ietf.org/html/rfc7517#appendix-A.1
	exp := big.NewInt(0).SetBytes(exponent).Uint64()
	if exp > math.MaxInt {
		return nil, ErrOverflowDetected
	}
	return &rsa.PublicKey{
		E: int(exp),
		N: big.NewInt(0).SetBytes(modulus),
	}, nil
}

// base64urlTrailingPadding removes trailing padding before decoding a string from base64url. Some non-RFC compliant
// JWKS contain padding at the end values for base64url encoded public keys.
//
// Trailing padding is required to be removed from base64url encoded keys.
// RFC 7517 defines base64url the same as RFC 7515 Section 2:
// https://datatracker.ietf.org/doc/html/rfc7517#section-1.1
// https://datatracker.ietf.org/doc/html/rfc7515#section-2
func base64urlTrailingPadding(s string) ([]byte, error) {
	s = strings.TrimRight(s, "=")
	return base64.RawURLEncoding.DecodeString(s)
}
