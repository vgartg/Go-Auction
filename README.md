# GoAuction – Real-time auction engine in Go

[![CI](https://github.com/vgartg/goauction/actions/workflows/ci.yml/badge.svg)](https://github.com/vgartg/goauction/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vgartg/goauction)](https://goreportcard.com/report/github.com/vgartg/goauction)

A small but production-shaped auction service: concurrent bidding with optimistic
locking, automatic lot closing, anti-sniping, JWT auth, real-time updates over
WebSocket and a server-rendered UI on **templ + HTMX + Tailwind**.

## Features

- **REST + Web UI in one binary.** Same engine drives `/api/...` JSON and a
  fully-server-rendered HTML site on `/`.
- **Real-time bidding** over WebSocket — broadcasts `new_bid`,
  `lot_extended` and `lot_closed` events to all subscribers of a lot.
- **Optimistic locking** on lots (`version` column) with bounded retry. No
  ghost bids on contention.
- **Anti-sniping** — a bid placed within the configured window before close
  automatically extends `closing_at`. Configurable via `SNIPING_WINDOW` /
  `SNIPING_EXTENSION`.
- **Auto-close timers** restored after restart from `GetActiveLots`,
  rescheduled on extension.
- **JWT auth** (HS256, HttpOnly cookie or `Authorization: Bearer …`)
  with bcrypt password hashing.
- **User stats** — derived from bids/wins/spend, exposed at
  `/api/users/{id}/stats` and as a public profile page.
- **Prometheus metrics** at `/metrics`: `auction_bids_total{lot_id}`,
  `auction_active_lots`, `auction_lot_closures_total`.
- **PostgreSQL** with `golang-migrate`, indexes, FK constraints.
- **Docker / docker-compose / GitHub Actions CI**.

## Architecture

```
cmd/goauction/main.go            ── wiring (config → repo → engine → web/api → http)

internal/
├── config/                       env-driven config (caarlos0/env)
├── models/                       Lot, Bid, User, UserStats
├── repository/                   LotRepository, UserRepository (+ Postgres impl)
├── auth/                         bcrypt, JWT issue/parse, middleware, session cookie
├── auction/engine.go             business rules: bidding, optimistic lock, anti-sniping, timers
├── metrics/                      Prometheus counters/gauges
├── api/                          JSON REST handlers + WebSocket manager
└── web/
    ├── handlers.go               server-rendered HTML handlers (HTMX-friendly)
    ├── routes.go                 chi routing
    └── views/*.templ             type-safe templ components (layout, lot, auth, profile)

migrations/                       golang-migrate up/down SQL
```

The engine has no knowledge of HTTP or templ — it depends only on a
`LotRepository` and a `WSBroadcaster` interface, which keeps the bidding
logic unit-testable with plain testify mocks.

## Quick start

```bash
git clone https://github.com/vgartg/goauction.git
cd goauction
docker-compose up --build
# open http://localhost:8080
```

The web UI ships on `/`. The JSON API is on `/api/...`. WebSocket is at
`/ws/lots/{lot_id}`. Metrics at `/metrics`.

## Local dev

```bash
make tools          # installs the templ CLI
make templ          # codegen for templ views
make run            # go run ./cmd/goauction (needs Postgres at $DATABASE_URL)
make test           # go test -race ./...
make templ-watch    # background codegen during development
```

## Configuration

| Variable             | Default                                                                      | Purpose                                    |
|----------------------|------------------------------------------------------------------------------|--------------------------------------------|
| `PORT`               | `8080`                                                                       | HTTP listen port                           |
| `DATABASE_URL`       | `postgres://postgres:postgres@localhost:5432/goauction?sslmode=disable`      | PG connection string                       |
| `JWT_SECRET`         | `dev-insecure-secret-please-change`                                          | HS256 signing key — override in production |
| `SNIPING_WINDOW`     | `30s`                                                                        | If bid lands within this window before close, extend |
| `SNIPING_EXTENSION`  | `30s`                                                                        | How far to push `closing_at` on extension  |
| `METRICS_ENABLED`    | `true`                                                                       | Expose `/metrics`                          |

## API surface

```
POST   /api/auth/register      { username, email, password } → { token, ... }
POST   /api/auth/login         { email, password }            → { token, ... }
GET    /api/auth/me                                            → User (auth required)

GET    /api/lots               → []Lot
GET    /api/lots/{id}          → Lot
POST   /api/lots               (auth) { title, start_price, min_step, closing_at } → Lot
POST   /api/lots/{id}/bids     (auth) { amount } → Lot

GET    /api/users/{id}/stats   → { bids_count, wins_count, total_spent, ... }

GET    /ws/lots/{id}           → WebSocket: { type: "new_bid" | "lot_extended" | "lot_closed", ... }
GET    /metrics                → Prometheus exposition format
```

### Example: create a lot, bid on it

```bash
TOKEN=$(curl -s -X POST localhost:8080/api/auth/register \
  -H 'content-type: application/json' \
  -d '{"username":"alice","email":"a@x","password":"hunter22"}' | jq -r .token)

LOT=$(curl -s -X POST localhost:8080/api/lots \
  -H "Authorization: Bearer $TOKEN" -H 'content-type: application/json' \
  -d '{"title":"Vintage guitar","start_price":100,"min_step":10,"closing_at":"2099-01-01T00:00:00Z"}' \
  | jq -r .id)

curl -X POST "localhost:8080/api/lots/$LOT/bids" \
  -H "Authorization: Bearer $TOKEN" -H 'content-type: application/json' \
  -d '{"amount":120}'
```

### WebSocket

```js
const ws = new WebSocket(`ws://localhost:8080/ws/lots/${LOT}`);
ws.onmessage = (e) => console.log(JSON.parse(e.data));
// { type: "new_bid",    lot_id, user_id, amount, new_price, timestamp }
// { type: "lot_extended", lot_id, closing_at, extended_count }
// { type: "lot_closed",  lot_id, winner_id, final_price }
```

## Design notes

- **Why optimistic locking instead of `SELECT ... FOR UPDATE` only?**
  Both are used: row lock guarantees we serialize observers of the current
  state inside one DB connection; the `version` column lets us safely
  short-retry from the application layer when two transactions raced on the
  same lot. Anti-sniping makes the retry path real, not theoretical.
- **Why server-rendered templ + HTMX rather than an SPA?**
  Real-time price updates need partial DOM patching, which HTMX and a tiny
  WebSocket listener handle with no JS toolchain. Closer in feel to Hotwire
  Turbo, idiomatic to Go.
- **Timer management.** Each active lot has one `*time.Timer` registered in
  the engine. On bid extension we stop and replace it; on server restart we
  rebuild the map from `GetActiveLots`.

## Roadmap

- Redis: rate-limit bids, cache the active-lots list, distributed lock for
  multi-instance deployments.
- OpenTelemetry tracing + Grafana dashboards shipped in `/deploy`.
- Outbox table + Kafka publisher for downstream consumers.
- testcontainers-go integration tests against a real Postgres.
- Kubernetes manifests (Deployment, Service, HPA, Ingress).

## License

MIT — see [LICENSE](LICENSE).
