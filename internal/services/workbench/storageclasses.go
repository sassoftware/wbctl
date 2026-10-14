// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/maputil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetStorageClasses gets the storage classes in the given WorkbenchService. If
// given any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetStorageClasses(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.StorageClasses, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyName:    {constants.SortKeyDisplayName},
		constants.SortKeyDefault: {constants.SortKeyPropertiesDefault},
	})
	return result.Split(Get[types.StorageClasses](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplaneStorageclasses), IDsQuery(ids, sortSpec),
	))
}

// PrintStorageClasses returns a function that prints the given storage classes.
func PrintStorageClasses(_ui ui.UI, _ string, resources types.StorageClasses) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelDescription,
			constants.TableLabelDefault,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.DisplayName,
			resource.Description,
			maputil.SafeValueAsString(resource.Properties, constants.DefaultKey),
		)
	}
	tbl.Print()
	return nil
}
