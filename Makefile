run: builda
	@bin/gobank.exe

builda:
	@go build -o bin/gobank.exe ./cmd/gobank.go

test:
	@go test -v ./...