// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/exiter"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedExitCode int
	}{
		{
			name: "success",
			args: []string{
				"wbctl",
				"--help",
			},
			expectedExitCode: 0,
		},
		{
			name: "failure",
			args: []string{
				"wbctl",
				"get",
				"workbenches",
				"--endpoint", "http://localhost:0",
			},
			expectedExitCode: 1,
		},
		{
			name: "create failure",
			args: []string{
				"wbctl",
				"create",
				"workbench",
			},
			expectedExitCode: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualExitCode := 0
			exiter.OsExiter = func(code int) {
				actualExitCode = code
			}
			os.Args = tt.args
			main()
			assert.Equal(t, tt.expectedExitCode, actualExitCode)
		})
	}

	main()
}
