// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/auth0admin"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/types"
	"github.com/sassoftware/wbctl/internal/ui"
)

// GetUsers gets the users in the given WorkbenchService. If given any ids, only
// the matching resources will be returned. Otherwise, all resources are
// returned.
func GetUsers(
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	sortSpec string,
) (types.Auth0UserList, error) {
	sortSpec = strings.ToLower(sortSpec)
	unsupportedSortKeys := make([]string, 0, 2)
	if strings.Contains(sortSpec, constants.SortKeyRoles) {
		unsupportedSortKeys = append(unsupportedSortKeys, constants.SortKeyRoles)
	}
	if strings.Contains(sortSpec, constants.SortKeyMetadata) {
		unsupportedSortKeys = append(unsupportedSortKeys, constants.SortKeyMetadata)
	}
	if len(unsupportedSortKeys) > 0 {
		return types.Auth0UserList{}, errors.New(
			i18n.Format(
				i18n.UnsupportedSortKeys,
				i18n.A{
					constants.I18nArgTypeSortKeysKey: strings.Join(unsupportedSortKeys, ", "),
				},
			),
		)
	}
	sortSpec, err := normalizeAuth0SortSpec(sortSpec, map[string][]string{
		constants.SortKeyAge: {constants.SortKeyCreationTimeStamp},
	})
	if err != nil {
		return types.Auth0UserList{}, err
	}
	users, err := result.Split(Get[types.Auth0UserList](
		ctx,
		s,
		auth0admin.Endpoint(constants.WorkbenchEndpointUsers),
		Auth0IDsQuery(ids, sortSpec),
	))
	if err != nil {
		return types.Auth0UserList{}, err
	}
	return users, nil
}

// PrintUsers returns a function that prints the given users.
func PrintUsers(_ui ui.UI, _ string, resources types.Auth0UserList) error {
	tbl := _ui.Table(
		[]string{
			constants.TableLabelID,
			constants.TableLabelName,
			constants.TableLabelEmail,
			constants.TableLabelRoles,
			constants.TableLabelMetadata,
			constants.TableLabelAge,
		},
	)
	for _, resource := range resources.Items {
		tbl.Add(
			resource.ID,
			resource.Name,
			resource.Email,
			collectRoleNames(resource.Roles),
			collectSettings(resource.UserMetadata),
			timeutil.Since(pointer.Ptr(resource.CreationTimeStamp)),
		)
	}
	tbl.Print()
	return nil
}

// DescribeUsers returns a function that describes the given users.
func DescribeUsers(_ui ui.UI, resources types.Auth0UserList) error {
	var tabbedString strings.Builder
	for i, resource := range resources.Items {
		if i != 0 {
			tabbedString.WriteString("\n\n")
		}
		printutil.AppendUserDetails(&tabbedString, resource)
		printutil.AppendRolesDetails(&tabbedString, resource.Roles)
		printutil.AppendMetadataDetails(&tabbedString, resource.UserMetadata)
	}
	_ui.Display(tabbedString.String()) // Display ensures a final newline
	return nil
}

func collectRoleNames(roles []types.Role) string {
	names := []string{}
	overflow := false
	for i, role := range roles {
		if i > 4 {
			overflow = true
			break
		}
		names = append(names, role.Name)
	}
	slices.Sort(names)
	if overflow {
		names = append(names, "...")
	}
	return strings.Join(names, ", ")
}

func collectSettings(settings map[string]any) string {
	formattedSettings := make([]string, 0, len(settings))
	for _, key := range slices.Sorted(maps.Keys(settings)) {
		formattedSettings = append(formattedSettings, fmt.Sprintf("%s=%v", key, settings[key]))
	}
	return strings.Join(formattedSettings, ", ")
}
