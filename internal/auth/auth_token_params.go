// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"strings"

	"github.com/sas-institute-rnd-saas/wbctl/internal/constants"
)

// GetAccessTokenParams collects the inputs to GetAccessToken in a struct so
// they can be addressed by name.
type GetAccessTokenParams struct {
	ControlPlaneEndpoint string
	IngressHost          string
	Organization         string
	AuthMode             string
	Username             string
	Password             string
	ClientID             string
	ClientSecret         string
	IgnoreTokenCache     bool
	NoListen             bool
}

// NormalizedAuthMode returns the normalized auth mode. If the auth mode is
// "auto", it will be determined based on the presence of client ID and
// username.
func (p *GetAccessTokenParams) NormalizedAuthMode() string {
	authMode := strings.ToLower(strings.TrimSpace(p.AuthMode))
	if authMode == "" {
		authMode = constants.AuthModeAuto
	}
	if authMode == constants.AuthModeAuto {
		if p.Username != "" {
			return constants.AuthModePassword
		}
		if p.ClientSecret != "" {
			return constants.AuthModeClientCredentials
		}
		return constants.AuthModePKCE
	}
	return authMode
}
