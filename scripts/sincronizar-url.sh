#!/usr/bin/env bash
# Sincroniza la URL del quick tunnel (trycloudflare.com) con el resto del sistema.
#
# Se ejecuta en el servidor desde la carpeta del clon. Si la URL cambió:
#   1. Actualiza PUBLIC_URL en .env.deploy y recrea backend y saas (CORS).
#   2. Actualiza el enlace del README y lo sube con [skip ci] para no
#      disparar otro despliegue.
# Si el servidor no usa el override del quick tunnel, no hace nada.
set -euo pipefail

[ -f docker-compose.override.yml ] || { echo "Sin quick tunnel: nada que sincronizar"; exit 0; }

compose() { docker compose --env-file .env.deploy "$@"; }

URL=""
for intento in $(seq 1 30); do
  URL=$(compose logs cloudflared 2>&1 | grep -o 'https://[a-z0-9-]*\.trycloudflare\.com' | tail -1 || true)
  [ -n "$URL" ] && break
  sleep 2
done
[ -n "$URL" ] || { echo "No se encontró la URL del túnel"; exit 1; }
echo "URL del túnel: $URL"

if ! grep -q "^PUBLIC_URL=$URL$" .env.deploy; then
  sed -i "s|^PUBLIC_URL=.*|PUBLIC_URL=$URL|" .env.deploy
  compose up -d backend saas
  echo "PUBLIC_URL actualizada"
fi

if ! grep -qF "$URL" README.md; then
  sed -i -E "s|https://[a-z0-9-]+\.trycloudflare\.com|$URL|g" README.md
  git add README.md
  git commit -m "README: actualizar el enlace de la demo [skip ci]"
  # Si llegó otro push mientras tanto, se rebasa y se reintenta.
  for intento in 1 2 3; do
    git push origin HEAD:main && break
    git pull --rebase origin main
  done
  echo "README actualizado"
fi

[ -n "${GITHUB_STEP_SUMMARY:-}" ] && echo "URL pública: $URL" >> "$GITHUB_STEP_SUMMARY"
exit 0
