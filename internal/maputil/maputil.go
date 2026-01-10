// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package maputil

import (
	"encoding/json"
	"strings"
)

// FromEqualDelimitedLines converts key=value strings into a map.
func FromEqualDelimitedLines(equalDelimitedLines []string) map[string]string {
	parsedLines := map[string]string{}
	for _, label := range equalDelimitedLines {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) != 2 {
			parsedLines[parts[0]] = ""
		} else {
			parsedLines[parts[0]] = parts[1]
		}
	}
	if len(parsedLines) == 0 {
		return nil
	}
	return parsedLines
}

// SafeValueAsString returns a map value as a string when possible.
func SafeValueAsString(m map[string]interface{}, key string) string {
	val, present := m[key]
	if present {
		strVal, ok := val.(string)
		if ok {
			return strVal
		}
		jsonVal, err := json.Marshal(val)
		if err == nil {
			return string(jsonVal)
		}
	}
	return ""
}

// MapNoError maps src with f and returns the results.
func MapNoError[T, R any](f func(T) R, src []T) []R {
	length := len(src)
	if length == 0 {
		return nil
	}
	res := make([]R, 0, length)
	for _, v := range src {
		res = append(res, f(v))
	}
	return res
}
