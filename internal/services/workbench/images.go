// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/printutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/sliceutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/timeutil"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sas-institute-rnd-saas/wbctl/internal/ui"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// GetImages gets the images in the given WorkbenchService. If given any ids,
// only the matching resources will be returned. Otherwise, all resources are
// returned.
func GetImages(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Images, error) {
	sortSpec = normalizeSortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge:          {constants.SortKeyCreatedAt},
		constants.SortKeyImageBuildID: {constants.SortKeyLatestImageBuildID},
		constants.SortKeyState:        {constants.SortKeyCurrentState},
		constants.SortKeyImage: {
			constants.SortKeyImageRegistry,
			constants.SortKeyImageName,
			constants.SortKeyImageTag,
		},
	})
	return result.Split(Get[types.Images](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointImages), IDsOrControllerIDsQuery(ids, sortSpec),
	))
}

// BulkDeleteImages deletes the images with the given ids.
func BulkDeleteImages(ctx context.Context, s *WorkbenchService, ids []string) ([]string, error) {
	return result.Split(Post[[]string](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkDelete),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids},
	))
}

// CreateImage creates an image.
func CreateImage(ctx context.Context, s *WorkbenchService, body types.Image) (types.Image, error) {
	return result.Split(Post[types.Image](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCImageCreate),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
}

// RebuildImage rebuilds the image with the given id.
func RebuildImage(ctx context.Context, s *WorkbenchService, id string) (bool, error) {
	return result.Split(Post[bool](
		ctx, s, postgrest.Endpoint(constants.WorkbenchEndpointRPCImageRebuild), map[string]interface{}{"id": id},
	))
}

// BulkUpdateImages applies the given types.Image to all the images with the
// given ids.
func BulkUpdateImages(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body types.Image,
) (types.Images, error) {
	return result.Split(Post[types.Images](
		ctx,
		s,
		postgrest.Endpoint(constants.WorkbenchEndpointRPCImageBulkUpdate),
		map[string]interface{}{constants.RPCBodyKeyIDs: ids, constants.RPCBodyKeyParam: body},
	))
}

func imageHeaders(isAdmin bool) []string {
	if isAdmin {
		return []string{
			constants.TableLabelID,
			constants.TableLabelControllerID,
			constants.TableLabelName,
			constants.TableLabelUserName,
			constants.TableLabelDescription,
			constants.TableLabelImage,
			constants.TableLabelImageBuildID,
			constants.TableLabelState,
			constants.TableLabelMessage,
			constants.TableLabelAge,
		}
	}
	return []string{
		constants.TableLabelID,
		constants.TableLabelControllerID,
		constants.TableLabelName,
		constants.TableLabelDescription,
		constants.TableLabelImage,
		constants.TableLabelImageBuildID,
		constants.TableLabelState,
		constants.TableLabelMessage,
		constants.TableLabelAge,
	}
}

func imageRow(isAdmin bool, resource types.Image) []string {
	tag := ""
	if resource.ImageTag != "" {
		tag = ":" + resource.ImageTag
	}
	if isAdmin {
		return []string{
			resource.ID,
			sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
			resource.Name,
			resource.Username,
			resource.Description,
			resource.ImageRegistry + "/" + resource.ImageName + tag,
			resource.LatestImageBuildID,
			resource.CurrentState,
			resource.Message,
			timeutil.Since(resource.CreatedAt),
		}
	}
	return []string{
		resource.ID,
		sliceutil.Coalesce(resource.ControllerID, constants.ValueUnknown),
		resource.Name,
		resource.Description,
		resource.ImageRegistry + "/" + resource.ImageName + tag,
		resource.LatestImageBuildID,
		resource.CurrentState,
		resource.Message,
		timeutil.Since(resource.CreatedAt),
	}
}

// PrintImages returns a function that prints the given images.
func PrintImages(_ui ui.UI, token string, resources types.Images) error {
	isAdmin := auth.IsRole(token, constants.WorkbenchAdminRole)
	tbl := _ui.Table(imageHeaders(isAdmin))
	for _, resource := range resources {
		tbl.Add(imageRow(isAdmin, resource)...)
	}
	tbl.Print()
	return nil
}

// DescribeImages returns a function that describes the given images.
func DescribeImages(_ui ui.UI, resources types.Images) error {
	var tabbedString strings.Builder
	for i, resource := range resources {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendImageDetails(&tabbedString, resource)
		printutil.AppendLabelDetails(&tabbedString, resource.Labels)
		printutil.AppendPodStatusDetails(&tabbedString, resource.PodStatus)
	}
	_ui.Display(tabbedString.String())
	return nil
}
