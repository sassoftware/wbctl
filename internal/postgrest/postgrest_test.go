// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package postgrest_test

import (
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/stretchr/testify/require"
)

func TestEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "empty", input: "", expected: "/postgrest"},
		{name: "slash", input: "/", expected: "/postgrest/"},
		{name: "plain segment", input: "rpc/job_create", expected: "/postgrest/rpc/job_create"},
		{name: "leading slash", input: "/rpc/job_create", expected: "/postgrest/rpc/job_create"},
		{name: "control character is escaped", input: "\n", expected: "/postgrest/%0A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, postgrest.Endpoint(tt.input))
		})
	}
}
