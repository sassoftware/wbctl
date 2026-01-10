// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package wbctljson contains wbctl-specific JSON marshaling and unmarshaling
// helpers used to bridge user-facing CLI payloads and canonical API structs.
//
// The package provides:
//   - MakeMarshaler, a reflection-based marshaler that can emit full view JSON
//     or edit-only payloads.
//   - MakeUnmarshaler, a normalizing unmarshaler that rewrites wbctl field
//     names before typed decoding.
//   - Shared tag parsing and normalization helpers that apply wbctl metadata
//     consistently for both encode and decode paths.
//
// Tag handling:
//   - Standard `json` tags remain the canonical schema contract, including
//     field naming, `omitempty`, and `json:"-"` exclusion.
//   - `wbctl:"editable"` gates fields included in edit payload output, letting
//     view output stay complete while edit output stays intentionally narrow.
//   - `wbctl:"seedfield:<FieldName>"` allows an editable field to use a fallback
//     value when the primary field is zero-valued, supporting stable edit
//     payloads for mirrored state fields.
//   - `wbctl:"-"` excludes a field from wbctl-specific processing even if other
//     wbctl metadata is present.
//
// Recursion:
//   - Marshal recursion follows three root/container kinds (struct, map,
//     slice), and applies wbctl filtering recursively for nested struct values.
//   - Map/slice recursion intentionally recurses only into struct entries and
//     leaves scalar/non-struct values to encoding/json, preserving native JSON
//     behavior where wbctl metadata does not apply.
//   - Pointer/interface wrappers are unwrapped before recursion decisions so
//     traversal is based on concrete payload shape instead of wrapper kinds.
package wbctljson
