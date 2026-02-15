print("Starting Full-Stack SaaS Generation...")

local files = {
	"models.py",
	"backend/routes.py",
	"frontend/types.ts",
	"frontend/api.ts",
}

local jobs = {}

function process_file(file)
	print("  Processing " .. file)
	return cogeni.process(file)
end

for _, file in ipairs(files) do
	jobs[#jobs + 1] = async(process_file, file)
end

function generate_sql_models()
	local ast = cogeni.read_ast("models.py", "python")
	cogeni.outfile("sql", "db/schema.sql")
	local sql_query = [[
    [.children[] | select(.type == "class_definition") | select(.fields.name.content != "BaseModel")]
    | map(
        "CREATE TABLE " + (.fields.name.content | ascii_downcase) + "s (\n" +
        ([.fields.body.children[] | select(.type == "assignment")]
        | map(
            "  " + .fields.left.content + " " +
            (if .fields.type.children[0].content == "int" then "INTEGER"
                elif .fields.type.children[0].content == "float" then "REAL"
                elif .fields.type.children[0].content == "bool" then "BOOLEAN"
                else "TEXT" end) +
            (if .fields.left.content == "id" then " PRIMARY KEY" else "" end)
            )
        | join(",\n")
        ) + "\n);"
        )
    | join("\n\n")
    ]]
	write("sql", jq.query(ast, sql_query))
end

jobs[#jobs + 1] = async(generate_sql_models)

for _, job in ipairs(jobs) do
	await(job)
end
