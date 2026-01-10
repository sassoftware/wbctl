// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package osutil

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHomeEnvVar(t *testing.T) {
	require.Equal(t, homeEnvVar, HomeEnvVar())
}

func TestHomeEnvVarName(t *testing.T) {
	require.Equal(t, homeEnvVarName, HomeEnvVarName())
}

func TestPathEnvVar(t *testing.T) {
	require.Equal(t, pathEnvVar, PathEnvVar())
}

func TestPathEnvVarName(t *testing.T) {
	require.Equal(t, pathEnvVarName, PathEnvVarName())
}

func TestPathSeparatorVar(t *testing.T) {
	require.Equal(t, string(os.PathSeparator), PathSeparator())
}

func TestExecFormatError(t *testing.T) {
	require.Equal(t, execFormatError, ExecFormatError())
}

func TestExecutableScript(t *testing.T) {
	elements := []string{"a", "b", "c"}
	require.Equal(t, strings.Join(elements, PathSeparator())+scriptExtension, ExecutableScript(elements...))
}

func TestExecutable(t *testing.T) {
	elements := []string{"a", "b", "c"}
	require.Equal(t, strings.Join(elements, PathSeparator())+executableExtension, Executable(elements...))
}

func TestDefaultEditor(t *testing.T) {
	require.Equal(t, defaultEditor, DefaultEditor())
}

func TestDefaultBinDir(t *testing.T) {
	require.Equal(t, defaultBinDir, DefaultBinDir())
}
