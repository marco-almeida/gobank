run: builda
	@bin/go-api-structure.exe

builda:
	@go build -o bin/go-api-structure.exe ./cmd/go-api-structure.go

test:
	@go test -v ./...