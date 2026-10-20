// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"strings"

	"github.com/sassoftware/wbctl/internal/constants"
)

// GetAccessTokenParams collects the inputs to GetAccessToken in a struct so
// they can be addressed by name.
type GetAccessTokenParams struct {
	ControlPlaneEndpoint string
	IngressHost          string
	Organization         string
	AuthMode             string
	ClientID             string
	IgnoreTokenCache     bool
	NoListen             bool
}

// NormalizedAuthMode returns the normalized auth mode. Anything other than pkce
// is returned as device.
func (p *GetAccessTokenParams) NormalizedAuthMode() string {
	authMode := strings.ToLower(strings.TrimSpace(p.AuthMode))
	if authMode == "" {
		return constants.AuthModeDevice
	}
	return authMode
}
