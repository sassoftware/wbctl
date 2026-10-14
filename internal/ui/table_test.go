// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ui_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	tbl := ui.NewTable(_ui, false, []string{"a", "b", "c"}...)
	tbl.Add([]string{"11111", "22222", "33333"}...)
	tbl.Add([]string{"1", "2", "3"}...)
	tbl.Print()
	expected := "A       B       C\n" +
		"11111   22222   33333\n" +
		"1       2       3\n"
	require.Equal(t, expected, w.String())
	require.Empty(t, e.String())
}

func TestTableWithColor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Colors not supported on Windows")
	}
	r := strings.NewReader("")
	w := &strings.Builder{}
	e := &strings.Builder{}
	_ui := ui.NewUI(r, w, e)
	tbl := ui.NewTable(_ui, true, []string{"a", "b", "c"}...)
	tbl.Add([]string{"11111", "22222", "33333"}...)
	tbl.Add([]string{"1", "2", "3"}...)
	tbl.Print()
	expected := "\x1b[1;38mA\x1b[0m       \x1b[1;38mB\x1b[0m       \x1b[1;38mC\x1b[0m\n" +
		"\x1b[96m11111\x1b[0m   22222   33333\n" +
		"\x1b[96m1    \x1b[0m   2       3\n"
	require.Equal(t, expected, w.String())
	require.Empty(t, e.String())
}
