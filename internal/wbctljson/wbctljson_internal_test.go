// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wbctljson

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	expectedJSONEmptyValue    = true
	expectedJSONNotEmptyValue = false

	testEmptyString    = ""
	testNonEmptyString = "value"
	testZeroInt        = 0
	testNonZeroInt     = 1
	testZeroUint       = uint(0)
	testNonZeroUint    = uint(1)
	testZeroFloat      = 0.0
	testNonZeroFloat   = 1.25
	testFalseBool      = false
	testTrueBool       = true
)

var (
	testEmptyArrayValue     = reflect.ValueOf([0]int{})
	testNonEmptyArrayValue  = reflect.ValueOf([1]int{testNonZeroInt})
	testEmptyMapValue       = reflect.ValueOf(map[string]string{})
	testNonEmptyMapValue    = reflect.ValueOf(map[string]string{"k": testNonEmptyString})
	testEmptySliceValue     = reflect.ValueOf([]string{})
	testNonEmptySliceValue  = reflect.ValueOf([]string{testNonEmptyString})
	testEmptyStringValue    = reflect.ValueOf(testEmptyString)
	testNonEmptyStringValue = reflect.ValueOf(testNonEmptyString)
	testFalseBoolValue      = reflect.ValueOf(testFalseBool)
	testTrueBoolValue       = reflect.ValueOf(testTrueBool)
	testZeroIntValue        = reflect.ValueOf(testZeroInt)
	testNonZeroIntValue     = reflect.ValueOf(testNonZeroInt)
	testZeroUintValue       = reflect.ValueOf(testZeroUint)
	testNonZeroUintValue    = reflect.ValueOf(testNonZeroUint)
	testZeroFloatValue      = reflect.ValueOf(testZeroFloat)
	testNonZeroFloatValue   = reflect.ValueOf(testNonZeroFloat)

	testNilPointerInt *int
	testPointerInt    = testNonZeroInt
	testPointerValue  = &testPointerInt

	testNilInterfaceValue    = reflect.ValueOf([]any{nil}).Index(0)
	testNonNilInterfaceValue = reflect.ValueOf([]any{testNonEmptyString}).Index(0)
	testNilPointerValue      = reflect.ValueOf(testNilPointerInt)
	testNonNilPointerValue   = reflect.ValueOf(testPointerValue)

	testStructValue = reflect.ValueOf(struct{ Value int }{Value: testZeroInt})
)

func TestIsJSONEmptyValue(t *testing.T) {
	testCases := []struct {
		name     string
		input    reflect.Value
		expected bool
	}{
		{
			name:     "empty array is empty",
			input:    testEmptyArrayValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "empty map is empty",
			input:    testEmptyMapValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "empty slice is empty",
			input:    testEmptySliceValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "empty string is empty",
			input:    testEmptyStringValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "false bool is empty",
			input:    testFalseBoolValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "zero int is empty",
			input:    testZeroIntValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "zero uint is empty",
			input:    testZeroUintValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "zero float is empty",
			input:    testZeroFloatValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "nil pointer is empty",
			input:    testNilPointerValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "nil interface is empty",
			input:    testNilInterfaceValue,
			expected: expectedJSONEmptyValue,
		},
		{
			name:     "non-empty array is not empty",
			input:    testNonEmptyArrayValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-empty map is not empty",
			input:    testNonEmptyMapValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-empty slice is not empty",
			input:    testNonEmptySliceValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-empty string is not empty",
			input:    testNonEmptyStringValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "true bool is not empty",
			input:    testTrueBoolValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-zero int is not empty",
			input:    testNonZeroIntValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-zero uint is not empty",
			input:    testNonZeroUintValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-zero float is not empty",
			input:    testNonZeroFloatValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-nil pointer is not empty",
			input:    testNonNilPointerValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "non-nil interface is not empty",
			input:    testNonNilInterfaceValue,
			expected: expectedJSONNotEmptyValue,
		},
		{
			name:     "struct is not empty",
			input:    testStructValue,
			expected: expectedJSONNotEmptyValue,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := isJSONEmptyValue(testCase.input)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

type testParseJSONTagFixture struct {
	Editable        string `json:"renamed,omitempty" wbctl:"editable,seedfield:FallbackField"`
	UntaggedIgnored string `wbctl:"-"`
}

type testCanonicalFieldFixture struct {
	Named   string `json:"named,omitempty"`
	Unnamed string
	Skipped string `json:"-"`
}

type testNormalizeChild struct {
	Value string `json:"value,omitempty" wbctl:"editable"`
}

type testNormalizeParent struct {
	Child   testNormalizeChild `json:"child,omitempty"   wbctl:"editable"`
	Missing string             `json:"missing,omitempty" wbctl:"editable"`
	Ignored string             `json:"ignored,omitempty" wbctl:"-"`
	Plain   string             `json:"plain,omitempty"`
	Skipped string             `json:"-"`
}

var (
	testParseJSONTagEditableField = reflect.TypeOf(testParseJSONTagFixture{}).Field(0)
	testParseJSONTagIgnoredField  = reflect.TypeOf(testParseJSONTagFixture{}).Field(1)

	testCanonicalNamedField   = reflect.TypeOf(testCanonicalFieldFixture{}).Field(0)
	testCanonicalUnnamedField = reflect.TypeOf(testCanonicalFieldFixture{}).Field(1)
	testCanonicalSkippedField = reflect.TypeOf(testCanonicalFieldFixture{}).Field(2)

	testNormalizeParentType        = reflect.TypeOf(testNormalizeParent{})
	testNormalizeParentPointerType = reflect.TypeOf((*testNormalizeParent)(nil))
	testNormalizeChildSliceType    = reflect.TypeOf([]testNormalizeChild{})
	testNormalizeChildMapType      = reflect.TypeOf(map[string]testNormalizeChild{})
	testNormalizeStringType        = reflect.TypeOf("")
	testNormalizeStructInput       = map[string]any{
		"child":   map[string]any{"value": "child-value"},
		"ignored": "kept",
		"extra":   "preserved",
	}
	testNormalizeStructExpected = map[string]any{
		"child":   map[string]any{"value": "child-value"},
		"ignored": "kept",
		"extra":   "preserved",
	}
	testNormalizePassthroughInput    = map[string]any{"other": "value"}
	testNormalizePassthroughExpected = map[string]any{"other": "value"}
	testNormalizeSliceInput          = []any{map[string]any{"value": "slice-value"}, "plain"}
	testNormalizeSliceExpected       = []any{map[string]any{"value": "slice-value"}, "plain"}
	testNormalizeMapInput            = map[string]any{"first": map[string]any{"value": "map-value"}}
	testNormalizeMapExpected         = map[string]any{"first": map[string]any{"value": "map-value"}}
	testNormalizeStringInput         = "plain"
	testNormalizeStringExpected      = "plain"
)

func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected wbctlTag
	}{
		{
			name:  "parses explicit name editable seedfield",
			field: testParseJSONTagEditableField,
			expected: wbctlTag{
				name:             "renamed",
				omitempty:        true,
				editable:         true,
				defaultFieldName: "FallbackField",
			},
		},
		{
			name:  "falls back to field name and ignores field",
			field: testParseJSONTagIgnoredField,
			expected: wbctlTag{
				name:   "UntaggedIgnored",
				ignore: true,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := parseJSONTag(testCase.field)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestParseCanonicalJSONFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
		included bool
	}{
		{
			name:     "uses explicit json name",
			field:    testCanonicalNamedField,
			expected: "named",
			included: true,
		},
		{
			name:     "falls back to field name",
			field:    testCanonicalUnnamedField,
			expected: "Unnamed",
			included: true,
		},
		{
			name:     "skips json excluded field",
			field:    testCanonicalSkippedField,
			expected: "",
			included: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actualName, actualIncluded := parseCanonicalJSONFieldName(testCase.field)
			require.Equal(t, testCase.expected, actualName)
			require.Equal(t, testCase.included, actualIncluded)
		})
	}
}

func TestNormalizeWbctlValue(t *testing.T) {
	tests := []struct {
		name       string
		input      any
		targetType reflect.Type
		expected   any
	}{
		{
			name:       "pointer to struct normalizes nested map",
			input:      testNormalizeStructInput,
			targetType: testNormalizeParentPointerType,
			expected:   testNormalizeStructExpected,
		},
		{
			name:       "struct target keeps scalar input",
			input:      testNormalizeStringInput,
			targetType: testNormalizeParentType,
			expected:   testNormalizeStringExpected,
		},
		{
			name:       "slice target normalizes list input",
			input:      testNormalizeSliceInput,
			targetType: testNormalizeChildSliceType,
			expected:   testNormalizeSliceExpected,
		},
		{
			name:       "slice target keeps scalar input",
			input:      testNormalizeStringInput,
			targetType: testNormalizeChildSliceType,
			expected:   testNormalizeStringExpected,
		},
		{
			name:       "map target normalizes nested input",
			input:      testNormalizeMapInput,
			targetType: testNormalizeChildMapType,
			expected:   testNormalizeMapExpected,
		},
		{
			name:       "map target keeps scalar input",
			input:      testNormalizeStringInput,
			targetType: testNormalizeChildMapType,
			expected:   testNormalizeStringExpected,
		},
		{
			name:       "scalar target keeps scalar input",
			input:      testNormalizeStringInput,
			targetType: testNormalizeStringType,
			expected:   testNormalizeStringExpected,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := normalizeWbctlValue(testCase.input, testCase.targetType)
			require.Equal(t, testCase.expected, actual)
		})
	}
}

func TestNormalizeWbctlStructMap(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]any
	}{
		{
			name:     "recurses nested struct and preserves unknown keys",
			input:    testNormalizeStructInput,
			expected: testNormalizeStructExpected,
		},
		{
			name:     "passes through when no fields match",
			input:    testNormalizePassthroughInput,
			expected: testNormalizePassthroughExpected,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := normalizeWbctlStructMap(testCase.input, testNormalizeParentType)
			require.Equal(t, testCase.expected, actual)
		})
	}
}
