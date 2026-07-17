# SPDX-License-Identifier: MIT
# Install NSIS for Windows packaging CI.
# Prefer an existing makensis, then Chocolatey with retries, then a pinned zip
# download so a Chocolatey.org 503 does not fail the desktop build.
[CmdletBinding()]
param(
  [string] $Version = "3.10",
  [int] $ChocoAttempts = 3
)

$ErrorActionPreference = "Stop"

function Find-MakeNsis {
  $candidates = @(
    (Join-Path ${env:ProgramFiles(x86)} "NSIS\makensis.exe"),
    (Join-Path $env:ProgramFiles "NSIS\makensis.exe")
  )
  if ($env:ChocolateyInstall) {
    $candidates += @(
      (Join-Path $env:ChocolateyInstall "bin\makensis.exe"),
      (Join-Path $env:ChocolateyInstall "lib\nsis\tools\makensis.exe")
    )
  }
  if ($env:NSIS_HOME) {
    $candidates += (Join-Path $env:NSIS_HOME "makensis.exe")
  }
  foreach ($path in $candidates) {
    if ($path -and (Test-Path -LiteralPath $path)) {
      return (Resolve-Path -LiteralPath $path).Path
    }
  }
  $cmd = Get-Command makensis.exe -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd.Source
  }
  return $null
}

function Add-MakeNsisToPath([string] $MakeNsisPath) {
  $dir = Split-Path -Parent $MakeNsisPath
  if ($env:GITHUB_PATH) {
    $dir | Out-File -FilePath $env:GITHUB_PATH -Encoding utf8 -Append
  }
  $env:Path = "$dir;$env:Path"
}

function Install-NsisViaChoco {
  if (-not (Get-Command choco.exe -ErrorAction SilentlyContinue)) {
    Write-Host "choco.exe not found; skipping Chocolatey install"
    return $false
  }
  for ($i = 1; $i -le $ChocoAttempts; $i++) {
    Write-Host "Chocolatey install attempt $i/$ChocoAttempts"
    & choco.exe install nsis -y --no-progress
    if ($LASTEXITCODE -eq 0) {
      $found = Find-MakeNsis
      if ($found) {
        return $true
      }
    }
    Write-Warning "Chocolatey NSIS install failed (exit $LASTEXITCODE)"
    if ($i -lt $ChocoAttempts) {
      Start-Sleep -Seconds (5 * $i)
    }
  }
  return $false
}

function Install-NsisFromZip {
  $tempRoot = if ($env:RUNNER_TEMP) { $env:RUNNER_TEMP } else { [System.IO.Path]::GetTempPath() }
  $zipPath = Join-Path $tempRoot "nsis-$Version.zip"
  $extractDir = Join-Path $tempRoot "nsis-$Version"
  $stamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
  $url = "https://downloads.sourceforge.net/project/nsis/NSIS%203/$Version/nsis-$Version.zip?ts=$stamp&use_mirror=autoselect"

  Write-Host "Downloading NSIS $Version zip from SourceForge"
  Invoke-WebRequest -Uri $url -OutFile $zipPath -UseBasicParsing

  if (Test-Path -LiteralPath $extractDir) {
    Remove-Item -LiteralPath $extractDir -Recurse -Force
  }
  Expand-Archive -LiteralPath $zipPath -DestinationPath $extractDir -Force

  $makensis = Get-ChildItem -LiteralPath $extractDir -Filter makensis.exe -Recurse -File |
    Select-Object -First 1
  if (-not $makensis) {
    throw "makensis.exe missing after extracting $zipPath"
  }
  return $makensis.FullName
}

$makensis = Find-MakeNsis
if (-not $makensis) {
  if (-not (Install-NsisViaChoco)) {
    Write-Warning "Chocolatey unavailable or failed; falling back to SourceForge zip"
    $makensis = Install-NsisFromZip
  } else {
    $makensis = Find-MakeNsis
  }
}

if (-not $makensis -or -not (Test-Path -LiteralPath $makensis)) {
  throw "makensis.exe not found after NSIS install attempts"
}

Add-MakeNsisToPath $makensis
Write-Host "Using makensis: $makensis"
& $makensis @('/VERSION')
if ($LASTEXITCODE -ne 0) {
  throw "makensis /VERSION failed with exit $LASTEXITCODE"
}
