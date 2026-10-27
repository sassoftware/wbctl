// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/osutil"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	// This is not a hardcoded credential value, but rather the expected error
	// messages when the token cache is malformed.
	// nolint:gosec
	tokenStoreInvalidCharacterErr = "invalid character 'o' in literal null (expecting 'u')"
)

func TestSaveFuncMissingHomeDir(t *testing.T) {
	fsys := afero.NewMemMapFs()
	t.Setenv(osutil.HomeEnvVar(), "")
	// Call the cache save fun. Should fail.
	saveResult := saveCache(fsys, tokenCache{}, "path")
	expectedSaveErr := osutil.HomeEnvVarName() + " is not defined"
	require.EqualError(t, saveResult, expectedSaveErr)
}

func TestSaveFuncRemovesTempFileOnRenameFailure(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv(osutil.HomeEnvVar(), homeDir)
	cachePath := filepath.Join(homeDir, constants.DefaultTokenFileName)
	fsys := stubs.StubFS{
		Fs: afero.NewMemMapFs(),
	}
	var removedPath string
	fsys.RemoveFunc = func(
		base afero.Fs,
		name string,
	) error {
		removedPath = name
		return base.Remove(name)
	}
	fsys.RenameFunc = func(
		_ afero.Fs,
		oldname, _ string,
	) error {
		return fmt.Errorf("rename %s: operation not permitted", oldname)
	}

	saveResult := saveCache(fsys, tokenCache{}, cachePath)
	require.EqualError(t, saveResult, "rename "+removedPath+": operation not permitted")
	require.NotEmpty(t, removedPath)
	_, statErr := fsys.Stat(removedPath)
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestSaveFuncRemovesTempFileOnWriteFailure(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv(osutil.HomeEnvVar(), homeDir)
	cachePath := filepath.Join(homeDir, constants.DefaultTokenFileName)
	fsys := stubs.StubFS{
		Fs: afero.NewMemMapFs(),
	}
	var removedPath string
	fsys.OpenFileFunc = func(
		base afero.Fs,
		name string,
		flag int,
		perm os.FileMode,
	) (afero.File, error) {
		if flag&os.O_EXCL != 0 {
			return base.OpenFile(name, flag, perm)
		}
		if flag&os.O_TRUNC != 0 {
			return nil, fmt.Errorf("write %s: operation not permitted", name)
		}
		return base.OpenFile(name, flag, perm)
	}
	fsys.RemoveFunc = func(
		base afero.Fs,
		name string,
	) error {
		removedPath = name
		return base.Remove(name)
	}

	saveResult := saveCache(fsys, tokenCache{}, cachePath)
	require.EqualError(t, saveResult, "write "+removedPath+": operation not permitted")
	require.NotEmpty(t, removedPath)
	_, statErr := fsys.Stat(removedPath)
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestLoadTokenCacheMissingFile(t *testing.T) {
	endpoint := "https://endpoint.test"
	tempDir := t.TempDir()
	t.Setenv(osutil.HomeEnvVar(), tempDir)

	fsys := afero.NewMemMapFs()
	result, err := getCachedTokenFor(fsys, viper.New(), endpoint)
	require.NoError(t, err)
	require.Equal(t, cachedToken{}, result)

	// Verify file does not exist
	expectedPath := filepath.Join(tempDir, constants.DefaultTokenFileName)
	_, err = fsys.Stat(expectedPath)
	require.ErrorIs(t, err, os.ErrNotExist, "expected file to not exist")
}

func TestLoadTokenCacheMalformedJSON(t *testing.T) {
	endpoint := "https://endpoint.test"
	tempDir := t.TempDir()
	t.Setenv(osutil.HomeEnvVar(), tempDir)

	// Create the directory and write malformed JSON.
	fsys := afero.NewMemMapFs()
	tokenDir := filepath.Join(tempDir, filepath.Dir(constants.DefaultTokenFileName))
	err := fsys.MkdirAll(tokenDir, 0o700)
	require.NoError(t, err, "failed to create token directory")

	tokenPath := filepath.Join(tempDir, constants.DefaultTokenFileName)
	err = afero.WriteFile(fsys, tokenPath, []byte("not valid json"), 0o600)
	require.NoError(t, err, "failed to write token to file")

	_, err = getCachedTokenFor(fsys, viper.New(), endpoint)
	expectedLoadErr := "failed to load token cache\n" +
		" Caused by:\n" +
		"\t* " + tokenStoreInvalidCharacterErr
	require.EqualError(t, err, expectedLoadErr)
}
