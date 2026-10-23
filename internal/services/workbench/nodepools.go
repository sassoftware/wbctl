// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"slices"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
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
	logutil.GetLogger().Info("getting nodepools", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyName: {constants.SortKeyDisplayName},
	})
	return result.Split(Get[types.Nodepools](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneNodepools), IDsQuery(ids, sortSpec),
	))
}

// GetNodepoolIDs returns all nodepool IDs present in the WorkbenchService.
func GetNodepoolIDs(
	ctx context.Context,
	s *WorkbenchService,
) ([]string, error) {
	nodepools, err := GetNodepools(ctx, s, nil, "")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(nodepools))
	for _, nodepool := range nodepools {
		ids = append(ids, nodepool.ID)
	}
	slices.Sort(ids)
	return ids, nil
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
