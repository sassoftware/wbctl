// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"encoding/json"
	"fmt"
)

// AsString returns the given value as a json marshaled string.
func AsString[T any](
	v T,
) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("FAILED TO MARSHAL: %v", v)
	}
	return string(jsonBytes)
}
