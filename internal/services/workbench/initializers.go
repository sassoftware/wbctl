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

// GetInitializers gets the initializers in the given WorkbenchService. If given
// any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetInitializers(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Initializers, error) {
	return result.Split(Get[types.Initializers](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneWorkspaceServices), IDsQuery(ids, sortSpec),
	))
}

// PrintInitializers returns a function that prints the given initializers.
func PrintInitializers(_ui ui.UI, _ string, resources types.Initializers) error {
	tbl := _ui.Table([]string{constants.TableLabelID, constants.TableLabelName, constants.TableLabelType})
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Name,
		)
	}
	tbl.Print()
	return nil
}
