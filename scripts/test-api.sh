#!/bin/bash

API_BASE="http://localhost:8080"

echo "Testing MES System APIs..."
echo "================================"

# Test 1: Login
echo -e "\n1. Testing Login API..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "❌ Login failed"
  echo $LOGIN_RESPONSE
  exit 1
else
  echo "✅ Login successful"
  echo "Token: ${TOKEN:0:20}..."
fi

# Test 2: Get User Info
echo -e "\n2. Testing Get User Info API..."
USER_INFO=$(curl -s "$API_BASE/api/v1/auth/user-info" \
  -H "Authorization: Bearer $TOKEN")
echo $USER_INFO | grep -q "admin" && echo "✅ Get user info successful" || echo "❌ Get user info failed"

# Test 3: Get Menus
echo -e "\n3. Testing Get Menus API..."
MENUS=$(curl -s "$API_BASE/api/v1/auth/menus" \
  -H "Authorization: Bearer $TOKEN")
echo $MENUS | grep -q "menu.dashboard" && echo "✅ Get menus successful" || echo "❌ Get menus failed"

# Test 4: List Factories
echo -e "\n4. Testing List Factories API..."
FACTORIES=$(curl -s "$API_BASE/api/v1/base-data/factories?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
echo $FACTORIES | grep -q '"list"' && echo "✅ List factories successful" || echo "❌ List factories failed"

echo -e "\n================================"
echo "API Tests Completed!"
