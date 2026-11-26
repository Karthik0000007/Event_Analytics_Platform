# web: Event Analytics dashboard

React 18 + TypeScript + Vite dashboard for the Event Analytics Platform.
It reads from the Go ingestion API (`/v1/...`) and shows events, analytics and dead letters.

## Run with Docker (recommended)

From the repo root:

```bash
docker compose up --build
```

The dashboard is served by nginx at http://localhost:3000.
nginx proxies `/v1`, `/healthz` and `/readyz` to the `api` service.

## Run for development

The Go API must be running on port 8080 (for example from `docker compose up`).

```bash
cd web
npm install
npm run dev      # http://localhost:3000, proxies API calls to localhost:8080
npm run build    # type-check and build into dist/
```

## Structure

- `src/pages/`: Dashboard, Events, EventDetail, Analytics, DLQ
- `src/components/`: layout, tables, charts, filters, toast
- `src/services/api.ts`: typed fetch client for the API
- `src/types/`: TypeScript types that mirror the API responses
- `Dockerfile` and `nginx.conf`: production image (build with node, serve with nginx)
