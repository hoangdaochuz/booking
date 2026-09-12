# Running TicketBox Locally with Podman

This guide runs the full stack (8 Go services, 7 Postgres databases, Redis, Kafka + Zookeeper + Kafka UI) on **rootless Podman**, on the WSL2 Ubuntu box where podman replaced docker-compose.

Everything podman-specific below is already wired into the repo (`backend/Makefile`, `backend/scripts/podman-up.sh`) — the day-to-day workflow is just `make up → make migrate → make seed`.

---

## 1. Prerequisites

| Tool | Version on this box | Used for |
|---|---|---|
| podman | 5.7.0 (rootless) | Container runtime |
| podman-compose | 1.5.0 (`podman compose`) | Compose the stack |
| golang-migrate CLI | `/usr/local/bin/migrate` | `make migrate` |
| curl + jq | — | `make seed` |
| Go + protoc | — | Only if regenerating proto (`make proto`) |
| Node (via nvm, v24 LTS) | — | Frontend dev server |

## 2. One-time host setup (WSL2 Podman quirks)

Already in place on this machine — keep this section for a fresh machine or if something breaks after a WSL reset. The root problem: WSL2 here has `systemd=true` but **no working systemd user session** (`user@1000.service` fails with "Device or resource busy"), and rootless podman's healthchecks, cgroups, and DNS helpers all assume that user bus exists.

### 2.1 containers.conf — cgroup manager

`~/.config/containers/containers.conf`:

```toml
[engine]
cgroup_manager = "cgroupfs"
```

Without this, `podman build` RUN steps die with `sd-bus call: Access denied`.

### 2.2 registries.conf — short image names

`~/.config/containers/registries.conf`:

```toml
unqualified-search-registries = ["docker.io"]
```

Lets `image: postgres:16-alpine` resolve without the `docker.io/library/` prefix.

### 2.3 systemd-run shim (netavark / aardvark-dns)

netavark launches its DNS helper via `systemd-run --user`, which fails without a user session (netavark issue #473). A shim that strips the flags and execs directly lives at:

```
~/.local/share/podman-wsl-shims/systemd-run
```

The Makefile prepends `~/.local/share/podman-wsl-shims` to `PATH` (see `backend/Makefile:8`), so as long as the shim exists there you don't need to do anything.

### 2.4 What the Makefile handles for you

`backend/Makefile` already exports, on every invocation:

- `unexport DBUS_SESSION_BUS_ADDRESS` — otherwise conmon/crun tries the systemd path even with cgroupfs
- `DISABLE_HC_SYSTEMD := true` — podman's documented escape hatch so containers start without systemd healthcheck timers
- `PATH` with the shim dir first

And `make up` runs `scripts/podman-up.sh` instead of plain `podman compose up`: it polls `podman healthcheck run` in a background loop while compose waits on `service_healthy` dependencies (podman-compose waits via `podman wait --condition=healthy`, which never fires without systemd timers — the loop makes it fire).

### 2.5 Podman Desktop on Windows (optional)

Podman Desktop only sees `podman-machine-default`, not this distro. `make up` depends on target `podman-api`, which starts `podman system service tcp://127.0.0.1:8090`. From Windows once:

```
podman system connection add ubuntu-wsl tcp://localhost:8090
```

⚠️ That TCP API is unauthenticated — fine on a localhost dev box, nothing more.

## 3. Run the backend

```bash
cd backend

# 1. Env file (already exists; on a fresh clone:)
cp .env.example .env
# Fill in: STRIPE_PUBLISHABLE_KEY, STRIPE_SECRET_KEY, STRIPE_SECRET_WEBHOOK
# POSTGRES_PASSWORD / JWT_SECRET have dev defaults.

# 2. Start the whole stack (~18 containers)
make up

# 3. Apply migrations to all 7 databases (localhost ports 5433–5439)
make migrate

# 4. Seed admin + test user + sample events
make seed
```

Credentials created by the seed:

- Admin: `admin@ticketbox.com` / `admin123`
- User: `user@example.com` / `user123`

### Scheduler job config (one-time)

The scheduler service reads job configs from `ticketbox_scheduler`. The `reservation-cleaner` job is **not** seeded by `make seed` — insert it manually:

```bash
podman exec backend_postgres-scheduler-1 psql -U ticketbox -d ticketbox_scheduler -c \
  "INSERT INTO scheduler_configs (name, interval_expression, is_enable)
   VALUES ('reservation-cleaner-job', '*/30 * * * * *', true)
   ON CONFLICT (name) DO NOTHING;"
```

Note the cron is 6-field (robfig/cron with seconds): `sec min hour dom mon dow` — `*/30 * * * * *` fires every 30 seconds.

## 4. Run the frontend

```bash
cd frontend
npm install
npm run dev        # http://localhost:3000
```

`frontend/.env` already sets `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` (pk_test_...). Next.js rewrites `/api/*` → gateway `:8000`, so no extra proxy config is needed. If the backend is down, the app falls back to mock data.

## 5. Stripe webhooks in local dev

The payment service exposes `POST :8081/webhooks/stripe`. Forward test-mode events from your Stripe account:

```bash
stripe listen --forward-to localhost:8081/webhooks/stripe
```

`STRIPE_SECRET_WEBHOOK` in `backend/.env` must be the `whsec_...` the listener prints, or signature verification rejects everything. This webhook is what resumes the saga after payment — without it, bookings stay paused at the payment step.

## 6. Ports & URLs

| Port | What |
|---|---|
| :3000 | Frontend (Next.js dev) |
| :8000 | Gateway (REST API — everything the frontend calls) |
| :8080 | Kafka UI |
| :8081 | Payment HTTP (Stripe webhook receiver) |
| :9092 | Kafka (host listener) |
| :5433–5439 | Postgres: user, event, booking, notification, payment, saga, scheduler |
| :6379 | Redis |

## 7. Day-to-day operations

```bash
cd backend
make logs                          # tail all container logs
make down                          # stop everything (volumes persist)
make up                            # start again

# Rebuild + restart a single service after code changes:
podman compose build booking-service && podman compose up -d booking-service
# ⚠️ If the new code still isn't live, the container may be pinned to the old
# (untagged) image ID — restart and even --force-recreate can keep resolving it.
# Check: podman inspect <cid> --format '{{.ImageID}}' vs `podman images`.
# Fallback: make down && make up to recreate every container from current tags.

# Watch saga events / payment outcomes:
#   open http://localhost:8080 (Kafka UI), cluster "ticketbox"

# Connect to a database:
podman exec -it backend_postgres-booking-1 psql -U ticketbox -d ticketbox_booking
```

Full reset (destroys all data):

```bash
make down
podman compose down -v   # removes pgdata-* and redis-data volumes
make up && make migrate && make seed
# + re-insert the scheduler job config (section 3)
```

## 8. Troubleshooting

**`make up` hangs forever bringing up postgres/kafka**
The healthcheck poll loop in `scripts/podman-up.sh` isn't running, or `DISABLE_HC_SYSTEMD` isn't exported (only happens if you run `podman compose up` directly instead of `make up`). Kill it and use `make up`.

**`podman build` fails with `sd-bus call: Access denied`**
`cgroup_manager = "cgroupfs"` missing from `~/.config/containers/containers.conf`, or `DBUS_SESSION_BUS_ADDRESS` leaked into your shell — the Makefile unexports it, but a direct `podman build` in a shell that sets it will still trip.

**`migrate: no migration found for version 0`**
Someone ran `migrate force 0`, which leaves a `version=0` row in `schema_migrations`. Fix by deleting the row, not by forcing another version:

```bash
podman exec backend_postgres-<svc>-1 psql -U ticketbox -d ticketbox_<svc> \
  -c "DELETE FROM schema_migrations WHERE version = 0;"
```

**Rebuilt an image but the service still runs old code** (`Unimplemented` RPCs, old logs after restart)
`podman restart` only re-runs the existing container — it never picks up a rebuilt image. Worse, when a rebuild moves the `latest` tag, the old image becomes untagged and the container (and sometimes even `podman compose up -d --force-recreate`) stays pinned to the old image ID. Verify with `podman inspect <cid> --format '{{.ImageID}}'` vs `podman images`; if they differ, `make down && make up`. Services with `depends_on` (booking/payment/saga/scheduler → event-service) can't be removed individually.

**Container DNS / name resolution flakes inside the compose network**
That's the netavark/aardvark-dns `systemd-run` issue — verify the shim still exists at `~/.local/share/podman-wsl-shims/systemd-run` and is executable.

**Frontend: `Cannot find module '../lightningcss.linux-x64-gnu.node'`**
Stale Turbopack cache after a cross-platform npm install (Windows ↔ WSL on `/mnt/d`). Fix: `rm -rf .next` and restart the dev server — don't reinstall deps.

**Bookings stuck in `PENDING` / saga paused forever**
Check that the Stripe webhook listener (section 5) is running and that `STRIPE_SECRET_WEBHOOK` matches the printed `whsec_...`; then check the saga DB (`backend_postgres-saga-1`, table with `current_step_index`) and Kafka UI for the payment-outcome event.
