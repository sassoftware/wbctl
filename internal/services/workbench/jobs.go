// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/printutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetJobs gets the jobs in the given WorkbenchService. If given any ids, only
// the matching resources will be returned. Otherwise, all resources are
// returned.
func GetJobs(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Jobs, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:   {constants.SortKeyCreatedAt},
		constants.SortKeyState: {constants.SortKeyCurrentState},
	})
	return result.Split(Get[types.Jobs](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointJobs), IDsQuery(ids, sortSpec),
	))
}

// BulkDeleteJobs deletes the job with the given id.
func BulkDeleteJobs(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateJob creates a job.
func CreateJob(ctx context.Context, s *WorkbenchService, body types.Job) (types.Job, error) {
	return result.Split(Post[types.Job](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

// BulkUpdateJobs  applies the given types.Job to all the jobs with the given
// ids.
func BulkUpdateJobs(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body types.Job,
) (types.Jobs, error) {
	return result.Split(Post[types.Jobs](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobBulkUpdate),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids, constants.RPCBodyKeyParam: body},
	))
}

func jobHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelUserName,
			constants.TableLabelApplicationID,
			constants.TableLabelServices,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelName,
		constants.TableLabelApplicationID,
		constants.TableLabelServices,
		constants.TableLabelAge,
	}
}

func jobRow(isAdmin bool, resource types.Job) []string {
	if isAdmin {
		return []string{
			resource.ID,
			resource.Name,
			resource.Username,
			resource.ApplicationID,
			getServiceIDs(resource.Services),
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		resource.Name,
		resource.ApplicationID,
		getServiceIDs(resource.Services),
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintJobs returns a function that prints the given jobs.
func PrintJobs(_ui ui.UI, token string, resources types.Jobs) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(jobHeaders(isAdmin))
	for _, resource := range resources {
		tbl.Add(jobRow(isAdmin, resource)...)
	}
	tbl.Print()
	return nil
}

// DescribeJobs returns a function that describes the given jobs.
func DescribeJobs(_ui ui.UI, resources types.Jobs) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendJobDetails(&tabbedString, resource)
		printutil.AppendLabelDetails(&tabbedString, resource.Labels)
		printutil.AppendServiceRequestDetails(&tabbedString, "Services", resource.Services)
		printutil.AppendWorkspaceRequestsDetails(&tabbedString, resource.Workspaces)
		printutil.AppendPropertiesDetails(&tabbedString, resource.Properties)
	}
	_ui.Display(tabbedString.String())
	return nil
}
