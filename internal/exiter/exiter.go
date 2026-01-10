// Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exiter

import "os"

// OsExiter is a hook for tests to capture exit codes.
var OsExiter = os.Exit
