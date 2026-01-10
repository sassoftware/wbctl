// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package jsonutil

import (
	"os"

	"github.com/sas-institute-rnd-saas/wbctl/internal/wbctljson"
	"github.com/spf13/afero"
)

// GetFromFileOrStdin loads JSON from a file or stdin, normalizing wbctl alias
// field names before decoding.
func GetFromFileOrStdin[T any](
	fsys afero.Fs,
	inputPath string,
) (T, error) {
	unmarshalF := wbctljson.MakeUnmarshaler[T]()
	var val T
	if inputPath == "-" {
		return val, unmarshalF(os.Stdin, &val)
	}
	f, err := fsys.Open(inputPath)
	if err != nil {
		return val, err
	}
	defer func() { _ = f.Close() }()
	return val, unmarshalF(f, &val)
}
