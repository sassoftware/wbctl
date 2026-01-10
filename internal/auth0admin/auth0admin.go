// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth0admin

import (
	"net/url"
	"path"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
)

var endpointBase = constants.Auth0AdminAPIEndpoint

// Endpoint prepends the given url part with the postgrest endpoint. If there is
// a problem joining as a url, it joins as a path.
func Endpoint(endpoint string) string {
	joinedURL, err := url.JoinPath(endpointBase, endpoint)
	if err != nil {
		return path.Join(endpointBase, endpoint)
	}
	return joinedURL
}
