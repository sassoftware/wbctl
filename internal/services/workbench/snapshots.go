// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strconv"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/auth"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/logutil"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/sliceutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetSnapshots gets the snapshots in the given WorkbenchService. If given any
// ids, only the matching resources will be returned. Otherwise, all resources
// are returned.
func GetSnapshots(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Snapshots, error) {
	logutil.GetLogger().Info("getting snapshots", "ids", ids, "sort", sortSpec)
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:    {constants.SortKeyCreatedAt},
		constants.SortKeySource: {constants.SortKeySpecSourcePVCName},
		constants.SortKeySize:   {constants.SortKeyVolumeStatusRestoreSize},
		constants.SortKeyReady:  {constants.SortKeyVolumeStatusReadyToUse},
	})
	return result.Split(Get[types.Snapshots](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointSnapshots), IDsControllerIDsOrNamesQuery(ids, sortSpec),
	))
}

// BulkDeleteSnapshots deletes the snapshot with the given id.
func BulkDeleteSnapshots(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	logutil.GetLogger().Info("deleting snapshots", "ids", ids)
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateSnapshot creates a snapshot.
func CreateSnapshot(
	ctx context.Context,
	s *WorkbenchService,
	body types.Snapshot,
) (types.Snapshot, error) {
	logutil.GetLogger().Info("creating snapshot", "body", body)
	return result.Split(Post[types.Snapshot](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

// BulkUpdateSnapshots applies the given types.Snapshot to all the snapshots
// with the given ids.
func BulkUpdateSnapshots(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body types.Snapshot,
) (types.Snapshots, error) {
	logutil.GetLogger().Info("updating snapshots", "ids", ids, "body", body)
	return result.Split(Post[types.Snapshots](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCSnapshotBulkUpdate),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids, constants.RPCBodyKeyParam: body},
	))
}

func snapshotHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelName,
			constants.TableLabelUserName,
			constants.TableLabelDescription,
			constants.TableLabelSource,
			constants.TableLabelSize,
			constants.TableLabelReady,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelControllerID,
		constants.TableLabelName,
		constants.TableLabelDescription,
		constants.TableLabelSource,
		constants.TableLabelSize,
		constants.TableLabelReady,
		constants.TableLabelAge,
	}
}

func snapshotRow(isAdmin bool, resource types.Snapshot) []string {
	if isAdmin {
		return []string{
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.Name,
			resource.Username,
			resource.Description,
			pointer.UnPtr(resource.Spec.Source.PersistentVolumeClaimName),
			printutil.QuantityPtrToString(resource.VolumeStatus.Status.RestoreSize),
			strconv.FormatBool(pointer.UnPtr(resource.VolumeStatus.Status.ReadyToUse)),
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
		resource.Name,
		resource.Description,
		pointer.UnPtr(resource.Spec.Source.PersistentVolumeClaimName),
		printutil.QuantityPtrToString(resource.VolumeStatus.Status.RestoreSize),
		strconv.FormatBool(pointer.UnPtr(resource.VolumeStatus.Status.ReadyToUse)),
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintSnapshots returns a function that prints the given snapshots.
func PrintSnapshots(_ui ui.UI, token string, resources types.Snapshots) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(snapshotHeaders(isAdmin))
	for _, rsrc := range resources {
		tbl.Add(snapshotRow(isAdmin, rsrc)...)
	}
	tbl.Print()
	return nil
}

// DescribeSnapshots returns a function that describes the given snapshots.
func DescribeSnapshots(_ui ui.UI, resources types.Snapshots) error {
	var tabbedString strings.Builder
	for i, rsrc := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendSnapshotDetails(&tabbedString, rsrc)
		printutil.AppendPropertiesDetails(&tabbedString, rsrc.Properties)
		printutil.AppendVolumeDetails(&tabbedString, rsrc.VolumeStatus)
	}
	_ui.Display(tabbedString.String()) // ensure a final newline
	return nil
}
