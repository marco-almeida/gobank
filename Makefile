run: builda
	@bin/mybank/main.exe

builda:
	@go build -o bin/mybank/main.exe ./cmd/mybank/main.go

test:
	@go test -v ./...

sqlc:
	sqlc generate -f ./internal/postgresql/sqlc.yaml

new_migration:
	migrate create -ext sql -dir internal/postgresql/migrations -seq $(name)

migrateup:
	migrate -path db/migrations -database "$(DB_URL)" -verbose up

.PHONY: run builda test sqlc new_migration