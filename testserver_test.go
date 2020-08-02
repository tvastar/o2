package o2_test

import (
	"encoding/json"
	"fmt"
	"github.com/tvastar/o2"
	"net/http"
)

type testServer struct {
	clients map[string]o2.Client
}

func (ts *testServer) register(clientID, clientSecret, scope, redirectURL string) {
	if ts.clients == nil {
		ts.clients = map[string]o2.Client{}
	}
	ts.clients[clientID] = &client{clientID, clientSecret, scope, redirectURL, nil}
}

func (ts *testServer) Client(r *http.Request, clientID string) (o2.Client, error) {
	c, _ := ts.clients[clientID]
	return c, nil
}

type client struct {
	clientID, clientSecret, scope, redirectURL string
	codes                                      map[string]string
}

func (c *client) RedirectURI(uri string) (string, error) {
	return c.redirectURL, nil
}

func (c *client) Authorized(w http.ResponseWriter, r *http.Request, scope string) *bool {
	authorized := true
	return &authorized
}

func (c *client) Code(r *http.Request, state interface{}) (string, error) {
	if c.codes == nil {
		c.codes = map[string]string{}
	}
	code := fmt.Sprintf("%d > 5 ? a", len(c.codes)+1)
	serialized, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	c.codes[code] = string(serialized)
	return code, nil
}
