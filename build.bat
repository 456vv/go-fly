set GIN_MODE=release

set GOOS=windows
set GOARCH=amd64
go build -o gofly-win-amd64.exe -trimpath -ldflags "-s -w" ./cmd/main/
go clean -cache

set GOOS=linux
set GOARCH=amd64
go build -o gofly-linux-amd64 -trimpath -ldflags "-s -w" ./cmd/main/
go clean -cache