# Copyright © 2026, SAS Institute Inc., Cary, NC, USA.  All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# This script installs the wbctl binary for Windows x64, verifying checksums.

param(
    [string]$Version = 'latest',
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA 'Programs\wbctl'),
    [string]$Repository = 'sassoftware/wbctl'
)

$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT' -or [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -ne 'X64') {
    throw 'This installer requires Windows x64.'
}

[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12
if ($Version -eq 'latest') {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest"
    $Version = $release.tag_name
}
if ($Version -notmatch '^[a-zA-Z0-9][a-zA-Z0-9._+-]*$' -or $Version -eq 'latest') {
    throw 'Expected a release tag, or latest.'
}

$asset = 'wbctl-windows-amd64.exe'
$temporary = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $temporary | Out-Null
try {
    foreach ($filename in @($asset, 'checksums.txt')) {
        Invoke-WebRequest -UseBasicParsing -Uri "https://github.com/$Repository/releases/download/$Version/$filename" -OutFile (Join-Path $temporary $filename)
    }
    $pattern = '^[a-fA-F0-9]{64}\s+\*?' + [regex]::Escape($asset) + '$'
    $entries = @(Get-Content (Join-Path $temporary 'checksums.txt') | Where-Object { $_ -match $pattern })
    if ($entries.Count -ne 1) {
        throw 'Missing or duplicate checksum; nothing was installed.'
    }
    $expected = ($entries[0] -split '\s+')[0]
    $actual = (Get-FileHash -Algorithm SHA256 -Path (Join-Path $temporary $asset)).Hash
    if ($actual -ne $expected) {
        throw 'Checksum verification failed; nothing was installed.'
    }
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $temporary $asset) -Destination (Join-Path $InstallDir 'wbctl.exe') -Force

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($InstallDir -notin ($userPath -split ';')) {
        [Environment]::SetEnvironmentVariable('Path', "$InstallDir;$userPath", 'User')
    }
    if ($InstallDir -notin ($env:Path -split ';')) {
        $env:Path = "$InstallDir;$env:Path"
    }
    Write-Output "Installed wbctl $Version to $InstallDir\wbctl.exe"
}
finally {
    Remove-Item -Recurse -Force -Path $temporary
}
