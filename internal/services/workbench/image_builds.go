// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/printutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetImageBuilds gets the image builds in the given WorkbenchService. If given
// any ids, only the matching resources will be returned. Otherwise, all
// resources are returned.
func GetImageBuilds(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.ImageBuilds, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge: {constants.SortKeyCreatedAt},
	})
	return result.Split(Get[types.ImageBuilds](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointImageBuilds), IDsOrControllerIDsQuery(ids, sortSpec),
	))
}

// PrintImageBuilds returns a function that prints the given image builds.
func PrintImageBuilds(_ui ui.UI, _ string, resources types.ImageBuilds) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelName,
			constants.TableLabelImageID,
			constants.TableLabelAge,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.Name,
			resource.ImageID,
			timeutil.Since(resource.CreatedAt),
		)
	}
	tbl.Print()
	return nil
}

// DescribeImageBuilds returns a function that describes the given image builds.
func DescribeImageBuilds(_ui ui.UI, resources types.ImageBuilds) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendImageBuildDetails(&tabbedString, resource)
		printutil.AppendPodStatusDetails(&tabbedString, resource.PodStatus)
	}
	_ui.Display(tabbedString.String())
	return nil
}
