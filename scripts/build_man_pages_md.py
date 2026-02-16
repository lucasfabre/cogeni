import os
import sys
import re

def parse_man_page(file_path):
    """
    Parses a man page file and returns its content as Markdown.
    Preserves indentation for code blocks and lists.
    """
    with open(file_path, 'r') as f:
        content = f.read()

    lines = content.splitlines()
    markdown_lines = []

    current_section = None
    in_code_block = False

    for line in lines:
        # Preserve indentation
        indent = len(line) - len(line.lstrip())
        stripped_line = line.strip()

        # Section Headers (.SH)
        if stripped_line.startswith('.SH'):
            section_title = stripped_line[4:].strip().replace('"', '')
            markdown_lines.append(f"## {section_title}\n")
            current_section = section_title
            continue

        # Sub-Section Headers (.SS)
        if stripped_line.startswith('.SS'):
            section_title = stripped_line[4:].strip().replace('"', '')
            markdown_lines.append(f"### {section_title}\n")
            continue

        # Bold (.B)
        if stripped_line.startswith('.B '):
            text = stripped_line[3:].strip()
            # Escape HTML brackets in bold text
            text = text.replace('<', '&lt;').replace('>', '&gt;')
            # Escape braces that might look like JSX expressions
            text = text.replace('{', '&#123;').replace('}', '&#125;')
            markdown_lines.append(f"**{text}**\n")
            continue

        # Italic (.I)
        if stripped_line.startswith('.I '):
            text = stripped_line[3:].strip()
            # Escape HTML brackets in italic text
            text = text.replace('<', '&lt;').replace('>', '&gt;')
            # Escape braces
            text = text.replace('{', '&#123;').replace('}', '&#125;')
            markdown_lines.append(f"*{text}*\n")
            continue

        # Paragraph (.P, .PP)
        if stripped_line.startswith('.P') or stripped_line.startswith('.PP'):
            markdown_lines.append("\n")
            continue

        # Title Header (.TH) - Skip
        if stripped_line.startswith('.TH'):
            continue

        # Comments (.\")
        if stripped_line.startswith('.\\"'):
            continue

        # Text lines
        if not stripped_line.startswith('.'):
            # Handle some basic inline formatting if possible, e.g., \fBbold\fR
            line_content = line # Keep original line with indentation

            # Escape HTML brackets
            line_content = line_content.replace('<', '&lt;').replace('>', '&gt;')
            # Escape braces
            line_content = line_content.replace('{', '&#123;').replace('}', '&#125;')

            line_content = re.sub(r'\\fB(.*?)\\fR', r'**\1**', line_content)
            line_content = re.sub(r'\\fI(.*?)\\fR', r'*\1*', line_content)

            markdown_lines.append(f"{line_content}\n")

    return "".join(markdown_lines)

def main():
    src_man_dir = "docs/man"
    out_dir = "docs/docusaurus/docs/man"

    if not os.path.exists(out_dir):
        os.makedirs(out_dir)

    # Create category JSON
    with open(os.path.join(out_dir, "_category_.json"), "w") as f:
        f.write('{"label": "Man Pages", "position": 4}')

    for root, dirs, files in os.walk(src_man_dir):
        for file in files:
            if file.endswith((".1", ".7")): # Man page extensions
                src_path = os.path.join(root, file)
                name = os.path.splitext(file)[0]
                section = os.path.splitext(file)[1][1:]

                print(f"Converting {src_path} -> {name}.md")

                markdown_content = parse_man_page(src_path)

                # Add Frontmatter
                frontmatter = f"""---
sidebar_position: {section}
title: {name}({section})
---

# {name}({section})

"""
                out_path = os.path.join(out_dir, f"{name}.md")
                with open(out_path, "w") as f:
                    f.write(frontmatter + markdown_content)

if __name__ == "__main__":
    main()
