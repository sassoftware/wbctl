// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package sliceutil

import (
	"errors"

	"github.com/sas-institute-rnd-saas/wbctl/internal/i18n"
)

// Coalesce returns the first non-zero value.
func Coalesce[T comparable](values ...T) T {
	var zeroValue T
	for _, v := range values {
		if v != zeroValue {
			return v
		}
	}
	return zeroValue
}

// GetAndEnsureOnlyOne returns the only element or an error.
func GetAndEnsureOnlyOne[T any](slice []T) (T, error) {
	var t T
	if len(slice) == 0 {
		return t, errors.New(i18n.Format(i18n.NoElementsFound))
	}
	if len(slice) > 1 {
		return t, errors.New(i18n.Format(i18n.MultipleElementsFound))
	}
	return slice[0], nil
}
