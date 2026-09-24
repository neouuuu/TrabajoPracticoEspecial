#!/bin/bash

BASE_URL="http://localhost:8080"

echo "1. CREAR USUARIO (POST /usuarios)"
USUARIO_RESP=$(curl -s -X POST "$BASE_URL/usuarios" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Neo Allende", "email": "neo@example.com"}')
echo "$USUARIO_RESP"

USUARIO_ID=$(echo "$USUARIO_RESP" | sed -n 's/.*"id":\s*\([0-9]*\).*/\1/p')
USUARIO_ID=${USUARIO_ID:-1}
echo "-> ID de usuario creado: $USUARIO_ID"

echo ""

echo "2. CREAR EVENTO (POST /eventos)"
EVENTO_RESP=$(curl -s -X POST "$BASE_URL/eventos" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Recital de Rock", "capacidad": 100, "fecha": "2026-10-15T21:00:00Z"}')
echo "$EVENTO_RESP"

EVENTO_ID=$(echo "$EVENTO_RESP" | sed -n 's/.*"id":\s*\([0-9]*\).*/\1/p')
EVENTO_ID=${EVENTO_ID:-1}
echo "-> ID de evento creado: $EVENTO_ID"

echo ""

echo "3. LISTAR EVENTOS (GET /eventos)"
curl -s -X GET "$BASE_URL/eventos"

echo -e "\n"

echo "4. OBTENER EVENTO POR ID (GET /eventos/{id})"
curl -s -X GET "$BASE_URL/eventos/$EVENTO_ID"

echo -e "\n"

echo "5. EDITAR EVENTO (PUT /eventos/{id})"
curl -s -X PUT "$BASE_URL/eventos/$EVENTO_ID" \
  -H "Content-Type: application/json" \
  -d '{"nombre": "Recital de Rock - Modificado", "capacidad": 150, "fecha": "2026-10-15T22:00:00Z"}'

echo -e "\n"

echo "6. RESERVAR TICKET (POST /tickets)"
TICKET_RESP=$(curl -s -X POST "$BASE_URL/tickets" \
  -H "Content-Type: application/json" \
  -d "{\"id_usuario\": $USUARIO_ID, \"id_evento\": $EVENTO_ID}")
echo "$TICKET_RESP"

TICKET_ID=$(echo "$TICKET_RESP" | sed -n 's/.*"id":\s*\([0-9]*\).*/\1/p')
TICKET_ID=${TICKET_ID:-1}
echo "-> ID de ticket creado: $TICKET_ID"

echo ""

echo "7. OBTENER RESERVAS DE EVENTO (GET /eventos/{id}/reservas)"
curl -s -X GET "$BASE_URL/eventos/$EVENTO_ID/reservas"

echo -e "\n"

echo "8. LISTAR TICKETS POR USUARIO (GET /usuarios/{id}/tickets)"
curl -s -X GET "$BASE_URL/usuarios/$USUARIO_ID/tickets"

echo -e "\n"

echo "9. ELIMINAR TICKET (DELETE /tickets/{id})"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/tickets/$TICKET_ID")
echo "HTTP Status Code: $HTTP_CODE (esperado: 204)"

echo ""

echo "10. VERIFICAR RESERVAS TRAS BORRADO (GET /eventos/{id}/reservas)"
curl -s -X GET "$BASE_URL/eventos/$EVENTO_ID/reservas"