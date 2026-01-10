// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPServerPost(t *testing.T) {
	srv, hitCounts := stubs.NewStubHTTPServerWithMultipleResponsesPerEndpoint(t, []stubs.StubHTTPEndpoints{
		{
			Request: stubs.StubHTTPRequests{
				Method: http.MethodPost,
				URL:    "/resource",
				ExpectedHeaders: []map[string]string{
					{"Content-Type": "application/json"},
					{"Content-Type": "application/json"},
					{"Content-Type": "application/json"},
				},
				ExpectedRequestBodies: []string{`{"id": "a"}`, `{"id": "b"}`, `{"id": "c"}`},
			},
			Response: stubs.StubHTTPResponses{
				Bodies:   []string{`{"id": "a"}`, `{"id": "b"}`, ``},
				Statuses: []int{http.StatusOK, http.StatusCreated, http.StatusBadRequest},
			},
		},
	})
	defer srv.Close()
	// First request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/resource",
		bytes.NewBufferString(`{"id": "a"}`),
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp := httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "a"}`, resp.MustGet())
	require.Equal(t, 1, (*hitCounts)["POST/resource"])
	require.Equal(t, 1, hitCounts.Count("POST/resource"))
	// second request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/resource",
		bytes.NewBufferString(`{"id": "b"}`),
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "b"}`, resp.MustGet())
	require.Equal(t, 2, (*hitCounts)["POST/resource"])
	require.Equal(t, 2, hitCounts.Count("POST/resource"))
	// third request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/resource",
		bytes.NewBufferString(`{"id": "c"}`),
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	errMessage := strings.ReplaceAll("POST ${URL}/resource failed\n Caused by:\n\t* 400 ", "${URL}", srv.URL)
	require.EqualError(t, resp.Error(), errMessage)
	require.Equal(t, 3, (*hitCounts)["POST/resource"])
	require.Equal(t, 3, hitCounts.Count("POST/resource"))
	// fourth request (beyond the bounds, should respond with defaults)
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodPost,
		srv.URL+"/resource",
		bytes.NewBufferString(`{"id": "d"}`),
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.Equal(t, 4, (*hitCounts)["POST/resource"])
	require.Equal(t, 4, hitCounts.Count("POST/resource"))
	require.Equal(t, 4, hitCounts.CountPrefix("POST/re"))
}

func TestNewHTTPServerGetMultiple(t *testing.T) {
	srv, hitCounts := stubs.NewStubHTTPServerWithMultipleResponsesPerEndpoint(t, []stubs.StubHTTPEndpoints{
		{
			Request: stubs.StubHTTPRequests{URL: "/resource",
				ExpectedHeaders: []map[string]string{
					{"Content-Type": "application/json"},
					{"Content-Type": "application/json"},
					{"Content-Type": "application/json"},
				},
				ExpectedQueries: []string{"id=a", "id=b", "id=c"},
			},
			Response: stubs.StubHTTPResponses{
				Bodies:      []string{`{"id": "a"}`, `{"id": "b"}`, `${HOST}`},
				Statuses:    []int{http.StatusOK, http.StatusOK, http.StatusBadRequest},
				BodyOptions: [][]stubs.StubHTTPBodyOption{{}, {}, {stubs.StubHTTPBodyReplaceHOST()}},
			},
		},
	})
	defer srv.Close()
	// first request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=a",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp := httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "a"}`, resp.MustGet())
	require.Equal(t, 1, (*hitCounts)["GET/resource"])
	// second request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=b",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "b"}`, resp.MustGet())
	require.Equal(t, 2, (*hitCounts)["GET/resource"])
	// third request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=c",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	errMessage := strings.ReplaceAll("GET ${URL}/resource?id=c failed\n Caused by:\n\t* 400 ${HOST}", "${URL}", srv.URL)
	errMessage = strings.ReplaceAll(errMessage, "${HOST}", getHostAndPort(t, srv.URL))
	require.EqualError(t, resp.Error(), errMessage)
	require.Equal(t, 3, (*hitCounts)["GET/resource"])
	// fourth request (beyond the bounds, should respond with defaults)
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=d",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
}

func TestNewHTTPServerGet(t *testing.T) {
	srv, hitCounts := stubs.NewStubHTTPServer(t, []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				URL: "/resource",
				ExpectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
				ExpectedQuery: "id=a",
			},
			Response: stubs.StubHTTPResponse{
				Body:   `{"id": "a"}`,
				Status: http.StatusOK,
			},
		},
		{
			Request: stubs.StubHTTPRequest{
				URL: "/resource",
				ExpectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
				ExpectedQuery: "id=b",
			},
			Response: stubs.StubHTTPResponse{
				Body:   `{"id": "b"}`,
				Status: http.StatusOK,
			},
		}, {
			Request: stubs.StubHTTPRequest{
				URL: "/resource",
				ExpectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
				ExpectedQuery: "id=c",
			},
			Response: stubs.StubHTTPResponse{
				Body:        `${HOST}`,
				Status:      http.StatusBadRequest,
				BodyOptions: []stubs.StubHTTPBodyOption{stubs.StubHTTPBodyReplaceHOST()},
			},
		},
	})
	defer srv.Close()
	// first request
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=a",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp := httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "a"}`, resp.MustGet())
	require.Equal(t, 1, (*hitCounts)["GET/resource"])
	// second request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=b",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
	require.JSONEq(t, `{"id": "b"}`, resp.MustGet())
	require.Equal(t, 2, (*hitCounts)["GET/resource"])
	// third request
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=c",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	errMessage := strings.ReplaceAll("GET ${URL}/resource?id=c failed\n Caused by:\n\t* 400 ${HOST}", "${URL}", srv.URL)
	errMessage = strings.ReplaceAll(errMessage, "${HOST}", getHostAndPort(t, srv.URL))
	require.EqualError(t, resp.Error(), errMessage)
	require.Equal(t, 3, (*hitCounts)["GET/resource"])
	// fourth request (beyond the bounds, should respond with defaults)
	req, err = http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=d",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	resp = httputils.DoAsString(http.DefaultClient, req)
	require.NoError(t, resp.Error())
}

func TestNewHTTPServerGetWithResopnseHeaders(t *testing.T) {
	srv, _ := stubs.NewStubHTTPServer(t, []stubs.StubHTTPEndpoint{
		{
			Request: stubs.StubHTTPRequest{
				URL: "/resource",
				ExpectedHeaders: map[string]string{
					"Content-Type": "application/json",
				},
				ExpectedQuery: "id=a",
			},
			Response: stubs.StubHTTPResponse{
				Body:   `{"id": "a"}`,
				Status: http.StatusOK,
				Headers: map[string][]string{
					"X-Test-Header": {"test-value"},
				},
			},
		},
	})
	defer srv.Close()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/resource?id=a",
		nil,
	)
	require.NoError(t, err)
	req.Header.Add("Content-Type", "application/json")
	responseResult := httputils.Do(http.DefaultClient, req)
	require.NoError(t, responseResult.Error())
	response := responseResult.MustGet()
	require.Equal(t, "test-value", response.Header.Get("X-Test-Header"))
	require.NoError(t, response.Body.Close())
}

func TestNewHTTPServerUnknown(t *testing.T) {
	srv, hitCounts := stubs.NewStubHTTPServerWithMultipleResponsesPerEndpoint(t, []stubs.StubHTTPEndpoints{
		{
			Request: stubs.StubHTTPRequests{URL: "/resource",
				ExpectedHeaders: []map[string]string{
					{"Content-Type": "application/json"},
				},
				ExpectedQueries: []string{"id=a"},
			},
			Response: stubs.StubHTTPResponses{
				Bodies:   []string{`{"id", "a"}`},
				Statuses: []int{http.StatusOK},
			},
		},
	})
	defer srv.Close()
	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		srv.URL+"/missing?id=a",
		nil,
	)
	require.NoError(t, err)
	resp := httputils.DoAsString(http.DefaultClient, req)
	expectedErr := strings.ReplaceAll("GET ${URL}/missing?id=a failed\n"+
		" Caused by:\n"+
		"\t* 404 Endpoint /missing?id=a not found for method GET",
		"${URL}", srv.URL)
	require.EqualError(t, resp.Error(), expectedErr)
	require.Equal(t, 0, (*hitCounts)["GET/resource"])
}

func getHostAndPort(
	t *testing.T,
	_url string,
) string {
	p, err := url.Parse(_url)
	require.NoError(t, err)
	return p.Host
}
