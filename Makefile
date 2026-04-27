include .env
export


service-run:
	go run cmd/shortener/main.go

migrate-up:
	migrate -path migrations -database ${DATABASE_DSN} up

migrate-down:
	migrate -path migrations -database ${DATABASE_DSN} down

docker-up:
	docker compose up -d

docker-down:
	docker compose down