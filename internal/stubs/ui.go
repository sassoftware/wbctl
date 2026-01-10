// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"fmt"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
)

// StubUI is a test UI implementation that records display and warning messages.
// Configure AskResponses and AskForPasswordResponses with queued return values
// for Ask calls (FIFO; empty queue returns "").
type StubUI struct {
	Displays []string
	Warns    []string
	Errors   []string

	// Recorded questions
	Asks            []string
	AskForPasswords []string

	// Configured responses
	AskResponses            []string
	AskForPasswordResponses []string

	// Captured tables, one entry per Table(...) call.
	Tables []*CapturedTable
}

// Display records a formatted display message.
func (u *StubUI) Display(message string, args ...interface{}) {
	if len(args) != 0 {
		message = fmt.Sprintf(message, args...)
	}
	u.Displays = append(u.Displays, message)
}

// Warn records a formatted warning message.
func (u *StubUI) Warn(message string, args ...interface{}) {
	if len(args) != 0 {
		message = fmt.Sprintf(message, args...)
	}
	u.Warns = append(u.Warns, message)
}

// Error records a formatted error message.
func (u *StubUI) Error(message string, args ...interface{}) {
	if len(args) != 0 {
		message = fmt.Sprintf(message, args...)
	}
	u.Errors = append(u.Errors, message)
}

// Ask records a question and returns the next queued response.
func (u *StubUI) Ask(question string, args ...interface{}) string {
	u.Asks = append(u.Asks, fmt.Sprintf(question, args...))
	if len(u.AskResponses) == 0 {
		return ""
	}
	r := u.AskResponses[0]
	u.AskResponses = u.AskResponses[1:]
	return r
}

// AskForPassword records a password prompt and returns the next queued response.
func (u *StubUI) AskForPassword(question string, args ...interface{}) string {
	u.AskForPasswords = append(u.AskForPasswords, fmt.Sprintf(question, args...))
	if len(u.AskForPasswordResponses) == 0 {
		return ""
	}
	r := u.AskForPasswordResponses[0]
	u.AskForPasswordResponses = u.AskForPasswordResponses[1:]
	return r
}

// Table captures table rows for later assertions.
func (u *StubUI) Table(headers []string) ui.Table {
	t := &CapturedTable{
		Headers: append([]string(nil), headers...),
	}
	u.Tables = append(u.Tables, t)
	return t
}

// CapturedTable records all table interactions for assertions in tests.
type CapturedTable struct {
	Headers    []string
	PrintCalls int
	Rows       [][]string
}

// String is an exported method.
func (t *CapturedTable) String() string {
	lines := make([]string, 0, len(t.Rows)+1)
	if len(t.Headers) > 0 {
		lines = append(lines, strings.Join(t.Headers, "\t"))
	}
	for _, row := range t.Rows {
		lines = append(lines, strings.Join(row, "\t"))
	}
	return strings.Join(lines, "\n")
}

// Add appends a row to the captured table.
func (t *CapturedTable) Add(values ...string) {
	t.Rows = append(t.Rows, append([]string(nil), values...))
}

// Print records a print call.
func (t *CapturedTable) Print() {
	t.PrintCalls++
}

var _ ui.Table = (*CapturedTable)(nil)

var _ ui.UI = (*StubUI)(nil)
