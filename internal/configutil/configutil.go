// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configutil

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/spf13/viper"
)

// ResolvePath returns the config file path, preferring the $WBCONFIG
// environment variable, falling back to the default path in the user's home
// directory. isDefault is true when no $WBCONFIG value was set, which lets
// callers decide whether a missing file should be treated as "nothing to
// load/bootstrap" rather than an error.
func ResolvePath() (configPath string, isDefault bool, err error) {
	configPath = os.Getenv(constants.ConfigFileEnvVar)
	if configPath != "" {
		return configPath, false, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", false, err
	}
	return filepath.Join(home, constants.DefaultConfigFileName), true, nil
}

// IsKnownFormat reports whether the given path's extension is a format type
// viper knows how to parse.
func IsKnownFormat(configPath string) bool {
	ext := filepath.Ext(configPath)
	if ext == "" {
		return false
	}
	ext = ext[1:]
	return slices.Contains(viper.SupportedExts, ext)
}
