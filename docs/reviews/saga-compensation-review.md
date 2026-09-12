# Saga Compensation & Resume — Code Review

**Date:** 2026-07-05
**Scope:** `backend/services/saga/internal/service/service.go` and closely related files (registry, repository, kafka consumer, `pkg/cronjob`, event seat repository).
**Reviewed against:**
- **Goal:** a payment failure or timeout must never leave seats stuck in `reserved`.
- **Acceptance:** every failure path runs compensating actions in reverse order and the booking ends `FAILED`.
- **Success criteria:** correct under concurrent webhook delivery and at-least-once redelivery.

> Review produced with the `go-backend-reviewer` skill (fresh-context, anti-trust stance). The reviewer formed its own model of the code before reading any justification, and treats comments/commit messages as checkable claims. Line numbers reference the state of the branch at review time (`feat/add-reservation-timeout`).

---

## Verdict: BLOCK

This change implements the order saga for the ticket-booking flow: `StartOrderSaga` runs synchronously through "reserve seats" (Step 0) and "create payment intent" (Step 1), then pauses for payment; on the Stripe webhook the saga Kafka consumer drives `HandleSagaAferPaymentSuccess` → `ContinueSagaAfterPaymentSuccess` → `reBuildSagaHandler`, which re-attaches Steps 2-4 (mark seats `booked`, set booking `CONFIRMED`, decrement tier availability) and resumes. `HandleSagaAfterPaymentFailure` is the dedicated failure path. The acceptance bar — "a payment failure or timeout must never leave seats stuck in `reserved`; every failure path runs compensating actions in reverse order and the booking ends `FAILED`, correct under concurrent webhook delivery and at-least-once redelivery" — is **not met**. The failure path is broken in multiple independent ways (no compensation ordering, swallowed errors, an idempotency hole that re-collects seats), the success path is non-idempotent and races with itself on redelivery, and there is no timeout path at all (the cronjob that the commit message claims to introduce is a no-op stub and the saga service never starts it). The defects below are data-integrity bugs that leave seats locked or double-booked; this should not merge.

## Findings

### Blocker 1 — `HandleSagaAfterPaymentFailure` runs no saga compensation; a mid-way failure leaves seats in `reserved`
**Location:** `backend/services/saga/internal/service/service.go:528-592`
**Criterion:** Reliability / acceptance criterion ("every failure path runs compensating actions in reverse order and the booking ends FAILED").
**What's wrong:** The failure handler does *not* touch the saga machinery at all. It does not load the saga (`GetSagaByBookingId`), does not call `compensate(...)`, does not set the saga status to `ROLLING_BACK`/`ROLLED_BACK`/`FAIL`. Instead it inlines four gRPC calls sequentially: mark payment `fail` → mark booking `FAILED` → fetch booking → set seats `available`. Every one of those calls is on the hot error path with `return err` on failure. If the second call (`UpdateBookingStatusById`) succeeds but the later `UpdateBatchSeatStatus` fails (network blip to event service, event DB deadlock, anything), the function returns early at line 590 with the seats **still `reserved`** and the saga row **still `WAIT_FOR_PAYMENT`** — so neither the inline recovery nor a future resume will ever release them, because the resume path is keyed off the saga state, which was never updated.
**Why it matters:** This is exactly the stuck-seat failure the acceptance criterion forbids. The most likely real-world trigger (event service briefly unavailable while payment service is up) is the one that strands seats. There is also no retry and no DLQ — once the handler returns `err`, the Kafka reader in `pkg/kafka/consumer.go:54-57` just logs and the message is gone (see the at-least-once Blocker 4 for the offset semantics).
**Recommendation:** Drive failure through the same saga object as success: load the saga, build the handler via `reBuildSagaHandler` (or a failure-mode variant), and call its `compensate(ctx, CurrentStepIndex+1)` so compensations run in reverse order, persisting per-step status. Set the saga and booking to `FAILED` only after the seat-release compensation succeeds. Compensations must be retried until they succeed or land in a DLQ — never logged-and-dropped.
**Tradeoff:** Centralising both outcomes through the saga state machine is more code and one extra DB round-trip on failure, and it requires the failure handler to know the step layout; the payoff is a single termination path you can reason about.

### Blocker 2 — `updateBookingToConfirmed.Execute` swallows the error and returns `nil`; a failed "set CONFIRMED" proceeds to decrement availability and never compensates
**Location:** `backend/services/saga/internal/service/service.go:449-458` (and the mirror in `Compensate` at `:459-468`)
**Criterion:** Reliability / acceptance criterion ("runs compensating actions in reverse order").
**What's wrong:** Inside the `Execute` closure, after the gRPC call fails the code logs the error but the `if err != nil { ... }` block is missing a `return err` — control falls through to `return nil`. The saga's `Execute` loop (`registry/saga_registry.go:110`) therefore treats the step as successful, marks it `COMPLETED`, and advances to the next step. The `Compensate` closure has the identical bug (`:464-467`), so when a later step fails and compensation runs, the booking is never moved to `FAILED` even though seats were released.
**Why it matters:** The booking stays in whatever pre-confirmation status it had while the saga believes the step succeeded, so a downstream failure (e.g. `reduceAvailableSeatNumber`) triggers compensation of the *later* steps only — the booking-status step is "completed" and skipped. The acceptance criterion ("booking ends FAILED") is violated on the exact path it was written for. This is the canonical "comment says handled, code returns nil" bug.
**Recommendation:** Return the error in both closures: `if err != nil { s.logger.Error(...); return err }`. Add a unit test that injects a `BookingServiceClient` failure at this step and asserts both the booking status and the saga outcome.
**Tradeoff:** None — this is a straight correctness fix; the only "cost" is that the saga will now correctly compensate instead of silently corrupting state.

### Blocker 3 — payment-refund compensation is fired in a detached goroutine using the request `ctx`, then the parent step returns `nil` before it finishes; refunds are lost on crash and unobservable
**Location:** `backend/services/saga/internal/service/service.go:387-407`
**Criterion:** Reliability (resource lifecycle, transactions, idempotency).
**What's wrong:** `createPaymentStep.Compensate` launches `go func(ctx context.Context){ ... refund ... }(ctx)`. The `ctx` is the request-scoped context from the webhook handler; the moment `HandleSagaAfterPaymentFailure` / `compensate` returns, that context is cancelled, so `MakeRefundPayment` races against context cancellation. Worse, the goroutine's result is never waited on (`compensate` in `saga_registry.go:157` calls `step.Compensate(ctx)` and immediately proceeds), so (a) the step is marked `COMPENSATED` regardless of whether the refund actually happened, (b) on process restart the refund is gone with no record, and (c) the subsequent `UpdatePaymentStatusByPaymentIntentId` call inside the same closure uses `sagaHandler.GetPaymentResponse().PaymentIntentId` — but `GetPaymentResponse()` is never set on the rebuild path (only `InitializeSagaHandler`'s live Execute sets it), so it dereferences a nil `paymentResponse` and panics inside the goroutine.
**Why it matters:** Three independent failure modes, each sufficient to strand the refund: lost-on-crash, cancelled-by-parent-context, and nil-pointer panic. The compensation ledger (saga step status) is lied to — it says `COMPENSATED` for a refund that never completed. Money is left on the table while seats are released.
**Recommendation:** Make the refund a synchronous call inside `Compensate` (no goroutine), propagate the error so the step is only marked compensated on success, and source the `PaymentIntentId` from `payment.PaymentIntentId` (the `reBuildSagaHandler` argument that is already in scope) rather than the nil `GetPaymentResponse()`. Long term, move refunds to an outbox/DLQ-driven worker keyed on `PaymentIntentId`.
**Tradeoff:** Synchronous refund lengthens the compensation path by one Stripe round-trip; that is the correct cost — a compensation that "succeeds" without doing the work is worse than a slow one.

### Blocker 4 — success path is not idempotent; at-least-once redelivery of `PaymentSucceed` re-executes Steps 2-4 and can double-decrement availability / re-mark seats
**Location:** `backend/services/saga/internal/service/service.go:258-317`, `registry/saga_registry.go:103-147`; redelivery via `pkg/kafka/consumer.go:39-58`
**Criterion:** Reliability / success criterion ("correct under at-least-once redelivery").
**What's wrong:** `ContinueSagaAfterPaymentSuccess` parses the payment, loads the saga, rebuilds the handler, and calls `Execute(ctx, saga.CurrentStepIndex+1)` **unconditionally** — it never checks the saga's terminal status. If the saga is already `COMPLETED` (a prior redelivery finished it) the code happily re-runs `updateSeatAfterPaymentStep` (`status="booked"`, which the SQL at `postgres_event_repository.go:699-718` will re-apply) and, more damagingly, re-runs `reduceAvailableSeatNumber` which issues `UpdateTicketAvailability` with `QuantityDelta: -item.Quantity` again — double-decrementing the tier counter. There is no idempotency key, no `WHERE status = 'reserved'` guard on the seat update, and no version/`updated_at` check. On the Kafka side the reader uses `ReadMessage` (segmentio kafka-go's auto-commit-on-read semantics for consumer groups), and the handler's return error is only logged (`consumer.go:54-57`) — but a crash between handler-completion and the implicit offset commit will replay `PaymentSucceed`, and Stripe itself can deliver the same event twice.
**Why it matters:** The acceptance criterion explicitly requires correctness under at-least-once redelivery. As written, two deliveries of the same `PaymentSucceed` corrupt the availability counter (tier shows fewer seats than reality) — a silent, hard-to-detect data bug. Concurrent delivery from two consumer instances (no partition key on the producer side is shown, so the same payment may not even land on one partition) makes it worse: both can pass the `WAIT_FOR_PAYMENT` read and both re-execute.
**Recommendation:** Gate resume on status: in `ContinueSagaAfterPaymentSuccess`, after loading the saga, `if saga.Status == SAGA_COMPLETED || saga.Status == SAGA_ROLLED_BACK { return nil }` (idempotent ack). Make Steps 2 and 4 idempotent at the target: seat update needs `WHERE id = ANY($1) AND status = 'reserved' AND reserved_by_booking_id = $4` so a re-run is a no-op; tier decrement needs a `booking_id` ledger or an insert into a `tier_deltas` table with a unique constraint so the delta is applied once per booking. Producers should key `PaymentSucceed`/`PaymentFail` by `payment_intent_id` so the same payment always lands on the same partition and is serialised.
**Tradeoff:** Idempotency guards add a `WHERE` clause and a small ledger table; that is the price of at-least-once delivery, which is the stated contract.

### Blocker 5 — `paymentIntentId := paymentDataObj["id"].(string)` panics on any shape variance, killing the consumer goroutine and leaving the saga stuck
**Location:** `backend/services/saga/internal/service/service.go:266`, `:537`; raw-message handling shared with `pkg/typed/type.go`
**Criterion:** Reliability (error handling, robustness under adversarial input).
**What's wrong:** Both handlers do an unchecked type assertion on a `map[string]interface{}` produced by `json.Unmarshal` of arbitrary Stripe webhook JSON. If the `id` field is missing or non-string (Stripe sends nested objects, and the code reads `paymentEvent.Data.Object["id"]` — but `Object` is itself unmarshalled into a `map[string]interface{}`, so any change in payload, a test event, or a refund event routed to the same topic will trigger a panic). There is no `recover` in `ConsumerHandler` or in the kafka `Start` loop, so the panic tears down the goroutine running `sagaComsumer.Start(ctx)` in `cmd/main.go:113`, and since that goroutine's fatal-exit is the only thing keeping the consumer alive, the saga stops processing *all* future webhooks — every in-flight `reserved` seat then times out with no one to release it.
**Why it matters:** A single malformed message takes down the entire saga resume loop. Combined with the missing timeout path (Blocker 6), every pending booking becomes a stuck seat.
**Recommendation:** Use the two-value type assertion `id, ok := obj["id"].(string); if !ok { return fmt.Errorf(...) }`, or unmarshal into a typed struct. Wrap the handler body (and the kafka loop) in a `recover` that logs and continues, so one bad message cannot silence the consumer.
**Tradeoff:** Defensive parsing is a few extra lines; the recover hides programmer errors if overused, so pair it with explicit `ok` checks at every assertion.

### Blocker 6 — there is no timeout / `reservation_expired_at` sweep; seats are stranded the moment a user abandons checkout or the webhook never arrives
**Location:** `backend/services/saga/cmd/main.go` (no cron wiring); `backend/pkg/cronjob/cronjob.go:43-53` (all methods return `nil`); migration `backend/services/event/migrations/000003_add_2_column_reservation_timeout_to_seat_table.up.sql`
**Criterion:** Acceptance criterion ("a payment failure **or timeout** must never leave seats stuck in `reserved`").
**What's wrong:** The commit subject advertises a "stub cronjob pkg" but: (1) `CronJobManager.Start`, `.Stop`, `.RegisterJob` are all empty no-ops; (2) `cmd/main.go` never instantiates `cronjob.New(...)` and never registers any job that scans `seats WHERE status='reserved' AND reservation_expired_at < NOW()`; (3) `ReservedOrCompensateBatchSeats` only ever runs from the saga's reserve/compensate step, never from a timer. So if a user closes the browser after Step 1 (payment intent created, saga paused at `WAIT_FOR_PAYMENT`) and the Stripe webhook never comes (declined card with no event, network loss, webhook misconfiguration), the seats stay `reserved` forever — `reservation_expired_at` is set but nothing reads it. The acceptance criterion explicitly names the timeout case, and there is no code path that handles it.
**Why it matters:** This is the dominant real-world failure mode for any ticketing system (cart abandonment), and it is entirely unhandled. Even with the failure path fixed, abandoned checkouts leak seats.
**Recommendation:** Implement a sweeper (cron or a long-running ticker) that runs `UPDATE seats SET status='available', reservation_expired_at=NULL, reserved_by_booking_id=NULL WHERE status='reserved' AND reservation_expired_at < NOW() RETURNING reserved_by_booking_id`, and for each distinct booking id also drives the saga to `FAIL` (so the booking row and any half-created payment are cleaned up). The sweeper update must be idempotent and ordered before any saga-Failed transition. Wire `cronjob` in `cmd/main.go` (and either implement the manager or replace the stub with `robfig/cron` directly).
**Tradeoff:** Adds a background worker with its own failure modes (it must not race with an in-flight webhook — see Blocker 4); the mitigation is to have the sweeper mark the saga `FAIL` *before* releasing seats, and have the success-resume path check saga status before proceeding.

### Major 1 — `HandleSagaAfterPaymentFailure` and the success path can race on the same booking and produce inconsistent seat state
**Location:** `backend/services/saga/internal/service/service.go:258` vs `:528`; no saga-level lock
**Criterion:** Reliability / success criterion ("correct under concurrent webhook delivery").
**What's wrong:** Stripe can deliver both a `payment_intent.payment_failed` and a late `payment_intent.succeeded` (or the reverse) for the same intent, and there is no mutual exclusion on the saga. `HandleSagaAfterPaymentFailure` releases seats and sets booking `FAILED` without re-reading saga status; the concurrent `ContinueSagaAfterPaymentSuccess` re-reads the saga (still `WAIT_FOR_PAYMENT`) and proceeds to mark seats `booked`. The two interleave against the same seat rows with no row-level lock in the success path's `UpdateBatchSeatStatus` (the SQL has no `FOR UPDATE` or `WHERE status=` guard). Outcomes include: seats marked `available` then `booked` (sold to a user whose payment actually failed), or `booked` then `available` (released after being sold).
**Why it matters:** Double booking / phantom release — the core problem the project exists to solve. The saga is supposed to be the single arbiter; bypassing it in the failure handler defeats that.
**Recommendation:** Both entry points must take the saga row lock (`SELECT ... FOR UPDATE` on the saga row, or a Redis lock keyed on `booking_id`) and re-read status. Resume must `return nil` if the saga is already terminal. The failure path must update saga status to `ROLLING_BACK` *before* doing work so a concurrent success sees the terminal state and bails out.
**Tradeoff:** A row lock serialises the two paths for the same booking (microseconds) — negligible vs. the consistency gain.

### Major 2 — `compensate` continues past failures and reports `isRolledback=false`/`SAGA_FAIL`, but the function still returns `nil`; callers cannot distinguish partial rollback from full success
**Location:** `backend/services/saga/internal/registry/saga_registry.go:149-187`
**Criterion:** Reliability (error handling, idempotency/retry of compensations).
**What's wrong:** When a `Compensate` closure errors, the loop logs, sets `isRolledback=false`, and `continue`s — then the function ends with `s.saga.Status = SAGA_FAIL`, persists, and returns `nil`. `Execute` (`:118`) returns whatever `compensate` returned, so `StartOrderSaga`/`ContinueSagaAfterPaymentSuccess` get `nil` and report success to the booking service and the user. There is no retry, no DLQ enqueue, no signal to a human. The seats remain `reserved` (the failed compensation's step never released them) and nobody is alerted.
**Why it matters:** Compensation failures are precisely when seats get stuck. Silently swallowing them and reporting success guarantees the stuck-seat condition persists indefinitely.
**Recommendation:** Return a typed error (e.g. `ErrCompensationIncomplete`) when `!isRolledback`, and have the consumer not commit the offset (or enqueue a retry). Track per-step compensation attempts in the DB so a sweeper can retry failed compensations to completion.
**Tradeoff:** Surfaces failures that today are hidden — operationally noisier until the retry/DLQ is in place, but that noise is the *correct* behaviour.

### Major 3 — `reBuildSagaHandler` indexes `saga.Steps[0]`/`[1]` without checking length, and `GetSagaByBookingId` returns steps in unspecified order
**Location:** `backend/services/saga/internal/service/service.go:302, 361-407`; SQL has no `ORDER BY` at `repository/postgres_saga_repository.go:195`
**Criterion:** Reliability (robustness, idempotency of rebuild).
**What's wrong:** `GetSagaByBookingId`'s step query (`:195-198`) has no `ORDER BY "order"`, so the slice order depends on physical storage; `reBuildSagaHandler` then assumes `saga.Steps[0]` is the reserve step and `saga.Steps[1]` is the payment step (after `FreeUpSteps` it has wiped the slice and re-adds them in fixed order, but the *initial* reads `saga.Steps[0].ID` etc. for the rebuild — so it copies whichever step happened to land at index 0). On any reorder, partial insert, or future schema change, this mis-attributes the step ID and either the wrong step is "completed" or the upsert collides on the wrong `id`. There is also no length guard — a saga whose steps were never persisted (e.g. `Create` partial failure) panics on index out of range.
**Why it matters:** Silent mis-attribution of step state is exactly the kind of bug that strands seats: the compensations are wired to closures keyed by name, but the persisted `id`s come from positional reads that may not match.
**Recommendation:** Add `ORDER BY "order" ASC` to both step queries, and in `reBuildSagaHandler` look up steps by `Name` (the constants from `pkg/saga`) rather than by index. Guard `len(saga.Steps) >= 2` and return a typed error otherwise.
**Tradeoff:** Name-based lookup is a few lines and removes a whole class of positional bugs; the cost is requiring `Name` to be unique within a saga, which it already is by construction.

### Minor 1 — `ContinueSagaAfterPaymentSuccess` logs `saga.CurrentStepIndex` and dereferences `saga.ID` *before* checking the load error
**Location:** `backend/services/saga/internal/service/service.go:302-306`
**Criterion:** Reliability (error handling, nil safety).
**What's wrong:** `GetSagaByBookingId` returns `(nil, err)` on miss; the next line (`:303`) does `saga.ID.String(), saga.CurrentStepIndex` via `Infof` before the `if err != nil` check on `:304`, which is a nil-pointer panic when the booking has no saga. The `err` check is also logically inverted (logging happens regardless of error).
**Why it matters:** A malformed payment event whose `BookingId` has no saga (a refund for a deleted booking, a test event) panics the consumer goroutine, with the same blast radius as the type-assertion Blocker 5.
**Recommendation:** Move the log line after the error check: `if err != nil { return err }; if saga == nil { ... }; s.logger...`.
**Tradeoff:** None.

### Minor 2 — `UpdateBatchSeatStatus` compensate ignores `reserved_by_booking_id`, so it can clobber a different booking's hold
**Location:** `backend/services/event/internal/repository/postgres_event_repository.go:707-712` vs `backend/services/saga/internal/service/service.go:188-198`
**Criterion:** Reliability (idempotency, correctness of compensation).
**What's wrong:** The compensate closure sends `Action: "compensate"` and only `SeatIds` — it does not pass `ReservedByBookingId`. The SQL therefore unconditionally sets seats to `available` for the given IDs. If (due to any of the races above) those seat IDs have already been re-reserved by a *different* booking (legitimate concurrent purchase after a sweeper released them), this compensate will silently clobber the new reservation and re-open double booking.
**Why it matters:** Defeats the seat-level double-booking guard.
**Recommendation:** The compensate SQL must be `WHERE id = ANY($1) AND reserved_by_booking_id = $bookingId AND status='reserved'` so it only releases *this* booking's hold. Pass `ReservedByBookingId` in the compensate request.
**Tradeoff:** One extra `WHERE` predicate; releases that target an already-released seat become no-ops, which is the desired behaviour.

### Minor 3 — hardcoded user email, currency comment drift, `Price` truncation
**Location:** `backend/services/saga/internal/service/service.go:211-214`
**Criterion:** Maintainability.
**What's wrong:** `UserEmail: "nhkhai2805@gmail.com"` is a developer address baked into the payment intent for every customer. The comment `// Convert to USD` next to `Price: int32(req.TotalCents)` is wrong — it's converting cents to an `int32` (losing precision above ~$21M and dropping the cents/dollars semantics entirely), not converting currency. `Currency: "usd"` is hardcoded.
**Why it matters:** Receipts/Stripe metadata go to the wrong user; a large event overflows `int32`; the misleading comment will mislead the next reader.
**Recommendation:** Plumb the real email from the booking/user; carry `TotalCents` as `int64`; delete the misleading comment or make the conversion explicit.
**Tradeoff:** Requires the booking context to carry the user email — one extra field on the booking record or a user-service lookup.

### Nit — typo `HandleSagaAferPaymentSuccess` (missing `t` in `After`)
**Location:** `backend/services/saga/internal/service/service.go:258`, `internal/kafka/consumer.go:43`
**Criterion:** Maintainability (naming).
**What's wrong:** Public method name misspelled; rippled to the consumer switch.
**Recommendation:** Rename to `HandleSagaAfterPaymentSuccess`. Renaming a method on an internal struct is cheap; do it before an external caller depends on the misspelling.
**Tradeoff:** Touches two files and any future callers — trivial now, costly later.

## Strengths

- Persisting per-step status and `CurrentStepIndex` and rebuilding the handler on resume (`reBuildSagaHandler`) is the right shape for a saga that must survive a process restart between "pause for payment" and "webhook resumes." The bones of the state machine are correct; the bugs are in how the state is driven, not in the persistence model.
- Centralising the gRPC clients behind `SagaService` thin wrappers and registering step processors in a registry keeps `service.go` decoupled from the concrete proto clients — the seam is clean enough to inject a fake `BookingServiceClient`/`PaymentServiceClient` and test the failure paths directly (which is exactly what's needed to fix the Blockers above).
- Wrapping every gRPC/DB error with `fmt.Errorf("...: %w", err)` is applied consistently in the service layer, so once the swallowed-error bugs are fixed, errors will propagate cleanly to the consumer for retry/DLQ handling.

---

## Suggested remediation order

A suggested order to tackle the findings, prioritising the acceptance criterion:

1. **Blocker 6 (timeout sweep)** — this is the headline feature of the branch and currently does nothing; without it the criterion is unreachable regardless of the other fixes.
2. **Blocker 2 + Blocker 3 (swallowed errors, detached refund goroutine)** — straight correctness fixes; unblock proper compensation.
3. **Blocker 1 (drive failure through the saga)** — depends on 2 and 3 being correct.
4. **Blocker 4 + Major 1 (idempotency + success/failure race)** — the at-least-once + concurrency guards; needed for the success criteria.
5. **Blocker 5 (panic on shape variance)** — cheap, high blast-radius; pair with a `recover`.
6. **Major 2 + Major 3 + Minors** — hardening and correctness of rebuild/compensation accounting.
