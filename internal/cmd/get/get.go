// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package get

import (
	"context"
	"errors"
	"strings"

	"github.com/sassoftware/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the get
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(getFunctions)
}

type getterF func(
	context.Context,
	*viper.Viper,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error

var getFunctions = []types.FunctionMetadata[getterF]{
	{
		Singular: constants.ApplicationType,
		Plural:   constants.ApplicationsType,
		F:        getImpl[types.Applications](workbench.GetApplications, workbench.PrintApplications),
	},
	{
		Singular: constants.DataplaneType,
		Plural:   constants.DataplanesType,
		F:        getImpl[types.Dataplanes](workbench.GetDataplanes, workbench.PrintDataplanes),
	},
	{
		Singular: constants.EventType,
		Plural:   constants.EventsType,
		F:        getImpl[types.Events](workbench.GetEvents, workbench.PrintEvents),
	},
	{
		Singular: constants.ImageBuildType,
		Plural:   constants.ImageBuildsType,
		F:        getImpl[types.ImageBuilds](workbench.GetImageBuilds, workbench.PrintImageBuilds),
	},
	{
		Singular: constants.ImageType,
		Plural:   constants.ImagesType,
		F:        getImpl[types.Images](workbench.GetImages, workbench.PrintImages),
	},
	{
		Singular: constants.InitializerType,
		Plural:   constants.InitializersType,
		F:        getImpl[types.Initializers](workbench.GetInitializers, workbench.PrintInitializers),
	},
	{
		Singular: constants.JobSubmissionType,
		Plural:   constants.JobSubmissionsType,
		F:        getImpl[types.JobSubmissions](workbench.GetJobSubmissions, workbench.PrintJobSubmissions),
	},
	{
		Singular: constants.JobType,
		Plural:   constants.JobsType,
		F:        getImpl[types.Jobs](workbench.GetJobs, workbench.PrintJobs),
	},
	{
		Singular: constants.NodepoolType,
		Plural:   constants.NodepoolsType,
		F:        getImpl[types.Nodepools](workbench.GetNodepools, workbench.PrintNodepools),
	},
	{
		Singular: constants.PingType,
		Plural:   constants.PingsType,
		F:        getImpl[types.Pings](workbench.GetPings, workbench.PrintPings),
	},
	{
		Singular: constants.ServiceType,
		Plural:   constants.ServicesType,
		F:        getImpl[types.Services](workbench.GetServices, workbench.PrintServices),
	},
	{
		Singular: constants.SnapshotType,
		Plural:   constants.SnapshotsType,
		F:        getImpl[types.Snapshots](workbench.GetSnapshots, workbench.PrintSnapshots),
	},
	{
		Singular: constants.StorageclassType,
		Plural:   constants.StorageclassesType,
		F:        getImpl[types.StorageClasses](workbench.GetStorageClasses, workbench.PrintStorageClasses),
	},
	{
		Singular: constants.UserType,
		Plural:   constants.UsersType,
		F:        getImpl[types.Auth0UserList](workbench.GetUsers, workbench.PrintUsers),
	},
	{
		Singular: constants.WorkbenchType,
		Plural:   constants.WorkbenchesType,
		F:        getImpl[types.Workbenches](workbench.GetWorkbenches, workbench.PrintWorkbenches),
	},
	{
		Singular: constants.VolumeType,
		Plural:   constants.VolumesType,
		F:        getImpl[types.Workspaces](workbench.GetWorkspaces, workbench.PrintWorkspaces),
	},
	{
		Singular: constants.VolumeConflictType,
		Plural:   constants.VolumeConflictsType,
		F:        getImpl[types.Conflicts](workbench.GetConflicts, workbench.PrintConflicts),
	},
}

// newGetFunction creates a new cli.ActionFunc for the get command.
func newGetFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		getter, ok := types.GetFunction(resourceType, getFunctions)
		if !ok {
			supportedTypes := SupportedResourceTypes()
			msg := i18n.Format(i18n.ResourceTypeIsRequired)
			if resourceType != "" {
				msg = i18n.Format(
					i18n.UnknownResourceType,
					i18n.A{constants.I18nArgKeyType: resourceType},
				)
			}
			supportedTypesMsg := i18n.Format(
				i18n.SupportedTypes,
				i18n.A{constants.I18nArgKeyTypes: strings.Join(supportedTypes, "\n    ")},
			)
			return errors.New(msg + "\n  " + supportedTypesMsg)
		}
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		return getter(ctx, v, args, _ui, svc)
	}
}

type resourceGetter[T any] func(
	context.Context,
	*workbench.WorkbenchService,
	[]string,
	string,
) (T, error)

type resourcePrinter[T any] func(
	ui.UI,
	string,
	T,
) error

func getImpl[T any](
	getter resourceGetter[T],
	printer resourcePrinter[T],
) func(
	context.Context,
	*viper.Viper,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error {
	return func(
		ctx context.Context,
		v *viper.Viper,
		args []string,
		_ui ui.UI,
		svc *workbench.WorkbenchService,
	) error {
		resources, err := getter(ctx, svc, args, v.GetString(constants.SortFlag))
		if err != nil {
			return err
		}
		token := workbench.GetAccessToken(svc)
		if constants.JSONOutputValue == viper.GetString(constants.OutputFlag) {
			response, err := wbctljson.MakeMarshaler[T](wbctljson.ViewOptions)(resources)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		return printer(_ui, token, resources)
	}
}
