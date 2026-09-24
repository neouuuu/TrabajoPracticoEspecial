#!/bin/bash

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