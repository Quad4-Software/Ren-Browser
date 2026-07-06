param(
    [string]$Exe = "bin/renbrowser.exe",
    [string]$Manifest = "build/windows/msix/app_manifest.xml",
    [string]$Icon = "build/appicon.png",
    [string]$Out = "bin/renbrowser-amd64.msix"
)

$ErrorActionPreference = "Stop"

$root = (Resolve-Path (Join-Path $PSScriptRoot "../..")).Path
Set-Location $root

function Resolve-ProjectPath([string]$path) {
    if ([string]::IsNullOrWhiteSpace($path)) {
        throw "Path must not be empty."
    }
    if ([System.IO.Path]::IsPathRooted($path)) {
        return (Resolve-Path $path).Path
    }
    return (Join-Path $root $path)
}

function Find-MakeAppx {
    $kitsRoot = Join-Path ${env:ProgramFiles(x86)} "Windows Kits\10\bin"
    if (Test-Path $kitsRoot) {
        $versionDir = Get-ChildItem $kitsRoot -Directory |
            Where-Object { $_.Name -match '^\d' } |
            Sort-Object { [version]$_.Name } -Descending |
            Select-Object -First 1
        if ($versionDir) {
            $candidate = Join-Path $versionDir.FullName "x64\MakeAppx.exe"
            if (Test-Path $candidate) {
                return $candidate
            }
        }
    }

    $cmd = Get-Command MakeAppx.exe -ErrorAction SilentlyContinue
    if ($cmd) {
        return $cmd.Source
    }

    throw "MakeAppx.exe not found. Install the Windows SDK (App Certification Kit)."
}

$exePath = Resolve-ProjectPath $Exe
$manifestPath = Resolve-ProjectPath $Manifest
$iconPath = Resolve-ProjectPath $Icon
$outPath = Resolve-ProjectPath $Out

if (-not (Test-Path $exePath)) {
    throw "Executable not found: $exePath"
}
if (-not (Test-Path $manifestPath)) {
    throw "Appx manifest not found: $manifestPath"
}
if (-not (Test-Path $iconPath)) {
    throw "Icon not found: $iconPath"
}

$staging = Join-Path $env:TEMP ("renbrowser-msix-" + [guid]::NewGuid().ToString("n"))
New-Item -ItemType Directory -Path $staging -Force | Out-Null
$assetsDir = Join-Path $staging "Assets"
New-Item -ItemType Directory -Path $assetsDir -Force | Out-Null

Copy-Item $exePath (Join-Path $staging "renbrowser.exe")
Copy-Item $manifestPath (Join-Path $staging "AppxManifest.xml")

@(
    "Square150x150Logo.png",
    "Square44x44Logo.png",
    "Wide310x150Logo.png",
    "SplashScreen.png",
    "StoreLogo.png"
) | ForEach-Object {
    Copy-Item $iconPath (Join-Path $assetsDir $_)
}

$makeappx = Find-MakeAppx
New-Item -ItemType Directory -Path (Split-Path $outPath -Parent) -Force | Out-Null
if (Test-Path $outPath) {
    Remove-Item $outPath -Force
}

& $makeappx pack /d $staging /p $outPath /o
if ($LASTEXITCODE -ne 0) {
    throw "MakeAppx.exe failed with exit code $LASTEXITCODE"
}

Remove-Item $staging -Recurse -Force
Write-Host "MSIX package created: $outPath"
