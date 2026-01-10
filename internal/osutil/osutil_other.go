// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package osutil

const (
	defaultBinDir       = "/usr/bin"
	defaultEditor       = "vi"
	execFormatError     = "exec format error"
	executableExtension = ""
	homeEnvVar          = "HOME"
	homeEnvVarName      = "$HOME"
	pathEnvVar          = "PATH"
	pathEnvVarName      = "$PATH"
	scriptExtension     = ".sh"
)
