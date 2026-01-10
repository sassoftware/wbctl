// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package workbench

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
	"github.com/sas-institute-rnd-saas/wbctl/internal/postgrest"
	"github.com/sas-institute-rnd-saas/wbctl/internal/types"
	"github.com/sassoftware/sas-ggdk/pkg/result"
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

// IDsOrControllerIDsQuery returns a PostgREST compatible query for the given
// IDs and sort spec that will match either the actual id or the controller_id
// columns.
func IDsOrControllerIDsQuery(ids []string, sortSpec string) url.Values {
	query := url.Values{}
	if sortSpec != "" {
		query["order"] = []string{sortSpec}
	}
	uuidIDs := []string{}
	controllerIDs := []string{}
	for _, id := range ids {
		if isUUID(id) {
			uuidIDs = append(uuidIDs, id)
		} else {
			controllerIDs = append(controllerIDs, id)
		}
	}
	switch {
	case len(uuidIDs) > 0 && len(controllerIDs) > 0:
		query["or"] = []string{
			"(id.in.(" + strings.Join(
				uuidIDs,
				",",
			) + "),controller_id.in.(" + strings.Join(
				controllerIDs,
				",",
			) + "))",
		}
	case len(uuidIDs) > 0:
		query["id"] = []string{"in.(" + strings.Join(uuidIDs, ",") + ")"}
	case len(controllerIDs) > 0:
		query["controller_id"] = []string{"in.(" + strings.Join(controllerIDs, ",") + ")"}
	}
	return query
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
