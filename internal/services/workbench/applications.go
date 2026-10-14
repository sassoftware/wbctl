// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetApplications gets the Applications supported by the WorkbenchService. If
// given any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetApplications(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Applications, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyName: {constants.SortKeyDisplayName},
	})
	return result.Split(Get[types.Applications](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchApplications), IDsQuery(ids, sortSpec),
	))
}

// PrintApplications returns a function that prints the given applications.
func PrintApplications(_ui ui.UI, _ string, resources types.Applications) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelDescription,
			constants.TableLabelServices,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.DisplayName,
			resource.Description,
			strings.Join(resource.Services, ", "),
		)
	}
	tbl.Print()
	return nil
}
