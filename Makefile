run: builda
	@bin/golang_api_project_layout.exe $(filter-out $@, $(MAKECMDGOALS))

builda:
	@go build -o bin/golang_api_project_layout.exe ./cmd/golang_api_project_layout.go

test:
	@go test -v ./...