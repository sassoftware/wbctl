// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package stubs

import (
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port returns the port part of the given url.
func Port(t *testing.T, _url string) string {
	p, err := url.Parse(_url)
	require.NoError(t, err)
	return p.Port()
}

// Host returns the host part of the given url.
func Host(t *testing.T, _url string) string {
	p, err := url.Parse(_url)
	require.NoError(t, err)
	return p.Hostname()
}

// SanitizeSlice calls Sanitize on each pair of strings in the expectedSlice and
// actualSlice slices. If there are a different number of elements in each slice
// then they are returned unchanged.
func SanitizeSlice(expectedSlice, actualSlice []string, replacements map[string]string) ([]string, []string) {
	if len(expectedSlice) != len(actualSlice) {
		return expectedSlice, actualSlice
	}
	dstExpected := make([]string, 0, len(expectedSlice))
	dstActual := make([]string, 0, len(actualSlice))
	for i := range expectedSlice {
		expected := expectedSlice[i]
		actual := actualSlice[i]
		expected, actual = Sanitize(expected, actual, replacements)
		dstExpected = append(dstExpected, expected)
		dstActual = append(dstActual, actual)
	}
	return dstExpected, dstActual
}

// Sanitize is a testing utility function. It takes an expected string and
// actual string and a list of name/value pairs. First, any occurrences of
// {SANITIZE:regexp:replacement} are found in the expected string. Any matches
// of the regexp in the actual string are replaced with the replacement value.
// Then the entire {SANITIZE:regexp:replacement} string in the expected string
// is replaced with the replacement value. Lastly, any occurrences of the keys
// of the replacements map are replaced in the expected string with the
// corresponding values. See TestExampleSanitizeUsage for an example.
func Sanitize(expected, actual string, replacements map[string]string) (string, string) {
	re := regexp.MustCompile(`{SANITIZE:(.*?)}`)
	replacementSpecs := re.FindAllString(expected, -1)
	for _, spec := range replacementSpecs {
		firstIdx := strings.Index(spec, ":")
		lastIdx := strings.LastIndex(spec, ":")
		pattern := spec[firstIdx+1 : lastIdx]
		replacement := strings.TrimSuffix(spec[lastIdx+1:], "}")
		san := regexp.MustCompile(pattern)
		actual = san.ReplaceAllString(actual, replacement)
		expected = strings.ReplaceAll(expected, spec, replacement)
	}
	expected = Replace(expected, replacements)
	return expected, actual
}

// Replace returns a copy fo the given src string with all instances of keys in
// the given replacements map with the corresponding values.
func Replace(src string, replacements map[string]string) string {
	for k, v := range replacements {
		src = strings.ReplaceAll(src, k, v)
	}
	return src
}
