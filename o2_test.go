package o2_test

import (
	"github.com/tvastar/o2"
	"github.com/tvastar/o2/o2mem"
	"github.com/tvastar/o2/o2test"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestCode tests the code grant flow.
func TestCode(t *testing.T) {
	clientID, clientSecret := "some_id", "some_secret"
	redirectURL := "http://localhost:7272/pqr?a=b"

	mem := &o2mem.Server{
		UserID: func(r *http.Request) string {
			return "foo"
		},
	}
	mem.AddClient(clientID, clientSecret, "scope1 scope2", redirectURL)
	mem.AuthorizeClient(clientID, "foo", "scope2")

	server := httptest.NewServer(o2.Handler(mem, "/authorize", "/token"))
	defer server.Close()

	suite := o2test.CodeSuite{
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		AuthorizationURL: server.URL + "/authorize",
		TokenURL:         server.URL + "/token",
		RedirectURL:      redirectURL,
	}

	suite.TestAll(t)
}
