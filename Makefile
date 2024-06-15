server:
	go run main.go

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

build:
	GOOS=linux GOARCH=amd64 go build -o bin/application

.PHONY: server sqlc, build
