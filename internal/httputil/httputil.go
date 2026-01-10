// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package httputil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// MakeHeadRequest builds a HEAD request constructor for the given context.
func MakeHeadRequest(ctx context.Context) func(*url.URL) (*http.Request, error) {
	return func(url *url.URL) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodHead, url.String(), nil)
	}
}

// MakeGetRequestContextFunc builds a GET request constructor for the given context.
func MakeGetRequestContextFunc(ctx context.Context) func(*url.URL) (*http.Request, error) {
	return func(url *url.URL) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	}
}

// MakePostRequestContextFunc builds a POST request constructor for the given context.
func MakePostRequestContextFunc(ctx context.Context) func(*url.URL) (*http.Request, error) {
	return func(url *url.URL) (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodPost, url.String(), nil)
	}
}

// AddHeaders adds the provided headers to a request.
func AddHeaders(headers map[string]string, req *http.Request) *http.Request {
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req
}

// AddBody marshals body into a request body when it is non-nil.
func AddBody(body any) func(*http.Request) (*http.Request, error) {
	return func(req *http.Request) (*http.Request, error) {
		if body == nil {
			return req, nil
		}
		content, err := json.Marshal(body)
		if err != nil {
			return req, err
		}
		req.Body = io.NopCloser(bytes.NewReader(content))
		return req, nil
	}
}

// MakeURL joins and parses an endpoint, path, and optional query string.
func MakeURL(endpoint string, path string, query url.Values) result.Result[*url.URL] {
	stringURL := result.New(url.JoinPath(endpoint, path))
	parsedURL := result.Map(url.Parse, stringURL)
	return result.MapNoError(func(u *url.URL) *url.URL {
		if len(query) > 0 {
			u.RawQuery = query.Encode()
		}
		return u
	}, parsedURL)
}
