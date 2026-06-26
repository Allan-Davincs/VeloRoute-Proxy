#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! git remote | grep -q '^heroku-backend$'; then
  heroku git:remote -a veloroute-backend -r heroku-backend
fi

echo "Deploying backend subtree to Heroku (veloroute-backend)..."
git subtree push --prefix backend heroku-backend main

echo "Backend deployed: https://veloroute-backend-bd3434ba5cd4.herokuapp.com"
echo "Set frontend API URL:"
echo "  heroku config:set VITE_API_BASE_URL=https://veloroute-backend-bd3434ba5cd4.herokuapp.com -a veloroute-proxy-frontend"
