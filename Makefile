postgres:
	docker run --name new-postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it new-postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it new-postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go	

mock:
    mockgen -package mockdb -destination db/mock/store.go db.sqlc.dev/app/db/sqlc Store

.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock