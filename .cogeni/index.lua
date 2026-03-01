-- index.lua

local basedir = ""
if _CURRENT_FILE and _CURRENT_FILE:find("%.cogeni") then
	basedir = fs.basedir(fs.basedir(_CURRENT_FILE)) .. "/"
else
	if _CURRENT_FILE then
		basedir = fs.basedir(_CURRENT_FILE) .. "/"
	end
end

-- Load modules
local parse = loadfile(basedir .. ".cogeni/parser.lua")()
local render_man = loadfile(basedir .. ".cogeni/man.lua")()
local render_md = loadfile(basedir .. ".cogeni/md.lua")()

local function merge_docs(dest, src)
	for mod_name, docs in pairs(src) do
		if not dest[mod_name] then
			dest[mod_name] = {}
		end
		for _, doc in ipairs(docs) do
			table.insert(dest[mod_name], doc)
		end
	end
end

-- Find files
local files = fs.find(basedir .. "src/lua_runtime", { type = "f", name = "*.go" })

-- Parse docs synchronously
local docs_by_module = {}

for path, _ in pairs(files) do
	local result = parse(path)
	if result then
		merge_docs(docs_by_module, result)
	end
end

-- Sort module names
local modules = {}
for mod, _ in pairs(docs_by_module) do
	if mod ~= "global" then
		table.insert(modules, mod)
	end
end
table.sort(modules)

-- Generate outputs
local man_content = render_man(docs_by_module, modules)
cogeni.outfile("lua_man", basedir .. "docs/man/man7/cogeni-lua.7")
write("lua_man", man_content)

local md_content = render_md(docs_by_module, modules)
cogeni.outfile("lua_api_md", basedir .. "docs/docusaurus/docs/lua-api-reference.md")
write("lua_api_md", md_content)
