# Full-stack SaaS Example

This example demonstrates how to use `cogeni` to build a comprehensive code generation pipeline for a full-stack SaaS application.

## Project Structure

```
examples/fullstack/
├── backend/            # Python/FastAPI backend
├── db/                 # Database schema definitions
├── frontend/           # TypeScript/React frontend
├── models.py           # Single source of truth for data models
└── cogeni.lua          # Main generation script
```

## How It Works

1. **Source of Truth**: The `models.py` file defines the core data structures (e.g., `User`, `Subscription`) using Python type hints.
2. **Generation**:
   - `cogeni` reads `models.py` and parses the AST.
   - The `cogeni.lua` script iterates over the classes in `models.py`.
   - **Backend**: Generates Pydantic models in `backend/models.py`.
   - **Database**: Generates SQLAlchemy models in `db/models.py`.
   - **Frontend**: Generates TypeScript interfaces in `frontend/types.ts`.

## Running the Example

1. Navigate to the example directory:
   ```bash
   cd examples/fullstack
   ```

2. Run the generator:
   ```bash
   cogeni run cogeni.lua
   ```

3. Inspect the generated files in `backend/`, `db/`, and `frontend/`.

## Key Concepts

- **Cross-Language Generation**: One Python file generates Python, SQL, and TypeScript code.
- **Unified Logic**: All generation logic is centralized in a single Lua script, making it easy to maintain and update.
- **Dependency Management**: `cogeni` ensures that changes in `models.py` trigger updates across the entire stack.
