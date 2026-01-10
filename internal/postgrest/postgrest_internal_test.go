// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package postgrest

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEndpoint_FallbackWhenJoinPathErrors(t *testing.T) {
	originalBase := endpointBase
	t.Cleanup(func() {
		endpointBase = originalBase
	})

	// Control chars in the base make net/url parsing fail, which forces
	// fallback. Control chars in the path are substituted with their hex
	// representation.
	endpointBase = "\n"

	input := "/rpc/job_create"
	require.Equal(t, path.Join(endpointBase, input), Endpoint(input))
}
