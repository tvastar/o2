package o2test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// CodeSuite implements the CodeGrant flow test suite.
type CodeSuite struct {
	ClientID, ClientSecret, RedirectURL string
	AuthorizationURL, TokenURL          string

	// The Setup function is called before each of the tests.
	// The return value from this is function is called when the test finishes.
	Setup func(t *testing.T) func()
}

// TestAll runs all the tests.
func (s CodeSuite) TestAll(t *testing.T) {
	t.Helper()
	t.Run("MissingClientID", s.TestMissingClientID)
	t.Run("InvalidClientID", s.TestInvalidClientID)
	t.Run("SuccessWithDefaultRedirectURI", s.TestSuccessWithDefaultRedirectURI)
}

// TestSuccessful redirect.
func (s CodeSuite) TestSuccessWithDefaultRedirectURI(t *testing.T) {
	defer setup(t, s.Setup)()
	params := map[string]string{
		"response_type": "code",
		"client_id":     s.ClientID,
		"state":         "a < b ? c : d",
		"extra":         "validate extra query params",
	}

	url, err := makeURL(s.AuthorizationURL, params)
	if err != nil {
		t.Fatal("invalid auth url", err)
	}
	client := &http.Client{CheckRedirect: urlChecker(s.RedirectURL)}
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal("http error", err)
	}
	if resp.StatusCode < 300 || resp.StatusCode >= 400 {
		t.Fatal("Not a redirect", resp.Status, resp.StatusCode)
	}
	loc, err := resp.Location()
	if err != nil {
		t.Fatal("Not a redirect", err)
	}
	q := loc.Query()
	if len(q["state"]) != 1 || q.Get("state") != params["state"] {
		t.Fatal("unexpected state", q.Get("state"))
	}
	q.Del("state")
	if len(q["code"]) != 1 {
		t.Fatal("no code", q)
	}
	q.Del("code")
	loc.RawQuery = q.Encode()
	if loc.String() != s.RedirectURL {
		t.Fatal("unexpected redirect url", loc.String())
	}
}

// TestMissingClientID tests proper response to a missing clientID.
func (s CodeSuite) TestMissingClientID(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		t.Run(method, func(t *testing.T) {
			defer setup(t, s.Setup)()
			params := map[string]string{
				"response_type": "code",
				"extra":         "validate extra query params",
			}
			resp, err := request(s.AuthorizationURL, method, params)
			if err != nil {
				t.Fatal("http error", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatal("Unexpected resp", resp.Status, resp.StatusCode)
			}
		})
	}
}

// TestInvalidClientID tests proper response to an invalid clientID.
func (s CodeSuite) TestInvalidClientID(t *testing.T) {
	for _, method := range []string{"GET", "POST"} {
		t.Run(method, func(t *testing.T) {
			defer setup(t, s.Setup)()
			params := map[string]string{
				"response_type": "code",
				"client_id":     "boo",
				"extra":         "validate extra query params",
			}
			resp, err := request(s.AuthorizationURL, method, params)
			if err != nil {
				t.Fatal("http error", err)
			}
			if resp.StatusCode != http.StatusBadRequest {
				t.Fatal("Unexpected resp", resp.Status, resp.StatusCode)
			}
		})
	}
}

func setup(t *testing.T, setupf func(t *testing.T) func()) func() {
	if setupf == nil {
		return func() {}
	}
	return setupf(t)
}

func urlChecker(s string) func(r *http.Request, via []*http.Request) error {
	return func(r *http.Request, via []*http.Request) error {
		expected, err := url.Parse(s)
		if err != nil {
			return err
		}
		actual := *r.URL

		// just check if both urls are the same if their
		// queries are removed
		expected.RawQuery = ""
		actual.RawQuery = ""
		if expected.String() == actual.String() {
			return http.ErrUseLastResponse
		}
		if len(via) >= 10 {
			return fmt.Errorf("too many redirects: %d", len(via))
		}
		return nil
	}
}
