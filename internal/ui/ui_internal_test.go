// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/creack/pty"
	"github.com/stretchr/testify/require"
)

func TestWarnColors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Colors not supported on Windows")
	}
	isFdTerminal = func(_ int) bool { return true }
	defer func() { isFdTerminal = isFdTerminalDefaultImplementation }()
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := NewUI(r, w, e)
	require.NotNil(t, _ui)
	_ui.Warn("Hello, World!")
	_ui.Warn("Hello, %s!", "World")
	require.Empty(t, w.String())
	require.Equal(t, "\x1b[1;35mWARNING\x1b[0m: Hello, World!\n\x1b[1;35mWARNING\x1b[0m: Hello, World!\n", e.String())
}

func TestErrorColors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Colors not supported on Windows")
	}
	isFdTerminal = func(_ int) bool { return true }
	defer func() { isFdTerminal = isFdTerminalDefaultImplementation }()
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := NewUI(r, w, e)
	require.NotNil(t, _ui)
	_ui.Error("Hello, World!")
	_ui.Error("Hello, %s!", "World")
	require.Empty(t, w.String())
	expected := "\x1b[1;31mERROR\x1b[0m: Hello, World!\n\x1b[1;31mERROR\x1b[0m: Hello, World!\n"
	require.Equal(t, expected, e.String())
}

func TestAskColors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Colors not supported on Windows")
	}
	isFdTerminal = func(_ int) bool { return true }
	defer func() { isFdTerminal = isFdTerminalDefaultImplementation }()
	expected := "testuser"
	r := strings.NewReader(expected + "\n")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.Ask("Username")
	require.Equal(t, expected, actual)
	require.Equal(t, "\nUsername\x1b[1;36m>\x1b[0m ", w.String())
	require.Empty(t, e.String())
}

func TestAskForPasswordFromRealTerminal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test does not work on windows")
	}
	ptm, pts, err := pty.Open()
	require.NoError(t, err)
	defer func() { _ = ptm.Close() }()
	defer func() { _ = pts.Close() }()
	origStdin := os.Stdin
	origStdout := os.Stdout
	os.Stdin = pts
	os.Stdout = pts
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := NewUI(r, w, e)
	require.NotNil(t, _ui)
	expected := "testpassword"
	_, _ = ptm.WriteString(expected + "\n")
	actual := _ui.AskForPassword("Password")
	require.Equal(t, expected, actual)
	require.Equal(t, "\nPassword\x1b[1;36m>\x1b[0m ", w.String())
	require.Empty(t, e.String())
}

func TestAskForPasswordFromInvalidTerminal(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test does not work on windows")
	}
	// Set up a real terminal so the prompt has colors but return an invalid
	// file descriptor the second time so the read fails.
	calls := 0
	stdinFd = func() int {
		calls++
		if calls > 1 {
			return -1
		}
		return int(os.Stdin.Fd())
	}
	defer func() { stdinFd = stdinFdDefaultImplementation }()
	ptm, pts, err := pty.Open()
	require.NoError(t, err)
	defer func() { _ = ptm.Close() }()
	defer func() { _ = pts.Close() }()
	origStdin := os.Stdin
	origStdout := os.Stdout
	os.Stdin = pts
	os.Stdout = pts
	defer func() {
		os.Stdin = origStdin
		os.Stdout = origStdout
	}()
	// Because there is a real terminal, the reader is ignored.
	r := strings.NewReader("ignored\n")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.AskForPassword("Password")
	require.Empty(t, actual)
	require.Equal(t, "\nPassword\x1b[1;36m>\x1b[0m ", w.String())
	require.Empty(t, e.String())
}
