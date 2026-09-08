# WSL2: systemd user session can't start, so /run/user/1000/bus doesn't exist.
# If conmon/crun sees DBUS_SESSION_BUS_ADDRESS it tries the systemd path and fails
# with "sd-bus call: Access denied" even though cgroup_manager is cgroupfs.
unexport DBUS_SESSION_BUS_ADDRESS
# WSL2: netavark launches aardvark-dns via `systemd-run --user` when it finds
# systemd-run in PATH, but there is no user session here. Prepend a shim that
# strips systemd-run flags and execs directly.
export PATH := $(HOME)/.local/share/podman-wsl-shims:$(PATH)
# WSL2: without a systemd user session, podman cannot register healthcheck
# timers over the user bus ("unable to get systemd connection to add
# healthchecks"). DISABLE_HC_SYSTEMD is podman's documented escape hatch.
export DISABLE_HC_SYSTEMD := true

.PHONY: proto build up down podman-api migrate test test-load logs clean

SERVICES := gateway user event booking notification payment saga scheduler

proto:
	@echo "Generating protobuf code..."
	@for dir in proto/*/v1; do \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$$dir/*.proto; \
	done
	@echo "Copying generated files to pkg/proto/..."
	@for service in booking event payment user saga scheduler; do \
		mkdir -p pkg/proto/$$service/v1; \
		cp proto/$$service/v1/*.pb.go pkg/proto/$$service/v1/ 2>/dev/null || true; \
	done

build:
	@for svc in $(SERVICES); do \
		echo "Building $$svc..."; \
		cd services/$$svc && go build -o ../../bin/$$svc ./cmd && cd ../..; \
	done

up: podman-api
	@bash scripts/podman-up.sh

# Expose the podman API over TCP for Podman Desktop on Windows. The rootless
# podman.socket cannot be socket-activated here (no systemd user session on
# WSL2), so Podman Desktop cannot reach this distro without it.
# Windows side: podman system connection add ubuntu-wsl tcp://localhost:8090
podman-api:
	@curl -s http://127.0.0.1:8090/_ping >/dev/null 2>&1 \
		|| (nohup podman system service --time=0 tcp://127.0.0.1:8090 >/tmp/podman-api.log 2>&1 & \
			sleep 2 && echo "podman API listening on tcp://127.0.0.1:8090")

down:
	podman compose down

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
	podman compose logs -f

clean:
	rm -rf bin/
