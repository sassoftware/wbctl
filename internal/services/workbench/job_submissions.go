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
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetJobSubmissions  gets the job submissions in the given WorkbenchService. If
// given any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetJobSubmissions(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.JobSubmissions, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:   {constants.SortKeyCreatedAt},
		constants.SortKeyState: {constants.SortKeyCurrentState},
	})
	return result.Split(Get[types.JobSubmissions](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointJobRuns), IDsOrControllerIDsQuery(ids, sortSpec),
	))
}

// BulkDeleteJobSubmissions deletes the job submission with the given id.
func BulkDeleteJobSubmissions(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateJobSubmission creates a job submission.
func CreateJobSubmission(
	ctx context.Context,
	s *WorkbenchService,
	body types.JobSubmission,
) (types.JobSubmission, error) {
	return result.Split(Post[types.JobSubmission](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCJobRunCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

func jobSubmissionHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelJobName,
			constants.TableLabelUserName,
			constants.TableLabelLogURL,
			constants.TableLabelState,
			constants.TableLabelMessage,
			constants.TableLabelRuntime,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelControllerID,
		constants.TableLabelJobName,
		constants.TableLabelLogURL,
		constants.TableLabelState,
		constants.TableLabelMessage,
		constants.TableLabelRuntime,
		constants.TableLabelAge,
	}
}

func jobSubmissionRow(isAdmin bool, resource types.JobSubmission) []string {
	if isAdmin {
		return []string{
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.JobName,
			resource.Username,
			resource.LogURL,
			resource.CurrentState,
			resource.Message,
			fmt.Sprintf("%.2fs", resource.Runtime),
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
		resource.JobName,
		resource.LogURL,
		resource.CurrentState,
		resource.Message,
		fmt.Sprintf("%.2fs", resource.Runtime),
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintJobSubmissions returns a function that prints the given job submissions.
func PrintJobSubmissions(_ui ui.UI, token string, resources types.JobSubmissions) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(jobSubmissionHeaders(isAdmin))
	for _, resource := range resources {
		tbl.Add(jobSubmissionRow(isAdmin, resource)...)
	}
	tbl.Print()
	return nil
}

// DescribeJobSubmissions returns a function that describes the given job submissions.
func DescribeJobSubmissions(_ui ui.UI, resources types.JobSubmissions) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendJobSubmissionDetails(&tabbedString, resource)
		printutil.AppendLabelDetails(&tabbedString, resource.Labels)
		printutil.AppendEnvDetails(&tabbedString, resource.Env)
		printutil.AppendWorkspaceRequestsDetails(&tabbedString, resource.Workspaces)
		printutil.AppendPropertiesDetails(&tabbedString, resource.Properties)
		printutil.AppendPodStatusDetails(&tabbedString, resource.PodStatus)
	}
	_ui.Display(tabbedString.String())
	return nil
}
