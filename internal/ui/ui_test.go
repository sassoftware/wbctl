// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ui_test

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/stretchr/testify/require"
)

type errReader struct{}

func (errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("failed")
}

func TestNewUI(t *testing.T) {
	r := strings.NewReader("")
	w := bufio.NewWriter(io.Discard)
	e := bufio.NewWriter(io.Discard)
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
}

func TestDisplay(t *testing.T) {
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	_ui.Display("Hello, World!")
	_ui.Display("Hello, %s!", "World")
	require.Equal(t, "Hello, World!\nHello, World!\n", w.String())
	require.Empty(t, e.String())
}

func TestWarn(t *testing.T) {
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	_ui.Warn("Hello, World!")
	_ui.Warn("Hello, %s!", "World")
	require.Empty(t, w.String())
	require.Equal(t, "WARNING: Hello, World!\nWARNING: Hello, World!\n", e.String())
}

func TestError(t *testing.T) {
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	_ui.Error("Hello, World!")
	_ui.Error("Hello, %s!", "World")
	require.Empty(t, w.String())
	require.Equal(t, "ERROR: Hello, World!\nERROR: Hello, World!\n", e.String())
}

func TestAsk(t *testing.T) {
	expected := "testuser"
	r := strings.NewReader(expected + "\n")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.Ask("Username")
	require.Equal(t, expected, actual)
	require.Equal(t, "\nUsername> ", w.String())
	require.Empty(t, e.String())
}

func TestAskFailedRead(t *testing.T) {
	r := errReader{}
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.Ask("Username")
	require.Empty(t, actual)
	require.Equal(t, "\nUsername> ", w.String())
	require.Empty(t, e.String())
}

func TestAskForPassword(t *testing.T) {
	expected := "testpassword"
	r := strings.NewReader(expected + "\n")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.AskForPassword("Password")
	require.Equal(t, expected, actual)
	require.Equal(t, "\nPassword> ", w.String())
	require.Empty(t, e.String())
}

func TestAskForPasswordFailedRead(t *testing.T) {
	r := errReader{}
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	actual := _ui.AskForPassword("Password")
	require.Empty(t, actual)
	require.Equal(t, "\nPassword> ", w.String())
	require.Empty(t, e.String())
}

func TestGetTable(t *testing.T) {
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	require.NotNil(t, _ui)
	tbl := _ui.Table([]string{"one", "two"})
	require.NotNil(t, tbl)
	require.Empty(t, w.String())
	require.Empty(t, e.String())
}
