---
id: ref-jwt-auth-flow
c3-version: 4
title: jwt-auth-flow
goal: Enforce consistent JWT-based authentication across frontend and backend
scope: [c3-1, c3-2, c3-6]
---

# jwt-auth-flow

## Goal

Enforce consistent JWT-based authentication across frontend and backend.

## Choice

Stateless JWT with access+refresh token pair. User Service generates tokens, Gateway validates on every request, Frontend stores in localStorage and auto-refreshes on 401. Redis blacklist for logout.

## Why

Stateless auth scales horizontally without shared session store. Refresh tokens enable short-lived access tokens (security) without frequent re-login (UX). Redis blacklist handles logout without making JWTs stateful.

## How

| Guideline | Example |
|-----------|---------|
| User Service generates both tokens | `access_token` (short TTL) + `refresh_token` (long TTL) |
| Gateway validates JWT on protected routes | Middleware extracts and verifies token, sets user context |
| Gateway checks Redis blacklist | Rejected tokens return 401 immediately |
| Frontend stores tokens in localStorage | `access_token` and `refresh_token` keys |
| Frontend auto-refreshes on 401 | Intercepts 401, calls refresh endpoint, retries original request |
| Logout blacklists token in Redis | Gateway adds token to Redis with TTL matching token expiry |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| Session-based auth | Requires shared session store, doesn't scale as well |
| OAuth2/OIDC | Overkill for single-app platform, adds complexity |
| Cookie-based tokens | Cross-origin complications with microservices |

## Scope

**Applies to:**
- c3-1 (gateway middleware), c3-2 (token generation), c3-6 (client-side token management)

**Does NOT apply to:**
- Service-to-service gRPC calls (trusted internal network)

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-102, c3-110, c3-210, c3-601, c3-602, c3-614
