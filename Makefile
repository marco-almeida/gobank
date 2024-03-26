run: builda
	@bin/gobank/main.exe

builda:
	@go build -o bin/gobank/main.exe ./cmd/gobank/main.go

test:
	@go test -v ./...