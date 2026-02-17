-- examples/performance/cogeni.lua
-- This script coordinates the processing of the generated stress test files.

-- Find all Python files in the generated source directory
local base_dir = fs.basedir(_CURRENT_FILE)
local src_dir = fs.join(base_dir, "src")

-- Pre-load Python grammar to avoid race conditions during parallel build
print("Pre-loading Python grammar...")
local sample_file = fs.join(src_dir, "file_1.py")
-- We attempt to read the AST of one file. This triggers the grammar download/build
-- in the main thread before we spawn parallel workers.
pcall(function()
	cogeni.read_ast(sample_file, "python")
	print("Python grammar loaded successfully.")
end)

local files = fs.find(src_dir, { type = "f", name = "*.py" })

print("Found generated Python files in " .. src_dir .. ". Processing...")

local count = 0
for path, _ in pairs(files) do
	count = count + 1
	-- Process each file asynchronously
	-- This triggers the embedded <cogeni> blocks inside the Python files.
	-- These blocks perform AST analysis (using tree-sitter) and read companion JSON files.
	async(cogeni.process, path)
end

print(string.format("Scheduled processing for %d files.", count))

-- Schedule a separate async task to aggregate some data to demonstrate parallelism and cross-file logic
async(function()
	print("Starting aggregation task...")
	local configs = fs.find(src_dir, { type = "f", name = "*.json" })
	local total_id_sum = 0
	local file_count = 0

	for path, _ in pairs(configs) do
		local f = io.open(path, "r")
		if f then
			local content = f:read("*a")
			f:close()
			-- Parse JSON content
			local data = json.decode(content)
			if data and data.id then
				total_id_sum = total_id_sum + data.id
			end
			file_count = file_count + 1
		end

		-- Yield periodically to let other tasks run interleaved
		if file_count % 50 == 0 then
			sleep(0.001)
		end
	end
	print(string.format("Aggregation complete: Sum of IDs = %d (from %d files)", total_id_sum, file_count))
end)
