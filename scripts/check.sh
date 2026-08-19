#!/usr/bin/env bash
set -euo pipefail

project_root=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
cd "$project_root"

unformatted=$(gofmt -l cmd internal)
if [[ -n "$unformatted" ]]; then
  echo "Go files need formatting:"
  echo "$unformatted"
  exit 1
fi

go vet ./...
go test -race ./...
mkdir -p bin
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o bin/blind-dev-setup-windows-x64.exe ./cmd/blind-dev-setup
