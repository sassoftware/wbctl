// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDetectWritableFormat(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedFormat format
		expectedErr    bool
	}{
		{name: "yaml extension", path: "/x/config.yaml", expectedFormat: formatYAML},
		{name: "yml extension", path: "/x/config.yml", expectedFormat: formatYAML},
		{name: "json extension", path: "/x/config.json", expectedFormat: formatJSON},
		{name: "no extension defaults to json", path: "/x/config", expectedFormat: formatJSON},
		{name: "unrecognized extension defaults to json", path: "/x/config.conf", expectedFormat: formatJSON},
		{name: "known but unsupported extension errors", path: "/x/config.toml", expectedErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := detectWritableFormat(tt.path)
			if tt.expectedErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedFormat, format)
		})
	}
}
