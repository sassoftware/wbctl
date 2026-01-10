rem Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
rem SPDX-License-Identifier: Apache-2.0

@echo off
powershell.exe -NoLogo -NoProfile -NonInteractive -Command ^
  "$path = '%~1'; $content = Get-Content -LiteralPath $path -Raw; Set-Content -LiteralPath $path -Value ($content.Replace('replace-me','UPDATED')) -NoNewline"