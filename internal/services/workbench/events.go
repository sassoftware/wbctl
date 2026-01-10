// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetEvents gets the events from the given WorkbenchService. If given any ids,
// only the matching resources will be returned. Otherwise, all resources are
// returned.
func GetEvents(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Events, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge: {constants.SortKeyCreatedAt},
	})
	return result.Split(
		Get[types.Events](ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointEvents), IDsQuery(ids, sortSpec)),
	)
}

// PrintEvents returns a function that prints the given events.
func PrintEvents(_ui ui.UI, _ string, resources types.Events) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelType,
			constants.TableLabelResourceType,
			constants.TableLabelKind,
			constants.TableLabelAction,
			constants.TableLabelNote,
			constants.TableLabelReason,
			constants.TableLabelAge,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.ID,
			resource.Type,
			resource.ResourceType,
			resource.Kind,
			resource.Action,
			strings.ReplaceAll(resource.Note, "\n", " "),
			resource.Reason,
			timeutil.Since(resource.CreatedAt),
		)
	}
	tbl.Print()
	return nil
}
