param(
    [ValidatePattern('^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$')]
    [string]$Version = '0.1.0'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$ProjectRoot = Split-Path -Parent $PSScriptRoot
$DistDir = Join-Path $ProjectRoot 'dist'
$StageDir = Join-Path $DistDir "samsung-fus-cli-$Version-windows-amd64"
$ExePath = Join-Path $StageDir 'samsung-fus-cli.exe'
$ZipPath = Join-Path $DistDir "samsung-fus-cli-$Version-windows-amd64-portable.zip"

Push-Location $ProjectRoot
try {
    go test ./...

    if (Test-Path -LiteralPath $StageDir) {
        Remove-Item -LiteralPath $StageDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $StageDir -Force | Out-Null

    $env:CGO_ENABLED = '0'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    go build -trimpath -ldflags "-s -w -X main.version=$Version" -o $ExePath .

    $Files = @(
        'README.md',
        'README.en.md',
        'README.ko.md',
        'README.zh-CN.md',
        'README.zh-TW.md',
		'CHANGELOG.md',
        'LICENSE.txt',
        'THIRD_PARTY_NOTICES.txt',
        'DISCLAIMER.txt',
        'SECURITY.md'
    )
    foreach ($File in $Files) {
        Copy-Item -LiteralPath (Join-Path $ProjectRoot $File) -Destination $StageDir
    }
    Copy-Item -LiteralPath (Join-Path $ProjectRoot 'docs/ARCHITECTURE.md') -Destination $StageDir
    Copy-Item -LiteralPath (Join-Path $ProjectRoot 'packaging/windows/README-PORTABLE.txt') -Destination $StageDir

    if (Test-Path -LiteralPath $ZipPath) {
        Remove-Item -LiteralPath $ZipPath -Force
    }
    Compress-Archive -Path (Join-Path $StageDir '*') -DestinationPath $ZipPath -CompressionLevel Optimal

    & $ExePath version
    $Hash = Get-FileHash -LiteralPath $ZipPath -Algorithm SHA256
    $ChecksumPath = "$ZipPath.sha256"
    Set-Content -LiteralPath $ChecksumPath -Value "$($Hash.Hash.ToLowerInvariant())  $(Split-Path -Leaf $ZipPath)" -Encoding ascii
    $Hash
}
finally {
    Pop-Location
    Remove-Item Env:\CGO_ENABLED -ErrorAction SilentlyContinue
    Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue
}
