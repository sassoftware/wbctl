// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"os"

	"github.com/sassoftware/wbctl/internal/cmd"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/exiter"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
)

func main() {
	_ui := ui.NewUI(os.Stdin, os.Stdout, os.Stderr)
	ctx := context.Background()
	app := cmd.CreateApp(ctx, afero.NewOsFs(), _ui)
	err := app.ExecuteContext(ctx)
	if err != nil {
		_ui.Error(i18n.Format(i18n.Error, i18n.A{constants.I18nArgKeyError: err.Error()}))
		exiter.OsExiter(1)
	} else {
		exiter.OsExiter(0)
	}
}
