# 生成 proto Go 代码
# 用法: ./scripts/gen_proto.ps1

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$protoDir = Join-Path $repoRoot "pkg\proto"

# 把 GOBIN 和 winget 的 protoc 路径加到 PATH
$goBin = Join-Path (go env GOPATH) "bin"
$env:Path = "$goBin;" + [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

# 校验 protoc 是否可用
if (-not (Get-Command protoc -ErrorAction SilentlyContinue)) {
    Write-Error "未找到 protoc，请先安装: winget install --id Google.Protobuf"
    exit 1
}
if (-not (Test-Path (Join-Path $goBin "protoc-gen-go.exe"))) {
    Write-Error "未找到 protoc-gen-go，请先安装: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
}
if (-not (Test-Path (Join-Path $goBin "protoc-gen-go-grpc.exe"))) {
    Write-Error "未找到 protoc-gen-go-grpc，请先安装: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
}

Write-Host "proto 目录: $protoDir" -ForegroundColor Cyan
Push-Location $repoRoot
try {
    Get-ChildItem -Path $protoDir -Filter *.proto | ForEach-Object {
        $rel = Resolve-Path $_.FullName -Relative
        Write-Host "生成: $($_.Name)" -ForegroundColor Green
        protoc --proto_path=pkg/proto `
               --go_out=. --go_opt=paths=source_relative `
               --go-grpc_out=. --go-grpc_opt=paths=source_relative `
               $rel
        if ($LASTEXITCODE -ne 0) { throw "protoc 执行失败: $($_.Name)" }
    }
    Write-Host "完成" -ForegroundColor Green
} finally {
    Pop-Location
}