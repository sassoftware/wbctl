// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/jsonutils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/cmd/auth"
	"github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/jsonutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/services/auth0admin"
	"github.com/sassoftware/wbctl/internal/services/workbench"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newAPIFunction creates a new cli.ActionFunc for the delete command.
func newAPIFunction(
	ctx context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) types.ActionFunc {
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
		auth0svc := auth0admin.NewAuth0AdminService(endpoint, accessToken)
		if len(args) == 0 {
			return getAPI(ctx, v, _ui, svc, auth0svc)
		}
		method := v.GetString(constants.MethodFlag)
		data := v.GetString(constants.DataFlag)
		filename := v.GetString(constants.FilenameFlag)
		query, err := url.ParseQuery(v.GetString(constants.QueryFlag))
		if err != nil {
			return err
		}
		body := map[string]interface{}{}
		if data != "" {
			body, err = result.Split(
				jsonutils.UnmarshalAs[map[string]interface{}]([]byte(data)),
			)
		} else if filename != "" {
			body, err = jsonutil.GetFromFileOrStdin[map[string]interface{}](
				fsys,
				filename,
			)
		}
		if err != nil {
			return err
		}
		return callAPI(ctx, _ui, args[0], method, svc, query, body)
	}
}

func getAPI(
	ctx context.Context,
	v *viper.Viper,
	_ui ui.UI,
	svc *workbench.WorkbenchService,
	auth0svc *auth0admin.Auth0AdminService,
) error {
	resources, err := workbench.GetAPI(ctx, svc)
	if err != nil {
		return err
	}
	links, err := auth0admin.GetAPI(ctx, auth0svc)
	if err != nil {
		return err
	}
	resources = addLinksToSwaggerResources(resources, links)

	if constants.JSONOutputValue == v.GetString(constants.OutputFlag) {
		response, err := result.Split(jsonutils.ToJSON(resources))
		if err == nil {
			_ui.Display(response)
		}
		return err
	}
	return workbench.PrintAPI(_ui, resources)
}

func callAPI(
	ctx context.Context,
	_ui ui.UI,
	endpoint string,
	method string,
	svc *workbench.WorkbenchService,
	query url.Values,
	body map[string]interface{},
) error {
	method = strings.ToUpper(method)
	if !strings.HasPrefix(endpoint, constants.Auth0AdminAPIEndpoint) {
		endpoint = postgrest.Endpoint(endpoint)
	}
	var resources any
	var err error
	switch method {
	case http.MethodGet:
		resources, err = result.Split(workbench.Get[any](ctx, svc, endpoint, query))
	case http.MethodPost:
		resources, err = result.Split(workbench.Post[any](ctx, svc, endpoint, body))
	default:
		return errors.New(i18n.Format(i18n.UnsupportedMethod, i18n.A{"method": method}))
	}
	if err != nil {
		return err
	}
	response, err := result.Split(jsonutils.ToJSON(resources))
	if err == nil {
		_ui.Display(response)
	}
	return err
}

func addLinksToSwaggerResources(resources workbench.API, links auth0admin.API) workbench.API {
	for _, link := range links.Links {
		item, present := resources.Paths.Paths[link.URI]
		if !present {
			item = spec.PathItem{
				PathItemProps: spec.PathItemProps{},
			}
		}
		switch link.Method {
		case http.MethodGet:
			item.Get = &spec.Operation{
				OperationProps: spec.OperationProps{
					Summary:     link.Title,
					Description: link.Title,
				},
			}
		case http.MethodPost:
			item.Post = &spec.Operation{
				OperationProps: spec.OperationProps{
					Summary:     link.Title,
					Description: link.Title,
				},
			}
		}
		resources.Paths.Paths[link.URI] = item
	}
	return resources
}
