# Creating Your First Generator

This guide will walk you through creating a simple code generator using `cogeni`. We'll create a Python script that defines a data model, and use `cogeni` to automatically generate a corresponding JSON schema.

## Prerequisites

- `cogeni` installed and available in your PATH.
- A text editor.

## Step 1: Define the Source Model

Create a file named `models.py` with the following content:

```python
class User:
    name: str
    age: int
    email: str
```

## Step 2: Create the Generator Script

Create a file named `generate_schema.lua` in the same directory. This script will read `models.py`, extract the class definition, and generate a JSON schema.

```lua
-- Read the AST of models.py
local ast = cogeni.read_ast("models.py")

-- Find the class definition using JQ (or manual traversal)
-- Here we assume a simple structure for demonstration
local class_node = jq.query(ast, ".root.children[] | select(.type == \"class_definition\")")
local class_name = jq.query(class_node, ".children[] | select(.type == \"identifier\") | .text")

-- Extract fields
local fields = {}
-- Note: Real AST traversal would be more robust. This is a simplified example.
-- We iterate over the body of the class
local body = jq.query(class_node, ".children[] | select(.type == \"block\")")
-- ... implementation details for field extraction would go here ...

-- For simplicity, let's hardcode the extraction based on known structure
local schema = {
    title = class_name,
    type = "object",
    properties = {
        name = { type = "string" },
        age = { type = "integer" },
        email = { type = "string" }
    }
}

-- Write the schema to a file
cogeni.outfile("schema", "schema.json")
write("schema", json.encode(schema))
```

## Step 3: Run the Generator

Execute the following command in your terminal:

```bash
cogeni run generate_schema.lua
```

## Step 4: Verify the Output

Check the generated `schema.json` file. It should contain the JSON schema corresponding to your Python class.

```json
{"title":"User","type":"object","properties":{"name":{"type":"string"},"age":{"type":"integer"},"email":{"type":"string"}}}
```

## Next Steps

- Explore the `cogeni.read_ast` documentation to learn more about AST structure.
- Use `cogeni ast models.py` to inspect the raw AST and refine your JQ queries.
