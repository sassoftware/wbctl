// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package flags

import (
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// AddOutputFlag adds the output format flag to the command.
func AddOutputFlag(cmd *cobra.Command) {
	key := constants.OutputFlag
	cmd.Flags().StringP(
		key,
		constants.OutputShortFlag,
		constants.TextOutputValue,
		i18n.Format(i18n.FlagOutputUsage),
	)
}

// AddSortFlag adds the sort expression flag to the command.
func AddSortFlag(cmd *cobra.Command) {
	key := constants.SortFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSortUsage),
	)
}

// AddDryRunFlag adds the dry-run flag to the command.
func AddDryRunFlag(cmd *cobra.Command) {
	key := constants.DryRunFlag
	cmd.Flags().BoolP(
		key,
		constants.DryRunShortFlag,
		false,
		i18n.Format(i18n.FlagDryRunUsage),
	)
}

// AddFilenameFlag adds a filename flag for the provided resource type.
func AddFilenameFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.FilenameFlag
	cmd.Flags().StringP(
		key,
		constants.FilenameShortFlag,
		"",
		i18n.Format(i18n.FlagFilenameUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddDataplaneFlag adds a dataplane flag for the provided resource type.
func AddDataplaneFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.DataplaneFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagDataplaneUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddNodepoolFlag adds a nodepool flag for the provided resource type.
func AddNodepoolFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.NodepoolFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagNodepoolUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddNameFlag adds a name flag for the provided resource type.
func AddNameFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.NameFlag
	cmd.Flags().StringP(
		key,
		constants.NameShortFlag,
		"",
		i18n.Format(i18n.FlagNameUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddDescriptionFlag adds a description flag for the provided resource type.
func AddDescriptionFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.DescriptionFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagDescriptionUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddContextFlag adds a context flag for the provided resource type.
func AddContextFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.ContextFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagContextUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddLabelFlag adds a label flag for the provided resource type.
func AddLabelFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.LabelFlag
	cmd.Flags().StringSliceP(
		key,
		constants.LabelShortFlag,
		[]string{},
		i18n.Format(i18n.FlagLabelUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddCommandFlag adds the command flag to the command.
func AddCommandFlag(cmd *cobra.Command) {
	key := constants.CommandFlag
	cmd.Flags().StringP(
		key,
		constants.CommandShortFlag,
		"",
		i18n.Format(i18n.FlagCommandUsage),
	)
}

// AddSubPathFlag adds a subpath flag for the provided resource type.
func AddSubPathFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.SubPathFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSubPathUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddScheduleFlag adds the schedule flag to the command.
func AddScheduleFlag(cmd *cobra.Command) {
	key := constants.ScheduleFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagScheduleUsage),
	)
}

// AddVolumeFlag adds workspace and volume flags for the provided resource
// type.
func AddVolumeFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.VolumeFlag
	cmd.Flags().StringSlice(
		key,
		[]string{},
		i18n.Format(i18n.FlagVolumeUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
	key = constants.WorkspaceFlag
	cmd.Flags().StringSliceP(
		key,
		constants.WorkspaceShortFlag,
		[]string{},
		i18n.Format(i18n.FlagWorkspaceUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
	_ = cmd.Flags().MarkDeprecated(key, i18n.Format(i18n.FlagWorkspaceDeprecationUsage))
}

// AddApplicationFlag adds an application flag for the provided resource type.
func AddApplicationFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.ApplicationFlag
	cmd.Flags().StringP(
		key,
		constants.ApplicationShortFlag,
		"",
		i18n.Format(i18n.FlagApplicationUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddServiceFlag adds a service flag for the provided resource type.
func AddServiceFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.ServiceFlag
	cmd.Flags().StringSliceP(
		key,
		constants.ServiceShortFlag,
		[]string{},
		i18n.Format(i18n.FlagServiceUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddEnvironmentFlag adds an environment flag for the provided resource type.
func AddEnvironmentFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.EnvironmentFlag
	cmd.Flags().StringSlice(
		key,
		[]string{},
		i18n.Format(i18n.FlagEnvironmentUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddInitializerFlag adds the initializer flag to the command.
func AddInitializerFlag(cmd *cobra.Command) {
	key := constants.InitializerFlag
	cmd.Flags().StringSlice(
		key,
		[]string{},
		i18n.Format(i18n.FlagInitializerUsage),
	)
}

// AddJobIDFlag adds the job ID flag to the command.
func AddJobIDFlag(cmd *cobra.Command) {
	key := constants.JobIDFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagJobIDUsage),
	)
}

// AddDesiredStateFlag adds a desired state flag for the provided resource type.
func AddDesiredStateFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.DesiredStateFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagDesiredStateUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddStorageClassFlag adds the storage class flag to the command.
func AddStorageClassFlag(cmd *cobra.Command) {
	key := constants.StorageClassFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagStorageClassUsage),
	)
}

// AddSizeFlag adds the size flag to the command.
func AddSizeFlag(cmd *cobra.Command) {
	key := constants.SizeFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSizeUsage),
	)
}

// AddHomeDirFlag adds the home directory flag to the command.
func AddHomeDirFlag(cmd *cobra.Command) {
	key := constants.HomeDirFlag
	cmd.Flags().Bool(
		key,
		false,
		i18n.Format(i18n.FlagHomeDirUsage),
	)
}

// AddSnapshotFlag adds the snapshot flag to the command.
func AddSnapshotFlag(cmd *cobra.Command) {
	key := constants.SnapshotFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSnapshotUsage),
	)
}

// AddSnapshotVolumeFlag adds snapshot workspace and volume flags to the
// command.
func AddSnapshotVolumeFlag(cmd *cobra.Command) {
	key := constants.VolumeFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSnapshotVolumeUsage),
	)
	key = constants.WorkspaceFlag
	cmd.Flags().StringP(
		key,
		constants.WorkspaceShortFlag,
		"",
		i18n.Format(i18n.FlagSnapshotWorkspaceUsage),
	)
	_ = cmd.Flags().MarkDeprecated(key, i18n.Format(i18n.FlagWorkspaceDeprecationUsage))
}

// AddImageFlag adds the image flag to the command.
func AddImageFlag(cmd *cobra.Command) {
	key := constants.ImageFlag
	cmd.Flags().StringP(
		key,
		constants.ImageShortFlag,
		"",
		i18n.Format(i18n.FlagImageUsage),
	)
}

// AddSpecFlag adds a spec flag for the provided resource type.
func AddSpecFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.SpecFlag
	cmd.Flags().StringP(
		key,
		constants.SpecShortFlag,
		"",
		i18n.Format(i18n.FlagSpecUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddDataFlag adds a data flag for the provided resource type.
func AddDataFlag(
	cmd *cobra.Command,
	_type string,
) {
	key := constants.DataFlag
	cmd.Flags().StringP(
		key,
		constants.DataShortFlag,
		"",
		i18n.Format(i18n.FlagDataUsage, i18n.A{constants.I18nArgTypeKey: _type}),
	)
}

// AddMethodFlag adds the HTTP method flag to the command.
func AddMethodFlag(cmd *cobra.Command) {
	key := constants.MethodFlag
	cmd.Flags().StringP(
		key,
		"X",
		"GET",
		i18n.Format(i18n.FlagMethodUsage),
	)
}

// AddQueryFlag adds the query flag to the command.
func AddQueryFlag(cmd *cobra.Command) {
	key := constants.QueryFlag
	cmd.Flags().StringP(
		key,
		constants.QueryShortFlag,
		"",
		i18n.Format(i18n.FlagQueryUsage),
	)
}

// AddSetFlag adds the set flag to the command.
func AddSetFlag(cmd *cobra.Command) {
	key := constants.SetFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagSetUsage),
	)
}

// AddValueFlag adds the value flag to the command.
func AddValueFlag(cmd *cobra.Command) {
	key := constants.ValueFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagValueUsage),
	)
}

// AddTenantFlag adds the tenant flag to the command.
func AddTenantFlag(cmd *cobra.Command) {
	key := constants.TenantFlag
	cmd.Flags().Bool(
		key,
		false,
		i18n.Format(i18n.FlagTenantUsage),
	)
}

// AddUserIDFlag adds the user-id flag to the command.
func AddUserIDFlag(cmd *cobra.Command) {
	key := constants.UserIDFlag
	cmd.Flags().String(
		key,
		"",
		i18n.Format(i18n.FlagUserIDUsage),
	)
}

// AddAuthModeFlag adds the persistent auth mode flag to the command.
func AddAuthModeFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(
		constants.AuthModeFlag,
		constants.AuthModeShortFlag,
		constants.AuthModeDevice,
		i18n.Format(i18n.FlagAuthModeUsage),
	)
}

// AddEndpointFlag adds the persistent endpoint flag to the command.
func AddEndpointFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(
		constants.EndpointFlag,
		constants.EndpointShortFlag,
		"",
		i18n.Format(i18n.FlagEndpointUsage),
	)
}

// AddIgnoreTokenCacheFlag adds the ignore token cache flag to the command.
func AddIgnoreTokenCacheFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(
		constants.IgnoreTokenCacheFlag,
		false,
		i18n.Format(i18n.FlagIgnoreTokenCacheFlagUsage),
	)
}

// AddNoListenFlag adds the persistent no-browser flag to the command.
func AddNoListenFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(
		constants.NoListenFlag,
		false,
		i18n.Format(i18n.FlagNoListenUsage),
	)
}

// AddIngressHostFlag adds the persistent ingress-host flag to the command.
func AddIngressHostFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String(
		constants.IngressHostFlag,
		"",
		i18n.Format(i18n.FlagIngressHostUsage),
	)
	// MarkHidden can only fail if the flag does not exist, but we just added
	// the flag so the error can be ignored.
	_ = cmd.PersistentFlags().MarkHidden(constants.IngressHostFlag)
}

// AddOrganizationFlag adds the persistent ingress-host flag to the command.
func AddOrganizationFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String(
		constants.OrganizationFlag,
		"",
		i18n.Format(i18n.FlagOrganizationUsage),
	)
}

// AddClientIDFlag adds the persistent client-id flag to the command.
func AddClientIDFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String(
		constants.ClientIDFlag,
		"",
		i18n.Format(i18n.FlagClientIDUsage),
	)
}

// BindToViper binds all the flags to viper. This cannot be done when the flags
// are created because multiple commands have the same flags and viper maintains
// a flat list. This means that `viper.GetString("name")` will look in the last
// `--name` flag that was defined, not the name flag for the current command.
func BindToViper(
	cmd *cobra.Command,
	_ []string,
) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// viper.BindPFlag can only return an error if flag is nil so we catch
		// that condition here and ignore the error.
		if flag != nil {
			_ = viper.BindPFlag(flag.Name, flag)
		}
	})
	// BindEnv can only fail if given zero arguments.
	_ = viper.BindEnv(constants.ClientIDFlag, constants.ClientIDEnvVar)
}
