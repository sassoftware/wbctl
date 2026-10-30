// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/postgrest"
	"github.com/sassoftware/wbctl/internal/types"
	corev1 "k8s.io/api/core/v1"
)

// IDsQuery returns a PostgREST compatible query for the given IDs and sort
// spec.
func IDsQuery(ids []string, sortSpec string) url.Values {
	if len(ids) == 0 && sortSpec == "" {
		return url.Values{}
	}
	query := url.Values{}
	if len(ids) > 0 {
		query["id"] = []string{"in.(" + strings.Join(ids, ",") + ")"}
	}
	if sortSpec != "" {
		query["order"] = []string{sortSpec}
	}
	return query
}

// IDsControllerIDsOrNamesQuery returns a PostgREST compatible query for the
// given IDs/names and sort spec that will match the id, controller_id, or name
// columns.
func IDsControllerIDsOrNamesQuery(ids []string, sortSpec string) url.Values {
	return buildFilterQuery(ids, "controller_id", "name", sortSpec)
}

// IDsOrNamesQuery returns a PostgREST compatible query for the given IDs/names
// and sort spec that will match the id or name columns.
func IDsOrNamesQuery(ids []string, sortSpec string) url.Values {
	return buildFilterQuery(ids, "", "name", sortSpec)
}

// IDsOrControllerIDsOrJobNamesQuery returns a PostgREST compatible query for the
// given IDs/names and sort spec that will match the id, controller_id, or job_name
// columns.
func IDsOrControllerIDsOrJobNamesQuery(ids []string, sortSpec string) url.Values {
	return buildFilterQuery(ids, "controller_id", "job_name", sortSpec)
}

// buildFilterQuery generates the query for the given IDs/names. UUID values
// are matched against the id column, while other values are matched against
// the controllerIDColumn if provided. Every value is matched against the
// nameColumn.
func buildFilterQuery(ids []string, controllerIDColumn, nameColumn, sortSpec string) url.Values {
	query := url.Values{}
	if sortSpec != "" {
		query["order"] = []string{sortSpec}
	}
	if len(ids) == 0 {
		return query
	}
	uuidIDs := []string{}
	nonUUIDs := []string{}
	for _, id := range ids {
		if isUUID(id) {
			uuidIDs = append(uuidIDs, id)
		} else {
			nonUUIDs = append(nonUUIDs, id)
		}
	}
	clauses := []string{}
	if len(uuidIDs) > 0 {
		clauses = append(clauses, "id.in.("+strings.Join(uuidIDs, ",")+")")
	}
	if controllerIDColumn != "" && len(nonUUIDs) > 0 {
		clauses = append(clauses, controllerIDColumn+".in.("+strings.Join(nonUUIDs, ",")+")")
	}
	clauses = append(clauses, nameColumn+".in.("+strings.Join(ids, ",")+")")
	query["or"] = []string{"(" + strings.Join(clauses, ",") + ")"}
	return query
}

// selectByIDThenName narrows resources to those requested by args. It provides precedence
// for selecting by ID > Name. Each arg is resolved independendently, resource order is
// preserved, and no resource is returned twice.
func selectByIDThenName[T any, TS ~[]T](
	args []string,
	resources TS,
	getIDFunc func(T) string,
	getControllerIDFunc func(T) string,
	getNameFunc func(T) string,
) TS {
	if len(args) == 0 {
		return resources
	}
	selected := make([]bool, len(resources))
	for _, arg := range args {
		if arg == "" {
			continue
		}
		matchedID := false
		for i, resource := range resources {
			if getIDFunc(resource) == arg || (getControllerIDFunc != nil && getControllerIDFunc(resource) == arg) {
				selected[i] = true
				matchedID = true
			}
		}
		if matchedID {
			continue
		}
		for i, resource := range resources {
			if getNameFunc(resource) == arg {
				selected[i] = true
			}
		}
	}
	filtered := make(TS, 0, len(resources))
	for i, resource := range resources {
		if selected[i] {
			filtered = append(filtered, resource)
		}
	}
	return filtered
}

// resolveIDs converts the given args, which may be ids, controller_ids, or names,
// into the database ids of the matching resources.
func resolveIDs[T any, TS ~[]T](
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	getter func(context.Context, *WorkbenchService, []string, string) (TS, error),
	getIDFunc func(T) string,
	resourceType string,
) ([]string, error) {
	if len(ids) == 0 {
		return ids, nil
	}
	resources, err := getter(ctx, s, ids, "")
	if err != nil {
		return nil, err
	}
	if len(resources) == 0 {
		return nil, errors.New(
			i18n.Format(
				i18n.ResourceNotFoundWithID,
				i18n.A{
					constants.I18nArgTypeKey: resourceType,
					constants.I18nArgKeyID:   strings.Join(ids, ", "),
				},
			),
		)
	}
	resolved := make([]string, 0, len(resources))
	for _, resource := range resources {
		resolved = append(resolved, getIDFunc(resource))
	}
	return resolved, nil
}

// Auth0IDsQuery returns an Auth0 compatible query for the given IDs.
func Auth0IDsQuery(ids []string, sortSpec string) url.Values {
	if len(ids) == 0 && sortSpec == "" {
		return url.Values{}
	}
	query := url.Values{}
	if len(ids) > 0 {
		query["id"] = ids
	}
	if sortSpec != "" {
		query["sortBy"] = []string{sortSpec}
	}
	return query
}

func isUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func storageClass(s *corev1.PersistentVolumeClaimSpec) string {
	if s == nil {
		return ""
	}
	if s.StorageClassName == nil {
		return ""
	}
	return *s.StorageClassName
}

func getServiceIDs(services []types.ServiceRequest) string {
	ids := []string{}
	for _, s := range services {
		id := s.ID
		image := s.Image
		if id == "" {
			id = image
			if id == "" {
				continue
			}
		}
		ids = append(ids, id)
	}
	return strings.Join(ids, ", ")
}

// bulkUpdateForUniqueNameConstrainedResources extracts the logic for updating
// resources with a uniqueness constraint on the name. If the name is not being
// updated then it does a bulk update. If the name is being updated then it does
// a single update for each resource and returns an error if there is more than
// one resource to update.
func bulkUpdateForUniqueNameConstrainedResources[T any, TS ~[]T](
	ctx context.Context,
	s *WorkbenchService,
	ids []string,
	body T,
	getNameFunc func(T) string,
	setIDFunc func(*T, string),
	resourceType string,
	bulkUpdateEndpoint string,
	updateEndpoint string,
	bulkBodyKey string,
) (TS, error) {
	// If not updating the name then we can do all the updates in one call.
	// There is a uniqueness constraint on names, so if we are updating the name
	// then we have to do each update in a separate call.
	if getNameFunc(body) == "" {
		return result.Split(Post[TS](
			ctx,
			s,
			postgrest.Endpoint(bulkUpdateEndpoint),
			map[string]interface{}{constants.RPCBodyKeyIDs: ids, bulkBodyKey: body},
		))
	}
	// If we are updating the name then we can only be updating one resource.
	// Otherwise, it will fail in the server.
	if len(ids) != 1 {
		err := errors.New(
			i18n.Format(
				i18n.BulkUpdateNameError,
				i18n.A{constants.I18nArgKeyType: resourceType},
			),
		)
		var ts TS
		return ts, err
	}
	setIDFunc(&body, ids[0])
	res, err := result.Split(Post[T](
		ctx,
		s,
		postgrest.Endpoint(updateEndpoint),
		map[string]interface{}{constants.RPCBodyKeyParam: body},
	))
	if err != nil {
		var ts TS
		return ts, err
	}
	return TS{res}, nil
}
