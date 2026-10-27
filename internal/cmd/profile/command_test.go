// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package profile_test

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd/profile"
	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCommand(t *testing.T) {
	cmd := profile.Command(context.Background(), afero.NewMemMapFs(), &stubs.StubUI{})
	require.Equal(t, constants.ProfileCommandName, cmd.Name())
	require.Contains(t, cmd.Aliases, constants.ProfileAliasName)

	sub, _, err := cmd.Find([]string{constants.InitCommandName})
	require.NoError(t, err)
	require.Equal(t, constants.InitCommandName, sub.Name())
}
