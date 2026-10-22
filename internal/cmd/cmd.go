// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"os"
	"path"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/cmd/api"
	"github.com/sassoftware/wbctl/internal/cmd/cancel"
	"github.com/sassoftware/wbctl/internal/cmd/create"
	"github.com/sassoftware/wbctl/internal/cmd/delete"
	"github.com/sassoftware/wbctl/internal/cmd/describe"
	"github.com/sassoftware/wbctl/internal/cmd/edit"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/cmd/get"
	"github.com/sassoftware/wbctl/internal/cmd/log"
	"github.com/sassoftware/wbctl/internal/cmd/login"
	"github.com/sassoftware/wbctl/internal/cmd/logout"
	"github.com/sassoftware/wbctl/internal/cmd/rebuild"
	"github.com/sassoftware/wbctl/internal/cmd/settings"
	"github.com/sassoftware/wbctl/internal/cmd/state"
	"github.com/sassoftware/wbctl/internal/cmd/version"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateApp returns the root app for the sas application.
func CreateApp(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	// Enable running of parent hooks from child commands. This allows us to
	// keep the config file handling in one place. We keep the viper binding in
	// the individual commands to make unit testing easier.
	cobra.EnableTraverseRunHooks = true
	app := &cobra.Command{
		Use:   constants.SASAppName,
		Short: constants.SASAppUsage,
		PersistentPreRunE: func(
			_ *cobra.Command,
			_ []string,
		) error {
			return loadConfigFile(fsys)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	AddRootFlags(app)
	app.AddCommand(createCommands(ctx, fsys, _ui)...)
	return app
}

// loadConfigFile loads values from the config file specified in the
// WBCONFIG env var.
func loadConfigFile(fsys afero.Fs) error {
	ignoreMissing := false
	configPath := os.Getenv(constants.ConfigFileEnvVar)
	if configPath == "" {
		ignoreMissing = true
		home, err := os.UserHomeDir()
		// If we can't get the home dir then just return. Ignore the error.
		if err != nil {
			return nil // nolint:nilerr
		}
		configPath = path.Join(home, constants.DefaultConfigFileName)
	}
	// Ignore a missing config default config file. Otherwise, return an error
	// for an explicit config file that is missing.
	exists, err := afero.Exists(fsys, configPath)
	if ignoreMissing && (err != nil || !exists) {
		return nil // nolint:nilerr
	}
	// If viper does not know what it is then try json
	if !isKnownViperType(configPath) {
		viper.SetConfigType(constants.DefaultConfigFileType)
	}
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(
			err,
			i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}),
		)
	}
	return nil
}

func isKnownViperType(configPath string) bool {
	ext := path.Ext(configPath)
	if len(ext) < 1 {
		return false
	}
	ext = ext[1:] // remove the dot
	for _, _type := range viper.SupportedExts {
		if ext == _type {
			return true
		}
	}
	return false
}

// AddRootFlags adds the root flags common to all commands.
func AddRootFlags(cmd *cobra.Command) {
	flags.AddAuthModeFlag(cmd)
	flags.AddEndpointFlag(cmd)
	flags.AddIgnoreTokenCacheFlag(cmd)
	flags.AddNoListenFlag(cmd)
	flags.AddIngressHostFlag(cmd)
	flags.AddOrganizationFlag(cmd)
	flags.AddClientIDFlag(cmd)
	flags.AddClientSecretFlag(cmd)
}

func createCommands(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) []*cobra.Command {
	return []*cobra.Command{
		api.Command(ctx, fsys, _ui),
		get.Command(ctx, fsys, _ui),
		describe.Command(ctx, fsys, _ui),
		delete.Command(ctx, fsys, _ui),
		create.CreateCommand(ctx, fsys, _ui),
		create.UpdateCommand(ctx, fsys, _ui),
		edit.Command(ctx, fsys, _ui),
		login.Command(ctx, fsys, _ui),
		logout.Command(ctx, fsys, _ui),
		log.Command(ctx, fsys, _ui),
		state.StartCommand(ctx, fsys, _ui),
		state.StopCommand(ctx, fsys, _ui),
		cancel.Command(ctx, fsys, _ui),
		rebuild.Command(ctx, fsys, _ui),
		settings.Command(ctx, fsys, _ui),
		version.Command(ctx, fsys, _ui),
	}
}
