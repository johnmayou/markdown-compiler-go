#!/usr/bin/env bash

set -euo pipefail

go vet ./...

go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
golangci-lint run ./...