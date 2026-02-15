#!/usr/bin/env bash
set -e

# ANSI color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}Starting End-to-End Full-Stack Verification${NC}"

# 0. Setup UV if missing
if ! command -v uv >/dev/null 2>&1; then
	if [[ ! -f "$HOME/.local/bin/uv" ]]; then
		echo "UV not found. Installing UV locally..."
		curl -LsSf https://astral.sh/uv/install.sh | sh
	fi
	export PATH="$HOME/.local/bin:$PATH"
fi

# 2. Test SQL Schema with SQLite
echo "Step 2: Testing SQL schema..."
rm -f test.db
if command -v sqlite3 >/dev/null 2>&1; then
	sqlite3 test.db <db/schema.sql
	echo -e "  ${GREEN}[OK]${NC} SQLite schema verified"
else
	echo "  [SKIP] sqlite3 missing"
fi

# 3. Start Backend Server
echo "Step 3: Starting FastAPI server with UV..."
# Create a dummy main.py to run the app
cat >main.py <<EOF
import uvicorn
from backend.routes import app

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
EOF

# Use UV to run the server with dependencies
# We set PYTHONPATH to . so it can find models.py and backend/
export PYTHONPATH=.
uv run --with fastapi --with uvicorn --with pydantic python3 main.py &
SERVER_PID=$!

# Wait for server to be ready
echo "Waiting for server to start..."
MAX_RETRIES=10
COUNT=0
while ! curl -s http://localhost:8000/todos >/dev/null; do
	sleep 1
	COUNT=$((COUNT + 1))
	if [ $COUNT -ge $MAX_RETRIES ]; then
		echo -e "${RED}Server failed to start${NC}"
		kill $SERVER_PID || true
		exit 1
	fi
done
echo -e "  ${GREEN}[OK]${NC} Server is up"

# 4. Run TS SDK tests
echo "Step 4: Running TS SDK tests..."
if npx --yes tsx frontend/test_sdk.ts; then
	echo -e "  ${GREEN}[OK]${NC} SDK tests passed"
else
	echo -e "  ${RED}[FAIL]${NC} SDK tests failed"
	kill $SERVER_PID || true
	exit 1
fi

# Cleanup
echo "Cleaning up..."
kill $SERVER_PID || true
rm -f main.py test.db
rm -rf .venv

echo -e "\n${GREEN}End-to-End verification successful!${NC}"
