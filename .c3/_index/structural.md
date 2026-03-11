# C3 Structural Index
<!-- hash: sha256:f1eac36347510c596e5698262b48da707ad1899f18087ab040f9d4a87630554a -->

## adr-00000000-c3-adoption — C3 Architecture Documentation Adoption (adr)
blocks: Goal ✓

## c3-0 — ${PROJECT} (context)
reverse deps: adr-00000000-c3-adoption, c3-1, c3-2, c3-3, c3-4, c3-5, c3-6, c3-7
blocks: Abstract Constraints ✓, Containers ○, Goal ✓

## c3-1 — gateway (container)
context: c3-0
reverse deps: c3-101, c3-102, c3-110, c3-111, c3-112, c3-113
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-101 — http-router (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-102 — auth-middleware (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-110 — auth-handler (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-111 — event-handler (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-112 — booking-handler (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-113 — user-handler (component)
container: c3-1 | context: c3-0
constraints from: c3-0, c3-1
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-2 — user-service (container)
context: c3-0
reverse deps: c3-201, c3-202, c3-210, c3-211
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-201 — grpc-server (component)
container: c3-2 | context: c3-0
constraints from: c3-0, c3-2
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-202 — user-repository (component)
container: c3-2 | context: c3-0
constraints from: c3-0, c3-2
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-210 — auth-logic (component)
container: c3-2 | context: c3-0
constraints from: c3-0, c3-2
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-211 — profile-management (component)
container: c3-2 | context: c3-0
constraints from: c3-0, c3-2
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-3 — event-service (container)
context: c3-0
reverse deps: c3-301, c3-302, c3-310, c3-311
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-301 — grpc-server (component)
container: c3-3 | context: c3-0
constraints from: c3-0, c3-3
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-302 — event-repository (component)
container: c3-3 | context: c3-0
constraints from: c3-0, c3-3
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-310 — event-crud (component)
container: c3-3 | context: c3-0
constraints from: c3-0, c3-3
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-311 — ticket-availability (component)
container: c3-3 | context: c3-0
constraints from: c3-0, c3-3
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-4 — booking-service (container)
context: c3-0
reverse deps: c3-401, c3-402, c3-410, c3-411
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-401 — grpc-server (component)
container: c3-4 | context: c3-0
constraints from: c3-0, c3-4
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-402 — booking-repository (component)
container: c3-4 | context: c3-0
constraints from: c3-0, c3-4
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-410 — booking-creation (component)
container: c3-4 | context: c3-0
constraints from: c3-0, c3-4
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-411 — booking-management (component)
container: c3-4 | context: c3-0
constraints from: c3-0, c3-4
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-5 — notification-service (container)
context: c3-0
reverse deps: c3-501, c3-502, c3-510
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-501 — kafka-consumer (component)
container: c3-5 | context: c3-0
constraints from: c3-0, c3-5
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-502 — notification-repository (component)
container: c3-5 | context: c3-0
constraints from: c3-0, c3-5
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-510 — notification-processing (component)
container: c3-5 | context: c3-0
constraints from: c3-0, c3-5
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-6 — frontend (container)
context: c3-0
reverse deps: c3-601, c3-602, c3-603, c3-604, c3-610, c3-611, c3-612, c3-613, c3-614
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-601 — api-client (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-602 — auth-state (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-603 — booking-state (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-604 — data-transformers (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-610 — event-browsing (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-611 — event-detail (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-612 — checkout (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-613 — my-tickets (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-614 — auth-pages (component)
container: c3-6 | context: c3-0
constraints from: c3-0, c3-6
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-7 — shared-backend (container)
context: c3-0
reverse deps: c3-701, c3-702, c3-703, c3-704, c3-705
constraints from: c3-0
blocks: Complexity Assessment ✓, Components ○, Goal ○, Responsibilities ✓

## c3-701 — config (component)
container: c3-7 | context: c3-0
constraints from: c3-0, c3-7
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-702 — database (component)
container: c3-7 | context: c3-0
constraints from: c3-0, c3-7
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-703 — kafka-pkg (component)
container: c3-7 | context: c3-0
constraints from: c3-0, c3-7
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-704 — grpc-middleware (component)
container: c3-7 | context: c3-0
constraints from: c3-0, c3-7
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## c3-705 — proto-generated (component)
container: c3-7 | context: c3-0
constraints from: c3-0, c3-7
blocks: Container Connection ✓, Dependencies ✓, Goal ○, Related Refs ○

## ref-concurrency-control — concurrency-control (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-data-shape-mapping — data-shape-mapping (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-grpc-service-structure — grpc-service-structure (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-jwt-auth-flow — jwt-auth-flow (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-outbox-pattern — outbox-pattern (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## ref-repository-pattern — repository-pattern (ref)
blocks: Choice ✓, Goal ✓, How ✓, Why ✓

## Ref Map
ref-concurrency-control
ref-data-shape-mapping
ref-grpc-service-structure
ref-jwt-auth-flow
ref-outbox-pattern
ref-repository-pattern
