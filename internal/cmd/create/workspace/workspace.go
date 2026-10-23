// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/wbctl/internal/cmd/auth"
	"github.com/sassoftware/wbctl/internal/cmd/flags"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/jsonutil"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// CreateCommand returns the create workspace Command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use: constants.VolumeCommandName,
		Aliases: []string{
			constants.VolumesCommandName,
			constants.WorkspaceCommandName,
			constants.WorkspacesCommandName,
		},
		Args: cobra.NoArgs,
		RunE: newCreateWorkspaceFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.VolumeType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.VolumeType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddFilenameFlag(cmd, constants.VolumeType)
	flags.AddDryRunFlag(cmd)
	flags.AddDataplaneFlag(cmd, constants.VolumeType)
	flags.AddNameFlag(cmd, constants.VolumeType)
	flags.AddDescriptionFlag(cmd, constants.VolumeType)
	flags.AddStorageClassFlag(cmd)
	flags.AddSizeFlag(cmd)
	flags.AddInitializerFlag(cmd)
	flags.AddHomeDirFlag(cmd)
	flags.AddSnapshotFlag(cmd)
	return cmd
}

// UpdateCommand returns the update workspace Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use: constants.VolumesCommandName,
		Aliases: []string{
			constants.VolumeCommandName,
			constants.WorkspaceCommandName,
			constants.WorkspacesCommandName,
		},
		Args: cobra.ArbitraryArgs,
		RunE: newUpdateWorkspaceFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdUpdateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.VolumeType)},
		),
		Long: i18n.Format(
			i18n.CmdUpdateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.VolumeType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddFilenameFlag(cmd, constants.VolumeType)
	flags.AddDryRunFlag(cmd)
	flags.AddNameFlag(cmd, constants.VolumeType)
	flags.AddDescriptionFlag(cmd, constants.VolumeType)
	flags.AddSizeFlag(cmd)
	flags.AddInitializerFlag(cmd)
	flags.AddHomeDirFlag(cmd)
	return cmd
}

// newCreateWorkspaceFunction creates a new cli.ActionFunc for the `create
// workspace` command.
func newCreateWorkspaceFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		body, err := getCreateWorkspaceRequestBody(ctx, _ui, v, fsys, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := wbctljson.MakeMarshaler[types.Workspace](wbctljson.EditOptions)(body)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.CreateWorkspace(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

// newUpdateWorkspaceFunction creates a new cli.ActionFunc for the `update
// workspace` command.
func newUpdateWorkspaceFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		body, err := getUpdateWorkspaceRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		ids, err := resolveUpdateWorkspaceIDs(args, body)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			payload := makeUpdateWorkspacePayload(ids, body)
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.EditOptions)(payload)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.BulkUpdateWorkspaces(ctx, svc, ids, body)
		if err == nil {
			for _, resource := range res {
				_ui.Display(resource.ID)
			}
		}
		return err
	}
}

func getWorkspaceRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Workspace, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.Workspace](fsys, filename)
	}
	name := v.GetString(constants.NameFlag)
	description := v.GetString(constants.DescriptionFlag)
	storageClass := v.GetString(constants.StorageClassFlag)
	storageSize := v.GetString(constants.SizeFlag)
	snapshot := v.GetString(constants.SnapshotFlag)
	initializers := v.GetStringSlice(constants.InitializerFlag)
	homeDir := v.GetBool(constants.HomeDirFlag)

	serviceRequests, err := workbench.GetInitializerRequests(ctx, svc, initializers)
	if err != nil {
		return types.Workspace{}, err
	}
	wksp := types.Workspace{
		Name:         name,
		Description:  description,
		Initializers: serviceRequests,
		StorageType:  constants.PersistentVolumeClaim,
	}
	if snapshot != "" {
		err = setSnapshot(ctx, svc, snapshot, &wksp)
	} else {
		err = setStorage(ctx, svc, homeDir, &wksp, storageClass, storageSize)
	}
	return wksp, err
}

func getCreateWorkspaceRequestBody(
	ctx context.Context,
	_ui ui.UI,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Workspace, error) {
	err := validateCreateOptions(ctx, _ui, v, svc)
	if err != nil {
		return types.Workspace{}, err
	}
	wksp, err := getWorkspaceRequestBody(ctx, v, fsys, svc)
	if err == nil {
		// Dataplane is not supported for update but is for create. So we have to
		// resolve the dataplane id here instead of in the common implementation
		// function.
		dataplaneID := v.GetString(constants.DataplaneFlag)
		wksp.DataplaneID, err = workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
		if err != nil {
			return types.Workspace{}, err
		}
	}
	return wksp, err
}

func setSnapshot(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	snapshotID string,
	wksp *types.Workspace,
) error {
	snapshots, err := workbench.GetSnapshots(ctx, svc, []string{snapshotID}, "")
	if err != nil {
		return err
	}
	snapshot, err := sliceutil.GetAndEnsureOnlyOne(snapshots)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			i18n.Format(i18n.FailedToAddSnapshot, i18n.A{constants.I18nArgKeyID: snapshotID}),
			err,
		)
	}
	err = addSnapshotStorage(wksp, snapshot)
	if err != nil {
		return fmt.Errorf(
			"%s: %w",
			i18n.Format(i18n.FailedToAddSnapshot, i18n.A{constants.I18nArgKeyID: snapshotID}),
			err,
		)
	}
	return nil
}

func addSnapshotStorage(
	wksp *types.Workspace,
	snapshot types.Snapshot,
) error {
	controllerID := snapshot.ControllerID
	if controllerID == "" || snapshot.VolumeStatus == nil ||
		!pointer.UnPtr(snapshot.VolumeStatus.Status.ReadyToUse) {
		return errors.New(i18n.Format(i18n.SnapshotNotReadyVolumeStatusMissing))
	}
	if snapshot.WorkspaceSource == nil || snapshot.WorkspaceSource.StorageSpec == nil {
		return errors.New(i18n.Format(i18n.SnapshotNotReadySourceStorageMissing))
	}
	size := snapshot.VolumeStatus.Status.RestoreSize
	if size == nil {
		return errors.New(i18n.Format(i18n.SnapshotNotReadyRestoreSizeMissing))
	}
	if len(snapshot.WorkspaceSource.StorageSpec.AccessModes) < 1 {
		return errors.New(i18n.Format(i18n.InvalidSnapshotAccessModeMissing))
	}
	accessMode := snapshot.WorkspaceSource.StorageSpec.AccessModes[0]
	storageClassName := snapshot.WorkspaceSource.StorageSpec.StorageClassName
	wksp.StorageSpec = &corev1.PersistentVolumeClaimSpec{
		AccessModes:      []corev1.PersistentVolumeAccessMode{accessMode},
		StorageClassName: storageClassName,
		Resources: corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: pointer.UnPtr(size),
			},
		},
		DataSource: &corev1.TypedLocalObjectReference{
			Kind:     constants.K8sVolumeSnapshotKind,
			Name:     snapshot.ControllerID,
			APIGroup: pointer.Ptr(constants.K8sVolumeSnapshotAPIGroup),
		},
	}
	return nil
}

func getStorageClass(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	storageClassName string,
) (types.StorageClass, error) {
	if storageClassName == "" {
		return types.StorageClass{}, nil
	}
	storageClasses, err := workbench.GetStorageClasses(ctx, svc, []string{storageClassName}, "")
	if err != nil {
		return types.StorageClass{}, err
	}
	storageClass, err := sliceutil.GetAndEnsureOnlyOne[types.StorageClass](storageClasses)
	if err != nil {
		return types.StorageClass{},
			fmt.Errorf(
				"%s: %w",
				i18n.Format(i18n.StorageClassNotFound, i18n.A{constants.I18nArgKeyName: storageClassName}),
				err,
			)
	}
	return storageClass, nil
}

func getQuantity(storageSizeString string) (resource.Quantity, error) {
	if storageSizeString == "" {
		return resource.Quantity{}, nil
	}
	return resource.ParseQuantity(storageSizeString)
}

func setStorage(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	homeDir bool,
	wksp *types.Workspace,
	storageClassName string,
	storageSizeString string,
) error {
	storageClass, err := getStorageClass(ctx, svc, storageClassName)
	if err != nil {
		return err
	}
	storageQuantity, err := getQuantity(storageSizeString)
	if err != nil {
		return err
	}
	pvcClaimSpec := &corev1.PersistentVolumeClaimSpec{}
	pvcClaimSpecUpdated := false
	if storageClass.AccessMode != "" {
		pvcClaimSpec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.PersistentVolumeAccessMode(storageClass.AccessMode),
		}
		pvcClaimSpecUpdated = true
	}
	if storageClass.ID != "" {
		pvcClaimSpec.StorageClassName = pointer.Ptr(storageClass.ID)
		pvcClaimSpecUpdated = true
	}
	if !storageQuantity.IsZero() {
		pvcClaimSpec.Resources = corev1.VolumeResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: storageQuantity,
			},
		}
		pvcClaimSpecUpdated = true
	}
	if pvcClaimSpecUpdated {
		wksp.StorageSpec = pvcClaimSpec
	}
	if homeDir {
		wksp.Properties = map[string]interface{}{
			constants.HomeDirPropertyKey: "TRUE",
		}
	}
	return nil
}

func getUpdateWorkspaceRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Workspace, error) {
	wksp, err := getWorkspaceRequestBody(ctx, v, fsys, svc)
	if err == nil {
		// Storage Type is never set on update. It is not one of the command line
		// options so we should never change it.
		wksp.StorageType = ""
	}
	return wksp, err
}

func resolveUpdateWorkspaceIDs(
	args []string,
	body types.Workspace,
) ([]string, error) {
	ids := args
	if len(ids) > 0 {
		return ids, nil
	}
	if body.ID == "" {
		return nil, errors.New(
			i18n.Format(
				i18n.AtLeastOneResourceIDForCommand,
				i18n.A{constants.I18nArgKeyCommandName: constants.UpdateCommandName},
			),
		)
	}
	return []string{body.ID}, nil
}

func makeUpdateWorkspacePayload(
	ids []string,
	body types.Workspace,
) map[string]interface{} {
	return map[string]interface{}{
		constants.RPCBodyKeyIDs:       ids,
		constants.RPCBodyKeyWorkspace: body,
	}
}

// nolint:cyclop
func validateCreateOptions(
	ctx context.Context,
	_ui ui.UI,
	v *viper.Viper,
	svc *workbench.WorkbenchService,
) error {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return nil
	}
	name := v.GetString(constants.NameFlag)
	storageClass := v.GetString(constants.StorageClassFlag)
	storageSize := v.GetString(constants.SizeFlag)
	snapshot := v.GetString(constants.SnapshotFlag)
	if name == "" {
		return errors.New(
			i18n.Format(
				i18n.FlagIsRequired,
				i18n.A{constants.I18nArgKeyFlagName: constants.NameFlag},
			),
		)
	}
	// If we are not using a snapshot and the storage class was not specified,
	// then find the default storage class. Ignore any error.
	if snapshot == "" && storageClass == "" {
		scs, err := workbench.GetStorageClasses(ctx, svc, nil, "")
		if err == nil {
			for _, sc := range scs {
				if sc.Properties[constants.DefaultKey] == true {
					storageClass = sc.ID
					v.Set(constants.StorageClassFlag, storageClass)
					warning := i18n.Format(i18n.WarningDefaultStorageClass, i18n.A{"storageclass": storageClass})
					logutil.GetLogger().Warn(warning)
					_ui.Warn(warning)
					break
				}
			}
		}
	}
	if (storageClass != "" && snapshot != "") || (storageClass == "" && snapshot == "") {
		return errors.New(i18n.Format(i18n.OneOfFlagNamesAreRequired, i18n.A{
			constants.I18nArgKeyFlagNameOne: constants.SnapshotFlag,
			constants.I18nArgKeyFlagNameTwo: constants.StorageClassFlag,
		}))
	}
	if storageClass != "" && storageSize == "" {
		return errors.New(i18n.Format(i18n.FlagRequiredWhen, i18n.A{
			constants.I18nArgKeyChild:  constants.SizeFlag,
			constants.I18nArgKeyParent: constants.StorageClassFlag,
		}))
	}
	if storageSize != "" && storageClass == "" {
		return errors.New(i18n.Format(i18n.FlagRequiredWhen, i18n.A{
			constants.I18nArgKeyChild:  constants.StorageClassFlag,
			constants.I18nArgKeyParent: constants.SizeFlag,
		}))
	}
	return nil
}
