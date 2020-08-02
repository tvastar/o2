package o2_test

import (
	"github.com/tvastar/o2"
	"github.com/tvastar/o2/o2test"
	"net/http/httptest"
	"testing"
)

// TestCode tests the code grant flow.
func TestCode(t *testing.T) {
	clientID, clientSecret := "some_id", "some_secret"
	redirectURL := "http://localhost:7272/pqr?a=b"

	ts := &testServer{}
	server := httptest.NewServer(o2.Handler(ts, "/authorize", "/token"))
	defer server.Close()

	ts.register(clientID, clientSecret, "scope", redirectURL)
	suite := o2test.CodeSuite{
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		AuthorizationURL: server.URL + "/authorize",
		TokenURL:         server.URL + "/token",
		RedirectURL:      redirectURL,
	}

	suite.TestAll(t)
}
