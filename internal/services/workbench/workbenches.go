// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetWorkbenches gets the workbenches in the given WorkbenchService. If given
// any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetWorkbenches(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Workbenches, error) {
	logutil.GetLogger().Info("getting workbenches", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:   {constants.SortKeyCreatedAt},
		constants.SortKeyState: {constants.SortKeyCurrentState},
	})
	return result.Split(Get[types.Workbenches](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointWorkbenches), IDsOrControllerIDsQuery(ids, sortSpec),
	))
}

// BulkDeleteWorkbenches deletes the workbench with the given id.
func BulkDeleteWorkbenches(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	logutil.GetLogger().Info("deleting workbenches", "ids", ids)
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateWorkbench creates a workbench.
func CreateWorkbench(
	ctx context.Context,
	s *WorkbenchService,
	body types.Workbench,
) (types.Workbench, error) {
	logutil.GetLogger().Info("creating workbench", "body", body)
	return result.Split(Post[types.Workbench](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkbenchCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

// BulkUpdateWorkbenches  applies the given types.Workbench to all the
// workbenches with the given ids.
func BulkUpdateWorkbenches(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body types.Workbench,
) (types.Workbenches, error) {
	logutil.GetLogger().Info("updating workbenches", "ids", ids, "body", body)
	return bulkUpdateForUniqueNameConstrainedResources[types.Workbench, types.Workbenches](
		ctx,
		s,
		ids,
		body,
		func(workbench types.Workbench) string { return workbench.Name },
		func(workbench *types.Workbench, id string) { workbench.ID = id },
		constants.WorkbenchesType,
		constants.WorkbenchEndpointRPCWorkbenchBulkUpdate,
		constants.WorkbenchEndpointRPCWorkbenchUpdate,
		constants.RPCBodyKeyWorkbench,
	)
}

func workbenchHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelName,
			constants.TableLabelUserName,
			constants.TableLabelApplicationID,
			constants.TableLabelServices,
			constants.TableLabelURL,
			constants.TableLabelState,
			constants.TableLabelMessage,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelControllerID,
		constants.TableLabelName,
		constants.TableLabelApplicationID,
		constants.TableLabelServices,
		constants.TableLabelURL,
		constants.TableLabelState,
		constants.TableLabelMessage,
		constants.TableLabelAge,
	}
}

func workbenchRow(isAdmin bool, resource types.Workbench) []string {
	if isAdmin {
		return []string{
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.Name,
			resource.Username,
			resource.ApplicationID,
			getServiceIDs(resource.Services),
			resource.URL,
			resource.CurrentState,
			resource.Message,
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
		resource.Name,
		resource.ApplicationID,
		getServiceIDs(resource.Services),
		resource.URL,
		resource.CurrentState,
		resource.Message,
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintWorkbenches returns a function that prints the given workbenches.
func PrintWorkbenches(_ui ui.UI, token string, resources types.Workbenches) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(workbenchHeaders(isAdmin))
	for _, resource := range resources {
		tbl.Add(workbenchRow(isAdmin, resource)...)
	}
	tbl.Print()
	return nil
}

// DescribeWorkbenches returns a function that describes the given workbenches.
func DescribeWorkbenches(_ui ui.UI, resources types.Workbenches) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendWorkbenchDetails(&tabbedString, resource)
		printutil.AppendLabelDetails(&tabbedString, resource.Labels)
		printutil.AppendServiceRequestDetails(&tabbedString, "Services", resource.Services)
		printutil.AppendWorkspaceRequestsDetails(&tabbedString, resource.Workspaces)
		printutil.AppendPropertiesDetails(&tabbedString, resource.Properties)
		printutil.AppendPodStatusDetails(&tabbedString, resource.PodStatus)
	}
	_ui.Display(tabbedString.String())
	return nil
}
