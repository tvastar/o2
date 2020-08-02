// Package o2mem implements an in memory version of the  o2.Server
// interface.
package o2mem

import (
	"encoding/json"
	"fmt"
	"github.com/tvastar/o2"
	"net/http"
	"strings"
	"sync"
)

// Server implements o2.Server using an in memory store.
//
// To create new client credentials, use AddClient.
// To authorize a specific ClientID, use AuthorizeClient.
type Server struct {
	UserID  func(r *http.Request) string
	clients map[string]*client
	mu      sync.Mutex
}

func (s *Server) AddClient(id, secret, scope, redirectURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clients == nil {
		s.clients = map[string]*client{}
	}
	s.clients[id] = &client{userID: s.UserID, id: id, secret: secret, redirectURL: redirectURL}
}

func (s *Server) AuthorizeClient(clientID, userID, scope string) {
	s.mu.Lock()
	c := s.clients[clientID]
	s.mu.Unlock()
	c.authorize(userID, scope)
}

func (s *Server) Client(r *http.Request, clientID string) (o2.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if c, ok := s.clients[clientID]; ok {
		return c, nil
	}
	return nil, nil
}

type client struct {
	userID func(r *http.Request) string

	id, secret, scope, redirectURL string

	codes      map[string]string
	codesCount int
	authorized map[string]string

	mu sync.Mutex
}

func (c *client) RedirectURI(r *http.Request, uri string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if uri == c.redirectURL || uri == "" {
		return c.redirectURL, nil
	}
	return "", fmt.Errorf("invalid redirect_uri")
}

func (c *client) Authorized(w http.ResponseWriter, r *http.Request, scope string) *bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	userID := c.userID(r)
	// TODO: if userID is not present, authenticate!
	// TODO: if a prior authorization isn't present, throw some UX
	// for it.
	authorized := scope == ""
	for _, s := range strings.Split(c.authorized[userID], " ") {
		if s == scope {
			authorized = true
			break
		}
	}
	return &authorized
}

func (c *client) authorize(userID, scope string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.authorized == nil {
		c.authorized = map[string]string{}
	}
	c.authorized[userID] = scope
}

func (c *client) Code(r *http.Request, state interface{}) (string, error) {
	serialized, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.codes == nil {
		c.codes = map[string]string{}
	}
	code := fmt.Sprintf("code%d", c.codesCount)
	c.codesCount++
	c.codes[code] = string(serialized)
	return code, nil
}
