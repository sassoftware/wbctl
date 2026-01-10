// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package osutil

const (
	defaultBinDir       = "C:\\Windows\\system32"
	defaultEditor       = "edit"
	execFormatError     = "%1 is not a valid Win32 application."
	executableExtension = ".exe"
	homeEnvVar          = "USERPROFILE"
	homeEnvVarName      = "%userprofile%"
	pathEnvVar          = "PATH"
	pathEnvVarName      = "%PATH%"
	scriptExtension     = ".bat"
)
