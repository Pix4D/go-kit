// This file contains test utilities.

package github_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
)

// ghTestCfg contains the secrets needed to run integration tests against the
// GitHub Commit Status API.
type ghTestCfg struct {
	Token               string
	GhAppClientID       string
	GhAppInstallationID string
	GhAppPrivateKey     string
	Owner               string
	Repo                string
	SHA                 string
}

// fakeTestCfg is a fake test configuration that can be used in some tests that need
// configuration but don't really use any external service.
var fakeTestCfg = ghTestCfg{
	Token: "fakeToken",
	Owner: "fakeOwner",
	Repo:  "fakeRepo",
	SHA:   "0123456789012345678901234567890123456789",
}

// gitHubSecretsOrFail returns the secrets needed to run integration tests against the
// GitHub Commit Status API. If the secrets are missing, gitHubSecretsOrFail fails the test.
func gitHubSecretsOrFail(t *testing.T) ghTestCfg {
	t.Helper()

	return ghTestCfg{
		Token:               getEnvOrFail(t, "GOKIT_TEST_OAUTH_TOKEN"),
		GhAppClientID:       getEnvOrFail(t, "GOKIT_TEST_GH_APP_CLIENT_ID"),
		GhAppInstallationID: getEnvOrFail(t, "GOKIT_TEST_GH_APP_INSTALLATION_ID"),
		GhAppPrivateKey:     getEnvOrFail(t, "GOKIT_TEST_GH_APP_PRIVATE_KEY"),
		Owner:               getEnvOrFail(t, "GOKIT_TEST_REPO_OWNER"),
		Repo:                getEnvOrFail(t, "GOKIT_TEST_REPO_NAME"),
		SHA:                 getEnvOrFail(t, "GOKIT_TEST_COMMIT_SHA"),
	}
}

// getEnvOrFail returns the value of environment variable key. If key is missing,
// getEnvOrFail fails the test.
func getEnvOrFail(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if len(value) == 0 {
		t.Fatalf("Missing environment variable (see CONTRIBUTING): %s", key)
	}
	return value
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func generatePrivateKey(t *testing.T, bitSize int) *rsa.PrivateKey {
	t.Helper()

	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		t.Fatal("generating private key:", err)
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		t.Fatal("validating private key:", err)
	}

	return privateKey
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	return pem.EncodeToMemory(&privBlock)
}

// decodeJWT decodes the HTTP request authorization header with the given RSA key
// and returns the registered claims of the decoded token.
func decodeJWT(t *testing.T, r *http.Request, key *rsa.PrivateKey) *jwt.RegisteredClaims {
	t.Helper()

	token := strings.Fields(r.Header.Get("Authorization"))[1]
	tok, err := jwt.ParseWithClaims(token, &jwt.RegisteredClaims{},
		func(tk *jwt.Token) (any, error) {
			if tk.Header["alg"] != "RS256" {
				return nil, fmt.Errorf("unexpected signing method: %v, expected: %v",
					tk.Header["alg"], "RS256")
			}
			return &key.PublicKey, nil
		})
	if err != nil {
		t.Fatal("parsing JWT claims:", err)
	}

	return tok.Claims.(*jwt.RegisteredClaims)
}

// diff compares have and want line by line, and return difference if exists.
func diff[T any](have, want T) string {
	if diff := cmp.Diff(want, have); diff != "" {
		return fmt.Sprintf("--- want\n+++ have\n%s", diff)
	}
	return ""
}
