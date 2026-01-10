// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Titler is used to title case resource type names for help strings. Resource
// type names are domain specific and therefore always in English.
var Titler = cases.Title(language.English)

// ActionFunc is a typedef for a cobra.Command action function.
type ActionFunc func(
	cmd *cobra.Command,
	args []string,
) error
