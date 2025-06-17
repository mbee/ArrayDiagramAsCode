#!/bin/bash

# Script to generate PNG images for README examples

# Ensure diagramgen can be run. Using go run for now.
# If you have a binary, change this to: DIAGRAMGEN_CMD="./path/to/diagramgen_binary"
DIAGRAMGEN_CMD="go run cmd/diagramgen/main.go"
INPUT_FILE="examples_for_readme.txt"
OUTPUT_DIR="doc/images"
VERBOSE_GEN="--verbose" # or "" to disable

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Temporary file to hold segments of the input
TEMP_INPUT_FILE="_temp_diag_input.txt"

# Counter for unique filenames if needed
COUNT=0

# Process sections based on 'main_table:' directives found in the input file
grep -E "^main_table:" "$INPUT_FILE" | while read -r main_table_line ; do
    # Extract the table ID from the main_table_line (e.g., "main_table: [my-table-id]")
    # Remove "main_table: [" prefix and "]" suffix
    TABLE_ID=$(echo "$main_table_line" | sed -e 's/main_table: \[\([^]]*\)\]/\1/')

    if [ -z "$TABLE_ID" ]; then
        echo "Warning: Could not extract TABLE_ID from line: $main_table_line"
        continue
    fi

    # Sanitize TABLE_ID for use in filename (replace non-alphanumeric with underscore)
    FILENAME_BASE=$(echo "$TABLE_ID" | sed 's/[^a-zA-Z0-9_-]/_/g')
    OUTPUT_PNG="${OUTPUT_DIR}/${FILENAME_BASE}.png"

    echo "Preparing to generate $OUTPUT_PNG for main table: $TABLE_ID..."

    # Create temporary input file for diagramgen
    # It contains the current main_table directive and all table definitions from the original file (excluding other main_table lines)
    {
        echo "$main_table_line"
        grep -v "^main_table:" "$INPUT_FILE"
    } > "$TEMP_INPUT_FILE"

    # Run diagramgen
    echo "Running: $DIAGRAMGEN_CMD -i $TEMP_INPUT_FILE -o $OUTPUT_PNG $VERBOSE_GEN"
    $DIAGRAMGEN_CMD -i "$TEMP_INPUT_FILE" -o "$OUTPUT_PNG" $VERBOSE_GEN

    if [ $? -eq 0 ]; then
        echo "Successfully generated $OUTPUT_PNG"
    else
        echo "Error generating $OUTPUT_PNG"
    fi
    echo "---"
    COUNT=$((COUNT + 1))
done

# Clean up temporary file
if [ -f "$TEMP_INPUT_FILE" ]; then
    rm "$TEMP_INPUT_FILE"
fi

echo "Image generation process complete."
echo "Total diagrams processed: $COUNT"
echo "Please check the '$OUTPUT_DIR' directory."

# Make the script executable by the subtask runner if it's not already.
# The subtask itself will ensure this by saving it and then running chmod.
