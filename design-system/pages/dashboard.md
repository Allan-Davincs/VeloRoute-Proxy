# Dashboard Page Overrides

## Layout
1. Header: VeloRoute logo, algorithm badge, connection status
2. Metrics row: 4 cards (Total Requests, Req/sec, Error Rate, P95 Latency)
3. Full-width Requests/sec chart (60s window)
4. Split row: Latency chart (P50/P95/P99) | Backend table
5. Full-width live access log feed (SSE)

## Chart Colors
- Requests line: `#6366f1` (indigo)
- P50 latency: `#22c55e` (green)
- P95 latency: `#f59e0b` (amber)
- P99 latency: `#ef4444` (red)

## Thresholds
- Error rate: green < 0.5%, red > 1%
- P95 latency: amber > 200ms, red > 500ms

## Log Feed
- Monospace font
- Method badges: GET=indigo, POST=green, PUT=amber, DELETE=red
- Status badges: 2xx=green, 3xx=blue, 4xx=amber, 5xx=red
- Auto-scroll unless user scrolled up
