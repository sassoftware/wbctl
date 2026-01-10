// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/sassoftware/wbctl/internal/httputil"
	"github.com/stretchr/testify/require"
)

type contextKey string

const (
	testContextKey                = contextKey("test-key")
	httputilMakeURLInvalidHostErr = "parse \"http://[::1\": missing ']' in host"
	httputilUnsupportedTypeErr    = "json: unsupported type: chan int"
)

func TestAddBody(t *testing.T) {
	type bodyValue struct {
		Name string `json:"name"`
	}
	type unsupportedBodyValue struct {
		Channel chan int `json:"channel"`
	}

	tests := []struct {
		name         string
		body         any
		expectedBody string
		expectedErr  string
	}{
		{
			name:         "nil body returns same request and no error",
			body:         nil,
			expectedBody: "",
		},
		{
			name:         "normal body marshals JSON into request body",
			body:         bodyValue{Name: "alpha"},
			expectedBody: `{"name":"alpha"}`,
		},
		{
			name:        "marshal failure returns error",
			body:        unsupportedBodyValue{Channel: make(chan int)},
			expectedErr: httputilUnsupportedTypeErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, "https://example.com/resources", nil)
			require.NoError(t, err)

			actualReq, actualErr := httputil.AddBody(tt.body)(req)

			require.Same(t, req, actualReq)
			if tt.expectedErr != "" {
				require.Error(t, actualErr)
				require.EqualError(t, actualErr, tt.expectedErr)
				return
			}

			require.NoError(t, actualErr)
			if tt.body == nil {
				require.Nil(t, actualReq.Body)
				return
			}

			require.NotNil(t, actualReq.Body)
			content, readErr := io.ReadAll(actualReq.Body)
			require.NoError(t, readErr)
			require.Equal(t, tt.expectedBody, string(content))
		})
	}
}

func TestAddHeaders(t *testing.T) {
	tests := []struct {
		name            string
		headers         map[string]string
		expectedHeaders map[string]string
	}{
		{
			name: "adds all provided headers to request",
			headers: map[string]string{
				"Authorization": "Bearer abc",
				"X-Trace-ID":    "trace-123",
			},
			expectedHeaders: map[string]string{
				"Authorization": "Bearer abc",
				"X-Trace-Id":    "trace-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/resources", nil)
			require.NoError(t, err)

			actualReq := httputil.AddHeaders(tt.headers, req)

			require.Same(t, req, actualReq)
			for key, expectedValue := range tt.expectedHeaders {
				require.Equal(t, expectedValue, actualReq.Header.Get(key))
			}
		})
	}
}

func TestMakeGetRequestContextFunc(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       *url.URL
		expectedMethod string
		expectedURL    string
	}{
		{
			name:           "returned function uses supplied context and GET method",
			inputURL:       mustParseURL(t, "https://example.com/resources?x=1"),
			expectedMethod: http.MethodGet,
			expectedURL:    "https://example.com/resources?x=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), testContextKey, "ctx-value")

			makeRequest := httputil.MakeGetRequestContextFunc(ctx)
			req, err := makeRequest(tt.inputURL)

			require.NoError(t, err)
			require.Equal(t, tt.expectedMethod, req.Method)
			require.Equal(t, tt.expectedURL, req.URL.String())
			require.Nil(t, req.Body)
			require.Equal(t, "ctx-value", req.Context().Value(testContextKey))
		})
	}
}

func TestMakeHeadRequest(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       *url.URL
		expectedMethod string
		expectedURL    string
	}{
		{
			name:           "creates HEAD request with URL and no body",
			inputURL:       mustParseURL(t, "https://example.com/resources"),
			expectedMethod: http.MethodHead,
			expectedURL:    "https://example.com/resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := httputil.MakeHeadRequest(context.Background())(tt.inputURL)

			require.NoError(t, err)
			require.Equal(t, tt.expectedMethod, req.Method)
			require.Equal(t, tt.expectedURL, req.URL.String())
			require.Nil(t, req.Body)
		})
	}
}

func TestMakePostRequest(t *testing.T) {
	tests := []struct {
		name           string
		inputURL       *url.URL
		expectedMethod string
		expectedURL    string
	}{
		{
			name:           "creates POST request with URL and no body",
			inputURL:       mustParseURL(t, "https://example.com/resources"),
			expectedMethod: http.MethodPost,
			expectedURL:    "https://example.com/resources",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := httputil.MakePostRequestContextFunc(context.Background())(tt.inputURL)

			require.NoError(t, err)
			require.Equal(t, tt.expectedMethod, req.Method)
			require.Equal(t, tt.expectedURL, req.URL.String())
			require.Nil(t, req.Body)
		})
	}
}

func TestMakeURL(t *testing.T) {
	tests := []struct {
		name             string
		endpoint         string
		path             string
		query            url.Values
		expectedRawQuery string
		expectedURL      string
		expectedErr      string
	}{
		{
			name:             "query absent keeps raw query empty",
			endpoint:         "https://example.com/base",
			path:             "/resource",
			query:            nil,
			expectedRawQuery: "",
			expectedURL:      "https://example.com/base/resource",
		},
		{
			name:     "query present sets encoded query",
			endpoint: "https://example.com/base",
			path:     "/resource",
			query: url.Values{
				"name":  []string{"value"},
				"token": []string{"abc"},
			},
			expectedRawQuery: "name=value&token=abc",
			expectedURL:      "https://example.com/base/resource?name=value&token=abc",
		},
		{
			name:        "invalid endpoint returns error result",
			endpoint:    "http://[::1",
			path:        "/resource",
			expectedErr: httputilMakeURLInvalidHostErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := httputil.MakeURL(tt.endpoint, tt.path, tt.query)

			if tt.expectedErr != "" {
				require.Error(t, actual.Error())
				require.EqualError(t, actual.Error(), tt.expectedErr)
				return
			}

			require.NoError(t, actual.Error())
			actualURL := actual.MustGet()
			require.Equal(t, tt.expectedRawQuery, actualURL.RawQuery)
			require.Equal(t, tt.expectedURL, actualURL.String())
		})
	}
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	return parsedURL
}
