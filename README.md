# TicketBox

A full-stack event ticket booking platform built with Go microservices and Next.js. Users can browse events, view interactive seat maps, select individual seats, and purchase tickets with real-time availability.

## Architecture

```
                         ┌─────────────┐
                         │   Browser   │
                         └──────┬──────┘
                                │
                         ┌──────▼──────┐
                         │   Next.js   │  :3000
                         │  (Frontend) │
                         └──────┬──────┘
                                │ /api/* rewrite
                         ┌──────▼──────┐
                         │   Gateway   │  :8000  (Gin HTTP)
                         │             │
                         └──┬───┬───┬──┘
                     gRPC   │   │   │   gRPC
               ┌────────────┘   │   └────────────┐
               ▼                ▼                 ▼
        ┌────────────┐  ┌────────────┐  ┌──────────────┐
        │   User     │  │   Event    │  │   Booking    │
        │  Service   │  │  Service   │  │   Service    │
        │  :50051    │  │  :50052    │  │   :50053     │
        └─────┬──────┘  └─────┬──────┘  └──────┬───────┘
              │               │                 │
              ▼               ▼                 ▼
        ┌──────────┐  ┌──────────┐       ┌──────────┐
        │ Postgres │  │ Postgres │       │ Postgres │
        │  :5433   │  │  :5434   │       │  :5435   │
        └──────────┘  └──────────┘       └──────────┘

        ┌─────────────────────────────────────────┐
        │              Kafka  :9092               │
        │         (Event-driven messaging)        │
        └────────────────────┬────────────────────┘
                             │
                     ┌───────▼────────┐
                     │  Notification  │
                     │    Service     │
                     └───────┬────────┘
                             │
                       ┌─────▼─────┐
                       │ Postgres  │
                       │  :5436    │
                       └───────────┘
```

## Tech Stack

### Backend
- **Language:** Go 1.25
- **Inter-service communication:** gRPC with Protocol Buffers
- **HTTP Gateway:** Gin
- **Databases:** PostgreSQL 16 (one per service)
- **Message Queue:** Apache Kafka (Confluent 7.6)
- **Cache:** Redis 7
- **Auth:** JWT (access + refresh tokens)

### Frontend
- **Framework:** Next.js 16 (App Router)
- **UI:** React 19, TypeScript 5, Tailwind CSS v4
- **Icons:** Lucide React
- **State:** React Context (AuthContext + BookingContext)

### Infrastructure
- **Containerization:** Docker & Docker Compose
- **Services:** 13 containers (4 Postgres, Redis, Zookeeper, Kafka, Kafka UI, 5 app services)

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Node.js](https://nodejs.org/) 18+ (for frontend)
- [Go](https://go.dev/) 1.25+ (for backend development only)

### 1. Start the Backend

```bash
cd backend

# Start all infrastructure and microservices
make up

# Run database migrations
make migrate
```

This starts 13 Docker containers. Wait for all services to be healthy (~30 seconds).

### 2. Seed Sample Data

```bash
cd backend
./scripts/seed.sh
```

This creates:
- **Admin user:** `admin@ticketbox.com` / `admin123`
- **Test user:** `user@example.com` / `user123`
- **3 sample events:** a concert, a sports match, and a film premiere

### 3. Start the Frontend

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

```
booking/
├── backend/
│   ├── docker-compose.yml        # All infrastructure services
│   ├── Makefile                  # Build, run, migrate, test commands
│   ├── .env.example              # Environment variables template
│   ├── proto/                    # Protobuf service definitions
│   │   ├── user/v1/user.proto
│   │   ├── event/v1/event.proto
│   │   └── booking/v1/booking.proto
│   ├── pkg/                      # Shared Go packages
│   │   ├── config/               # Configuration loading
│   │   ├── database/             # Database connection helpers
│   │   ├── kafka/                # Kafka producer/consumer
│   │   ├── middleware/           # gRPC interceptors
│   │   └── proto/                # Generated protobuf Go code
│   ├── services/
│   │   ├── gateway/              # HTTP API gateway (Gin)
│   │   ├── user/                 # User & auth service
│   │   ├── event/                # Event & ticket tier service
│   │   ├── booking/              # Booking service
│   │   └── notification/         # Kafka-driven notification service
│   └── scripts/
│       ├── migrate.sh            # Database migrations
│       ├── seed.sh               # Development seed data
│       └── load-test.sh          # Load testing
│
├── frontend/
│   ├── app/                      # Next.js App Router pages
│   │   ├── page.tsx              # Homepage with featured event + grid
│   │   ├── login/page.tsx        # Login page
│   │   ├── register/page.tsx     # Registration page
│   │   ├── events/[id]/          # Event detail page
│   │   ├── events/[id]/seats/    # Interactive seat selection
│   │   ├── checkout/page.tsx     # Checkout with auth guard
│   │   └── my-tickets/page.tsx   # User's bookings
│   ├── components/
│   │   ├── navbar.tsx            # Auth-aware navigation
│   │   ├── event-card.tsx        # Event card component
│   │   ├── venue-map.tsx         # SVG venue map renderer
│   │   ├── section-picker.tsx    # Seat grid for a section
│   │   ├── seat-legend.tsx       # Seat status legend
│   │   └── seat-tooltip.tsx      # Seat hover tooltip
│   ├── lib/
│   │   ├── api/
│   │   │   ├── client.ts         # HTTP client with JWT refresh
│   │   │   ├── types.ts          # Backend API response types
│   │   │   ├── transformers.ts   # Backend → frontend type mapping
│   │   │   └── generate-venue-layout.ts  # Seat map generator from tiers
│   │   ├── auth-context.tsx      # Authentication state & methods
│   │   ├── booking-context.tsx   # Events, cart, and booking state
│   │   ├── types.ts              # Frontend domain types
│   │   └── mock-data.ts          # Fallback data when backend is offline
│   └── next.config.ts            # API proxy rewrites
│
└── docs/plans/                   # Architecture & implementation plans
```

## API Endpoints

All endpoints are served through the gateway at `http://localhost:8000`.

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register a new user |
| POST | `/api/auth/login` | Login and receive JWT tokens |
| POST | `/api/auth/refresh` | Refresh access token |
| POST | `/api/auth/logout` | Invalidate tokens |
| GET | `/api/users/me` | Get current user profile |

### Events
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/events` | List events (supports `?category=`, `?search=`, `?page=`, `?page_size=`) |
| GET | `/api/events/:id` | Get event details with ticket tiers |
| POST | `/api/events` | Create event (admin only) |

### Bookings
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/bookings` | Create a booking |
| GET | `/api/bookings` | List current user's bookings |
| GET | `/api/bookings/:id` | Get booking details |
| POST | `/api/bookings/:id/cancel` | Cancel a booking |

## Key Features

### Interactive Seat Maps
Every event displays an interactive SVG venue map. Users click a section to see individual seats, then select specific seats before checkout. For API-sourced events, seat maps are automatically generated from ticket tier data using concert arena or sports stadium layouts.

### JWT Authentication
The API client handles the full token lifecycle:
- Stores access/refresh tokens in localStorage
- Automatically retries requests with a refreshed token on 401
- Clears tokens and redirects on session expiry

### Graceful Degradation
When the backend is unavailable, the frontend falls back to built-in mock data so the UI remains functional for development and demos.

### Event-Driven Notifications
Booking events are published to Kafka. The notification service consumes these events to handle confirmation emails and other notifications asynchronously.

## Development

### Backend Commands

```bash
cd backend

make up          # Start all services
make down        # Stop all services
make migrate     # Run database migrations
make logs        # Tail service logs
make test        # Run all tests
make test-load   # Run load tests
make build       # Build all service binaries
make proto       # Regenerate protobuf code
make clean       # Remove build artifacts
```

### Frontend Commands

```bash
cd frontend

npm run dev      # Start dev server on :3000
npm run build    # Production build
npm run start    # Start production server
npm run lint     # Run ESLint
npx tsc --noEmit # Type check without emitting
```

### Ports Reference

| Service | Port | Protocol |
|---------|------|----------|
| Frontend (Next.js) | 3000 | HTTP |
| Gateway | 8000 | HTTP |
| Kafka UI | 8080 | HTTP |
| User Service | 50051 | gRPC |
| Event Service | 50052 | gRPC |
| Booking Service | 50053 | gRPC |
| Postgres (User) | 5433 | TCP |
| Postgres (Event) | 5434 | TCP |
| Postgres (Booking) | 5435 | TCP |
| Postgres (Notification) | 5436 | TCP |
| Redis | 6379 | TCP |
| Kafka | 9092 | TCP |
| Zookeeper | 2181 | TCP |

### Environment Variables

Copy the example env file for backend configuration:

```bash
cp backend/.env.example backend/.env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_PASSWORD` | `ticketbox_secret` | PostgreSQL password for all databases |
| `JWT_SECRET` | `your-jwt-secret-change-in-production` | Secret for signing JWT tokens |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `KAFKA_BROKERS` | `localhost:9092` | Kafka broker addresses |
| `BOOKING_MODE` | `pessimistic` | Booking concurrency strategy |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8000` | Backend gateway URL (frontend) |

## License

This project is for educational and demonstration purposes.
