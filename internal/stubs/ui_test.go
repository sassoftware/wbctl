// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs_test

import (
	"testing"

	"github.com/sas-institute-rnd-saas/wbctl/internal/stubs"
	"github.com/stretchr/testify/require"
)

func TestStubUI_RecordsAndNoopMethods(t *testing.T) {
	u := &stubs.StubUI{}

	u.Display("hello %s", "world")
	u.Display("count=%d", 2)
	u.Warn("warn %d", 7)
	u.Warn("warn %s", "again")
	u.Error("error %d", 7)
	u.Error("error %s", "again")

	require.Equal(t, []string{"hello world", "count=2"}, u.Displays)
	require.Equal(t, []string{"warn 7", "warn again"}, u.Warns)
	require.Equal(t, []string{"error 7", "error again"}, u.Errors)

	// Ask with empty queue returns "" and records the question.
	require.Empty(t, u.Ask("question %s", "one"))
	require.Equal(t, []string{"question one"}, u.Asks)

	// Ask with queued responses pops in FIFO order.
	u.AskResponses = []string{"first", "second"}
	require.Equal(t, "first", u.Ask("q2"))
	require.Equal(t, "second", u.Ask("q3"))
	require.Empty(t, u.Ask("q4")) // queue exhausted

	// AskForPassword with empty queue returns "" and records the question.
	require.Empty(t, u.AskForPassword("password %s", "prompt"))
	require.Equal(t, []string{"password prompt"}, u.AskForPasswords)

	// AskForPassword with queued responses pops in FIFO order.
	u.AskForPasswordResponses = []string{"secret1", "secret2"}
	require.Equal(t, "secret1", u.AskForPassword("q2"))
	require.Equal(t, "secret2", u.AskForPassword("q3"))
	require.Empty(t, u.AskForPassword("q4")) // queue exhausted

	tbl := u.Table([]string{"h1", "h2"})
	require.IsType(t, &stubs.CapturedTable{}, tbl)
	require.Len(t, u.Tables, 1)
	require.Equal(t, []string{"h1", "h2"}, u.Tables[0].Headers)

	captured := u.Tables[0]
	captured.Add("r1c1", "r1c2")
	require.Equal(t, "h1\th2\nr1c1\tr1c2", captured.String())
	captured.Print()
	require.Equal(t, 1, captured.PrintCalls)
}
