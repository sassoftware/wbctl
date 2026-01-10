// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package osutil

import (
	"os"
	"strings"
)

// HomeEnvVar returns the os specifiv home environment variable.
func HomeEnvVar() string {
	return homeEnvVar
}

// HomeEnvVarName returns the os specific name of the home environment variable.
func HomeEnvVarName() string {
	return homeEnvVarName
}

// PathEnvVar returns the os specific path environment variable.
func PathEnvVar() string {
	return pathEnvVar
}

// PathEnvVarName returns the os specific name of the path environment variable.
func PathEnvVarName() string {
	return pathEnvVarName
}

// PathSeparator returns the os.PathSeparator as a string.
func PathSeparator() string {
	return string(os.PathSeparator)
}

// ExecFormatError returns the os specific error for a binary that is not
// executable.
func ExecFormatError() string {
	return execFormatError
}

// ExecutableScript returns the given path parts joined with the os specific path
// separator and with the os specific executable script extension appended.
func ExecutableScript(elements ...string) string {
	return strings.Join(elements, PathSeparator()) + scriptExtension
}

// Executable returns the given path parts joined with the os specific path
// separator and with the os specific executable script extension appended.
func Executable(elements ...string) string {
	return strings.Join(elements, PathSeparator()) + executableExtension
}

// DefaultEditor returns the os specific default text editor.
func DefaultEditor() string {
	return defaultEditor
}

// DefaultBinDir returns the os specific default bin directory. For test
// benchmarks.
func DefaultBinDir() string {
	return defaultBinDir
}
