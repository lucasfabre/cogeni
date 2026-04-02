---
sidebar_position: 2
title: "First Generator"
---

# Creating Your First Generator

This guide walks through a simple artifact-generation workflow with `cogeni`. We will define a Python model and generate a JSON schema from it. The same pattern applies to docs, CLIs, SDKs, and other derived artifacts.

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

Create a file named `cogeni.lua` in the same directory. This script reads `models.py`, extracts the class definition, and writes a derived artifact: `schema.json`.

```lua
-- Read the AST of models.py
-- cogeni detects the language automatically from the .py extension
local ast = cogeni.read_ast("models.py")

-- Find the class definition and its name using JQ
-- We query the root children and use named fields
local class_node = jq.query(ast, ".children[] | select(.type == \"class_definition\")")
local class_name = jq.query(class_node, ".fields.name.content")

-- Map Python types to JSON Schema types
local type_mapping = {
    ["str"] = "string",
    ["int"] = "integer",
    ["float"] = "number",
    ["bool"] = "boolean"
}

-- Extract fields by iterating over the body of the class
local properties = {}
local body_children = jq.query(class_node, ".fields.body.children[]")

-- JQ returns a single node if there's only one child, or a list.
-- We ensure we're working with a list for the iterator.
if body_children and body_children.type then body_children = { body_children } end

for _, statement in ipairs(body_children or {}) do
    -- We only care about assignments (which include type annotations in Python)
    if statement.type == "assignment" then
        -- Use named fields: 'left' for the variable and 'type' for its annotation
        local field_name = jq.query(statement, ".fields.left.content")
        -- The type field is a wrapper node, its first child is the actual type identifier
        local field_type = jq.query(statement, ".fields.type.children[0].content")

        if field_name and field_type then
            properties[field_name] = { type = type_mapping[field_type] or "string" }
        end
    end
end

local schema = {
    title = class_name,
    type = "object",
    properties = properties
}

-- Write the schema to a file with indentation for readability
cogeni.outfile("schema", "schema.json")
write("schema", json.encode(schema, { indent = true }))
```

## Step 3: Run the Generator

Execute the following command in your terminal:

```bash
cogeni
```

## Step 4: Verify the Output

Check the generated `schema.json` file. It should contain a JSON schema derived from your Python class.

```json
{
  "properties": {
    "age": {
      "type": "integer"
    },
    "email": {
      "type": "string"
    },
    "name": {
      "type": "string"
    }
  },
  "title": "User",
  "type": "object"
}
```

## Next Steps

- Explore the [Lua API Reference](../lua-api-reference) to learn more about available modules.
- Use `cogeni ast models.py` to inspect the raw AST and refine your JQ queries.
- Adapt the same workflow to generate docs, CLIs, or other project-specific artifacts.
