set GOOS=windows
set GOARCH=amd64
go build -o gofly-win-amd64.exe  -gcflags "-N -l" ./cmd/main/
go clean -cache
