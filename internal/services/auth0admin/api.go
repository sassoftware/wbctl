// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth0admin

import (
	"context"

	"github.com/sas-institute-rnd-saas/wbctl/internal/auth0admin"
	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
	"github.com/sassoftware/sas-ggdk/pkg/result"
)

// API is the API specification for the Auth0Admin service.
type API struct {
	Version int    `json:"version,omitempty"`
	Links   []Link `json:"links,omitempty"`
}

// Link is the respresentation of a service link.
type Link struct {
	Method           string `json:"method,omitempty"`
	Rel              string `json:"rel,omitempty"`
	Href             string `json:"href,omitempty"`
	URI              string `json:"uri,omitempty"`
	MediaType        string `json:"type,omitempty"`
	ResponseType     string `json:"responseType,omitempty"`
	ItemType         string `json:"itemType,omitempty"`
	ResponseItemType string `json:"responseItemType,omitempty"`
	Title            string `json:"title,omitempty"`
}

// GetAPI returns the API from the WorkbenchService.
func GetAPI(
	ctx context.Context,
	s *Auth0AdminService,
) (API, error) {
	return result.Split(Get[API](ctx, s, auth0admin.Endpoint(constants.Auth0AdminEndpointRoot), nil))
}
