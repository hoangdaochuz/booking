# Task 1: Project Scaffold & Go Workspace

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Set up the Go multi-module workspace with all 5 microservice stubs.

**Files:**
- Create: `backend/go.work`
- Create: `backend/Makefile`
- Create: `backend/.gitignore`
- Create: `backend/pkg/go.mod`
- Create: `backend/services/gateway/go.mod` + `cmd/main.go`
- Create: `backend/services/user/go.mod` + `cmd/main.go`
- Create: `backend/services/event/go.mod` + `cmd/main.go`
- Create: `backend/services/booking/go.mod` + `cmd/main.go`
- Create: `backend/services/notification/go.mod` + `cmd/main.go`

---

### Step 1: Create directory structure

```bash
cd /Users/dev/work/booking
mkdir -p backend/{proto/{user/v1,event/v1,booking/v1},pkg/{kafka,database,config,middleware},services/{gateway/{cmd,internal/{handler,middleware,router}},user/{cmd,internal/{domain,repository,service,grpc,kafka},migrations},event/{cmd,internal/{domain,repository,service,grpc,kafka},migrations},booking/{cmd,internal/{domain,repository,service,grpc,kafka},migrations},notification/{cmd,internal/{domain,repository,service,kafka},migrations}},scripts}
```

### Step 2: Initialize Go workspace

`backend/go.work`:
```go
go 1.22

use (
    ./pkg
    ./services/gateway
    ./services/user
    ./services/event
    ./services/booking
    ./services/notification
)
```

Initialize each module:
```bash
cd backend/pkg && go mod init github.com/ticketbox/pkg
cd ../services/gateway && go mod init github.com/ticketbox/gateway
cd ../user && go mod init github.com/ticketbox/user
cd ../event && go mod init github.com/ticketbox/event
cd ../booking && go mod init github.com/ticketbox/booking
cd ../notification && go mod init github.com/ticketbox/notification
```

### Step 3: Create placeholder main.go for each service

Each `services/<name>/cmd/main.go`:
```go
package main

import "fmt"

func main() {
    fmt.Println("<name>-service starting...")
}
```

### Step 4: Create .gitignore

`backend/.gitignore`:
```
bin/
vendor/
*.exe
*.dll
*.so
*.dylib
*.test
.idea/
.vscode/
*.swp
.env
.env.*
!.env.example
proto/**/*.pb.go
```

### Step 5: Create Makefile

`backend/Makefile`:
```makefile
.PHONY: proto build up down migrate test test-load logs clean

SERVICES := gateway user event booking notification

proto:
	@echo "Generating protobuf code..."
	@for dir in proto/*/v1; do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$$dir/*.proto; \
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
```

### Step 6: Verify each module compiles

```bash
cd /Users/dev/work/booking/backend
for svc in gateway user event booking notification; do
    cd services/$svc && go build ./cmd && cd ../..
done
```
Expected: All compile with no errors.

### Step 7: Commit

```bash
git add backend/
git commit -m "feat(backend): scaffold Go workspace with 5 microservices"
```
