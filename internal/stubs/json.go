// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"encoding/json"
	"testing"
)

// AsString returns the given value as a json marshaled string.
func AsString[T any](
	t *testing.T,
	v T,
) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return string(jsonBytes)
}
