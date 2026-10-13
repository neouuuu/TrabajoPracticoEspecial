#!/bin/bash
set -u
cd "$(dirname "$0")"
 
# Código generado por sqlc (db/sqlc está en el .gitignore y el build de Docker lo necesita)
if [ ! -d db/sqlc ]; then
  SQLC="$(command -v sqlc || true)"
  [ -z "$SQLC" ] && [ -x "$HOME/go/bin/sqlc" ] && SQLC="$HOME/go/bin/sqlc"
  if [ -z "$SQLC" ]; then
    echo "ERROR: falta db/sqlc y no encuentro sqlc. Instalalo con:"
    echo "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
    exit 1
  fi
  echo "Generando código con sqlc..."
  "$SQLC" generate
fi
 
cleanup() {
  docker compose down -v
}
 
# Garantiza que docker compose down se ejecute al salir (exito o error)
trap cleanup EXIT
 
echo "Levantando servicios..."
docker compose up -d --build
 
echo "Esperando que la API esté lista..."
until curl -s http://localhost:8080/eventos > /dev/null; do
  sleep 1
done
 
echo "Ejecutando pruebas de integración..."
./requests.sh