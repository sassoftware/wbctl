// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strconv"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetCapabilities gets the Capabilities supported by the WorkbenchService. If
// given any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetCapabilities(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Capabilities, error) {
	logutil.GetLogger().Info("getting capabilities", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyEnabled: {constants.SortKeyDisabled},
	})
	return result.Split(Get[types.Capabilities](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchCapabilities), IDsQuery(ids, sortSpec),
	))
}

// PrintCapabilities returns a function that prints the given capabilities.
func PrintCapabilities(_ui ui.UI, _ string, resources types.Capabilities) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelDescription,
			constants.TableLabelVersion,
			constants.TableLabeCategory,
			constants.TableLabelEnabled,
			constants.TableLabelRecommended,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Name,
			resource.Description,
			resource.Version,
			resource.Category,
			strconv.FormatBool(!resource.Disabled),
			strconv.FormatBool(resource.Recommended),
		)
	}
	tbl.Print()
	return nil
}
