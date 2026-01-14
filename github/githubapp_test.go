package github_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Pix4D/go-kit/github"
)

func TestGenerateInstallationToken(t *testing.T) {
	clientID := "abcd1234"
	installationID := 12345

	privateKey := generatePrivateKey(t, 2048)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "wrong HTTP method")
			return
		}

		claims := decodeJWT(t, r, privateKey)
		if claims.Issuer != clientID {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "unauthorized: wrong JWT token")
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, `{"token": "dummy_installation_token"}`)
	}

	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	gotToken, err := github.GenerateInstallationToken(
		ctx,
		ts.Client(),
		ts.URL,
		github.GitHubApp{
			ClientId:       clientID,
			InstallationId: installationID,
			PrivateKey:     string(encodePrivateKeyToPEM(privateKey)),
		},
	)
	if err != nil {
		t.Fatalf("%s\nhave: %v\nwant: %v", "token: error", err, "<no error>")
	}
	if have, want := gotToken, "dummy_installation_token"; have != want {
		t.Fatalf("%s\nhave: %v\nwant: %v", "token", have, want)
	}
}

func TestGitHubAppIsZero(t *testing.T) {
	type testCase struct {
		name string
		app  github.GitHubApp
		want bool
	}

	run := func(t *testing.T, tc testCase) {
		if have, want := tc.app.IsZero(), tc.want; have != want {
			t.Fatalf("%s\nhave: %v\nwant: %v", "IsZero", have, want)
		}
	}

	testCases := []testCase{
		{
			name: "empty app",
			app:  github.GitHubApp{},
			want: true,
		},
		{
			name: "one field set: client-id",
			app:  github.GitHubApp{ClientId: "client-id"},
			want: false,
		},
		{
			name: "all fields set",
			app: github.GitHubApp{
				ClientId:       "client-id",
				InstallationId: 12345,
				PrivateKey:     "dummy-private-key",
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { run(t, tc) })
	}
}
