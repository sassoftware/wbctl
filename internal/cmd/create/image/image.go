// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package image

import (
	"context"
	"errors"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/flags"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/jsonutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/maputil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sas-institute-rnd-saas/wbctl/internal/wbctljson"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// CreateCommand returns the create image Command.
func CreateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.ImageCommandName,
		Aliases: []string{constants.ImagesCommandName},
		Args:    cobra.NoArgs,
		RunE:    newCreateImageFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdCreateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.ImageType)},
		),
		Long: i18n.Format(
			i18n.CmdCreateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.ImageType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddDryRunFlag(cmd)
	flags.AddFilenameFlag(cmd, constants.ImageType)
	flags.AddDataplaneFlag(cmd, constants.ImageType)
	flags.AddNameFlag(cmd, constants.ImageType)
	flags.AddDescriptionFlag(cmd, constants.ImageType)
	flags.AddContextFlag(cmd, constants.ImageType)
	flags.AddImageFlag(cmd)
	flags.AddLabelFlag(cmd, constants.ImageType)
	flags.AddSpecFlag(cmd, constants.ImageType)
	flags.AddSubPathFlag(cmd, constants.ImageType)
	return cmd
}

// UpdateCommand returns the update image Command.
func UpdateCommand(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:     constants.ImagesCommandName,
		Aliases: []string{constants.ImageCommandName},
		Args:    cobra.ArbitraryArgs,
		RunE:    newUpdateImageFunction(ctx, fsys, _ui),
		Short: i18n.Format(
			i18n.CmdUpdateTypeUsage,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.ImageType)},
		),
		Long: i18n.Format(
			i18n.CmdUpdateTypeDescription,
			i18n.A{constants.I18nArgTypeKey: cmdtypes.Titler.String(constants.ImageType)},
		),
		PreRun: flags.BindToViper,
	}
	flags.AddDryRunFlag(cmd)
	flags.AddFilenameFlag(cmd, constants.ImageType)
	flags.AddNameFlag(cmd, constants.ImageType)
	flags.AddDescriptionFlag(cmd, constants.ImageType)
	flags.AddContextFlag(cmd, constants.ImageType)
	flags.AddImageFlag(cmd)
	flags.AddLabelFlag(cmd, constants.ImageType)
	flags.AddSpecFlag(cmd, constants.ImageType)
	flags.AddSubPathFlag(cmd, constants.ImageType)
	return cmd
}

// newCreateImageFunction creates a new cli.ActionFunc for the `create image`
// command.
func newCreateImageFunction(
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
		body, err := getCreateImageRequestBody(ctx, v, fsys, svc)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			response, err := wbctljson.MakeMarshaler[types.Image](wbctljson.EditOptions)(body)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.CreateImage(ctx, svc, body)
		if err == nil {
			_ui.Display(res.ID)
		}
		return err
	}
}

// newUpdateImageFunction creates a new cli.ActionFunc for the `update image`
// command.
func newUpdateImageFunction(
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
		body, err := getImageRequestBody(v, fsys)
		if err != nil {
			return err
		}
		ids, err := resolveUpdateImageIDs(args, body)
		if err != nil {
			return err
		}
		dryRun := v.GetBool(constants.DryRunFlag)
		if dryRun {
			payload := makeUpdateImagePayload(ids, body)
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.EditOptions)(payload)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		res, err := workbench.BulkUpdateImages(ctx, svc, ids, body)
		if err != nil {
			return err
		}
		for _, resource := range res {
			_ui.Display(resource.ID)
		}
		return nil
	}
}

func getImageRequestBody(
	v *viper.Viper,
	fsys afero.Fs,
) (types.Image, error) {
	filename := v.GetString(constants.FilenameFlag)
	if filename != "" {
		return jsonutil.GetFromFileOrStdin[types.Image](fsys, filename)
	}
	name := v.GetString(constants.NameFlag)
	description := v.GetString(constants.DescriptionFlag)
	imageContext := v.GetString(constants.ContextFlag)
	imageRegistry, imageName, imageTag := parseImageRef(v.GetString(constants.ImageFlag))
	specString := v.GetString(constants.SpecFlag)
	subPath := v.GetString(constants.SubPathFlag)
	labels := v.GetStringSlice(constants.LabelFlag)

	spec, err := parseSpec(fsys, specString)
	if err != nil {
		return types.Image{}, err
	}
	image := types.Image{
		Name:          name,
		Description:   description,
		Context:       imageContext,
		Labels:        maputil.FromEqualDelimitedLines(labels),
		SubPath:       subPath,
		ImageTag:      imageTag,
		ImageRegistry: imageRegistry,
		ImageName:     imageName,
	}
	setSpec(&image, spec)
	return image, nil
}

func getCreateImageRequestBody(
	ctx context.Context,
	v *viper.Viper,
	fsys afero.Fs,
	svc *workbench.WorkbenchService,
) (types.Image, error) {
	image, err := getImageRequestBody(v, fsys)
	if err != nil {
		return types.Image{}, err
	}
	// Dataplane is not supported for update but is for create. So we have to
	// resolve the dataplane id here instead of in the common implementation
	// function.
	dataplaneID := v.GetString(constants.DataplaneFlag)
	image.DataplaneID, err = workbench.GetValidDataplaneID(ctx, svc, dataplaneID)
	if err != nil {
		return types.Image{}, err
	}
	if image.Name == "" {
		return types.Image{},
			errors.New(
				i18n.Format(
					i18n.FlagIsRequired,
					i18n.A{constants.I18nArgKeyFlagName: constants.NameFlag},
				),
			)
	}
	if image.Context == "" && len(image.Spec) == 0 {
		return types.Image{}, errors.New(i18n.Format(i18n.EitherFlagNamesAreRequired, i18n.A{
			constants.I18nArgKeyFlagNameOne: constants.ContextFlag,
			constants.I18nArgKeyFlagNameTwo: constants.SpecFlag,
		}))
	}
	return image, nil
}

func setSpec(
	image *types.Image,
	spec map[string]interface{},
) {
	if len(spec) > 0 {
		image.Spec = spec
	}
}

func parseSpec(
	fsys afero.Fs,
	spec string,
) (map[string]interface{}, error) {
	if spec == "" {
		return map[string]interface{}{}, nil
	}
	_, err := fsys.Stat(spec)
	// If not a file then treat as literal spec
	if err != nil {
		return result.Split(jsonutils.UnmarshalAs[map[string]interface{}]([]byte(spec)))
	}
	// Otherwise load it from disk
	fs := afero.Afero{Fs: fsys}
	return result.Split(jsonutils.LoadWith[map[string]interface{}](fs.ReadFile, spec))
}

func parseImageRef(ref string) (string, string, string) {
	parts := strings.SplitN(ref, "/", 2)
	registry := parts[0]
	name := ""
	tag := ""
	if len(parts) == 2 {
		parts = strings.SplitN(parts[1], ":", 2)
		name = parts[0]
		tag = "latest"
		if len(parts) == 2 {
			tag = parts[1]
		}
	}
	return registry, name, tag
}

func resolveUpdateImageIDs(
	args []string,
	body types.Image,
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

func makeUpdateImagePayload(
	ids []string,
	body types.Image,
) map[string]interface{} {
	return map[string]interface{}{
		constants.RPCBodyKeyIDs:   ids,
		constants.RPCBodyKeyParam: body,
	}
}
