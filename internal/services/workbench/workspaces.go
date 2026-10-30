// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"fmt"
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

// GetWorkspaces gets the workspaces in the given WorkbenchService. If given any
// ids, only the matching resources will be returned. Otherwise, all resources
// are returned.
func GetWorkspaces(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Workspaces, error) {
	logutil.GetLogger().Info("getting volumes", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:              {constants.SortKeyCreatedAt},
		constants.SortKeySize:             {constants.SortKeyStorageSizeG},
		constants.SortKeyHomeDir:          {constants.SortKeyPropertiesHomeDirectory},
		constants.SortKeyStorageClassName: {constants.SortKeyStorageSpecStorageClassName},
	})
	workspaces, err := result.Split(Get[types.Workspaces](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointWorkspaces), IDsControllerIDsOrNamesQuery(ids, sortSpec),
	))
	if err != nil {
		return nil, err
	}
	return selectByIDThenName(
		ids,
		workspaces,
		func(w types.Workspace) string { return w.ID },
		func(w types.Workspace) string { return w.ControllerID },
		func(w types.Workspace) string { return w.Name },
	), nil
}

func resolveWorkspaceIDs(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	return resolveIDs(
		ctx, s, ids, GetWorkspaces, func(w types.Workspace) string { return w.ID }, constants.VolumeType,
	)
}

// BulkDeleteWorkspaces deletes the workspaces with the given ids,
// controller_ids, or names.
func BulkDeleteWorkspaces(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	logutil.GetLogger().Info("deleting volumes", "ids", ids)
	ids, err := resolveWorkspaceIDs(ctx, s, ids)
	if err != nil {
		return nil, err
	}
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateWorkspace creates a workspace.
func CreateWorkspace(
	ctx context.Context,
	s *WorkbenchService,
	body types.Workspace,
) (types.Workspace, error) {
	logutil.GetLogger().Info("creating volume", "body", body)
	return result.Split(Post[types.Workspace](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCWorkspaceCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

// BulkUpdateWorkspaces  applies the given types.Workspace to all the workspaces
// with the given ids.
func BulkUpdateWorkspaces(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body types.Workspace,
) (types.Workspaces, error) {
	logutil.GetLogger().Info("updating volumes", "ids", ids)
	ids, err := resolveWorkspaceIDs(ctx, s, ids)
	if err != nil {
		return nil, err
	}
	return bulkUpdateForUniqueNameConstrainedResources[types.Workspace, types.Workspaces](
		ctx,
		s,
		ids,
		body,
		func(workspace types.Workspace) string { return workspace.Name },
		func(workspace *types.Workspace, id string) { workspace.ID = id },
		constants.VolumesType,
		constants.WorkbenchEndpointRPCWorkspaceBulkUpdate,
		constants.WorkbenchEndpointRPCWorkspaceUpdate,
		constants.RPCBodyKeyWorkspace,
	)
}

func workspaceHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelName,
			constants.TableLabelUserName,
			constants.TableLabelSize,
			constants.TableLabelHomeDir,
			constants.TableLabelClass,
			constants.TableLabelState,
			constants.TableLabelMessage,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelControllerID,
		constants.TableLabelName,
		constants.TableLabelSize,
		constants.TableLabelHomeDir,
		constants.TableLabelClass,
		constants.TableLabelState,
		constants.TableLabelMessage,
		constants.TableLabelAge,
	}
}

func workspaceRow(isAdmin bool, resource types.Workspace) []string {
	homeDir := ""
	if isHomeDir(resource) {
		homeDir = "true"
	}
	if isAdmin {
		return []string{
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.Name,
			resource.Username,
			fmt.Sprintf("%vGi", resource.StorageSizeG),
			homeDir,
			storageClass(resource.StorageSpec),
			resource.State,
			resource.Message,
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
		resource.Name,
		fmt.Sprintf("%vGi", resource.StorageSizeG),
		homeDir,
		storageClass(resource.StorageSpec),
		resource.State,
		resource.Message,
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintWorkspaces returns a function that prints the given workspaces.
func PrintWorkspaces(_ui ui.UI, token string, resources types.Workspaces) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(workspaceHeaders(isAdmin))
	for _, resource := range resources {
		tbl.Add(workspaceRow(isAdmin, resource)...)
	}
	tbl.Print()
	return nil
}

// DescribeWorkspaces returns a function that describes the given workspaces.
func DescribeWorkspaces(_ui ui.UI, resources types.Workspaces) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendWorkspaceDetails(&tabbedString, resource)
		printutil.AppendStorageSpecDetails(&tabbedString, resource.StorageSpec, "")
		printutil.AppendServiceRequestDetails(&tabbedString, "Initializers", resource.Initializers)
		printutil.AppendPropertiesDetails(&tabbedString, resource.Properties)
		printutil.AppendPodStatusDetails(&tabbedString, resource.PodStatus)
	}
	_ui.Display(tabbedString.String())
	return nil
}
