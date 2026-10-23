// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetDataplanes gets the Dataplanes in the given WorkbenchService. If given any
// ids, only the matching resources will be returned. Otherwise, all resources
// are returned.
func GetDataplanes(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Dataplanes, error) {
	logutil.GetLogger().Info("getting dataplanes", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge: {constants.SortKeyCreatedAt},
	})
	return result.Split(Get[types.Dataplanes](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointDataplanes), IDsQuery(ids, sortSpec),
	))
}

// PrintDataplanes returns a function that prints the given dataplanes.
func PrintDataplanes(_ui ui.UI, _ string, resources types.Dataplanes) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelDescription,
			constants.TableLabelMessage,
			constants.TableLabelAge,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Name,
			resource.Description,
			resource.Message,
			timeutil.Since(resource.CreatedAt),
		)
	}
	tbl.Print()
	return nil
}

// DescribeDataplanes returns a function that describes the given dataplanes.
func DescribeDataplanes(_ui ui.UI, resources types.Dataplanes) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendDataplaneDetails(&tabbedString, resource)
		printutil.AppendPropertiesDetails(&tabbedString, resource.Properties)
	}
	_ui.Display(tabbedString.String())
	return nil
}
