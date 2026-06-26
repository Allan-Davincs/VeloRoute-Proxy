# Deploying VeloRoute (Heroku & Render)

<p align="center">
  <img
    src="https://res.cloudinary.com/ddlegxejs/image/upload/v1782450980/VeloRoute-logo_ywb2qe.png"
    alt="VeloRoute logo"
    width="180"
  />
</p>

Live demo apps (Heroku):

| App | URL |
|-----|-----|
| **Backend** (proxy + admin API + metrics) | https://veloroute-backend-bd3434ba5cd4.herokuapp.com |
| **Frontend** (dashboard) | https://veloroute-proxy-frontend-a5705033a4ef.herokuapp.com |

Try the proxy: `curl https://veloroute-backend-bd3434ba5cd4.herokuapp.com/get`  
Metrics API: `curl https://veloroute-backend-bd3434ba5cd4.herokuapp.com/api/metrics`

---

## How cloud deployment works

Heroku and Render expose **one HTTP port** (`$PORT`). VeloRoute detects `PORT` and runs in **cloud mode**:

| Path | Service |
|------|---------|
| `/` | Reverse proxy (load-balanced traffic) |
| `/api/*` | Admin REST API + dashboard data |
| `/metrics` | Prometheus scrape endpoint |

Cloud config uses public demo backends (`httpbin.org`, `postman-echo.com`, `jsonplaceholder.typicode.com`) — see [`backend/config.heroku.yaml`](backend/config.heroku.yaml).

---

## Deploy to Heroku

### Prerequisites

1. [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli) installed and logged in:
   ```bash
   heroku login
   ```
2. Two Heroku apps created (or use the existing ones):
   ```bash
   heroku apps:create veloroute-backend
   heroku apps:create veloroute-proxy-frontend
   ```
3. Buildpacks (usually auto-detected):
   - Backend: `heroku/go`
   - Frontend: `heroku/nodejs`

### Step 1 — Deploy the backend

From the **repository root**:

```bash
cd VeloRoute-Proxy

# Add Heroku remote (once)
heroku git:remote -a veloroute-backend -r heroku-backend

# Deploy only the backend/ folder
git subtree push --prefix backend heroku-backend main
```

Or use the helper script:

```bash
bash scripts/deploy-heroku-backend.sh
```

Verify:

```bash
curl https://veloroute-backend-bd3434ba5cd4.herokuapp.com/api/metrics
```

### Step 2 — Deploy the frontend

Set the API URL to your **backend** Heroku URL (required at build time for Vite):

```bash
heroku config:set \
  VITE_API_BASE_URL=https://veloroute-backend-bd3434ba5cd4.herokuapp.com \
  -a veloroute-proxy-frontend
```

Deploy the frontend subtree:

```bash
heroku git:remote -a veloroute-proxy-frontend -r heroku-frontend
git subtree push --prefix frontend heroku-frontend main
```

Or:

```bash
bash scripts/deploy-heroku-frontend.sh
```

Open the dashboard: https://veloroute-proxy-frontend-a5705033a4ef.herokuapp.com

### Heroku one-liner reference

```bash
# Backend
cd VeloRoute-Proxy
git init   # if not already a git repo
heroku git:remote -a veloroute-backend -r heroku-backend
git subtree push --prefix backend heroku-backend main

# Frontend (after backend is live)
heroku config:set VITE_API_BASE_URL=https://veloroute-backend-bd3434ba5cd4.herokuapp.com -a veloroute-proxy-frontend
heroku git:remote -a veloroute-proxy-frontend -r heroku-frontend
git subtree push --prefix frontend heroku-frontend main
```

### Scale dynos (required on new apps)

```bash
heroku ps:scale web=1 -a veloroute-backend
heroku ps:scale web=1 -a veloroute-proxy-frontend
```

### View logs

```bash
heroku logs --tail -a veloroute-backend
heroku logs --tail -a veloroute-proxy-frontend
```

---

## Deploy to Render

1. Push this repo to GitHub.
2. In [Render Dashboard](https://dashboard.render.com) → **New** → **Blueprint**.
3. Connect the repo — Render reads [`render.yaml`](render.yaml).
4. After the backend deploys, set on the **frontend** service:
   ```
   VITE_API_BASE_URL=https://<your-backend-service>.onrender.com
   ```
5. Trigger a manual redeploy of the frontend so Vite picks up the variable.

### Manual Render setup (without Blueprint)

**Backend**

| Setting | Value |
|---------|-------|
| Root Directory | `backend` |
| Build Command | `go build -o bin/veloroute ./cmd/veloroute` |
| Start Command | `./bin/veloroute --config config.heroku.yaml` |
| Health Check | `/api/metrics` |

**Frontend**

| Setting | Value |
|---------|-------|
| Root Directory | `frontend` |
| Build Command | `npm install && npm run build` |
| Start Command | `npm start` |
| Env `VITE_API_BASE_URL` | Backend URL |

---

## Environment variables

| Variable | App | Description |
|----------|-----|-------------|
| `PORT` | Backend | Set automatically by Heroku/Render |
| `VELOROUTE_CONFIG` | Backend | Default `config.heroku.yaml` in cloud |
| `VITE_API_BASE_URL` | Frontend | Backend URL (e.g. `https://veloroute-backend-....herokuapp.com`) |

---

## Troubleshooting

| Issue | Fix |
|-------|-----|
| Dashboard shows API errors | Ensure `VITE_API_BASE_URL` matches backend URL; redeploy frontend after changing it |
| `503` on proxy | Demo backends may be slow; check `/api/backends` for alive status |
| Heroku build fails (Go) | Ensure `go.mod` is inside `backend/` when using subtree push |
| CORS errors | Admin API sends `Access-Control-Allow-Origin: *` — backend URL must be correct |
| SSE log feed disconnected | Heroku free/hobby dynos sleep; upgrade dyno or use Render with always-on |

---

## Security note

The admin API has **no authentication** in this demo. Do not expose production backends without adding auth and TLS.
