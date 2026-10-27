// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/httputils"
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
	"github.com/sassoftware/wbctl/internal/cmd/profile"
	"github.com/sassoftware/wbctl/internal/cmd/rebuild"
	"github.com/sassoftware/wbctl/internal/cmd/settings"
	"github.com/sassoftware/wbctl/internal/cmd/state"
	"github.com/sassoftware/wbctl/internal/cmd/version"
	"github.com/sassoftware/wbctl/internal/cmd/whoami"
	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/logutil"
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
			cmd *cobra.Command,
			_ []string,
		) error {
			return configureApplication(cmd, fsys)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	AddRootFlags(app)
	app.AddCommand(createCommands(ctx, fsys, _ui)...)
	return app
}

// configureApplication handles the configuration of the application including
// loading any config file and setting up logging.
func configureApplication(
	cmd *cobra.Command,
	fsys afero.Fs,
) error {
	flags.BindToViper(cmd, nil)
	// Configure logging after binding so that loading the config file can be
	// logged based on command line options.
	configureLogging()
	err := loadConfigFile(fsys)
	if err != nil {
		return err
	}
	// Configure logging a second time to honor logging settings from the now
	// loaded config file.
	configureLogging()
	return nil
}

func configureLogging() {
	verbosityLevel := logutil.AsLevel(viper.Get(constants.VerboseFlag))
	// Log level set by command line is already handled but if the log level was
	// set by config file then logutil does not know about it and must be told.
	logutil.SetLevel(verbosityLevel)
	switch {
	case verbosityLevel < 0 && verbosityLevel > -5:
		httputils.SetDebug("true")
	case verbosityLevel <= -5:
		httputils.SetDebug("full")
	}
	viper.SetOptions(viper.WithLogger(logutil.GetLogger()))
}

// loadConfigFile loads values from the config file specified in the
// WBCONFIG env var.
func loadConfigFile(
	fsys afero.Fs,
) error {
	configPath, isDefault, err := configutil.ResolvePath()
	// If we can't get the home dir then just return. Ignore the error.
	if err != nil {
		logutil.GetLogger().Warn("failed to resolve config path", "error", err)
		return nil // nolint:nilerr
	}
	// Ignore a missing default config file. Otherwise, return an error for an
	// explicit config file that is missing.
	exists, existsErr := afero.Exists(fsys, configPath)
	if isDefault && (existsErr != nil || !exists) {
		return nil // nolint:nilerr
	}
	// If viper does not know what it is then try json
	if !configutil.IsKnownFormat(configPath) {
		viper.SetConfigType(constants.ConfigFileTypeDefault)
	}
	viper.SetConfigFile(configPath)
	err = viper.ReadInConfig()
	if err != nil {
		return errors.Wrap(
			err,
			i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}),
		)
	}
	// Resolve profile, if present. If a profile is selected and present, then
	// set all the profile's keys at the top level. Otherwise, leave viper
	// unchanged.
	profileName := viper.GetString(constants.ProfileFlag)
	logutil.GetLogger().Debug("loading profile", "profile", profileName)
	profile := viper.GetStringMap(profileName)
	if len(profile) == 0 {
		logutil.GetLogger().Debug("profile not found, using top level keys", "profile", profileName)
		return nil
	}
	for k, v := range profile {
		viper.Set(k, v)
	}
	return nil
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
	flags.AddVerboseFlag(cmd)
	flags.AddProfileFlag(cmd)
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
		whoami.Command(ctx, fsys, _ui),
		profile.Command(ctx, fsys, _ui),
	}
}
