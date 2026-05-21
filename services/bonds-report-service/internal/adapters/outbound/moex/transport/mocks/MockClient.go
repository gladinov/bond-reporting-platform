package mocks

import "net/http"

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewMockClient(fn RoundTripFunc) http.Client {
	return http.Client{Transport: fn}
}
