// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package timeutil

import (
	"time"

	"k8s.io/apimachinery/pkg/util/duration"
)

var now = time.Now

// SetNow overrides the current time function.
func SetNow(newNow func() time.Time) {
	now = newNow
}

// Now returns the result of calling the current time function.
func Now() time.Time {
	return now()
}

// Since returns a human-readable duration since the given time.
func Since(at *time.Time) string {
	if at == nil || at.IsZero() {
		return ""
	}
	return duration.HumanDuration(now().Sub(*at))
}
