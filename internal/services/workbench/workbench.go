// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/httputils"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/sas-ggdk/pkg/sliceutils"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/httputil"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/maputil"
	"github.com/sassoftware/wbctl/internal/types"
)

var defaultClient = result.Ok(&http.Client{
	Timeout: time.Second * 60,
})

// WorkbenchService is an exported type.
// nolint:revive
type WorkbenchService struct {
	endpoint    string
	accessToken string
	csrfToken   string
}

// NewWorkbenchService returns a new workbench service.
func NewWorkbenchService(endpoint, accessToken, csrfToken string) *WorkbenchService {
	return &WorkbenchService{
		endpoint:    endpoint,
		accessToken: accessToken,
		csrfToken:   csrfToken,
	}
}

// GetAccessToken gets the access token from the given WorkbenchService.
func GetAccessToken(s *WorkbenchService) string {
	return s.accessToken
}

// Get performs an http GET with the given path and query against the given
// WorkbenchService. The response body is marshaled from JSON into the given
// type T and returned as a result.
func Get[T any](ctx context.Context, s *WorkbenchService, path string, query url.Values) result.Result[T] {
	resourceURL := httputil.MakeURL(s.endpoint, path, query)
	request := result.Map(httputil.MakeGetRequestContextFunc(ctx), resourceURL)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey:      "Bearer " + s.accessToken,
			constants.PostgrestCSRFTokenHeaderKey: s.csrfToken,
			constants.CookieHeaderKey:             constants.PostgrestCSRFTokenKey + "=" + s.csrfToken + ";",
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsJSON[T], defaultClient, request)
}

// GetRaw performs an http GET with the given path and query against the given
// WorkbenchService. The response body is marshaled into a string and returned
// as a result.
func GetRaw(ctx context.Context, s *WorkbenchService, rawURL string) result.Result[string] {
	parsedURL := result.New(url.Parse(rawURL))
	request := result.Map(httputil.MakeGetRequestContextFunc(ctx), parsedURL)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey:      "Bearer " + s.accessToken,
			constants.PostgrestCSRFTokenHeaderKey: s.csrfToken,
			constants.CookieHeaderKey:             constants.PostgrestCSRFTokenKey + "=" + s.csrfToken + ";",
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsString, defaultClient, request)
}

// Post performs an http POST with the given path and query against the given
// WorkbenchService. The response body is marshaled from JSON into the given
// type T and returned as a result.
func Post[T any](ctx context.Context, s *WorkbenchService, path string, body any) result.Result[T] {
	resourceURL := httputil.MakeURL(s.endpoint, path, url.Values{})
	request := result.Map(httputil.MakePostRequestContextFunc(ctx), resourceURL)
	request = result.Map(httputil.AddBody(body), request)
	request = result.MapNoError(func(r *http.Request) *http.Request {
		httputil.AddHeaders(map[string]string{
			constants.AuthorizationHeaderKey:      "Bearer " + s.accessToken,
			constants.PostgrestCSRFTokenHeaderKey: s.csrfToken,
			constants.CookieHeaderKey:             constants.PostgrestCSRFTokenKey + "=" + s.csrfToken + ";",
		}, r)
		return r
	}, request)
	return result.FlatMap2(httputils.DoAsJSON[T], defaultClient, request)
}

// GetValidDataplaneID looks up a dataplane using the given dataplaneID. If the
// dataplane exists then its ID is returned. If given an empty dataplaneID then
// an empty string is returned with no error.
func GetValidDataplaneID(ctx context.Context, svc *WorkbenchService, dataplaneID string) (string, error) {
	query := []string{}
	if dataplaneID != "" {
		query = append(query, dataplaneID)
	}
	dataplanes, err := GetDataplanes(ctx, svc, query, "")
	if err != nil {
		return "", err
	}
	if len(dataplanes) > 1 {
		return "", errors.New(i18n.Format(i18n.MultipleDataplanesFoundSpecifyOne))
	}
	if len(dataplanes) == 0 {
		if dataplaneID != "" {
			return "", errors.New(
				i18n.Format(
					i18n.ResourceNotFoundWithID,
					i18n.A{constants.I18nArgTypeKey: constants.DataplaneType, constants.I18nArgKeyID: dataplaneID},
				),
			)
		}
		return "", errors.New(i18n.Format(i18n.DataplaneNotFound))
	}
	return dataplanes[0].ID, nil
}

// GetValidWorkspaceRequests takes the given worksaceIDs, fetches those
// workspaces from the server, and returns the full workspace body.
func GetValidWorkspaceRequests(
	ctx context.Context,
	svc *WorkbenchService,
	workspaceIDs []string,
) ([]types.HomeDirAwareWorkspaceRequest, error) {
	validWorkspaces := make([]types.HomeDirAwareWorkspaceRequest, 0, len(workspaceIDs))
	for _, workspaceID := range workspaceIDs {
		workspaces, err := GetWorkspaces(ctx, svc, []string{workspaceID}, "")
		if err != nil {
			return nil, err
		}
		if len(workspaces) == 0 {
			return nil, errors.New(
				i18n.Format(
					i18n.ResourceNotFoundWithID,
					i18n.A{constants.I18nArgTypeKey: constants.VolumeType, constants.I18nArgKeyID: workspaceID},
				),
			)
		}
		id := workspaces[0].ControllerID
		if id == "" {
			id = workspaces[0].ID
		}
		mountName := workspaces[0].Name
		homeDir := isHomeDir(workspaces[0])
		if homeDir {
			mountName = "sas"
		}
		validWorkspaces = append(validWorkspaces, types.HomeDirAwareWorkspaceRequest{
			WorkspaceRequest: types.WorkspaceRequest{
				ID:        id,
				MountName: mountName,
			},
			HomeDir: homeDir,
		})
	}
	return validWorkspaces, nil
}

func isHomeDir(workspace types.Workspace) bool {
	homeDir, ok := workspace.Properties[constants.HomeDirPropertyKey].(string)
	if ok && strings.ToLower(homeDir) == "true" {
		return true
	}
	return false
}

// AddEnvsToServiceRequests returns a function that will add the given
// environment strings, in the form `name=value`, to each service request.
func AddEnvsToServiceRequests(
	envs []string,
	serviceRequests []types.ServiceRequest,
) []types.ServiceRequest {
	envMap := maputil.FromEqualDelimitedLines(envs)
	for i := range serviceRequests {
		if serviceRequests[i].Env == nil {
			serviceRequests[i].Env = map[string]string{}
		}
		for k, v := range envMap {
			serviceRequests[i].Env[k] = v
		}
	}
	return serviceRequests
}

// GetServiceRequests converts the given application into a set of service
// requests that can be included in a Workbench request for that application.
func GetServiceRequests(
	ctx context.Context,
	svc *WorkbenchService,
	application string,
	serviceIDs []string,
) ([]types.ServiceRequest, error) {
	if application != "" {
		return applicationToServiceRequests(ctx, svc, application)
	}
	return result.Split(
		sliceutils.Map(
			result.MakeFlatMapper(serviceIDtoServiceRequest(ctx, svc)),
			serviceIDs,
		),
	)
}

func applicationToServiceRequests(
	ctx context.Context,
	svc *WorkbenchService,
	applicationID string,
) ([]types.ServiceRequest, error) {
	applications, err := GetApplications(ctx, svc, []string{applicationID}, "")
	if err != nil {
		return nil, err
	}
	if len(applications) == 0 {
		return nil, errors.New(
			i18n.Format(
				i18n.ResourceNotFoundWithID,
				i18n.A{constants.I18nArgTypeKey: constants.ApplicationType, constants.I18nArgKeyID: applicationID},
			),
		)
	}
	services := []types.ServiceRequest{}
	for _, service := range applications[0].Services {
		services = append(services, types.ServiceRequest{ID: service})
	}
	return services, nil
}

func serviceIDtoServiceRequest(
	ctx context.Context,
	svc *WorkbenchService,
) func(string) (types.ServiceRequest, error) {
	return func(serviceID string) (types.ServiceRequest, error) {
		services, err := GetServices(ctx, svc, []string{serviceID}, "")
		if err != nil {
			return types.ServiceRequest{}, err
		}
		if len(services) == 0 {
			return types.ServiceRequest{}, errors.New(
				i18n.Format(
					i18n.ResourceNotFoundWithID,
					i18n.A{constants.I18nArgTypeKey: constants.ServiceType, constants.I18nArgKeyID: serviceID},
				),
			)
		}
		return types.ServiceRequest{
			ID: services[0].ID,
		}, nil
	}
}

// GetInitializerRequests converts the given initializer IDs into a set of
// service requests that can be included in a Workspace request with those
// initializers.
func GetInitializerRequests(
	ctx context.Context,
	svc *WorkbenchService,
	initializerIDs []string,
) ([]types.ServiceRequest, error) {
	serviceRequests := make([]types.ServiceRequest, 0, len(initializerIDs))
	for _, initializerID := range initializerIDs {
		initializers, err := GetInitializers(ctx, svc, []string{initializerID}, "")
		if err != nil {
			return nil, err
		}
		if len(initializers) == 0 {
			return nil, errors.New(
				i18n.Format(
					i18n.ResourceNotFoundWithID,
					i18n.A{
						constants.I18nArgTypeKey: constants.InitializerType,
						constants.I18nArgKeyID:   initializerID,
					},
				),
			)
		}
		serviceRequests = append(serviceRequests, types.ServiceRequest{
			ID: initializers[0].ID,
		})
	}
	return serviceRequests, nil
}

func normalizeSortSpec(sortSpec string, substitutions map[string][]string) string {
	parts := strings.Split(sortSpec, ",")
	sortSpecs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = normalizeSortSpecPart(part, substitutions)
		sortSpecs = append(sortSpecs, part)
	}
	return strings.Join(sortSpecs, ",")
}

func normalizeSortSpecPart(sortSpecPart string, substitutions map[string][]string) string {
	sortSpecPart = strings.ToLower(sortSpecPart)
	parts := strings.SplitN(sortSpecPart, ".", 2)
	subs, present := substitutions[parts[0]]
	if present {
		sortSpecs := make([]string, 0, len(subs))
		for _, sub := range subs {
			subSpec := strings.Join(append([]string{sub}, parts[1:]...), ".")
			sortSpecs = append(sortSpecs, subSpec)
		}
		return strings.Join(sortSpecs, ",")
	}
	return sortSpecPart
}

func normalizeAuth0SortSpec(sortSpec string, substitutions map[string][]string) (string, error) {
	if sortSpec == "" {
		return "", nil
	}
	parts := strings.Split(sortSpec, ",")
	if len(parts) > 1 {
		return "", errors.New(i18n.Format(i18n.Auth0OnlyOneSortFieldAllowed))
	}
	sortSpec = strings.ToLower(parts[0])
	parts = strings.SplitN(sortSpec, ".", 2)
	field := parts[0]
	direction := ":ascending"
	if len(parts) > 1 && strings.HasPrefix(parts[1], "desc") {
		direction = ":descending"
	}
	subs, present := substitutions[field]
	if present {
		sortSpecs := make([]string, 0, len(subs))
		for _, sub := range subs {
			subSpec := sub + direction
			sortSpecs = append(sortSpecs, subSpec)
		}
		return strings.Join(sortSpecs, ","), nil
	}
	return field + direction, nil
}
