package o2test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func request(uri, method string, params map[string]string) (*http.Response, error) {
	switch method {
	case "GET":
		uri, err := makeURL(uri, params)
		if err != nil {
			return nil, err
		}
		return http.Get(uri)
	case "POST":
		values := url.Values{}
		for key, value := range params {
			values.Set(key, value)
		}
		body := strings.NewReader(values.Encode())
		return http.Post(uri, "application/x-www-form-urlencoded", body)
	default:
		return nil, fmt.Errorf("method %s not supported", method)
	}
}

func makeURL(uri string, params map[string]string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	v, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}

	for key, value := range params {
		v.Set(key, value)
	}
	u.RawQuery = v.Encode()
	return u.String(), nil
}
