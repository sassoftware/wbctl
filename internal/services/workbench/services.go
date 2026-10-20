// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"slices"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
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

// GetServiceIDs returns all service IDs present in the WorkbenchService.
func GetServiceIDs(
	ctx context.Context,
	s *WorkbenchService,
) ([]string, error) {
	services, err := GetServices(ctx, s, nil, "")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(services))
	for _, service := range services {
		ids = append(ids, service.ID)
	}
	slices.Sort(ids)
	return ids, nil
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
