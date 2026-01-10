// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"errors"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/cmd/auth"
	cmdtypes "github.com/sas-institute-rnd-saas/wbctl/internal/cmd/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/services/workbench"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// SupportedResourceTypes returns the resource types supported by the log
// command.
func SupportedResourceTypes() []string {
	return types.SupportedResourceTypes(logFunctions)
}

type logF func(
	context.Context,
	[]string,
	ui.UI,
	*workbench.WorkbenchService,
) error

var logFunctions = []types.FunctionMetadata[logF]{
	{
		Singular: constants.ImageBuildType,
		Plural:   constants.ImageBuildsType,
		F:        imageBuildLogs,
	},
	{
		Singular: constants.JobSubmissionType,
		Plural:   constants.JobSubmissionsType,
		F:        jobSubmissionLogs,
	},
}

// newLogFunction creates a new cli.ActionFunc for the delete command.
func newLogFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
	resourceType string,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		args []string,
	) error {
		logger, ok := types.GetFunction(resourceType, logFunctions)
		if !ok {
			supportedTypes := types.SupportedResourceTypes(logFunctions)
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
			return errors.New(
				i18n.Format(
					i18n.AtLeastOneResourceIDForCommand,
					i18n.A{constants.I18nArgKeyCommandName: constants.LogCommandName},
				),
			)
		}
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		accessToken, csrfToken, err := auth.GetAccessTokens(ctx, v, _ui, fsys)
		if err != nil {
			return err
		}
		svc := workbench.NewWorkbenchService(endpoint, accessToken, csrfToken)
		return logger(ctx, ids, _ui, svc)
	}
}

func jobSubmissionLogs(
	ctx context.Context,
	ids []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	jobSubmissions, err := workbench.GetJobSubmissions(ctx, svc, ids, "")
	if err != nil {
		return err
	}
	for _, jobSubmission := range jobSubmissions {
		err = printJobSubmissionLog(ctx, svc, _ui, jobSubmission)
		if err != nil {
			return err
		}
	}
	return nil
}

func printJobSubmissionLog(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	_ui ui.UI,
	jobSubmission types.JobSubmission,
) error {
	logURL := jobSubmission.LogURL
	if logURL == "" {
		return nil
	}
	logURL = strings.ReplaceAll(logURL, "/logs/", "/logs/raw/")
	log, err := result.Split(workbench.GetRaw(ctx, svc, logURL))
	if err != nil {
		return err
	}
	_ui.Display(log)
	return nil
}

func imageBuildLogs(
	ctx context.Context,
	ids []string,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
) error {
	imageBuilds, err := workbench.GetImageBuilds(ctx, svc, ids, "")
	if err != nil {
		return err
	}
	for _, imageBuild := range imageBuilds {
		err = printImageBuildLog(ctx, svc, _ui, imageBuild)
		if err != nil {
			return err
		}
	}
	return nil
}

func printImageBuildLog(
	ctx context.Context,
	svc *workbench.WorkbenchService,
	_ui ui.UI,
	imageBuild types.ImageBuild,
) error {
	logURL := imageBuild.LogURL
	if logURL == "" {
		return nil
	}
	logURL = strings.ReplaceAll(logURL, "/logs/", "/logs/raw/")
	log, err := result.Split(workbench.GetRaw(ctx, svc, logURL))
	if err != nil {
		return err
	}
	_ui.Display(log)
	return nil
}
