-- Helper function to parse a doc block
local function parse_doc_block(lines)
	local doc = {
		module = "",
		func = "",
		summary = "",
		usage = "",
		params = {},
		returns = "",
	}

	for _, line in ipairs(lines) do
		-- Strip leading // or spaces
		line = line:gsub("^%s*//%s*", ""):gsub("^%s*", "")

		if line:find("@module") then
			doc.module = line:match("@module%s+(.+)")
		elseif line:find("@function") then
			doc.func = line:match("@function%s+(.+)")
		elseif line:find("@summary") then
			doc.summary = line:match("@summary%s+(.+)")
		elseif line:find("@usage") then
			doc.usage = line:match("@usage%s+(.+)")
		elseif line:find("@param") then
			local p_name, p_type, p_desc = line:match("@param%s+(%S+)%s+(%S+)%s+(.+)")
			if not p_name then
				p_name, p_type = line:match("@param%s+(%S+)%s+(%S+)")
				p_desc = ""
			end
			if not p_name then
				p_name, p_type, p_desc = line:match("@param%s+(%.%.%.)%s+(%S+)%s+(.+)")
			end

			if p_name then
				table.insert(doc.params, { name = p_name, type = p_type, desc = p_desc })
			else
				-- Fallback
				local rest = line:match("@param%s+(.+)")
				table.insert(doc.params, { name = "...", type = "any", desc = rest })
			end
		elseif line:find("@returns") then
			doc.returns = line:match("@returns%s+(.+)")
		end
	end
	return doc
end

return function(path)
	local docs_by_module = {}

	local ast = cogeni.read_ast(path, "go")
	local comments = jq.query(ast, '.. | select(.type? == "comment")')

	if comments and comments.type == "comment" then
		comments = { comments }
	elseif type(comments) ~= "table" then
		comments = {}
	end

	local in_block = false
	local current_block = {}

	for _, comment in ipairs(comments) do
		local content = comment.content
		if content and content:find("<lua_api>") then
			in_block = true
			current_block = {}
		elseif content and content:find("</lua_api>") then
			in_block = false
			local doc = parse_doc_block(current_block)
			if doc.module ~= "" then
				if not docs_by_module[doc.module] then
					docs_by_module[doc.module] = {}
				end
				table.insert(docs_by_module[doc.module], doc)
			end
		elseif in_block then
			table.insert(current_block, content)
		end
	end

	return docs_by_module
end
