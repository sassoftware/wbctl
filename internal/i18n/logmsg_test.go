// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package i18n_test

import (
	"testing"

	"github.com/sassoftware/wbctl/internal/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBundle(t *testing.T) {
	b := i18n.Bundle()
	require.NotNil(t, b)
	actual := b.GetString(i18n.Error, i18n.Error)
	assert.Equal(t, "{error}", actual)
}

func TestFormat(t *testing.T) {
	msg := i18n.Format(i18n.Error, i18n.A{"error": "an error"})
	assert.Equal(t, "an error", msg)
}
