// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package list

import (
	"context"
	"slices"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/cmd/profile/config"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// newListFunction creates a command action that prints the profiles available
// in the current config file.
func newListFunction(
	_ context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		configPath, _, err := configutil.ResolvePath()
		if err != nil {
			return errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
		}
		doc, err := config.ReadDocument(fsys, configPath)
		if err != nil {
			return err
		}
		profileNames := make([]string, 0, len(doc))
		for profile := range doc {
			profileNames = append(profileNames, profile)
		}
		slices.Sort(profileNames)
		msg := i18n.Format(i18n.AvailableProfiles, i18n.A{"profiles": strings.Join(profileNames, ", ")})
		_ui.Display(msg)
		return nil
	}
}
