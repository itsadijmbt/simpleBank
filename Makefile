# this postgres has to be done once only for project setup rest use createdv and drop ones
postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	# Create the 'simple_bank' database inside the 'postgres12' container
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	# Drop the 'simple_bank' database inside the 'postgres12' container
	docker exec -it postgres12 dropdb --username=root simple_bank

migrateup:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migrateFixToStable:
	# here the fixation is for state 0004 init schema.sql so forecilby setting to this state
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" force 0004

migratedown1:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

migrateup1:
	migrate -path db/migration -database "postgres://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

#^ migrate upto "1" version up/down


sqlc:
	sqlc generate

test:
	go test -v -cover ./...

make server:
	go run main.go
mock:
	mockgen -package mockdb -destination  db/mock/store.go github.com/itsadijmbt/simple_bank/db/sqlc Store 

.PHONY: postgres createdb dropdb migrateup migratedown migrateFixToOne sqlc server mock migratedown1 migrateup1

#!migrate create -ext sql -dir db/migration -seq add_user
#! to create migration version