set GOOS=windows
set GOARCH=amd64
set GIN_MODE=release
go build -o gofly-win-amd64.exe  -gcflags "-N -l" ./cmd/main/
go clean -cache
