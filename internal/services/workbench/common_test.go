// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench_test

import (
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
)

// Get test table structs.
type getRequest struct {
	ids  []string
	sort string
}

type getExpected[T any] struct {
	err    string
	result T
}

type getTestCase[T any] struct {
	name          string
	request       getRequest
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      getExpected[T]
}

// Bulk update test table structs.
type bulkupdateRequest[T any] struct {
	ids  []string
	body T
}

type bulkupdateExpected[T any] struct {
	err    string
	result T
}

type bulkupdateTestCase[T any, TS any] struct {
	name          string
	request       bulkupdateRequest[T]
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      bulkupdateExpected[TS]
}

// Rebuild test table structs.
type rebuildRequest struct {
	rebuildID string
}

type rebuildExpected struct {
	err    string
	result bool
}

type rebuildTestCase struct {
	name          string
	request       rebuildRequest
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      rebuildExpected
}

// Delete test table structs.
type deleteRequest struct {
	ids []string
}

type deleteExpected getExpected[[]string]

type deleteTestCase struct {
	name          string
	request       deleteRequest
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      deleteExpected
}

// Create test table structs.
type createRequest[T any] struct {
	body T
}

type createExpected[T any] struct {
	err    string
	result T
}

type createTestCase[T any] struct {
	name          string
	request       createRequest[T]
	mockEndpoints []stubs.StubHTTPEndpoint
	expected      createExpected[T]
}

type paramPayload[T any] struct {
	Param T `json:"param"`
}

var (
	authHeaders = map[string]string{
		constants.AuthorizationHeaderKey:      "Bearer access-token",
		constants.PostgrestCSRFTokenHeaderKey: "csrf-token",
	}
)
