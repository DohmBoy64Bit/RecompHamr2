param(
    [Parameter(Mandatory = $true)]
    [string]$ReleaseDir,

    [Parameter(Mandatory = $true)]
    [string]$InstallDir,

    [string]$Artifact = "recomphamr_windows_amd64.zip",

    [switch]$SkipChecksum
)

$ErrorActionPreference = "Stop"

$manifest = Join-Path $ReleaseDir "SHA256SUMS"
$artifactPath = Join-Path $ReleaseDir $Artifact

if (-not (Test-Path -LiteralPath $artifactPath -PathType Leaf)) {
    throw "Artifact not found: $artifactPath"
}

if (-not $SkipChecksum) {
    if (-not (Test-Path -LiteralPath $manifest -PathType Leaf)) {
        throw "Checksum manifest not found: $manifest"
    }
    $expected = Select-String -LiteralPath $manifest -Pattern "^[a-fA-F0-9]{64}\s+\*?$([regex]::Escape($Artifact))$" | Select-Object -First 1
    if ($null -eq $expected) {
        throw "Artifact is missing from SHA256SUMS: $Artifact"
    }
    $expectedHash = ($expected.Line -split "\s+")[0].ToLowerInvariant()
    $actualHash = (Get-FileHash -LiteralPath $artifactPath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -ne $expectedHash) {
        throw "Checksum mismatch for $Artifact"
    }
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Expand-Archive -LiteralPath $artifactPath -DestinationPath $InstallDir -Force

$exe = Join-Path $InstallDir "recomphamr.exe"
if (-not (Test-Path -LiteralPath $exe -PathType Leaf)) {
    throw "Installed archive did not contain recomphamr.exe"
}

Write-Output "Installed RecompHamr to $exe"
