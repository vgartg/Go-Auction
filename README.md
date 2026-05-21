# GoAuction

[![CI](https://github.com/vgartg/goauction/actions/workflows/ci.yml/badge.svg)](https://github.com/vgartg/goauction/actions/workflows/ci.yml)
[![Deploy](https://github.com/vgartg/goauction/actions/workflows/deploy.yml/badge.svg)](https://github.com/vgartg/goauction/actions/workflows/deploy.yml)
[![Live Demo](https://img.shields.io/badge/demo-live-22C55E?logo=githubpages&logoColor=white)](https://vgartg.github.io/Go-Auction/)
[![Go Report Card](https://goreportcard.com/badge/github.com/vgartg/PET_Go-Auction)](https://goreportcard.com/report/github.com/vgartg/PET_Go-Auction)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Real-time auction engine in Go. Concurrent bidding with optimistic locking, anti-sniping, JWT auth, WebSocket updates and a server-rendered UI on **templ + HTMX + Tailwind**

> One binary serves the JSON REST API on `/api/*` and the full HTML UI on `/`. Same engine, same database, no SPA toolchain

**Live demo:** <https://vgartg.github.io/Go-Auction> — a Vite/TypeScript SPA in [`web/`](web/) that ports the auction engine to the browser so the demo runs without Postgres, JWT or the Go binary

---

<img width="1015" height="421" alt="image" src="https://github.com/user-attachments/assets/ac2bc892-6ca8-41bb-8559-9fc68e3da38b" />

---

## Features

- REST API and web UI in one binary
- Real-time bidding over WebSocket (`new_bid`, `lot_extended`, `lot_closed`)
- Optimistic locking + bounded retry, all inside one SQL transaction
- Anti-sniping that auto-extends `closing_at` near the wire
- Auto-close timers restored on restart
- JWT auth (HS256, HttpOnly cookie or `Authorization: Bearer …`)
- Per-IP token-bucket rate limiting on auth and bidding
- Prometheus metrics: bid counters, opt-lock retries, anti-sniping extensions, bid latency, HTTP duration, rate-limit rejections
- Structured `slog` JSON access log with `request_id`
- PostgreSQL with `golang-migrate`, FK constraints and indexes
- OpenAPI 3 spec, Docker, GitHub Actions CI, Kubernetes manifests

---

## Quick start

```bash
git clone https://github.com/vgartg/goauction.git
cd goauction
docker-compose up --build
```

Open [http://localhost:8080](http://localhost:8080)

| Path        | What it is                          |
|-------------|-------------------------------------|
| `/`         | Web UI                              |
| `/api/*`    | JSON REST                           |
| `/ws/lots/{id}` | WebSocket per lot               |
| `/metrics`  | Prometheus                          |
| `/healthz`  | Liveness                            |

---

## Architecture

```
cmd/goauction/main.go        wiring (config → repo → engine → web/api → http)

internal/
├── config/                  env-driven config (caarlos0/env)
├── models/                  Lot, Bid, User, UserStats
├── repository/              LotRepository, UserRepository + Postgres impl
├── auth/                    bcrypt, JWT, middleware, session cookie
├── auction/engine.go        bidding rules, optimistic lock, anti-sniping, timers
├── metrics/                 Prometheus counters / gauges / histograms
├── httpx/                   access-log middleware + token-bucket rate limiter
├── api/                     JSON REST handlers + WebSocket manager
└── web/                     server-rendered HTML (templ) + chi routes

migrations/                  golang-migrate up/down SQL
deploy/k8s/                  Postgres StatefulSet + app Deployment / HPA
api/openapi.yaml             OpenAPI 3 spec
```

The engine knows nothing about HTTP or templ — it depends on `repository.LotRepository` and a `WSBroadcaster`. `PlaceBid` opens one Postgres transaction via `repo.WithinTx(ctx, …)` that wraps `SELECT … FOR UPDATE`, the bid insert and the lot update; the whole tx restarts on `ErrOptimisticLock`

---

## Local development

```bash
make tools          # install templ CLI
make templ          # regenerate views (one-shot)
make run            # go run ./cmd/goauction
make test           # go test -race ./...
make lint           # golangci-lint v2
make templ-watch    # codegen on .templ changes
```

Requires Go 1.25+ and Postgres 16 if running outside Docker

---

## Configuration

| Variable             | Default                                                                 |
|----------------------|-------------------------------------------------------------------------|
| `PORT`               | `8080`                                                                  |
| `DATABASE_URL`       | `postgres://postgres:postgres@localhost:5432/goauction?sslmode=disable` |
| `JWT_SECRET`         | `dev-insecure-secret-please-change`                                     |
| `SNIPING_WINDOW`     | `30s`                                                                   |
| `SNIPING_EXTENSION`  | `30s`                                                                   |
| `BID_RATE_PER_SEC`   | `5`                                                                     |
| `BID_BURST`          | `10`                                                                    |
| `AUTH_RATE_PER_SEC`  | `1`                                                                     |
| `AUTH_BURST`         | `5`                                                                     |
| `METRICS_ENABLED`    | `true`                                                                  |

---

## API

```
POST   /api/auth/register           → { token, user_id, username }
POST   /api/auth/login              → { token, user_id, username }
GET    /api/auth/me                 (auth)

GET    /api/lots[?status=&limit=&offset=]
GET    /api/lots/{id}
POST   /api/lots                    (auth)
POST   /api/lots/{id}/bids          (auth, rate-limited)

GET    /api/users/{id}/stats

GET    /ws/lots/{id}                WebSocket
GET    /metrics, /healthz
```

Full schema in [`api/openapi.yaml`](api/openapi.yaml) — paste into <https://editor.swagger.io>

### WebSocket events

```js
const ws = new WebSocket(`ws://localhost:8080/ws/lots/${LOT}`);
ws.onmessage = (e) => console.log(JSON.parse(e.data));
// { type: "new_bid",      lot_id, user_id, amount, new_price, timestamp }
// { type: "lot_extended", lot_id, closing_at, extended_count }
// { type: "lot_closed",   lot_id, winner_id, final_price }
```

The lot page reconnects with exponential backoff if the socket drops

---

## Deployment

```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/postgres.yaml
kubectl apply -f deploy/k8s/goauction.yaml
```

Gets you Postgres `StatefulSet` with a 1Gi PVC, a 2-replica `Deployment` with readiness/liveness on `/healthz`, and an HPA scaling 2→6 pods at 70% CPU

---

## Roadmap

- Redis for distributed rate limiting and hot-list cache
- OpenTelemetry tracing + Grafana dashboards in `/deploy`
- Outbox + Kafka publisher
- `testcontainers-go` integration tests for the repo
