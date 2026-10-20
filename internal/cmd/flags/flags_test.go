// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package flags_test

import (
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd/flags"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func withType(
	f func(
		*cobra.Command,
		string,
	),
	_type string,
) func(*cobra.Command) {
	return func(cmd *cobra.Command) {
		f(cmd, _type)
	}
}

func TestExportedFlags(t *testing.T) {
	tests := []struct {
		name          string
		call          func(*cobra.Command)
		args          []string
		viperKey      string
		expectedValue any
	}{
		{
			name:          "OutputFlag",
			call:          flags.AddOutputFlag,
			args:          []string{"--" + constants.OutputFlag, "json"},
			viperKey:      constants.OutputFlag,
			expectedValue: "json",
		},
		{
			name:          "OutputFlagDefault",
			call:          flags.AddOutputFlag,
			viperKey:      constants.OutputFlag,
			expectedValue: "text",
		},
		{
			name:          "SortFlag",
			call:          flags.AddSortFlag,
			args:          []string{"--" + constants.SortFlag, "sort-value"},
			viperKey:      constants.SortFlag,
			expectedValue: "sort-value",
		},
		{
			name:          "SortFlagDefault",
			call:          flags.AddSortFlag,
			viperKey:      constants.SortFlag,
			expectedValue: "",
		},
		{
			name:          "DryRunFlag",
			call:          flags.AddDryRunFlag,
			args:          []string{"--" + constants.DryRunFlag},
			viperKey:      constants.DryRunFlag,
			expectedValue: true,
		},
		{
			name:          "DryRunFlagDefault",
			call:          flags.AddDryRunFlag,
			viperKey:      constants.DryRunFlag,
			expectedValue: false,
		},
		{
			name:          "FilenameFlag",
			call:          withType(flags.AddFilenameFlag, constants.JobType),
			args:          []string{"--" + constants.FilenameFlag, "filename-value"},
			viperKey:      constants.FilenameFlag,
			expectedValue: "filename-value",
		},
		{
			name:          "FilenameFlagDefault",
			call:          withType(flags.AddFilenameFlag, constants.JobType),
			viperKey:      constants.FilenameFlag,
			expectedValue: "",
		},
		{
			name:          "DataplaneFlag",
			call:          withType(flags.AddDataplaneFlag, constants.JobType),
			args:          []string{"--" + constants.DataplaneFlag, "dataplane-value"},
			viperKey:      constants.DataplaneFlag,
			expectedValue: "dataplane-value",
		},
		{
			name:          "DataplaneFlagDefault",
			call:          withType(flags.AddDataplaneFlag, constants.JobType),
			viperKey:      constants.DataplaneFlag,
			expectedValue: "",
		},
		{
			name:          "NodepoolFlag",
			call:          withType(flags.AddNodepoolFlag, constants.JobType),
			args:          []string{"--" + constants.NodepoolFlag, "nodepool-value"},
			viperKey:      constants.NodepoolFlag,
			expectedValue: "nodepool-value",
		},
		{
			name:          "NodepoolFlagDefault",
			call:          withType(flags.AddNodepoolFlag, constants.JobType),
			viperKey:      constants.NodepoolFlag,
			expectedValue: "",
		},
		{
			name:          "NameFlag",
			call:          withType(flags.AddNameFlag, constants.JobType),
			args:          []string{"--" + constants.NameFlag, "name-value"},
			viperKey:      constants.NameFlag,
			expectedValue: "name-value",
		},
		{
			name:          "NameFlagDefault",
			call:          withType(flags.AddNameFlag, constants.JobType),
			viperKey:      constants.NameFlag,
			expectedValue: "",
		},
		{
			name:          "DescriptionFlag",
			call:          withType(flags.AddDescriptionFlag, constants.JobType),
			args:          []string{"--" + constants.DescriptionFlag, "description-value"},
			viperKey:      constants.DescriptionFlag,
			expectedValue: "description-value",
		},
		{
			name:          "DescriptionFlagDefault",
			call:          withType(flags.AddDescriptionFlag, constants.JobType),
			viperKey:      constants.DescriptionFlag,
			expectedValue: "",
		},
		{
			name:          "ContextFlag",
			call:          withType(flags.AddContextFlag, constants.JobType),
			args:          []string{"--" + constants.ContextFlag, "context-value"},
			viperKey:      constants.ContextFlag,
			expectedValue: "context-value",
		},
		{
			name:          "ContextFlagDefault",
			call:          withType(flags.AddContextFlag, constants.JobType),
			viperKey:      constants.ContextFlag,
			expectedValue: "",
		},
		{
			name: "LabelFlag",
			call: withType(flags.AddLabelFlag, constants.JobType),
			args: []string{"--" + constants.LabelFlag, "label-value1",
				"--" + constants.LabelFlag, "label-value2",
				"--" + constants.LabelFlag, "label-value3"},
			viperKey:      constants.LabelFlag,
			expectedValue: []string{"label-value1", "label-value2", "label-value3"},
		},
		{
			name:          "LabelFlagDefault",
			call:          withType(flags.AddLabelFlag, constants.JobType),
			viperKey:      constants.LabelFlag,
			expectedValue: []string{},
		},
		{
			name:          "CommandFlag",
			call:          flags.AddCommandFlag,
			args:          []string{"--" + constants.CommandFlag, "command-value"},
			viperKey:      constants.CommandFlag,
			expectedValue: "command-value",
		},
		{
			name:          "CommandFlagDefault",
			call:          flags.AddCommandFlag,
			viperKey:      constants.CommandFlag,
			expectedValue: "",
		},
		{
			name:          "SubPathFlag",
			call:          withType(flags.AddSubPathFlag, constants.JobType),
			args:          []string{"--" + constants.SubPathFlag, "subpath-value"},
			viperKey:      constants.SubPathFlag,
			expectedValue: "subpath-value",
		},
		{
			name:          "SubPathFlagDefault",
			call:          withType(flags.AddSubPathFlag, constants.JobType),
			viperKey:      constants.SubPathFlag,
			expectedValue: "",
		},
		{
			name:          "ScheduleFlag",
			call:          flags.AddScheduleFlag,
			args:          []string{"--" + constants.ScheduleFlag, "schedule-value"},
			viperKey:      constants.ScheduleFlag,
			expectedValue: "schedule-value",
		},
		{
			name:          "ScheduleFlagDefault",
			call:          flags.AddScheduleFlag,
			viperKey:      constants.ScheduleFlag,
			expectedValue: "",
		},
		{
			name: "WorkspaceFlag",
			call: withType(flags.AddVolumeFlag, constants.JobType),
			args: []string{"--" + constants.WorkspaceFlag, "workspace-value1",
				"--" + constants.WorkspaceFlag, "workspace-value2",
				"--" + constants.WorkspaceFlag, "workspace-value3"},
			viperKey:      constants.WorkspaceFlag,
			expectedValue: []string{"workspace-value1", "workspace-value2", "workspace-value3"},
		},
		{
			name:          "WorkspaceFlagDefault",
			call:          withType(flags.AddVolumeFlag, constants.JobType),
			viperKey:      constants.WorkspaceFlag,
			expectedValue: []string{},
		},
		{
			name:          "ApplicationFlag",
			call:          withType(flags.AddApplicationFlag, constants.JobType),
			args:          []string{"--" + constants.ApplicationFlag, "application-value"},
			viperKey:      constants.ApplicationFlag,
			expectedValue: "application-value",
		},
		{
			name:          "ApplicationFlagDefault",
			call:          withType(flags.AddApplicationFlag, constants.JobType),
			viperKey:      constants.ApplicationFlag,
			expectedValue: "",
		},
		{
			name: "ServiceFlag",
			call: withType(flags.AddServiceFlag, constants.JobType),
			args: []string{"--" + constants.ServiceFlag, "service-value1",
				"--" + constants.ServiceFlag, "service-value2",
				"--" + constants.ServiceFlag, "service-value3"},
			viperKey:      constants.ServiceFlag,
			expectedValue: []string{"service-value1", "service-value2", "service-value3"},
		},
		{
			name:          "ServiceFlagDefault",
			call:          withType(flags.AddServiceFlag, constants.JobType),
			viperKey:      constants.ServiceFlag,
			expectedValue: []string{},
		},
		{
			name: "EnvironmentFlag",
			call: withType(flags.AddEnvironmentFlag, constants.JobType),
			args: []string{"--" + constants.EnvironmentFlag, "environment-value1",
				"--" + constants.EnvironmentFlag, "environment-value2",
				"--" + constants.EnvironmentFlag, "environment-value3"},
			viperKey:      constants.EnvironmentFlag,
			expectedValue: []string{"environment-value1", "environment-value2", "environment-value3"},
		},
		{
			name:          "EnvironmentFlagDefault",
			call:          withType(flags.AddEnvironmentFlag, constants.JobType),
			viperKey:      constants.EnvironmentFlag,
			expectedValue: []string{},
		},
		{
			name: "InitializerFlag",
			call: flags.AddInitializerFlag,
			args: []string{"--" + constants.InitializerFlag, "initializer-value1",
				"--" + constants.InitializerFlag, "initializer-value2",
				"--" + constants.InitializerFlag, "initializer-value3"},
			viperKey:      constants.InitializerFlag,
			expectedValue: []string{"initializer-value1", "initializer-value2", "initializer-value3"},
		},
		{
			name:          "InitializerFlagDefault",
			call:          flags.AddInitializerFlag,
			viperKey:      constants.InitializerFlag,
			expectedValue: []string{},
		},
		{
			name:          "JobIDFlag",
			call:          flags.AddJobIDFlag,
			args:          []string{"--" + constants.JobIDFlag, "jobid-value"},
			viperKey:      constants.JobIDFlag,
			expectedValue: "jobid-value",
		},
		{
			name:          "JobIDFlagDefault",
			call:          flags.AddJobIDFlag,
			viperKey:      constants.JobIDFlag,
			expectedValue: "",
		},
		{
			name:          "DesiredStateFlag",
			call:          withType(flags.AddDesiredStateFlag, constants.JobType),
			args:          []string{"--" + constants.DesiredStateFlag, "desiredstate-value"},
			viperKey:      constants.DesiredStateFlag,
			expectedValue: "desiredstate-value",
		},
		{
			name:          "StorageClassFlag",
			call:          flags.AddStorageClassFlag,
			args:          []string{"--" + constants.StorageClassFlag, "storageclass-value"},
			viperKey:      constants.StorageClassFlag,
			expectedValue: "storageclass-value",
		},
		{
			name:          "StorageClassFlagDefault",
			call:          flags.AddStorageClassFlag,
			viperKey:      constants.StorageClassFlag,
			expectedValue: "",
		},
		{
			name:          "SizeFlag",
			call:          flags.AddSizeFlag,
			args:          []string{"--" + constants.SizeFlag, "size-value"},
			viperKey:      constants.SizeFlag,
			expectedValue: "size-value",
		},
		{
			name:          "SizeFlag",
			call:          flags.AddSizeFlag,
			viperKey:      constants.SizeFlag,
			expectedValue: "",
		},
		{
			name:          "HomeDirFlag",
			call:          flags.AddHomeDirFlag,
			args:          []string{"--" + constants.HomeDirFlag},
			viperKey:      constants.HomeDirFlag,
			expectedValue: true,
		},
		{
			name:          "HomeDirFlagDefault",
			call:          flags.AddHomeDirFlag,
			viperKey:      constants.HomeDirFlag,
			expectedValue: false,
		},
		{
			name:          "SnapshotFlag",
			call:          flags.AddSnapshotFlag,
			args:          []string{"--" + constants.SnapshotFlag, "snapshot-value"},
			viperKey:      constants.SnapshotFlag,
			expectedValue: "snapshot-value",
		},
		{
			name:          "SnapshotFlagDefault",
			call:          flags.AddSnapshotFlag,
			viperKey:      constants.SnapshotFlag,
			expectedValue: "",
		},
		{
			name:          "SnapshotWorkspaceFlag",
			call:          flags.AddSnapshotVolumeFlag,
			args:          []string{"--" + constants.WorkspaceFlag, "snapshot-value"},
			viperKey:      constants.WorkspaceFlag,
			expectedValue: "snapshot-value",
		},
		{
			name:          "SnapshotWorkspaceFlagDefault",
			call:          flags.AddSnapshotVolumeFlag,
			viperKey:      constants.WorkspaceFlag,
			expectedValue: "",
		},
		{
			name:          "ImageFlag",
			call:          flags.AddImageFlag,
			args:          []string{"--" + constants.ImageFlag, "image-value"},
			viperKey:      constants.ImageFlag,
			expectedValue: "image-value",
		},
		{
			name:          "ImageFlagDefault",
			call:          flags.AddImageFlag,
			viperKey:      constants.ImageFlag,
			expectedValue: "",
		},
		{
			name:          "SpecFlag",
			call:          withType(flags.AddSpecFlag, constants.JobType),
			args:          []string{"--" + constants.SpecFlag, "spec-value"},
			viperKey:      constants.SpecFlag,
			expectedValue: "spec-value",
		},
		{
			name:          "SpecFlagDefault",
			call:          withType(flags.AddSpecFlag, constants.JobType),
			viperKey:      constants.SpecFlag,
			expectedValue: "",
		},
		{
			name:          "DataFlag",
			call:          withType(flags.AddDataFlag, constants.JobType),
			args:          []string{"--" + constants.DataFlag, "data-value"},
			viperKey:      constants.DataFlag,
			expectedValue: "data-value",
		},
		{
			name:          "DataFlagDefault",
			call:          withType(flags.AddDataFlag, constants.JobType),
			viperKey:      constants.DataFlag,
			expectedValue: "",
		},
		{
			name:          "MethodFlag",
			call:          flags.AddMethodFlag,
			args:          []string{"--" + constants.MethodFlag, "method-value"},
			viperKey:      constants.MethodFlag,
			expectedValue: "method-value",
		},
		{
			name:          "MethodFlagDefault",
			call:          flags.AddMethodFlag,
			viperKey:      constants.MethodFlag,
			expectedValue: "GET",
		},
		{
			name:          "QueryFlag",
			call:          flags.AddQueryFlag,
			args:          []string{"--" + constants.QueryFlag, "query-value"},
			viperKey:      constants.QueryFlag,
			expectedValue: "query-value",
		},
		{
			name:          "QueryFlagDefault",
			call:          flags.AddQueryFlag,
			viperKey:      constants.QueryFlag,
			expectedValue: "",
		},
		{
			name:          "SetFlag",
			call:          flags.AddSetFlag,
			args:          []string{"--" + constants.SetFlag, "set-value"},
			viperKey:      constants.SetFlag,
			expectedValue: "set-value",
		},
		{
			name:          "SetFlagDefault",
			call:          flags.AddSetFlag,
			viperKey:      constants.SetFlag,
			expectedValue: "",
		},
		{
			name:          "ValueFlag",
			call:          flags.AddValueFlag,
			args:          []string{"--" + constants.ValueFlag, "value-value"},
			viperKey:      constants.ValueFlag,
			expectedValue: "value-value",
		},
		{
			name:          "ValueFlagDefault",
			call:          flags.AddValueFlag,
			viperKey:      constants.ValueFlag,
			expectedValue: "",
		},
		{
			name:          "TenantFlag",
			call:          flags.AddTenantFlag,
			args:          []string{"--" + constants.TenantFlag},
			viperKey:      constants.TenantFlag,
			expectedValue: true,
		},
		{
			name:          "TenantFlagDefault",
			call:          flags.AddTenantFlag,
			viperKey:      constants.TenantFlag,
			expectedValue: false,
		},
		{
			name:          "UserIDFlag",
			call:          flags.AddUserIDFlag,
			args:          []string{"--" + constants.UserIDFlag, "user1"},
			viperKey:      constants.UserIDFlag,
			expectedValue: "user1",
		},
		{
			name:          "UserIDFlagDefault",
			call:          flags.AddUserIDFlag,
			viperKey:      constants.UserIDFlag,
			expectedValue: "",
		},
		{
			name:          "AuthModeFlag",
			call:          flags.AddAuthModeFlag,
			args:          []string{"--" + constants.AuthModeFlag, "authmode-value"},
			viperKey:      constants.AuthModeFlag,
			expectedValue: "authmode-value",
		},
		{
			name:          "AuthModeFlagDefault",
			call:          flags.AddAuthModeFlag,
			viperKey:      constants.AuthModeFlag,
			expectedValue: "device",
		},
		{
			name:          "EndpointFlag",
			call:          flags.AddEndpointFlag,
			args:          []string{"--" + constants.EndpointFlag, "endpoint-value"},
			viperKey:      constants.EndpointFlag,
			expectedValue: "endpoint-value",
		},
		{
			name:          "EndpointFlagDefault",
			call:          flags.AddEndpointFlag,
			viperKey:      constants.EndpointFlag,
			expectedValue: "",
		},
		{
			name:          "IgnoreTokenCacheFlag",
			call:          flags.AddIgnoreTokenCacheFlag,
			args:          []string{"--" + constants.IgnoreTokenCacheFlag},
			viperKey:      constants.IgnoreTokenCacheFlag,
			expectedValue: true,
		},
		{
			name:          "IgnoreTokenCacheFlagDefault",
			call:          flags.AddIgnoreTokenCacheFlag,
			viperKey:      constants.IgnoreTokenCacheFlag,
			expectedValue: false,
		},
		{
			name:          "NoListenFlag",
			call:          flags.AddNoListenFlag,
			args:          []string{"--" + constants.NoListenFlag},
			viperKey:      constants.NoListenFlag,
			expectedValue: true,
		},
		{
			name:          "NoListenFlagDefault",
			call:          flags.AddNoListenFlag,
			viperKey:      constants.NoListenFlag,
			expectedValue: false,
		},
		{
			name:          "IngressHostFlag",
			call:          flags.AddIngressHostFlag,
			args:          []string{"--" + constants.IngressHostFlag, "ingress-host-value"},
			viperKey:      constants.IngressHostFlag,
			expectedValue: "ingress-host-value",
		},
		{
			name:          "IngressHostFlagDefault",
			call:          flags.AddIngressHostFlag,
			viperKey:      constants.IngressHostFlag,
			expectedValue: "",
		},
		{
			name:          "OrganizationFlag",
			call:          flags.AddOrganizationFlag,
			args:          []string{"--" + constants.OrganizationFlag, "organization-value"},
			viperKey:      constants.OrganizationFlag,
			expectedValue: "organization-value",
		},
		{
			name:          "OrganizationFlagDefault",
			call:          flags.AddOrganizationFlag,
			viperKey:      constants.OrganizationFlag,
			expectedValue: "",
		},
		{
			name:          "ClientIDFlag",
			call:          flags.AddClientIDFlag,
			args:          []string{"--" + constants.ClientIDFlag, "client-id-value"},
			viperKey:      constants.ClientIDFlag,
			expectedValue: "client-id-value",
		},
		{
			name:          "ClientIDFlagDefault",
			call:          flags.AddClientIDFlag,
			viperKey:      constants.ClientIDFlag,
			expectedValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer viper.Reset()
			mockCmd := &cobra.Command{}
			tt.call(mockCmd)
			mockCmd.SetArgs(tt.args)
			// Execute first so that cobra combines the persistent flags with
			// the command flags so that we can bind them all to viper.
			err := mockCmd.Execute()
			require.NoError(t, err)
			flags.BindToViper(mockCmd, nil)
			require.NoError(t, err)
			actual := viper.Get(tt.viperKey)
			require.Equal(t, tt.expectedValue, actual)
		})
	}
}
