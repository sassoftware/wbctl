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

// GetServices gets the services in the given WorkbenchService. If given any
// ids, only the matching resources will be returned. Otherwise, all resources
// are returned.
func GetServices(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Services, error) {
	return result.Split(Get[types.Services](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkbenchServices), IDsQuery(ids, sortSpec),
	))
}

// PrintServices returns a function that prints the given services.
func PrintServices(_ui ui.UI, _ string, resources types.Services) error {
	tbl := _ui.Table([]string{constants.TableLabelID, constants.TableLabelName, constants.TableLabelType})
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Name,
			resource.Type,
		)
	}
	tbl.Print()
	return nil
}
