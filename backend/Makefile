include .env
export

migrate-up:
	migrate -path db/migrations \
	-database "$(DATABASE_URL)" \
	up

migrate-down:
	migrate -path db/migrations \
	-database "$(DATABASE_URL)" \
	down 1

sqlc:
	sqlc generate

# Opt-in: writes synthetic records and prints new credentials once.
seed-demo:
	go run ./cmd/seed-demo -allow-demo-data

# Testing commands
test:
	go test -v ./...

test-auth:
	go test -v ./internal/auth

test-user:
	go test -v ./internal/user

test-role:
	go test -v ./internal/role

test-device:
	go test -v ./internal/device

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-coverage-report:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
