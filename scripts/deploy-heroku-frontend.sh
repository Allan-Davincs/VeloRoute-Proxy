#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! git remote | grep -q '^heroku-frontend$'; then
  heroku git:remote -a veloroute-proxy-frontend -r heroku-frontend
fi

BACKEND_URL="${VITE_API_BASE_URL:-https://veloroute-backend-bd3434ba5cd4.herokuapp.com}"
echo "Using VITE_API_BASE_URL=$BACKEND_URL"
heroku config:set "VITE_API_BASE_URL=$BACKEND_URL" -a veloroute-proxy-frontend

echo "Deploying frontend subtree to Heroku (veloroute-proxy-frontend)..."
git subtree push --prefix frontend heroku-frontend main

echo "Frontend deployed: https://veloroute-proxy-frontend-a5705033a4ef.herokuapp.com"
