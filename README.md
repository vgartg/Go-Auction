# GoAuction – High‑performance auction engine

Real‑time auction engine on Go with WebSocket, optimistic locking, and Prometheus metrics.

## Features

- REST API for lots and bidding
- Optimistic locking for concurrent bids
- Automatic lot closing with timers (restored after restart)
- WebSocket real‑time updates (`new_bid`, `lot_closed`)
- Prometheus metrics endpoint `/metrics`
- Docker & docker‑compose ready
- GitHub Actions CI

## Quick start

```bash
git clone https://github.com/vgartg/goauction.git
cd goauction
docker-compose up --build
```

API runs at http://localhost:8080

## Example requests
### Create a lot:

```bash
curl -X POST http://localhost:8080/api/lots \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Vintage Guitar",
    "start_price": 100.0,
    "min_step": 10.0,
    "closing_at": "2026-12-31T23:59:59Z"
  }'
```


Place a bid:
```bash
curl -X POST http://localhost:8080/api/lots/{lot_id}/bids \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "amount": 120.0}'
```

WebSocket:
```javascript
const ws = new WebSocket("ws://localhost:8080/ws/lots/{lot_id}");
ws.onmessage = (event) => console.log(JSON.parse(event.data));
```
