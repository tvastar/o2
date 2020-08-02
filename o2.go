// Package o2 implements a simple Oauth2 server.
//
// An example OAuth2 server can be created via:
//
//     server := ... server implementation ...
//     handler := o2.Handler(server, "/authorize", "/token")
//     http.ListenAndServe(":80", handler)
//
// The server implementation provides access to the underlying
// storage.
package o2

import (
	"net/http"
	"net/url"
)

// Server is the interface an oauth2 server should implement.
type Server interface {
	// Client fetches the client by clientID
	//
	// If the client is not found, a nil client is returned. An
	// error return indicates an internal error.
	Client(r *http.Request, clientID string) (Client, error)
}

// Client is the interface than an oauth2 server should implement for
// operating on a clientID.
type Client interface {
	// Authorized checks if a client is authorized.
	//
	// If the authorization check can be answered immediately, the
	// return value is a non-nil bool (authorized or not).
	//
	// If the check requires a HTML form or some other response,
	// the Authorized call should render that and return a nil to
	// indicate it has taken care of it.
	Authorized(w http.ResponseWriter, r *http.Request, scope string) *bool

	// RedirectURI returns the URI to redirect to.
	//
	// The uri arg is the optional authorization request
	// parameteer  "redirect_uri".  If the URI is present and does
	// not match the configured value, an error should be
	// returned.
	RedirectURI(r *http.Request, uri string) (string, error)

	// Code fetches an authorization code grant.
	//
	// The provided state should be encoded in the return code (or
	// stored with the code as a key).
	Code(r *http.Request, state interface{}) (string, error)
}

// Params is used to pass response params.
type Params map[string]string

// Handler creates a http Handler for the authorize/token paths.
func Handler(s Server, authorizePath, tokenPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case authorizePath:
			Authorize(s).ServeHTTP(w, r)
		default:
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})
}

// Authorize implements just the authorize endpoint handler.
func Authorize(s Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch rtype := r.FormValue("response_type"); rtype {
		case "code":
			uri, params := authorizeCode(s, w, r)
			if uri != "" || params != nil {
				Redirect(w, uri, params)
			}
		default:
			client, uri := clientRedirectURI(s, w, r)
			if client != nil && uri != "" {
				Redirect(w, uri, Params{
					"state": r.FormValue("state"),
					"error": "unsupported_response_type",
				})
			}
		}
	})
}

// Redirect redirects the client.
//
// The redirect uri can be nil or invalid, in which case the current
// reques fails with an appropriate http status code derived from
// Params["error"] (or using a default internal server error message).
//
// If the redirect uri is valid, any provided query parameters are
// tacked onto that URI.
func Redirect(w http.ResponseWriter, uri string, params Params) {
	status, description := redirect(w, uri, params)
	if status == "" {
		return
	}

	if params["error"] != "" {
		status, description = params["error"], params["error_description"]
	}

	switch status {
	case "server_error":
		http.Error(w, description, http.StatusInternalServerError)
	default:
		http.Error(w, description, http.StatusBadRequest)
	}
}

func authorizeCode(s Server, w http.ResponseWriter, r *http.Request) (string, Params) {
	state := r.FormValue("state")
	scope := r.FormValue("sccope")
	challenge := r.FormValue("challenge")

	client, uri := clientRedirectURI(s, w, r)
	if client == nil {
		return "", nil
	}

	auth := client.Authorized(w, r, scope)
	if auth == nil {
		return "", nil
	}

	if !*auth {
		return uri, Params{"state": state, "error": "unauthorized"}
	}

	code, err := client.Code(r, authCodeState{state, challenge})
	if err != nil {
		return uri, Params{"state": state, "error": "server_error", "error_description": err.Error()}
	}

	return uri, Params{"state": state, "code": code}
}

func clientRedirectURI(s Server, w http.ResponseWriter, r *http.Request) (Client, string) {
	client, err := s.Client(r, r.FormValue("client_id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, ""
	}

	if client == nil {
		http.Error(w, "invalid client_id", http.StatusBadRequest)
		return nil, ""
	}

	uri, err := client.RedirectURI(r, r.FormValue("redirect_uri"))
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return nil, ""
	}

	return client, uri
}

func redirect(w http.ResponseWriter, uri string, params Params) (string, string) {
	u, err := url.Parse(uri)
	if err != nil {
		return "server_error", err.Error()
	}

	if u.Scheme == "" && u.Host == "" {
		return "server_error", "invalid request_uri"
	}

	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "invalid_request", err.Error()
	}

	for key, value := range params {
		if value != "" {
			q.Set(key, value)
		}
	}

	u.RawQuery = q.Encode()
	http.Redirect(w, &http.Request{}, u.String(), http.StatusFound)
	return "", ""
}

type authCodeState struct {
	State     string `json:"state,omitempty"`
	Challenge string `json:"code_challenge,omitempty"`
}
