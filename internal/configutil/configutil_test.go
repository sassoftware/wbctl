// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configutil_test

import (
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/stretchr/testify/require"
)

func TestResolvePath(t *testing.T) {
	t.Run("honors WBCONFIG", func(t *testing.T) {
		t.Setenv(constants.ConfigFileEnvVar, "/custom/path/config.yaml")
		p, isDefault, err := configutil.ResolvePath()
		require.NoError(t, err)
		require.False(t, isDefault)
		require.Equal(t, "/custom/path/config.yaml", p)
	})

	t.Run("falls back to home dir default", func(t *testing.T) {
		t.Setenv(constants.ConfigFileEnvVar, "")
		home := t.TempDir()
		t.Setenv(osutil.HomeEnvVar(), home)
		p, isDefault, err := configutil.ResolvePath()
		require.NoError(t, err)
		require.True(t, isDefault)
		require.Equal(t, filepath.Join(home, ".wbctl", "config.json"), p)
	})

	t.Run("error returned for missing home", func(t *testing.T) {
		t.Setenv(osutil.HomeEnvVar(), "")
		p, isDefault, err := configutil.ResolvePath()
		require.Error(t, err)
		require.Empty(t, p)
		require.False(t, isDefault)
	})
}

func TestIsKnownFormat(t *testing.T) {
	require.True(t, configutil.IsKnownFormat("filename."+constants.ConfigFileTypeYML))
	require.True(t, configutil.IsKnownFormat("filename."+constants.ConfigFileTypeYAML))
	require.True(t, configutil.IsKnownFormat("filename."+constants.ConfigFileTypeDefault))
	require.False(t, configutil.IsKnownFormat("filename"))
	require.False(t, configutil.IsKnownFormat("filename.unknown"))
}
