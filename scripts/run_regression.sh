#!/bin/bash
set -e

# Install bruno-cli if not present
if ! command -v bru &> /dev/null; then
    echo "Installing bruno-cli..."
    npm install -g @usebruno/cli
fi

# Load .env.local if present and not in CI
if [ -f .env.local ] && [ -z "$CI" ]; then
    echo "Loading .env.local..."
    export $(grep -v '^#' .env.local | xargs)
fi

# Determine BASE_URL
if [ -n "$GCLOUD_SERVICE_STAGING" ]; then
    export BASE_URL="$GCLOUD_SERVICE_STAGING"
elif [ -n "$BASE_URL" ]; then
    export BASE_URL="$BASE_URL"
else
    export BASE_URL="http://localhost:8080" # Default fallback
fi

echo "Running regression tests against: $BASE_URL"

# Run Bruno tests
bru run "docs/api/tests" --env "Staging" --env-var "baseUrl=$BASE_URL" --env-var "apiKey=$API_TEST_KEY"
