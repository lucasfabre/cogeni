local function render_md_details(md_output, doc)
	if #doc.params > 0 then
		table.insert(md_output, "\n**Parameters:**\n")
		for _, p in ipairs(doc.params) do
			local desc = p.desc or ""
			table.insert(md_output, "- `" .. p.name .. "` (" .. p.type .. "): " .. desc)
		end
		table.insert(md_output, "")
	end

	if doc.returns and doc.returns ~= "" then
		table.insert(md_output, "\n**Returns:**\n\n- " .. doc.returns .. "\n")
	end
end

return function(docs_by_module, modules)
	local md_output = {}
	table.insert(
		md_output,
		[[---
sidebar_position: 8
id: lua-api-reference
title: Lua API Reference
---

# Lua API Reference

The `cogeni` environment provides a suite of modules designed for AST-driven code generation.

## Global Variables

- `_CURRENT_FILE`: Absolute path to the currently executing Lua script.
- `_FILE_EXTENSION`: Extension of the currently executing script (including the dot).
]]
	)

	if docs_by_module["global"] then
		table.insert(md_output, "## Global Functions\n")

		for _, doc in ipairs(docs_by_module["global"]) do
			local func_name = doc.usage and doc.usage or doc.func
			table.insert(md_output, "### `" .. func_name .. "`\n")
			if doc.summary and doc.summary ~= "" then
				table.insert(md_output, doc.summary)
			end
			render_md_details(md_output, doc)
		end
	end

	table.insert(md_output, "## Modules\n")

	for _, mod in ipairs(modules) do
		table.insert(md_output, "### " .. mod .. "\n")
		local funcs = docs_by_module[mod]

		for _, doc in ipairs(funcs) do
			local func_name = doc.usage and doc.usage or (mod .. "." .. doc.func)
			table.insert(md_output, "#### `" .. func_name .. "`\n")
			if doc.summary then
				table.insert(md_output, doc.summary)
			end
			render_md_details(md_output, doc)
		end
	end

	return table.concat(md_output, "\n")
end
