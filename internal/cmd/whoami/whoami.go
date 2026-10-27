// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package whoami

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/sas-ggdk/pkg/pointer"
	"github.com/sassoftware/sas-ggdk/pkg/result"
	"github.com/sassoftware/wbctl/internal/auth"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/timeutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newWhoamiFunction creates a command action that prints information about the
// currently logged in user.
func newWhoamiFunction(
	_ context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		endpoint := v.GetString(constants.EndpointFlag)
		idtoken := auth.GetCachedIdentityToken(
			fsys,
			v,
			endpoint,
		)
		if constants.JSONOutputValue == v.GetString(constants.OutputFlag) {
			return printAsJSON(_ui, idtoken)
		}
		return printIDToken(_ui, idtoken)
	}
}

func printAsJSON(_ui ui.UI, idtoken string) error {
	if idtoken == "" {
		_ui.Display("{}")
		return nil
	}
	// parse the idtoken as a map[string]any to preserve all fields that may not
	// be preserved by the auth.JWT type.
	jwtPayloadBytes := result.New(auth.ParseTokenPayloadAsBytes(idtoken))
	jwtPayloadMap := result.Map(func(payloadBytes []byte) (map[string]any, error) {
		var asMap map[string]any
		err := json.Unmarshal(payloadBytes, &asMap)
		return asMap, err
	}, jwtPayloadBytes)
	bytes := result.Map(func(payloadMap map[string]any) ([]byte, error) {
		return wbctljson.MakeMarshaler[map[string]any](wbctljson.ViewOptions)(payloadMap)
	}, jwtPayloadMap)
	if bytes.IsError() {
		return errors.Wrap(bytes.Error(), i18n.Format(i18n.FailedToMarshalJWTPayloadToJSON))
	}
	_ui.Display(string(bytes.OrElse([]byte("{}"))))
	return nil
}

func printIDToken(
	_ui ui.UI,
	idtoken string,
) error {
	if idtoken == "" {
		_ui.Display(i18n.Format(i18n.NotAuthenticated))
		return nil
	}
	jwtPayload, err := auth.ParseTokenPayload(idtoken)
	if err != nil {
		return errors.Wrap(err, i18n.Format(i18n.FailedToMarshalJWTPayloadToJSON))
	}
	var tabbedString strings.Builder
	now := timeutil.Now()
	expires := time.Unix(int64(jwtPayload.Exp), 0)
	expiredLabel := ""
	if expires.Before(now) {
		expiredLabel = " (EXPIRED)"
	}
	printutil.AppendTabbedString(&tabbedString, func(w io.Writer) {
		_, _ = fmt.Fprintf(w, "User:\t%s (%s)\n", jwtPayload.Name, jwtPayload.Sas.UserID)
		_, _ = fmt.Fprintf(w, "Organization:\t%s (%s)\n", jwtPayload.Sas.OrgName, jwtPayload.Sas.OrgID)
		_, _ = fmt.Fprintf(w, "Roles:\t%s\n", strings.Join(jwtPayload.Sas.Roles, ", "))
		_, _ = fmt.Fprintf(w, "Issued:\t%s\n", printutil.TimePtrToString(pointer.Ptr(time.Unix(int64(jwtPayload.Iat), 0))))
		_, _ = fmt.Fprintf(w, "Expires:\t%s%s\n", printutil.TimePtrToString(pointer.Ptr(expires)), expiredLabel)
	})
	_ui.Display(tabbedString.String())
	return nil
}
