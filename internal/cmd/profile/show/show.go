// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package show

import (
	"context"
	"fmt"
	"io"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/sassoftware/sas-ggdk/pkg/errors"
	"github.com/sassoftware/wbctl/internal/cmd/profile/config"
	cmdtypes "github.com/sassoftware/wbctl/internal/cmd/types"
	"github.com/sassoftware/wbctl/internal/configutil"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/sassoftware/wbctl/internal/printutil"
	"github.com/sassoftware/wbctl/internal/ui"
	"github.com/sassoftware/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newShowFunction creates a command action that prints the values of the
// current profile.
func newShowFunction(
	_ context.Context,
	fsys afero.Fs,
	_ui ui.UI,
) cmdtypes.ActionFunc {
	return func(
		_ *cobra.Command,
		_ []string,
	) error {
		v := viper.GetViper()
		profileName := v.GetString(constants.ProfileFlag)
		configPath, _, err := configutil.ResolvePath()
		if err != nil {
			return errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
		}
		doc, err := config.ReadDocument(fsys, configPath)
		if err != nil {
			return err
		}
		// Grab any values that apply to all profiles
		values := copyNonMapValues(doc, map[string]any{})
		// Grab the values from the current profile
		for key := range doc {
			if key != profileName {
				continue
			}
			profileDoc, ok := (doc[key]).(map[string]any)
			if !ok {
				msg := i18n.Format(i18n.InvalidConfigFile, i18n.A{"path": configPath})
				err = errors.New(msg)
				return errors.Wrap(err, i18n.Format(i18n.ConfigFileReadFailed, i18n.A{constants.I18nArgKeyPath: configPath}))
			}
			values = copyNonMapValues(profileDoc, values)
		}
		if constants.JSONOutputValue == v.GetString(constants.OutputFlag) {
			response, err := wbctljson.MakeMarshaler[map[string]any](wbctljson.ViewOptions)(values)
			if err == nil {
				_ui.Display(string(response))
			}
			return err
		}
		printProfileValues(_ui, values)
		return nil
	}
}

func printProfileValues(_ui ui.UI, values map[string]any) {
	keys := slices.Collect(maps.Keys(values))
	slices.Sort(keys)
	var tabbedString strings.Builder
	printutil.AppendTabbedString(&tabbedString, func(w io.Writer) {
		for _, key := range keys {
			_, _ = fmt.Fprintf(w, "%s:\t%v\n", key, values[key])
		}
	})
	_ui.Display(tabbedString.String())
}

func copyNonMapValues(src, dst map[string]any) map[string]any {
	for key, value := range src {
		if isMap(value) {
			continue
		}
		dst[key] = value
	}
	return dst
}

func isMap(value any) bool {
	if value != nil {
		rv := reflect.ValueOf(value)
		return rv.Kind() == reflect.Map
	}
	return false
}
