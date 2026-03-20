.PHONY: proto build up down migrate test test-load logs clean

SERVICES := gateway user event booking notification

proto:
	@echo "Generating protobuf code..."
	@for dir in proto/*/v1; do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$$dir/*.proto; \
	done
	@echo "Copying generated files to pkg/proto/..."
	@for service in booking event payment user; do \
		mkdir -p pkg/proto/$$service/v1; \
		cp proto/$$service/v1/*.pb.go pkg/proto/$$service/v1/ 2>/dev/null || true; \
	done

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build -o ../../bin/$$svc ./cmd && cd ../..; \
	done

up:
	docker-compose up -d

down:
	docker-compose down

migrate:
	@bash scripts/migrate.sh

test:
	@for svc in $(SERVICES); do \
		echo "Testing $$svc..."; \
		cd services/$$svc && go test ./... -v && cd ../..; \
	done

test-load:
	@bash scripts/load-test.sh

logs:
	docker-compose logs -f

clean:
	rm -rf bin/
