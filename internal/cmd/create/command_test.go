// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package create_test

import (
	"context"
	"testing"

	"github.com/sassoftware/wbctl/internal/cmd/create"
	"github.com/sassoftware/wbctl/internal/stubs"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand(t *testing.T) {
	mockCmd := create.CreateCommand(context.Background(), afero.NewMemMapFs(), &stubs.StubUI{})
	require.True(t, mockCmd.HasSubCommands())
}

func TestUpdateCommand(t *testing.T) {
	mockCmd := create.UpdateCommand(context.Background(), afero.NewMemMapFs(), &stubs.StubUI{})
	require.True(t, mockCmd.HasSubCommands())
}
