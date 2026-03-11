# Task 4: Protobuf Definitions

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Define gRPC service contracts for User, Event, and Booking services.

**Files:**
- Create: `backend/proto/user/v1/user.proto`
- Create: `backend/proto/event/v1/event.proto`
- Create: `backend/proto/booking/v1/booking.proto`

---

### Step 1: Install protoc tools (if not installed)

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Step 2: Create user.proto

`backend/proto/user/v1/user.proto`:
```protobuf
syntax = "proto3";

package user.v1;

option go_package = "github.com/ticketbox/pkg/proto/user/v1;userv1";

import "google/protobuf/timestamp.proto";

service UserService {
    rpc Register(RegisterRequest) returns (AuthResponse);
    rpc Login(LoginRequest) returns (AuthResponse);
    rpc RefreshToken(RefreshTokenRequest) returns (AuthResponse);
    rpc Logout(LogoutRequest) returns (LogoutResponse);
    rpc GetProfile(GetProfileRequest) returns (UserProfile);
    rpc UpdateProfile(UpdateProfileRequest) returns (UserProfile);
    rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

message RegisterRequest {
    string email = 1;
    string password = 2;
    string name = 3;
}

message LoginRequest {
    string email = 1;
    string password = 2;
}

message AuthResponse {
    string access_token = 1;
    string refresh_token = 2;
    UserProfile user = 3;
}

message RefreshTokenRequest {
    string refresh_token = 1;
}

message LogoutRequest {
    string access_token = 1;
}

message LogoutResponse {}

message GetProfileRequest {
    string user_id = 1;
}

message UpdateProfileRequest {
    string user_id = 1;
    string name = 2;
    string email = 3;
}

message UserProfile {
    string id = 1;
    string email = 2;
    string name = 3;
    string role = 4;
    google.protobuf.Timestamp created_at = 5;
}

message ValidateTokenRequest {
    string access_token = 1;
}

message ValidateTokenResponse {
    bool valid = 1;
    string user_id = 2;
    string email = 3;
    string role = 4;
}
```

### Step 3: Create event.proto

`backend/proto/event/v1/event.proto`:
```protobuf
syntax = "proto3";

package event.v1;

option go_package = "github.com/ticketbox/pkg/proto/event/v1;eventv1";

import "google/protobuf/timestamp.proto";

service EventService {
    rpc CreateEvent(CreateEventRequest) returns (EventDetail);
    rpc GetEvent(GetEventRequest) returns (EventDetail);
    rpc ListEvents(ListEventsRequest) returns (ListEventsResponse);
    rpc UpdateEvent(UpdateEventRequest) returns (EventDetail);
    rpc DeleteEvent(DeleteEventRequest) returns (DeleteEventResponse);
    rpc GetTicketAvailability(GetTicketAvailabilityRequest) returns (TicketAvailabilityResponse);
    rpc UpdateTicketAvailability(UpdateTicketAvailabilityRequest) returns (TicketTier);
}

message CreateEventRequest {
    string title = 1;
    string description = 2;
    string category = 3;
    string venue = 4;
    string location = 5;
    google.protobuf.Timestamp date = 6;
    string image_url = 7;
    repeated CreateTicketTierRequest tiers = 8;
}

message CreateTicketTierRequest {
    string name = 1;
    int64 price_cents = 2;
    int32 total_quantity = 3;
}

message GetEventRequest {
    string event_id = 1;
}

message ListEventsRequest {
    string category = 1;
    string search = 2;
    int32 page = 3;
    int32 page_size = 4;
}

message ListEventsResponse {
    repeated EventDetail events = 1;
    int32 total = 2;
    int32 page = 3;
    int32 page_size = 4;
}

message UpdateEventRequest {
    string event_id = 1;
    string title = 2;
    string description = 3;
    string category = 4;
    string venue = 5;
    string location = 6;
    google.protobuf.Timestamp date = 7;
    string image_url = 8;
}

message DeleteEventRequest {
    string event_id = 1;
}

message DeleteEventResponse {}

message EventDetail {
    string id = 1;
    string title = 2;
    string description = 3;
    string category = 4;
    string venue = 5;
    string location = 6;
    google.protobuf.Timestamp date = 7;
    string image_url = 8;
    string status = 9;
    repeated TicketTier tiers = 10;
    google.protobuf.Timestamp created_at = 11;
}

message TicketTier {
    string id = 1;
    string event_id = 2;
    string name = 3;
    int64 price_cents = 4;
    int32 total_quantity = 5;
    int32 available_quantity = 6;
    int32 version = 7;
}

message GetTicketAvailabilityRequest {
    string tier_id = 1;
}

message TicketAvailabilityResponse {
    string tier_id = 1;
    int32 available_quantity = 2;
    int32 version = 3;
}

message UpdateTicketAvailabilityRequest {
    string tier_id = 1;
    int32 quantity_delta = 2;
    int32 expected_version = 3;
}
```

### Step 4: Create booking.proto

`backend/proto/booking/v1/booking.proto`:
```protobuf
syntax = "proto3";

package booking.v1;

option go_package = "github.com/ticketbox/pkg/proto/booking/v1;bookingv1";

import "google/protobuf/timestamp.proto";

service BookingService {
    rpc CreateBooking(CreateBookingRequest) returns (BookingDetail);
    rpc GetBooking(GetBookingRequest) returns (BookingDetail);
    rpc ListUserBookings(ListUserBookingsRequest) returns (ListUserBookingsResponse);
    rpc CancelBooking(CancelBookingRequest) returns (BookingDetail);
}

message CreateBookingRequest {
    string user_id = 1;
    string event_id = 2;
    repeated BookingItemRequest items = 3;
    string booking_mode = 4;
}

message BookingItemRequest {
    string ticket_tier_id = 1;
    int32 quantity = 2;
}

message GetBookingRequest {
    string booking_id = 1;
}

message ListUserBookingsRequest {
    string user_id = 1;
    int32 page = 2;
    int32 page_size = 3;
}

message ListUserBookingsResponse {
    repeated BookingDetail bookings = 1;
    int32 total = 2;
}

message CancelBookingRequest {
    string booking_id = 1;
    string user_id = 2;
}

message BookingDetail {
    string id = 1;
    string user_id = 2;
    string event_id = 3;
    string status = 4;
    int64 total_amount_cents = 5;
    repeated BookingItem items = 6;
    google.protobuf.Timestamp created_at = 7;
}

message BookingItem {
    string id = 1;
    string ticket_tier_id = 2;
    string tier_name = 3;
    int32 quantity = 4;
    int64 unit_price_cents = 5;
}
```

### Step 5: Generate Go code

```bash
cd /Users/dev/work/booking/backend
mkdir -p pkg/proto/{user/v1,event/v1,booking/v1}

protoc --proto_path=proto \
    --go_out=pkg/proto --go_opt=paths=source_relative \
    --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
    proto/user/v1/user.proto

protoc --proto_path=proto \
    --go_out=pkg/proto --go_opt=paths=source_relative \
    --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
    proto/event/v1/event.proto

protoc --proto_path=proto \
    --go_out=pkg/proto --go_opt=paths=source_relative \
    --go-grpc_out=pkg/proto --go-grpc_opt=paths=source_relative \
    proto/booking/v1/booking.proto
```

### Step 6: Update pkg/go.mod

```bash
cd /Users/dev/work/booking/backend/pkg
go get google.golang.org/protobuf
go get google.golang.org/grpc
go mod tidy
```

### Step 7: Verify generated code compiles

```bash
cd /Users/dev/work/booking/backend/pkg
go build ./...
```
Expected: No errors.

### Step 8: Commit

```bash
git add backend/proto/ backend/pkg/proto/
git commit -m "feat(backend): add protobuf definitions for user, event, booking services"
```
