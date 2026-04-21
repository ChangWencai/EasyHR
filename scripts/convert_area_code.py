#!/usr/bin/env python3
"""
Convert MySQL/MariaDB area_code dump to PostgreSQL compatible format.
Usage: python convert_area_code.py <input.sql> <output.sql>
"""
import re
import sys

def convert_mysql_to_pg(sql: str) -> str:
    """Process SQL line by line, handling multi-line CREATE TABLE."""
    lines = sql.split('\n')
    output = []
    i = 0
    in_create = False   # inside CREATE TABLE block
    create_lines = []   # collected CREATE TABLE lines

    def clean_sql_line(line: str) -> str:
        """Apply all PostgreSQL conversions to a single line."""
        # Remove column-level COMMENT '...' from CREATE TABLE
        line = re.sub(r"\s+COMMENT\s+'[^']*'", '', line)
        # Remove ENGINE, CHARSET, COLLATE from table options
        line = re.sub(r'\s*ENGINE\s*=\s*\w+', '', line)
        line = re.sub(r'\s*DEFAULT\s+CHARSET\s*=\s*[\w]+', '', line)
        line = re.sub(r'\s*COLLATE\s*=\s*[\w_]+', '', line)
        # Convert MySQL unsigned int to bigint (PostgreSQL has no unsigned)
        line = re.sub(r'\bbigint unsigned\b', 'bigint', line)
        # Convert tinyint to smallint, strip MySQL length specifiers from integer types
        line = re.sub(r'\btinyint\s*\(\d+\)', 'smallint', line)
        line = re.sub(r'\bint\s*\(\d+\)', 'integer', line)
        # Convert backticks to double quotes for identifiers
        line = re.sub(r'`', '"', line)
        # Convert KEY to INDEX (backtick already converted to quote)
        line = re.sub(r'^(\s*)KEY\s+"(\w+)"\s*\("(\w+)"\)', r'\1INDEX "\2" USING btree ("\3")', line)
        # Add USING btree so column name in parentheses is parsed as column, not type
        line = re.sub(r'^(\s*)INDEX\s+"(\w+)"\s*\("(\w+)"\)', r'\1INDEX "\2" USING btree ("\3")', line)
        return line

    def clean_create_table_line(line: str) -> str:
        """Apply conversions specific to CREATE TABLE column/constraint lines."""
        line = clean_sql_line(line)
        # Rename "name" column to avoid PostgreSQL built-in type conflict
        # Only replace the column name (appears as type specifier after column position)
        line = re.sub(r'(?<=\s)"name"(?=\s+varchar)', '"area_name"', line)
        # Also handle in PRIMARY/unique constraints
        line = re.sub(r'(?<=\s)"name"(?=\s*(?:,|\)))', '"area_name"', line)
        # Add USING btree so column name in parentheses is parsed as column, not type
        line = re.sub(r'^(\s*)INDEX\s+"(\w+)"\s*\("(\w+)"\)', r'\1INDEX "\2" USING btree ("\3")', line)
        return line

    while i < len(lines):
        line = lines[i]
        stripped = line.strip()
        i += 1

        # Skip MySQL conditional comments
        if stripped.startswith('/*!') and '*/' in stripped:
            continue
        if stripped.startswith('/*!'):
            while i < len(lines) and '*/' not in lines[i - 1]:
                i += 1
            continue

        # Skip SET statements
        if stripped.startswith(('SET @OLD_', 'SET TIME_ZONE', 'SET NAMES')):
            continue

        # Skip dump header comments
        if re.match(r"^-- MariaDB dump|^-- Host:|^-- Server version|^-- Dump completed", stripped):
            continue
        if stripped == '--':
            # Blank comment separator — keep only before data section
            if not in_create and not output:
                continue
            output.append('')
            continue

        # Skip LOCK / DISABLE KEYS
        if 'LOCK TABLES' in stripped or ('ALTER TABLE' in stripped and 'DISABLE KEYS' in stripped):
            continue

        # Skip character_set_client assignments
        if 'character_set_client' in stripped:
            continue

        # Data section marker
        if 'Dumping data for table' in stripped:
            output.append('')
            output.append('-- Area code data (2024)')
            output.append('')
            continue

        # Start of CREATE TABLE block
        if 'CREATE TABLE' in stripped and 'area_code' in stripped:
            in_create = True
            create_lines = [line]
            continue

        if in_create:
            create_lines.append(line)
            if stripped.endswith(';'):
                # Collect CREATE INDEX statements separately; clean remaining as table def
                extra_stmts = []
                table_lines = []
                for cl in create_lines:
                    cl_stripped = cl.strip()
                    if cl_stripped.startswith('INDEX ') or cl_stripped.startswith('KEY '):
                        idx_clean = clean_sql_line(cl)
                        idx_name_m = re.search(r'INDEX\s+"(\w+)"', idx_clean)
                        idx_col_m = re.search(r'\("(\w+)"\)', idx_clean)
                        if idx_name_m and idx_col_m:
                            col = idx_col_m.group(1)
                            # The "name" column was renamed to "area_name"
                            if col == 'name':
                                col = 'area_name'
                            extra_stmts.append(
                                f'CREATE INDEX "{idx_name_m.group(1)}" ON "area_code" '
                                f'USING btree ("{col}");'
                            )
                    elif cl_stripped == ')' or cl_stripped == ')' or 'ENGINE=' in cl_stripped:
                        pass  # Drop original closing line; emit our own below
                    else:
                        cleaned = clean_create_table_line(cl)
                        # Skip lines that collapse to just ")" (e.g. ") ENGINE=..." stripped to ")")
                        if cleaned.strip() == ')':
                            continue
                        table_lines.append(cleaned)
                # Close the table definition
                if table_lines:
                    # Fix trailing comma on last constraint line
                    last = table_lines[-1].rstrip()
                    if last.endswith(','):
                        table_lines[-1] = last[:-1]
                    table_lines.append(');')
                output.extend(table_lines)
                output.extend(extra_stmts)
                in_create = False
                create_lines = []
            continue

        # Regular line — clean all non-comment lines (DROP TABLE, etc.)
        if stripped.startswith('INSERT INTO'):
            output.append(re.sub(r'`', '"', line))
        elif stripped.startswith('--'):
            output.append(line)  # SQL comments: preserve text as-is
        else:
            output.append(clean_sql_line(line))

    return '\n'.join(output)

if __name__ == '__main__':
    input_file = sys.argv[1] if len(sys.argv) > 1 else 'area_code_2024.sql'
    output_file = sys.argv[2] if len(sys.argv) > 2 else 'migrations/00001_area_code.sql'

    with open(input_file, 'r', encoding='utf-8') as f:
        content = f.read()

    pg_sql = convert_mysql_to_pg(content)

    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(pg_sql)

    print(f"Converted: {input_file}")
    print(f"Output: {output_file}")
    print(f"Size: {len(pg_sql):,} bytes")
