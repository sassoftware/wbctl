// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package wbctljson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strings"

	"github.com/sassoftware/wbctl/internal/constants"
	"github.com/sassoftware/wbctl/internal/i18n"
)

const (
	wbctlTagName    = `wbctl`
	jsonTagName     = `json`
	omitemptyTag    = `omitempty`
	editableTag     = `editable`
	seedfieldTag    = `seedfield`
	seedfieldPrefix = seedfieldTag + `:`
)

// MarshalerOptions are the options that control the wbctljson marshaling
// functions.
type MarshalerOptions struct {
	EditOnly bool
	Pretty   bool
	Indent   string
}

// EditOptions is the set of options used when marshaling a resource for the
// user to edit. This is either via the edit command or via the --dry-run option
// passed to update or create.
var EditOptions = MarshalerOptions{
	EditOnly: true,
	Pretty:   true,
	Indent:   "    ",
}

// ViewOptions is the set of options used when marshaling a resource for the
// customer to view.
var ViewOptions = MarshalerOptions{
	EditOnly: false,
	Pretty:   true,
	Indent:   "    ",
}

// MakeMarshaler returns s function that marshals a struct (or struct pointer)
// to JSON, if editOnly is true then only fields tagged with `wbctl:"editable"`
// will be included. It also accepts map[string]any and applies the same
// editable-field filtering to struct values in the map at any nesting level as
// long as it is map[string]any all the way to the struct (see "nested valid
// map" test).
func MakeMarshaler[T any](opts MarshalerOptions) func(T) ([]byte, error) {
	return func(t T) ([]byte, error) {
		// The public entry point always starts at the root indent level.
		// Recursive calls add indentation as they descend into nested
		// structures.
		return marshalWbctlType(t, opts, "")
	}
}

// MakeUnmarshaler returns a function that accepts wbctl JSON and normalizes
// field names before decoding into the typed API struct.
func MakeUnmarshaler[T any]() func(io.Reader, *T) error {
	return func(r io.Reader, target *T) error {
		content, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		return unmarshalWbctlType(content, target)
	}
}

func marshalWbctlType[T any](
	t T,
	opts MarshalerOptions,
	currentPrefix string,
) ([]byte, error) {
	// Normalize pointers early so downstream code can reason in terms of the
	// underlying value kind instead of repeating pointer checks in each branch.
	value := reflect.Indirect(reflect.ValueOf(t))
	valueType := value.Type()

	// We only support two top-level shapes for this custom marshaller:
	// map[string]... and structs tagged with wbctl metadata.
	switch valueType.Kind() {
	case reflect.Map:
		return marshalMapWbctlType(value, opts, currentPrefix)
	case reflect.Struct:
		return marshalStructWbctlType(value, opts, currentPrefix)
	case reflect.Slice:
		return marshalSliceWbctlType(value, opts, currentPrefix)
	default:
		return nil, errors.New(
			i18n.Format(
				i18n.UnsupportedType,
				i18n.A{constants.I18nArgTypeKey: valueType.Kind()},
			),
		)
	}
}

func unmarshalWbctlType[T any](content []byte, target *T) error {
	// Match encoding/json behavior for nil targets so callers see the standard
	// InvalidUnmarshalError contract.
	if target == nil {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(target)}
	}

	// Decode to untyped JSON first so field names can be normalized before
	// strongly-typed decoding enforces struct field names.
	var intermediate any
	if err := json.Unmarshal(content, &intermediate); err != nil {
		return err
	}

	// Recursively normalize wbctl field names using the destination type as
	// schema guidance.
	normalized := normalizeWbctlValue(intermediate, reflect.TypeOf(*target))

	// Re-encode normalized data and let encoding/json perform the final typed
	// decode (including all native type checks and conversions).
	//
	// `normalized` is always composed of values produced by json.Unmarshal plus
	// key rewrites, so encoding/json marshaling cannot fail due to unsupported
	// value kinds.
	// nolint: errchkjson
	normalizedContent, _ := json.Marshal(normalized)

	return json.Unmarshal(normalizedContent, target)
}

func marshalMapWbctlType(
	mapValue reflect.Value,
	opts MarshalerOptions,
	currentPrefix string,
) ([]byte, error) {
	buf := bytes.Buffer{}
	// Begin the object
	buf.WriteString("{")
	if opts.Pretty {
		buf.WriteString("\n" + currentPrefix)
	}

	mapType := mapValue.Type()
	nextPrefix := currentPrefix + opts.Indent
	fieldWritten := false

	// JSON object keys must be strings.
	if mapType.Key().Kind() != reflect.String {
		return nil, errors.New(
			i18n.Format(
				i18n.UnsupportedMapKeyKind,
				i18n.A{constants.I18nArgKindKey: mapType.Key().Kind()},
			),
		)
	}

	// Sort keys so tests, diffs, and user edits are stable across runs.
	keys := sortedMapKeys(mapValue)
	for _, key := range keys {
		// Map values are often stored as interface{} or pointers. Unwrap it to
		// get the real type so we can decide whether to recurse with wbctl
		// filtering or marshal raw.
		mapEntry := unwrapInterfaceOrPointer(mapValue.MapIndex(reflect.ValueOf(key)))

		var valueJSON []byte
		var err error
		if mapEntry.Kind() == reflect.Struct {
			// Struct map values are treated like first-class resources, so recurse
			// and apply editable-field filtering to them too.
			valueJSON, err = marshalWbctlType(
				mapEntry.Interface(),
				opts,
				nextPrefix,
			)
		} else {
			// Non-struct values keep normal JSON behavior since they won't have
			// wbctl tags.
			if opts.Pretty {
				valueJSON, err = json.MarshalIndent(mapEntry.Interface(), nextPrefix, opts.Indent)
			} else {
				valueJSON, err = json.Marshal(mapEntry.Interface())
			}
		}
		if err != nil {
			return nil, err
		}

		// Write the field name with prefixing and comma handling.
		writeJSONField(&buf, opts.Pretty, fieldWritten, nextPrefix, key)

		// Write the field value.
		buf.Write(valueJSON)
		fieldWritten = true
	}

	// Close the object
	if opts.Pretty {
		buf.WriteString("\n" + currentPrefix)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

// nolint: cyclop
func marshalStructWbctlType(
	structValue reflect.Value,
	opts MarshalerOptions,
	currentPrefix string,
) ([]byte, error) {
	buf := bytes.Buffer{}
	// Begin the object
	buf.WriteString("{")
	if opts.Pretty {
		buf.WriteString("\n")
	}

	structType := structValue.Type()
	nextPrefix := currentPrefix + opts.Indent
	fieldWritten := false

	for i := range structValue.NumField() {
		typeField := structType.Field(i)
		// Figure out what the tags want us to do
		wbctlTagInfo := parseJSONTag(typeField)

		// If the field is ignored, or we are emitting for edit and it does not
		// have the `wbctl:"editable"` tag, then skip it.
		if wbctlTagInfo.ignore || (!wbctlTagInfo.editable && opts.EditOnly) {
			continue
		}
		valueField := structValue.Field(i)
		// If the primary field is zero-valued, seedfield allows using a fallback
		// field as the source of truth for edit payload generation.
		if valueField.IsZero() && wbctlTagInfo.defaultFieldName != "" {
			valueField = structValue.FieldByName(wbctlTagInfo.defaultFieldName)
			if !valueField.IsValid() {
				continue
			}
		}
		// Respect omitempty using encoding/json-style emptiness semantics.
		if wbctlTagInfo.omitempty && isJSONEmptyValue(valueField) {
			continue
		}

		// Marshal each editable field independently so we can preserve
		// object-level formatting while still delegating value encoding to
		// encoding/json.
		var valueJSON []byte
		var err error
		if opts.Pretty {
			valueJSON, err = json.MarshalIndent(valueField.Interface(), nextPrefix, opts.Indent)
		} else {
			valueJSON, err = json.Marshal(valueField.Interface())
		}
		if err != nil {
			return nil, err
		}

		// Write the field name with prefixing and comma handling.
		writeJSONField(&buf, opts.Pretty, fieldWritten, nextPrefix, wbctlTagInfo.name)

		// Write the field value
		buf.Write(valueJSON)
		fieldWritten = true
	}

	if !fieldWritten {
		return []byte("{}"), nil
	}

	// Close the object
	if opts.Pretty {
		buf.WriteString("\n" + currentPrefix)
	}
	buf.WriteString("}")
	return buf.Bytes(), nil
}

func marshalSliceWbctlType(
	sliceValue reflect.Value,
	opts MarshalerOptions,
	currentPrefix string,
) ([]byte, error) {
	buf := bytes.Buffer{}
	// Begin the array
	buf.WriteString("[")
	if opts.Pretty {
		buf.WriteString("\n")
	}

	nextPrefix := currentPrefix + opts.Indent
	itemWritten := false

	for i := range sliceValue.Len() {
		sliceEntry := unwrapInterfaceOrPointer(sliceValue.Index(i))

		var valueJSON []byte
		var err error
		if sliceEntry.Kind() == reflect.Struct {
			// Struct slice elements are treated like first-class resources, so recurse
			// and apply editable-field filtering to them too.
			valueJSON, err = marshalWbctlType(
				sliceEntry.Interface(),
				opts,
				nextPrefix,
			)
		} else {
			// Non-struct values keep normal JSON behavior since they won't have
			// wbctl tags.
			if opts.Pretty {
				valueJSON, err = json.MarshalIndent(sliceEntry.Interface(), nextPrefix, opts.Indent)
			} else {
				valueJSON, err = json.Marshal(sliceEntry.Interface())
			}
		}
		if err != nil {
			return nil, err
		}

		if itemWritten {
			buf.WriteString(",\n")
		}
		buf.WriteString(nextPrefix)
		buf.Write(valueJSON)
		itemWritten = true
	}

	// Close the array
	if opts.Pretty {
		buf.WriteString("\n" + currentPrefix)
	}
	buf.WriteString("]")
	return buf.Bytes(), nil
}

func isJSONEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	default:
		return false
	}
}

func sortedMapKeys(mapValue reflect.Value) []string {
	keys := make([]string, 0, mapValue.Len())
	iter := mapValue.MapRange()
	for iter.Next() {
		keys = append(keys, iter.Key().String())
	}
	slices.Sort(keys)
	return keys
}

func unwrapInterfaceOrPointer(value reflect.Value) reflect.Value {
	// Reflection frequently returns interface/pointer wrappers. Peel them away
	// until we hit a concrete value so kind checks reflect real payload shape.
	for (value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer) && !value.IsNil() {
		value = value.Elem()
	}
	return value
}

// nolint: cyclop
func normalizeWbctlValue(value any, targetType reflect.Type) any {
	// Normalize through pointer layers so recursion decisions are based on the
	// concrete destination shape.
	for targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	switch targetType.Kind() {
	case reflect.Struct:
		// Struct normalization is where field-name mapping happens; non-object
		// JSON values are left untouched.
		mapValue, ok := value.(map[string]any)
		if !ok {
			return value
		}
		return normalizeWbctlStructMap(mapValue, targetType)
	case reflect.Slice, reflect.Array:
		// Normalize each entry using the element type so nested lists of structs
		// are handled as well.
		listValue, ok := value.([]any)
		if !ok {
			return value
		}
		normalizedList := make([]any, len(listValue))
		for i := range listValue {
			normalizedList[i] = normalizeWbctlValue(listValue[i], targetType.Elem())
		}
		return normalizedList
	case reflect.Map:
		// Only string-keyed JSON objects are expected here; recurse into values
		// so map entries that contain structs are normalized too.
		mapValue, ok := value.(map[string]any)
		if !ok {
			return value
		}
		normalizedMap := make(map[string]any, len(mapValue))
		for key, rawValue := range mapValue {
			normalizedMap[key] = normalizeWbctlValue(rawValue, targetType.Elem())
		}
		return normalizedMap
	default:
		return value
	}
}

func normalizeWbctlStructMap(mapValue map[string]any, targetType reflect.Type) map[string]any {
	// Start from a copy so unknown fields are preserved exactly as provided while
	// known fields can be rewritten to canonical names.
	normalizedMap := make(map[string]any, len(mapValue))
	for key, value := range mapValue {
		normalizedMap[key] = value
	}

	for i := range targetType.NumField() {
		typeField := targetType.Field(i)
		wbctlTagInfo := parseJSONTag(typeField)
		if wbctlTagInfo.ignore {
			continue
		}

		jsonName, include := parseCanonicalJSONFieldName(typeField)
		if !include {
			continue
		}

		// Prefer canonical json keys and recurse into matching values.
		rawValue, exists := mapValue[jsonName]
		if !exists {
			continue
		}

		// Recurse using the field type so normalization also applies to nested
		// structs, slices, and maps.
		normalizedValue := normalizeWbctlValue(rawValue, typeField.Type)
		normalizedMap[jsonName] = normalizedValue
	}

	return normalizedMap
}

func writeJSONField(buf *bytes.Buffer, pretty bool, fieldWritten bool, indentPrefix, name string) {
	// JSON object entries require commas between fields but never before the
	// first field.
	if fieldWritten {
		buf.WriteString(",")
		if pretty {
			buf.WriteString("\n")
		}
	}
	fmt.Fprintf(buf, "%s\"%s\":", indentPrefix, name)
	if pretty {
		buf.WriteString(" ")
	}
}

type wbctlTag struct {
	name             string
	omitempty        bool
	editable         bool
	ignore           bool
	defaultFieldName string
}

func parseJSONTag(field reflect.StructField) wbctlTag {
	// Parse standard JSON tag first so we preserve exported API field names and
	// omitempty behavior expected by callers.
	jsonTagSpec := field.Tag.Get(jsonTagName)
	parts := strings.Split(jsonTagSpec, ",")
	name := parts[0]
	if name == "" {
		// Fall back to the Go field name when no explicit json tag name exists.
		name = field.Name
	}
	omitempty := false
	for _, part := range parts[1:] {
		if part == omitemptyTag {
			omitempty = true
		}
	}
	wbctlTagSpec := field.Tag.Get(wbctlTagName)
	parts = strings.Split(wbctlTagSpec, ",")
	edit := false
	ignore := false
	defaultFieldName := ""
	for _, part := range parts {
		// editable controls whether this field participates in edit payloads.
		if part == editableTag {
			edit = true
		}
		if part == "-" {
			ignore = true
		}
		// seedfield ties two fields together so one can provide a value when
		// the primary editable field is empty.
		if strings.HasPrefix(part, seedfieldPrefix) {
			defaultFieldName = strings.TrimPrefix(part, seedfieldPrefix)
		}
	}
	return wbctlTag{
		name:             name,
		omitempty:        omitempty,
		editable:         edit,
		ignore:           ignore,
		defaultFieldName: defaultFieldName,
	}
}

func parseCanonicalJSONFieldName(field reflect.StructField) (string, bool) {
	jsonTagSpec := field.Tag.Get(jsonTagName)
	parts := strings.Split(jsonTagSpec, ",")
	name := parts[0]
	if name == "-" {
		return "", false
	}
	if name == "" {
		name = field.Name
	}

	return name, true
}
