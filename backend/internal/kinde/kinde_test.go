package kinde

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/devi/bookleaf/internal/platform/config"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, tokenExpiresIn int) (*httptest.Server, *int, *[]string) {
	t.Helper()

	tokenCalls := 0
	var deleteTokens []string

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, _ *http.Request) {
		tokenCalls++
		fmt.Fprintf(w, `{"access_token":"token-%d","expires_in":%d}`, tokenCalls, tokenExpiresIn)
	})
	mux.HandleFunc("/api/v1/user", func(w http.ResponseWriter, r *http.Request) {
		deleteTokens = append(deleteTokens, strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	return server, &tokenCalls, &deleteTokens
}

func newTestClient(server *httptest.Server) *Client {
	return NewClient(config.KindeConfig{
		IssuerURL:          server.URL,
		M2MClientID:        "client-id",
		M2MClientSecret:    "client-secret",
		M2MTokenURL:        server.URL + "/oauth2/token",
		ManagementAudience: server.URL + "/api",
	})
}

func TestClient_DeleteUser_CachesAndReusesToken(t *testing.T) {
	server, tokenCalls, deleteTokens := newTestServer(t, 3600)
	client := newTestClient(server)

	require.NoError(t, client.DeleteUser(context.Background(), "kp_abc123"))
	require.NoError(t, client.DeleteUser(context.Background(), "kp_abc123"))

	require.Equal(t, 1, *tokenCalls, "token should be fetched once and reused")
	require.Equal(t, []string{"token-1", "token-1"}, *deleteTokens)
}

func TestClient_DeleteUser_RefetchesExpiredToken(t *testing.T) {
	server, tokenCalls, deleteTokens := newTestServer(t, 0)
	client := newTestClient(server)

	require.NoError(t, client.DeleteUser(context.Background(), "kp_abc123"))
	require.NoError(t, client.DeleteUser(context.Background(), "kp_abc123"))

	require.Equal(t, 2, *tokenCalls, "an expired token should be refetched")
	require.Equal(t, []string{"token-1", "token-2"}, *deleteTokens)
}
