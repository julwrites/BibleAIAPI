#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting Bible API Service Local Test...${NC}"

# 1. Build the server
echo "Building server..."
mkdir -p bin
go build -o bin/server cmd/server/main.go

# 2. Start the server
echo "Starting server in background..."
# Use a random port or default 8080
PORT=${PORT:-8080}

# Construct API_KEYS if not set, for local fallback
if [ -z "$API_KEYS" ]; then
    if [ -n "$API_KEY" ]; then
        echo -e "${GREEN}Converting legacy API_KEY to API_KEYS JSON format for server...${NC}"
        export API_KEYS="{\"local-test\": \"$API_KEY\"}"
    else
        echo -e "${RED}Warning: API_KEYS/API_KEY is not set. Authentication will be bypassed.${NC}"
    fi
fi

# We use the API_KEY variable for the client request header.
# If API_KEYS was provided but API_KEY wasn't, we need to extract one or warn.
if [ -z "$API_KEY" ] && [ -n "$API_KEYS" ]; then
    echo -e "${RED}Warning: API_KEYS is set but API_KEY (for curl) is not. Using 'ignored' key.${NC}"
fi

PORT=$PORT ./bin/server > server.log 2>&1 &
SERVER_PID=$!

# Function to cleanup server process
cleanup() {
    echo "Stopping server (PID: $SERVER_PID)..."
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null || true
    echo "Server stopped."
}
trap cleanup EXIT

# Wait for server to be ready
echo "Waiting for server to start on port $PORT..."
sleep 5

BASE_URL="http://localhost:$PORT/query"

# Helper function for requests
make_request() {
    local description=$1
    local payload=$2

    echo -e "\n${GREEN}=== Test: $description ===${NC}"
    echo "Payload: $payload"

    response=$(curl -s -X POST "$BASE_URL" \
        -H "Content-Type: application/json" \
        -H "X-API-KEY: ${API_KEY:-ignored}" \
        -d "$payload")

    # Check if response is valid JSON and print it
    if echo "$response" | jq . >/dev/null 2>&1; then
        echo "$response" | jq .
    else
        echo -e "${RED}Invalid JSON response:${NC}"
        echo "$response"
    fi
}

# Test 1: Prose Verse (John 3:16)
# Expecting <p> tags or similar prose formatting
make_request "Prose Verse (John 3:16)" '{"query": {"verses": ["John 3:16"]}, "context": {"user": {"version": "ESV"}}}'

# Test 2: Poetry Verse (Psalm 23:1)
# Expecting <br/> or poetry specific classes
make_request "Poetry Verse (Psalm 23:1)" '{"query": {"verses": ["Psalm 23:1"]}, "context": {"user": {"version": "ESV"}}}'

# Test 3: Word Search (Grace)
make_request "Word Search ('grace')" '{"query": {"words": ["grace"]}, "context": {"user": {"version": "ESV"}}}'

# Test 4: LLM Features (Conditional)
if [ -n "$OPENAI_API_KEY" ] || [ -n "$GEMINI_API_KEY" ] || [ -n "$DEEPSEEK_API_KEY" ]; then
    echo -e "\n${GREEN}=== Test: LLM Open Query (Who is Jesus?) ===${NC}"
    # Note: This requires external internet access and valid keys
    make_request "Open Query" '{"query": {"oquery": "Who is Jesus?"}, "context": {"user": {"version": "ESV"}}}'
else
    echo -e "\n${RED}Skipping LLM tests (No API keys found in environment).${NC}"
fi

echo -e "\n${GREEN}Tests completed. Check server.log for application logs.${NC}"
