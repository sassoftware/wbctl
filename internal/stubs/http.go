// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// HitCounts maps an endpoint key to the number of times it was invoked in
// tests.
type HitCounts map[string]int

// Count returns the recorded hit count for the specified key.
func (h *HitCounts) Count(key string) int {
	return (*h)[key]
}

// CountPrefix returns the total hits for keys that begin with the provided
// prefix.
func (h *HitCounts) CountPrefix(prefix string) int {
	count := 0
	for key, value := range *h {
		if strings.HasPrefix(key, prefix) {
			count += value
		}
	}
	return count
}

// StubHTTPBodyOption mutates or formats a response body before it is written.
// Receives the request and the current body and must return the body to send.
type StubHTTPBodyOption func(
	r *http.Request,
	body string,
) string

// StubHTTPBodyReplaceHOST returns a StubHTTPBodyOption that replaces ${HOST}
// placeholders in the body with the request Host value.
func StubHTTPBodyReplaceHOST() StubHTTPBodyOption {
	return func(
		r *http.Request,
		body string,
	) string {
		return strings.ReplaceAll(body, "${HOST}", r.Host)
	}
}

// StubHTTPRequests describes expectations for incoming HTTP requests in the
// test stub.
type StubHTTPRequests struct {
	Method                string
	URL                   string
	ExpectedQueries       []string
	ExpectedHeaders       []map[string]string
	ExpectedRequestBodies []string
}

// StubHTTPResponses configures the responses returned by a stub endpoint across
// multiple calls.
type StubHTTPResponses struct {
	Bodies      []string
	Headers     []map[string][]string
	Statuses    []int
	BodyOptions [][]StubHTTPBodyOption
}

// StubHTTPEndpoints groups request expectations and response configurations for
// a single route.
type StubHTTPEndpoints struct {
	Request  StubHTTPRequests
	Response StubHTTPResponses
}

// StubHTTPRequest contains expectations for a single HTTP request in the stub
// harness.
type StubHTTPRequest struct {
	Method          string
	URL             string
	ExpectedQuery   string
	ExpectedHeaders map[string]string
	ExpectedBody    string
}

// StubHTTPResponse describes a single HTTP response returned by a stub
// endpoint.
type StubHTTPResponse struct {
	Body        string
	Headers     map[string][]string
	Status      int
	BodyOptions []StubHTTPBodyOption
}

// StubHTTPEndpoint pairs a single expected request with the response to return.
type StubHTTPEndpoint struct {
	Request  StubHTTPRequest
	Response StubHTTPResponse
}

// NewStubHTTPServer returns a new httptest.Server configured with the given
// endpoints.
func NewStubHTTPServer(
	t *testing.T,
	endpoints []StubHTTPEndpoint,
) (*httptest.Server, *HitCounts) {
	combinedEndpoints := map[string]StubHTTPEndpoints{}
	for _, endpoint := range endpoints {
		currentEndpoint, present := combinedEndpoints[endpoint.Request.Method+endpoint.Request.URL]
		if !present {
			currentEndpoint = StubHTTPEndpoints{
				Request: StubHTTPRequests{
					Method: endpoint.Request.Method,
					URL:    endpoint.Request.URL,
				},
			}
		}
		currentEndpoint.Request.ExpectedHeaders = append(
			currentEndpoint.Request.ExpectedHeaders,
			endpoint.Request.ExpectedHeaders,
		)
		currentEndpoint.Request.ExpectedRequestBodies = append(
			currentEndpoint.Request.ExpectedRequestBodies,
			endpoint.Request.ExpectedBody,
		)
		currentEndpoint.Request.ExpectedQueries = append(
			currentEndpoint.Request.ExpectedQueries,
			endpoint.Request.ExpectedQuery,
		)
		currentEndpoint.Response.Bodies = append(
			currentEndpoint.Response.Bodies,
			endpoint.Response.Body,
		)
		currentEndpoint.Response.Headers = append(
			currentEndpoint.Response.Headers,
			endpoint.Response.Headers,
		)
		currentEndpoint.Response.Statuses = append(
			currentEndpoint.Response.Statuses,
			endpoint.Response.Status,
		)
		currentEndpoint.Response.BodyOptions = append(
			currentEndpoint.Response.BodyOptions,
			endpoint.Response.BodyOptions,
		)
		combinedEndpoints[endpoint.Request.Method+endpoint.Request.URL] = currentEndpoint
	}
	return NewStubHTTPServerWithMultipleResponsesPerEndpoint(t, slices.Collect(maps.Values(combinedEndpoints)))
}

// NewStubHTTPServerWithMultipleResponsesPerEndpoint returns a new
// httptest.Server configured with the given endpoints.
// nolint:cyclop
func NewStubHTTPServerWithMultipleResponsesPerEndpoint(
	t *testing.T,
	endpoints []StubHTTPEndpoints,
) (*httptest.Server, *HitCounts) {
	hitCounts := HitCounts{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		for _, endpoint := range endpoints {
			if endpoint.Request.Method == "" {
				endpoint.Request.Method = http.MethodGet
			}
			if endpoint.Request.Method != r.Method || endpoint.Request.URL != r.URL.Path {
				continue
			}
			hitCountKey := endpoint.Request.Method + endpoint.Request.URL
			_, present := hitCounts[hitCountKey]
			if !present {
				hitCounts[hitCountKey] = 0
			}
			currentHit := hitCounts[hitCountKey]
			hitCounts[hitCountKey]++
			if len(endpoint.Request.ExpectedQueries) > currentHit {
				query, err := url.QueryUnescape(r.URL.RawQuery)
				assert.NoError(t, err, "malformed query")
				assert.Equal(
					t,
					endpoint.Request.ExpectedQueries[currentHit],
					query,
					"query strings are not equal for "+hitCountKey,
				)
			}
			if len(endpoint.Request.ExpectedHeaders) > currentHit {
				for key, values := range endpoint.Request.ExpectedHeaders[currentHit] {
					assert.Equal(
						t,
						values,
						r.Header.Get(key),
						"header values are not equal for "+hitCountKey,
					)
				}
			}
			if len(endpoint.Request.ExpectedRequestBodies) > currentHit {
				actualBody, err := io.ReadAll(r.Body)
				assert.NoError(t, err, "error reading request body")
				expected, actual := Sanitize(endpoint.Request.ExpectedRequestBodies[currentHit], string(actualBody), nil)
				var v any
				err = json.Unmarshal(actualBody, &v)
				if err != nil {
					assert.Equal(
						t,
						expected,
						actual,
						"request bodies are not equal for "+hitCountKey,
					)
				} else {
					assert.JSONEq(
						t,
						expected,
						actual,
						"request bodies are not equal for "+hitCountKey,
					)
				}
			}
			status := http.StatusOK
			if len(endpoint.Response.Statuses) > currentHit {
				status = endpoint.Response.Statuses[currentHit]
			}
			if len(endpoint.Response.Headers) > currentHit {
				for key, values := range endpoint.Response.Headers[currentHit] {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
			}
			w.WriteHeader(status)
			if len(endpoint.Response.Bodies) > currentHit {
				body := endpoint.Response.Bodies[currentHit]
				if len(endpoint.Response.BodyOptions) > currentHit {
					bodyOptions := endpoint.Response.BodyOptions[currentHit]
					for _, option := range bodyOptions {
						if option != nil {
							body = option(r, body)
						}
					}
				}
				_, err := fmt.Fprint(w, body)
				assert.NoError(t, err, "error writing response body")
			}
			return
		}
		w.WriteHeader(http.StatusNotFound)
		// XSS is not a concern for this test utility
		// nolint:gosec
		_, err := fmt.Fprintf(w, "Endpoint %s not found for method %s", r.URL.String(), r.Method)
		assert.NoError(t, err, "error writing error message")
	})
	srv := httptest.NewServer(mux)
	return srv, &hitCounts
}
