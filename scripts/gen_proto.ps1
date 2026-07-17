# Generate Go code from .proto files.
# Usage: ./scripts/gen_proto.ps1

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$protoDir = Join-Path $repoRoot "pkg\proto"

# Put GOBIN and winget protoc on PATH.
$goBin = Join-Path (go env GOPATH) "bin"
$env:Path = "$goBin;" + [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

# Check protoc is available.
if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Error "protoc not found. Install: winget install --id Google.Protobuf"
    exit 1
}
if (-not (Test-Path (Join-Path $goBin "protoc-gen-go.exe"))) {
    Write-Error "protoc-gen-go not found. Install: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
}
if (-not (Test-Path (Join-Path $goBin "protoc-gen-go-grpc.exe"))) {
    Write-Error "protoc-gen-go-grpc not found. Install: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
}

Write-Host "proto dir: $protoDir" -ForegroundColor Cyan
Push-Location $repoRoot
try {
    Get-ChildItem -Path $protoDir -Filter *.proto | ForEach-Object {
        Write-Host "generating: $($_.Name)" -ForegroundColor Green
        # Outputs go straight into pkg/proto (source_relative keeps the base name),
        # which is what the code imports as shop/pkg/proto.
        & protoc --proto_path=pkg/proto --go_out=pkg/proto --go_opt=paths=source_relative --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative $_.Name
        if ($LASTEXITCODE -ne 0) { throw "protoc failed: $($_.Name)" }
    }
    Write-Host "done" -ForegroundColor Green
} finally {
    Pop-Location
}