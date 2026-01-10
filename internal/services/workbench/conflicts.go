// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetConflicts returns the Conflicts currently present in the Workbench
// Service. If given any ids, only the matching resources will be returned.
// Otherwise, all resources are returned.
func GetConflicts(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Conflicts, error) {
	sortSpec = strings.ToLower(sortSpec)
	sortSpec = strings.ReplaceAll(sortSpec, "volume_", "workspace_")
	return result.Split(Get[types.Conflicts](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointWorkspaceConflicts), IDsQuery(ids, sortSpec),
	))
}

// PrintConflicts returns a function that prints the given conflicts.
func PrintConflicts(_ui ui.UI, _ string, resources types.Conflicts) error {
	tbl := _ui.Table([]string{
		constants.TableLabelID,
		constants.TableLabelName,
		constants.TableLabelType,
		constants.TableLabelConflictID,
		constants.TableLabelConflictName,
		constants.TableLabelConflictType,
		constants.TableLabelVolumeID,
		constants.TableLabelVolumeName,
	})
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Name,
			resource.Type,
			resource.ConflictID,
			resource.ConflictName,
			resource.ConflictType,
			resource.WorkspaceID,
			resource.WorkspaceName,
		)
	}
	tbl.Print()
	return nil
}

// DescribeConflicts returns a function that describes the given conflicts.
func DescribeConflicts(_ui ui.UI, resources types.Conflicts) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendConflictDetails(&tabbedString, resource)
	}
	_ui.Display(tabbedString.String())
	return nil
}
