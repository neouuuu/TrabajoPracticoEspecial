#!/usr/bin/env bash
# Levanta solo el contenedor de Postgres y corre main.go en tu máquina
# (sin compilar el Go dentro de un contenedor).
#
# Uso:
#   ./dev.sh           # levanta la BD (conserva los datos) y corre el servidor
#   ./dev.sh --clean   # borra el contenedor y arranca con la BD vacía (re-ejecuta schema.sql)
#
# Al cortar con Ctrl+C se detiene el contenedor (los datos se conservan hasta el próximo --clean).

set -euo pipefail
cd "$(dirname "$0")"

CONTENEDOR="server_postgres"   # container_name del docker-compose.yml
SERVICIO="db"        # nombre del servicio en el docker-compose.yml
DB_USER="root"

# ---------- Chequeos ----------
command -v docker >/dev/null || { echo "ERROR: no encuentro docker."; exit 1; }
command -v go >/dev/null     || { echo "ERROR: no encuentro go."; exit 1; }

# ---------- Código generado por sqlc (db/sqlc está en el .gitignore) ----------
if [ ! -d db/sqlc ]; then
    SQLC="$(command -v sqlc || true)"
    [ -z "$SQLC" ] && [ -x "$HOME/go/bin/sqlc" ] && SQLC="$HOME/go/bin/sqlc"
    if [ -z "$SQLC" ]; then
        echo "ERROR: falta db/sqlc y no encuentro sqlc. Instalalo con:"
        echo "  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"
        exit 1
    fi
    echo "=== Generando código con sqlc ==="
    "$SQLC" generate
fi

# ---------- Base de datos ----------
if [ "${1:-}" = "--clean" ]; then
    echo "=== Limpiando contenedor y datos previos ==="
    docker compose down -v --remove-orphans
fi

echo "=== Levantando PostgreSQL ==="
docker compose up -d "$SERVICIO"

echo "=== Esperando a que Postgres acepte conexiones ==="
until docker exec "$CONTENEDOR" pg_isready -U "$DB_USER" >/dev/null 2>&1; do
    echo "  Postgres se está inicializando..."
    sleep 1
done

# Al salir (Ctrl+C o error) frenamos el contenedor
trap 'echo; echo "=== Deteniendo PostgreSQL ==="; docker compose stop "$SERVICIO"' EXIT

# ---------- Servidor ----------
echo "=== Iniciando servidor en http://localhost:8080 ==="
go run .