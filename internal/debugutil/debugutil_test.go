// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package debugutil_test

import (
	"runtime/debug"
	"testing"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/debugutil"
	"github.com/stretchr/testify/require"
)

func TestGetSetting(t *testing.T) {
	expected := ""
	actual := debugutil.GetSetting(nil, constants.BuildSettingGOARCHKey)
	require.Equal(t, expected, actual)

	expected = "value"
	actual = debugutil.GetSetting([]debug.BuildSetting{
		{Key: constants.BuildSettingGOOSKey, Value: expected},
	}, constants.BuildSettingGOOSKey)
	require.Equal(t, expected, actual)
}
