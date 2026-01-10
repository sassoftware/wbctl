// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetPings gets the pings in the given WorkbenchService. If given any ids, only
// the matching resources will be returned. Otherwise, all resources are
// returned.
func GetPings(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Pings, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:        {constants.SortKeyRequestAt},
		constants.SortKeyResponseAt: {constants.SortKeyStatusUpdatedAt},
		constants.SortKeyState:      {constants.SortKeyStatusState},
		constants.SortKeyMessage:    {constants.SortKeyStatusMessage},
	})
	return result.Split(
		Get[types.Pings](ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointPings), IDsQuery(ids, sortSpec)),
	)
}

// PrintPings returns a function that prints the given pings.
func PrintPings(_ui ui.UI, _ string, resources types.Pings) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelDataplaneID,
			constants.TableLabelRequestAt,
			constants.TableLabelResponseAt,
			constants.TableLabelResponseTime,
			constants.TableLabelState,
			constants.TableLabelMessage,
			constants.TableLabelAge,
		},
	)
	for _, resource := range resources {
		tbl.Add(
			resource.DataplaneID,
			printutil.TimePtrToString(resource.RequestAt),
			printutil.TimePtrToString(resource.ResponseAt),
			safeTimeDiff(resource.RequestAt, resource.ResponseAt),
			resource.State,
			resource.Message,
			timeutil.Since(resource.RequestAt),
		)
	}
	tbl.Print()
	return nil
}

func safeTimeDiff(start, end *time.Time) string {
	if start == nil || end == nil {
		return constants.ValueUnknown
	}
	return end.Sub(*start).String()
}
