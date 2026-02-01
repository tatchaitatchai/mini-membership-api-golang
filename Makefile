.PHONY: help run build test clean migrate-up migrate-down migrate-schema migrate-seed migrate-all docker-up docker-down

help:
	@echo "Available commands:"
	@echo "  make run          - Run the application"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make migrate-up   - Run database migrations"
	@echo "  make docker-up    - Start Docker containers"
	@echo "  make docker-down  - Stop Docker containers"

run:
	./dev.sh

build:
	go build -o bin/api cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

migrate-up:
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/001_initial_schema.sql

migrate-schema:
	@echo "Running schema migration (002)..."
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/002_initial_schema.sql
	@echo "Running schema additions (003)..."
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/003_addedit_schema.sql
	@echo "Schema migration complete!"

migrate-seed:
	@echo "Running seed data (004)..."
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/004_seed_data.sql
	@echo "Seed data complete!"

migrate-all:
	@echo "Running all migrations (002-004)..."
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/002_initial_schema.sql
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/003_addedit_schema.sql
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership < migrations/004_seed_data.sql
	@echo "All migrations complete!"

migrate-down:
	docker exec -i katom-membership-postgres psql -U katom -d katom_membership -c "DROP TABLE IF EXISTS member_point_transactions CASCADE; DROP TABLE IF EXISTS members CASCADE; DROP TABLE IF EXISTS staff_users CASCADE;"

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f postgres

tidy:
	go mod tidy
