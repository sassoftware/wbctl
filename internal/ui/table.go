// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package ui

import (
	prettytable "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Table provides an interface for displaying tabular information.
type Table interface {
	Add(row ...string)
	Print()
}
type table struct {
	ui            UI
	writer        prettytable.Writer
	colorsEnabled bool
}

// NewTable returns a new Table with the given headers that will render to the
// given UI.
func NewTable(ui UI, colorsEnabled bool, headers ...string) Table {
	writer := prettytable.NewWriter()
	writer.SuppressTrailingSpaces()
	writer.Style().Box.PaddingLeft = ""
	writer.Style().Box.PaddingRight = ""
	writer.Style().Box.LeftSeparator = " "
	writer.Style().Box.RightSeparator = " "
	writer.Style().Box.MiddleVertical = "   "
	writer.Style().Options.DrawBorder = false
	writer.Style().Options.SeparateColumns = true
	writer.Style().Options.SeparateHeader = false
	writer.SetIndexColumn(1)
	if colorsEnabled {
		// We can't use the header ColorOptions because SuppressTrailingSpaces
		// is applied after the color codes are added which makes the strings no
		// longer end in spaces. But we can use it for the index column because
		// it is not at the end of the string.
		writer.Style().Color = prettytable.ColorOptions{
			IndexColumn: text.Colors{text.FgHiCyan},
		}
		writer.AppendHeader(asSliceOfInterface(asSliceOfColorized(headers, white)...))
	} else {
		writer.AppendHeader(asSliceOfInterface(headers...))
	}
	return &table{
		ui:            ui,
		writer:        writer,
		colorsEnabled: colorsEnabled,
	}
}

// Add adds the given row to the table.
func (_table *table) Add(row ...string) {
	_table.writer.AppendRow(asSliceOfInterface(row...))
}

// Print renders the table and displays it to the UI.
func (_table *table) Print() {
	_table.ui.Display(_table.writer.Render())
}

func asSliceOfInterface(slice ...string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

func asSliceOfColorized(slice []string, color color) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[i] = colorizeBold(v, color)
	}
	return result
}
