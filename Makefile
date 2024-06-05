server:
	go run main.go

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

.PHONY: server sqlc
