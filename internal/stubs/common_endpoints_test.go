// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"net/http"
	"testing"

	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/stretchr/testify/require"
)

func TestGetStubAuthEndpoints(t *testing.T) {
	endpoints := stubs.GetStubAuthEndpoints()
	require.Len(t, endpoints, 3)
}

func TestGetStubWorkspaceEndpoints(t *testing.T) {
	defaultEndpoints := stubs.GetStubWorkspaceEndpoints(stubs.StubHTTPPartialGetConfig{})
	require.Len(t, defaultEndpoints, 2)

	endpoints := stubs.GetStubWorkspaceEndpoints(stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusNotFound,
		ResponseBody:   `not-found`,
	})
	require.Len(t, endpoints, 1)
	require.Equal(t, `not-found`, endpoints[0].Response.Body)
}

func TestGetStubDataplaneEndpoints(t *testing.T) {
	defaultEndpoints := stubs.GetStubDataplaneEndpoints(stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `ok`,
	})
	require.Len(t, defaultEndpoints, 1)
	require.Equal(t, `ok`, defaultEndpoints[0].Response.Body)
}

func TestGetStubApplicationEndpoints(t *testing.T) {
	defaultEndpoints := stubs.GetStubApplicationEndpoints(stubs.StubHTTPPartialGetConfig{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `ok`,
	})
	require.Len(t, defaultEndpoints, 1)
	require.Equal(t, `ok`, defaultEndpoints[0].Response.Body)
}
