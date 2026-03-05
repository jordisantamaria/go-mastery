#!/bin/bash
# Script para probar todos los endpoints del API Gateway.
# Requiere que los tres servicios esten corriendo.
#
# Uso:
#   ./scripts/test.sh                    # Gateway en localhost:8080
#   ./scripts/test.sh http://mi-host:9090  # Gateway en host personalizado

set -e

BASE_URL="${1:-http://localhost:8080}"

echo "=== Probando Microservices Platform ==="
echo "Gateway: $BASE_URL"
echo ""

# --- Health Check ---
echo "--- Health Check ---"
curl -s "$BASE_URL/health" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/health"
echo ""
echo ""

# --- Crear usuarios ---
echo "--- Crear usuario: Ana ---"
USER1=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Ana Garcia","email":"ana@example.com"}')
echo "$USER1" | python3 -m json.tool 2>/dev/null || echo "$USER1"
USER1_ID=$(echo "$USER1" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")
echo ""

echo "--- Crear usuario: Carlos ---"
USER2=$(curl -s -X POST "$BASE_URL/api/users" \
  -H "Content-Type: application/json" \
  -d '{"name":"Carlos Lopez","email":"carlos@example.com"}')
echo "$USER2" | python3 -m json.tool 2>/dev/null || echo "$USER2"
echo ""

# --- Listar usuarios ---
echo "--- Listar usuarios ---"
curl -s "$BASE_URL/api/users" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/users"
echo ""
echo ""

# --- Obtener usuario por ID ---
if [ -n "$USER1_ID" ]; then
  echo "--- Obtener usuario: $USER1_ID ---"
  curl -s "$BASE_URL/api/users/$USER1_ID" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/users/$USER1_ID"
  echo ""
  echo ""

  # --- Actualizar usuario ---
  echo "--- Actualizar usuario: $USER1_ID ---"
  curl -s -X PUT "$BASE_URL/api/users/$USER1_ID" \
    -H "Content-Type: application/json" \
    -d '{"name":"Ana Garcia Martinez","email":"ana.garcia@example.com"}' | python3 -m json.tool 2>/dev/null
  echo ""
  echo ""

  # --- Crear pedido ---
  echo "--- Crear pedido para usuario: $USER1_ID ---"
  ORDER=$(curl -s -X POST "$BASE_URL/api/orders" \
    -H "Content-Type: application/json" \
    -d "{\"user_id\":\"$USER1_ID\",\"items\":[{\"product_id\":\"prod-1\",\"name\":\"Teclado mecanico\",\"quantity\":1,\"price\":89.99},{\"product_id\":\"prod-2\",\"name\":\"Raton ergonomico\",\"quantity\":2,\"price\":34.99}]}")
  echo "$ORDER" | python3 -m json.tool 2>/dev/null || echo "$ORDER"
  ORDER_ID=$(echo "$ORDER" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null || echo "")
  echo ""

  # --- Obtener pedido ---
  if [ -n "$ORDER_ID" ]; then
    echo "--- Obtener pedido: $ORDER_ID ---"
    curl -s "$BASE_URL/api/orders/$ORDER_ID" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/orders/$ORDER_ID"
    echo ""
    echo ""

    # --- Actualizar estado del pedido ---
    echo "--- Actualizar estado a 'confirmed' ---"
    curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
      -H "Content-Type: application/json" \
      -d '{"status":"confirmed"}' | python3 -m json.tool 2>/dev/null
    echo ""
    echo ""

    echo "--- Actualizar estado a 'shipped' ---"
    curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
      -H "Content-Type: application/json" \
      -d '{"status":"shipped"}' | python3 -m json.tool 2>/dev/null
    echo ""
    echo ""

    echo "--- Actualizar estado a 'delivered' ---"
    curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
      -H "Content-Type: application/json" \
      -d '{"status":"delivered"}' | python3 -m json.tool 2>/dev/null
    echo ""
    echo ""
  fi

  # --- Listar pedidos del usuario ---
  echo "--- Listar pedidos del usuario: $USER1_ID ---"
  curl -s "$BASE_URL/api/users/$USER1_ID/orders" | python3 -m json.tool 2>/dev/null || curl -s "$BASE_URL/api/users/$USER1_ID/orders"
  echo ""
  echo ""

  # --- Eliminar usuario ---
  echo "--- Eliminar usuario: $USER1_ID ---"
  curl -s -X DELETE "$BASE_URL/api/users/$USER1_ID" | python3 -m json.tool 2>/dev/null
  echo ""
  echo ""
fi

# --- Probar error: pedido con usuario inexistente ---
echo "--- Error esperado: pedido con usuario inexistente ---"
curl -s -X POST "$BASE_URL/api/orders" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"no-existe","items":[{"product_id":"p1","name":"Item","quantity":1,"price":10}]}'
echo ""
echo ""

echo "=== Tests completados ==="
