// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth0admin_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/services/auth0admin"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/stretchr/testify/require"
)

const (
	auth0AdminGetSuccessBody  = `{"value":"ok"}`
	auth0AdminPostSuccessBody = `{"ok":"yes"}`
)

type auth0AdminGetResponse struct {
	Value string `json:"value"`
}

type auth0AdminPostResponse struct {
	OK string `json:"ok"`
}

func TestNewAuth0AdminService(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
	}{
		{name: "valid endpoint", endpoint: "http://example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := auth0admin.NewAuth0AdminService(tt.endpoint, "access-token")
			require.NotNil(t, svc)
		})
	}
}

func TestGetAccessToken(t *testing.T) {
	svc := auth0admin.NewAuth0AdminService("http://example.com", "access-token")
	require.Equal(t, "access-token", auth0admin.GetAccessToken(svc))
}

func TestGet(t *testing.T) {
	type expected struct {
		err    string
		result auth0AdminGetResponse
	}
	type testCase struct {
		name          string
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name: "get get success",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				auth0AdminGetSuccessBody,
			),
			expected: expected{
				result: auth0AdminGetResponse{Value: "ok"},
			},
		},
		{
			name: "get get invalid json",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				`[`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* unexpected EOF",
			},
		},
		{
			name: "get get non-200",
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* 401 unauthorized",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := auth0admin.NewAuth0AdminService(srv.URL, "access-token")

			actual := auth0admin.Get[auth0AdminGetResponse](
				context.Background(),
				svc,
				"/resource",
				map[string][]string{"id": {"id-a"}},
			)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.result, actual.MustGet())
		})
	}
}

func TestGetRaw(t *testing.T) {
	type request struct {
		rawURL string
	}
	type expected struct {
		err    string
		result string
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "get raw success",
			request: request{rawURL: "/resource?id=id-a"},
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusOK,
				`raw-body`,
			),
			expected: expected{
				result: "raw-body",
			},
		},
		{
			name:    "get raw non-200",
			request: request{rawURL: "/resource?id=id-a"},
			mockEndpoints: getEndpoints(
				"/resource",
				"id=id-a",
				http.StatusForbidden,
				`forbidden`,
			),
			expected: expected{
				err: "GET ${URL}/resource?id=id-a failed\n Caused by:\n\t* 403 forbidden",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := auth0admin.NewAuth0AdminService(srv.URL, "access-token")

			rawURL := tt.request.rawURL
			if len(rawURL) > 0 && rawURL[0] == '/' {
				rawURL = srv.URL + rawURL
			}
			actual := auth0admin.GetRaw(context.Background(), svc, rawURL)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.result, actual.MustGet())
		})
	}
}

func TestPost(t *testing.T) {
	type request struct {
		requestBody map[string]string
	}
	type expected struct {
		err   string
		value auth0AdminPostResponse
	}
	type testCase struct {
		name          string
		request       request
		mockEndpoints []stubs.StubHTTPEndpoint
		expected      expected
	}
	tests := []testCase{
		{
			name:    "success",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusOK,
				auth0AdminPostSuccessBody,
			),
			expected: expected{value: auth0AdminPostResponse{OK: "yes"}},
		},
		{
			name:    "invalid json",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusOK,
				`[`,
			),
			expected: expected{err: "POST ${URL}/resource failed\n Caused by:\n\t* unexpected EOF"},
		},
		{
			name:    "non-200",
			request: request{requestBody: map[string]string{"id": "x"}},
			mockEndpoints: postEndpoints(
				"/resource",
				`{"id":"x"}`,
				http.StatusUnauthorized,
				`unauthorized`,
			),
			expected: expected{err: "POST ${URL}/resource failed\n Caused by:\n\t* 401 unauthorized"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := stubs.NewStubHTTPServer(t, tt.mockEndpoints)
			defer srv.Close()
			svc := auth0admin.NewAuth0AdminService(srv.URL, "access-token")

			actual := auth0admin.Post[auth0AdminPostResponse](context.Background(), svc, "/resource", tt.request.requestBody)
			if tt.expected.err != "" {
				require.EqualError(t, actual.Error(), strings.ReplaceAll(tt.expected.err, "${URL}", srv.URL))
				return
			}
			require.NoError(t, actual.Error())
			require.Equal(t, tt.expected.value, actual.MustGet())
		})
	}
}

func getEndpoints(
	url string, // nolint: unparam
	expectedQuery string, // nolint: unparam
	status int,
	body string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{{
		Request: stubs.StubHTTPRequest{
			Method:          http.MethodGet,
			URL:             url,
			ExpectedHeaders: authHeaders,
			ExpectedQuery:   expectedQuery,
		},
		Response: stubs.StubHTTPResponse{
			Status: status,
			Body:   body,
		},
	}}
}

func postEndpoints(
	url string,
	expectedBody string,
	status int,
	body string,
) []stubs.StubHTTPEndpoint {
	return []stubs.StubHTTPEndpoint{{
		Request: stubs.StubHTTPRequest{
			Method:          http.MethodPost,
			URL:             url,
			ExpectedHeaders: authHeaders,
			ExpectedBody:    expectedBody,
		},
		Response: stubs.StubHTTPResponse{
			Status: status,
			Body:   body,
		},
	}}
}
