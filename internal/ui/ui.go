// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"golang.org/x/term"
)

// UI provides methods for interacting with the user in a terminal.
type UI interface {
	Display(message string, args ...interface{})
	Ask(prompt string, args ...interface{}) (answer string)
	AskForPassword(prompt string, args ...interface{}) (answer string)
	Table(headers []string) Table
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
}

type ui struct {
	stdinReader   *bufio.Reader
	stdout        io.Writer
	stderr        io.Writer
	colorsEnabled bool
}

// NewUI returns a new UI that will read from the given io.Reader and write to
// the given io.Writer.
func NewUI(
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) UI {
	return &ui{
		stdinReader:   bufio.NewReader(stdin),
		stdout:        stdout,
		stderr:        stderr,
		colorsEnabled: colorsSupported(),
	}
}

var _ UI = (*ui)(nil)

// Display writes the given message to the stdout io.Writer of the UI.
func (_ui *ui) Display(
	message string,
	args ...interface{},
) {
	display(_ui.stdout, message, args...)
}

// Warn writes the given message to the stderr io.Writer of the UI in a warning
// color.
func (_ui *ui) Warn(message string, args ...interface{}) {
	_ui.colorizedLevelMessage(magenta, "WARNING", message, args...)
}

// Error writes the given message to the stderr io.Writer of the UI. It also
// prepends the string "ERROR -" in an error color to the beginning of the
// string.
func (_ui *ui) Error(message string, args ...interface{}) {
	_ui.colorizedLevelMessage(red, "ERROR", message, args...)
}

// Ask displays the given prompt to the io.Writer of the UI and reads the user's
// answer from the io.Reader of the UI.
func (_ui *ui) Ask(prompt string, args ...interface{}) (answer string) {
	_, _ = fmt.Fprintln(_ui.stdout, "")
	separator := ">"
	if _ui.colorsEnabled {
		separator = colorizeBold(">", cyan)
	}
	_, _ = fmt.Fprintf(_ui.stdout, prompt+separator+" ", args...)

	line, err := _ui.stdinReader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

// AskForPassword displays the given prompt to the io.Writer of the UI. If
// stdout of the program is a terminal then the password is read directly from
// os.Stdin without echoing the input to the terminal. Otherwise, the password
// is read from the io.Reader of the UI.
func (_ui *ui) AskForPassword(prompt string, args ...interface{}) (answer string) {
	_, _ = fmt.Fprintln(_ui.stdout, "")
	separator := ">"
	if _ui.colorsEnabled {
		separator = colorizeBold(">", cyan)
	}
	_, _ = fmt.Fprintf(_ui.stdout, prompt+separator+" ", args...)
	var password []byte
	var err error
	if isFdTerminal(stdinFd()) {
		password, err = term.ReadPassword(stdinFd())
		if err != nil {
			return ""
		}
	} else {
		password, err = _ui.stdinReader.ReadBytes('\n')
		if err != nil {
			return ""
		}
	}
	return strings.TrimSpace(string(password))
}

// Table returns a new Table.
func (_ui *ui) Table(headers []string) Table {
	return NewTable(_ui, _ui.colorsEnabled, headers...)
}

func (_ui *ui) colorizedLevelMessage(color color, level, message string, args ...interface{}) {
	if len(args) != 0 {
		message = fmt.Sprintf(message, args...)
	}
	if _ui.colorsEnabled {
		level = colorizeBold(level, color)
	}
	display(_ui.stderr, "%s: %s", level, message)
}

func display(out io.Writer, message string, args ...interface{}) {
	if len(args) == 0 {
		_, _ = fmt.Fprintf(out, "%s\n", message)
	} else {
		_, _ = fmt.Fprintf(out, message+"\n", args...)
	}
}

func colorizeBold(message string, color color) string {
	return fmt.Sprintf("\033[1;%dm%s\033[0m", color, message)
}

type color uint

const (
	red     color = 31
	magenta color = 35
	cyan    color = 36
	white   color = 38
)

func colorsSupported() bool {
	if runtime.GOOS != "windows" &&
		isFdTerminal(int(os.Stdout.Fd())) &&
		os.Getenv("TERM") != "dumb" {
		return true
	}
	return false
}

// Testing hooks

var isFdTerminal = isFdTerminalDefaultImplementation

func isFdTerminalDefaultImplementation(fd int) bool {
	return term.IsTerminal(fd)
}

var stdinFd = stdinFdDefaultImplementation

func stdinFdDefaultImplementation() int {
	return int(os.Stdin.Fd())
}
