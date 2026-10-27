// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package init

import (
	"context"
	"fmt"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/cmd/profile/config"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// otherProfileFlags are the AddRootFlags flags, besides profile, endpoint,
// and organization, whose values are saved into the profile only when they
// were explicitly set on the command line.
var otherProfileFlags = []string{
	constants.AuthModeFlag,
	constants.IgnoreTokenCacheFlag,
	constants.NoListenFlag,
	constants.IngressHostFlag,
	constants.ClientIDFlag,
	constants.ClientSecretFlag,
	constants.VerboseFlag,
}

// newInitFunction creates a command action that prompts for (or defaults)
// the profile name, endpoint, and organization, then saves them, along with
// any other root flag explicitly set on the command line, into the named
// profile object in the config file.
func newInitFunction(
	_ context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		cmd *cobra.Command,
		_ []string,
	) error {
		logutil.GetLogger().Info("initializing profile")
		v := viper.GetViper()
		profileName, profile, err := buildProfile(_ui, v, cmd)
		if err != nil {
			return err
		}
		return saveProfile(_ui, fsys, profileName, profile)
	}
}

func buildProfile(
	_ui ui.UI,
	v *viper.Viper,
	cmd *cobra.Command,
) (string, map[string]any, error) {
	profileName := promptWithDefault(
		_ui,
		i18n.Format(i18n.ProfileNamePrompt),
		v.GetString(constants.ProfileFlag),
	)
	if profileName == "" {
		return "", nil, errors.New(
			i18n.Format(i18n.FlagIsRequired, i18n.A{constants.I18nArgKeyFlagName: constants.ProfileFlag}),
		)
	}
	endpoint := promptWithDefault(
		_ui,
		i18n.Format(i18n.EndpointPrompt),
		v.GetString(constants.EndpointFlag),
	)
	if endpoint == "" {
		return "", nil, errors.New(
			i18n.Format(i18n.FlagIsRequired, i18n.A{constants.I18nArgKeyFlagName: constants.EndpointFlag}),
		)
	}
	organization := promptWithDefault(
		_ui,
		i18n.Format(i18n.OrganizationPrompt),
		v.GetString(constants.OrganizationFlag),
	)
	profile := map[string]any{
		constants.EndpointFlag: endpoint,
	}
	if organization != "" {
		profile[constants.OrganizationFlag] = organization
	}
	for _, flagName := range otherProfileFlags {
		if cmd.Flags().Changed(flagName) {
			profile[flagName] = v.Get(flagName)
		}
	}
	return profileName, profile, nil
}

func saveProfile(
	_ui ui.UI,
	fsys afero.Fs,
	profileName string,
	profile map[string]any,
) error {
	logutil.GetLogger().Info("saving profile", "name", profileName)
	configPath, _, err := configutil.ResolvePath()
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.ConfigFileWriteFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
	}
	doc, err := config.ReadDocument(fsys, configPath)
	if err != nil {
		return err
	}
	doc[profileName] = profile
	err = config.WriteDocument(fsys, configPath, doc)
	if err != nil {
		return err
	}
	_ui.Display(i18n.Format(
		i18n.ProfileSaved,
		i18n.A{
			constants.I18nArgKeyName: profileName,
			constants.I18nArgKeyPath: configPath,
		},
	))
	return nil
}

// promptWithDefault asks the user for a value, showing defaultValue (if any)
// in the prompt, and returns defaultValue unchanged when the user enters a
// blank response.
func promptWithDefault(
	_ui ui.UI,
	prompt string,
	defaultValue string,
) string {
	displayPrompt := prompt
	if defaultValue != "" {
		displayPrompt = fmt.Sprintf("%s [%s]", prompt, defaultValue)
	}
	answer := _ui.Ask(displayPrompt)
	if answer == "" {
		return defaultValue
	}
	return answer
}
