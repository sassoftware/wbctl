// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package edit

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/osutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sas-institute-rnd-saas/wbctl/internal/wbctljson"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the get
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(editFunctions)
}

type editorF func(
	context.Context,
	string,
	afero.Fs,
	*workbench.WorkbenchService,
) error

var editFunctions = []types.FunctionMetadata[editorF]{
	{
		Singular: constants.ImageType,
		F: editImpl[types.Image, types.Images](
			workbench.GetImages,
			workbench.BulkUpdateImages,
			types.AsImageSlice,
		),
	},
	{
		Singular: constants.JobType,
		F: editImpl[types.Job, types.Jobs](
			workbench.GetJobs,
			workbench.BulkUpdateJobs,
			types.AsJobSlice,
		),
	},
	{
		Singular: constants.SnapshotType,
		F: editImpl[types.Snapshot, types.Snapshots](
			workbench.GetSnapshots,
			workbench.BulkUpdateSnapshots,
			types.AsSnapshotSlice,
		),
	},
	{
		Singular: constants.WorkbenchType,
		F: editImpl[types.Workbench, types.Workbenches](
			workbench.GetWorkbenches,
			workbench.BulkUpdateWorkbenches,
			types.AsWorkbenchSlice,
		),
	},
	{
		Singular: constants.VolumeType,
		F: editImpl[types.Workspace, types.Workspaces](
			workbench.GetWorkspaces,
			workbench.BulkUpdateWorkspaces,
			types.AsWorkspaceSlice,
		),
	},
}

// newEditFunction creates a new cli.ActionFunc for the edit command.
func newEditFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		editor, ok := types.GetFunction(resourceType, editFunctions)
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
		ids := args
		if len(ids) == 0 {
			msg := i18n.Format(i18n.IDRequired)
			return errors.New(msg)
		}
		if len(ids) > 1 {
			msg := i18n.Format(i18n.OnlyOneIDSupported)
			return errors.New(msg)
		}
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		return editor(ctx, ids[0], fsys, svc)
	}
}

type resourceGetter[TS any] func(
	ctx context.Context,
	s *workbench.WorkbenchService,
	ids []string,
	sortSpec string,
) (TS, error)

type resourceUpdater[T, TS any] func(
	ctx context.Context,
	s *workbench.WorkbenchService,
	ids []string,
	body T,
) (TS, error)

type slicer[T, TS any] func(slice TS) []T

func editImpl[T, TS any](
	getter resourceGetter[TS],
	updater resourceUpdater[T, TS],
	slicer slicer[T, TS],
) editorF {
	return func(
		ctx context.Context,
		id string,
		fsys afero.Fs,
		svc *workbench.WorkbenchService,
	) error {
		resources, err := getter(ctx, svc, []string{id}, "")
		if err != nil {
			return err
		}
		modifiedResource, err := edit(ctx, slicer, fsys, resources)
		if err != nil {
			return err
		}
		_, err = updater(ctx, svc, []string{id}, modifiedResource)
		return err
	}
}

func edit[T, TS any](
	ctx context.Context,
	slicer slicer[T, TS],
	fsys afero.Fs,
	resources TS,
) (T, error) {
	// Get the first resource.
	resourceSlice := slicer(resources)
	originalResource, err := sliceutil.GetAndEnsureOnlyOne(resourceSlice)
	if err != nil {
		var t T
		return t, err
	}
	// Marshal it to JSON
	originalJSON, err := wbctljson.MakeMarshaler[T](wbctljson.EditOptions)(originalResource)
	if err != nil {
		var t T
		return t, err
	}

	// Write JSON to tmpfile
	fs := afero.Afero{Fs: fsys}
	tmpFileName, err := writeToTmpfile(fs, originalJSON)
	if err != nil {
		var t T
		return t, err
	}
	defer func() {
		_ = fs.Remove(tmpFileName)
	}()

	// Open default editor
	err = editFile(ctx, tmpFileName)
	if err != nil {
		var t T
		return t, err
	}

	// Load and return edited file
	return loadEditedFile[T](fs, originalJSON, tmpFileName)
}

func getEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = osutil.DefaultEditor()
	}
	return exec.LookPath(editor)
}

func cleanJSON[T any](originalJSON []byte) []byte {
	var t T
	err := wbctljson.MakeUnmarshaler[T]()(bytes.NewReader(originalJSON), &t)
	if err != nil {
		return originalJSON
	}
	washedJSON, err := wbctljson.MakeMarshaler[T](wbctljson.EditOptions)(t)
	if err != nil {
		return originalJSON
	}
	return washedJSON
}

func writeToTmpfile(
	fs afero.Afero,
	originalJSON []byte,
) (string, error) {
	f, err := fs.TempFile("", "wbctl.edit.*")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	err = fs.WriteFile(f.Name(), originalJSON, 0644)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func editFile(
	ctx context.Context,
	fileName string,
) error {
	editor, err := getEditor()
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, editor, fileName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Start()
	if err != nil {
		msg := i18n.Format(i18n.FailedToStartEditor)
		return errors.Wrap(err, msg)
	}
	err = cmd.Wait()
	if err != nil {
		msg := i18n.Format(i18n.FailedToWaitForEditor)
		return errors.Wrap(err, msg)
	}
	return nil
}

func loadEditedFile[T any](
	fs afero.Afero,
	originalJSON []byte,
	fileName string,
) (T, error) {
	var t T
	updatedJSON, err := fs.ReadFile(fileName)
	if err != nil {
		return t, err
	}
	if len(bytes.TrimSpace(updatedJSON)) == 0 {
		return t, errors.New(i18n.Format(i18n.EditCanceledFileEmpty))
	}
	// wash edited JSON so that whitespace changes are ignored.
	updatedJSON = cleanJSON[T](updatedJSON)
	if slices.Equal(bytes.TrimSpace(updatedJSON), bytes.TrimSpace(originalJSON)) {
		return t, errors.New(i18n.Format(i18n.EditCanceledFileUnchanged))
	}
	err = wbctljson.MakeUnmarshaler[T]()(bytes.NewReader(updatedJSON), &t)
	return t, err
}
