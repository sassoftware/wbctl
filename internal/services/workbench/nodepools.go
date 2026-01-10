// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetNodepools gets the nodepools in the given WorkbenchService. If given any
// ids, only the matching resources will be returned. Otherwise, all resources
// are returned.
func GetNodepools(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Nodepools, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyName: {constants.SortKeyDisplayName},
	})
	return result.Split(Get[types.Nodepools](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools), IDsQuery(ids, sortSpec),
	))
}

// PrintNodepools returns a function that prints the given nodepools.
func PrintNodepools(_ui ui.UI, _ string, resources types.Nodepools) error {
	tbl := _ui.Table([]string{constants.TableLabelID, constants.TableLabelName, constants.TableLabelDescription})
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.DisplayName,
			resource.Description,
		)
	}
	tbl.Print()
	return nil
}
