// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth0admin

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/httputil"
)

var defaultClient = result.Ok(&http.Client{
	Timeout: time.Second * 60,
})

// Auth0AdminService is an exported type.
// nolint:revive
type Auth0AdminService struct {
	endpoint    string
	accessToken string
}

// NewAuth0AdminService returns a new workbench service.
func NewAuth0AdminService(endpoint, accessToken string) *Auth0AdminService {
	return &Auth0AdminService{
		endpoint:    endpoint,
		accessToken: accessToken,
	}
}

// GetAccessToken gets the access token from the given WorkbenchService.
func GetAccessToken(s *Auth0AdminService) string {
	return s.accessToken
}

// Get performs an http GET with the given path and query against the given
// WorkbenchService. The response body is marshaled from JSON into the given
// type T and returned as a result.
func Get[T any](ctx context.Context, s *Auth0AdminService, path string, query url.Values) result.Result[T] {
	resourceURL := httputil.MakeURL(s.endpoint, path, query)
	request := result.Map(httputil.MakeGetRequestContextFunc(ctx), resourceURL)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey: "Bearer " + s.accessToken,
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsJSON[T], defaultClient, request)
}

// GetRaw performs an http GET with the given path and query against the given
// WorkbenchService. The response body is marshaled into a string and returned
// as a result.
func GetRaw(ctx context.Context, s *Auth0AdminService, rawURL string) result.Result[string] {
	parsedURL := result.New(url.Parse(rawURL))
	request := result.Map(httputil.MakeGetRequestContextFunc(ctx), parsedURL)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey: "Bearer " + s.accessToken,
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsString, defaultClient, request)
}

// Post performs an http POST with the given path and query against the given
// WorkbenchService. The response body is marshaled from JSON into the given
// type T and returned as a result.
func Post[T any](ctx context.Context, s *Auth0AdminService, path string, body any) result.Result[T] {
	resourceURL := httputil.MakeURL(s.endpoint, path, url.Values{})
	request := result.Map(httputil.MakePostRequestContextFunc(ctx), resourceURL)
	request = result.Map(httputil.AddBody(body), request)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey: "Bearer " + s.accessToken,
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsJSON[T], defaultClient, request)
}
