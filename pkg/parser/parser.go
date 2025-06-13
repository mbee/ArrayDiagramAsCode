package parser

import (
	"diagramgen/pkg/table"
	"fmt" // For error wrapping
	"regexp"
	"strconv"
	"strings"
)

// parseCell parses a single cell string into a Cell object.
// It extracts title (e.g., "[My Title]"), colspan (e.g., "::colspan=2::"), and content.
func parseCell(cellStr string) (table.Cell, error) {
	cellStr = strings.TrimSpace(cellStr)
	title := ""
	content := cellStr // Initial content is the full cell string
	colspan := 1

	// 1. Parse Title: [Title] content
	if strings.HasPrefix(content, "[") {
		endIndex := strings.Index(content, "]")
		if endIndex != -1 {
			title = strings.TrimSpace(content[1:endIndex])
			content = strings.TrimSpace(content[endIndex+1:])
		}
		// If no closing ']', it's not a valid title, treat as content.
	}

	// 2. Parse Colspan: ::colspan=N:: content
	// The regex also handles cases where there's no content after colspan directive
	re := regexp.MustCompile(`^::colspan=(\d+)::(.*)`)
	matches := re.FindStringSubmatch(content)

	if len(matches) == 3 { // 0: full match, 1: N (colspan value), 2: rest of content
		parsedColspan, err := strconv.Atoi(matches[1])
		if err == nil && parsedColspan >= 1 {
			colspan = parsedColspan
		}
		// If Atoi fails or parsedColspan < 1, colspan remains 1 (default)
		content = strings.TrimSpace(matches[2]) // Update content to what's after the colspan directive
	}

	// 3. The remaining content is the cell's actual content
	// NewCell initializes Colspan and Rowspan to 1.
	finalCell := table.NewCell(title, content) // table.NewCell(title, content)
	finalCell.Colspan = colspan                // Override Colspan if parsed

	return finalCell, nil
}

// Parse takes a string input and attempts to parse it into a table.Table object.
func Parse(input string) (table.Table, error) {
	var t table.Table
	t.Rows = []table.Row{} // Initialize Rows slice

	trimmedInput := strings.TrimSpace(input)
	lines := strings.Split(trimmedInput, "\n")

	if len(lines) == 0 && trimmedInput == "" { // Handle completely empty or whitespace-only input
		return t, nil
	}
	if len(lines) == 0 { // Should not happen if trimmedInput was not empty, but as a safe guard
	    return t, nil
	}


	// Handle table title: "table: My Table Title"
	firstLineIdx := 0
	if strings.HasPrefix(lines[0], "table:") {
		t.Title = strings.TrimSpace(strings.TrimPrefix(lines[0], "table:"))
		firstLineIdx = 1
		if len(lines) <= 1 { // Only title line exists
			return t, nil
		}
	}

	for i := firstLineIdx; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue // Skip empty lines
		}

		var currentRow table.Row
		// Split by '|'. Standard markdown table lines might start/end with '|'
		cellStrings := strings.Split(trimmedLine, "|")

		// Determine actual cell content based on leading/trailing pipes
		startIdx := 0
		endIdx := len(cellStrings)

		if len(cellStrings) > 0 && cellStrings[0] == "" && strings.HasPrefix(trimmedLine, "|") {
			startIdx = 1
		}
		if len(cellStrings) > 1 && cellStrings[len(cellStrings)-1] == "" && strings.HasSuffix(trimmedLine, "|") {
			endIdx = len(cellStrings) - 1
		}

		actualCellStrings := cellStrings[startIdx:endIdx]

        // If the line was just "||" or "| |", actualCellStrings might be [""] or [" "].
        // If the line was just "|" or "", actualCellStrings might be empty.
		if len(actualCellStrings) == 0 {
		    // If the original line was just "|" or "||" etc. and not empty.
		    // A line like "||" means one cell. A line like "|" means zero cells if it's not "| content |"
		    // if cellStrings was ["", ""] (from "|"), actualCellStrings is empty.
		    // if cellStrings was ["", "", ""] (from "||"), actualCellStrings is [""]
		    if trimmedLine == "|" && len(t.Rows) == 0 && i == firstLineIdx && len(lines) == firstLineIdx+1 {
		        // Special case: if the entire table content is just a single "|"
		        // treat as an empty table rather than a row with no cells.
		        // Or, let it be a row with zero cells, consistent with below.
		        // Current logic: if actualCellStrings is empty, no cells are added,
		        // and if currentRow.Cells remains empty, the row is not added. This is fine.
		    } else if len(cellStrings) > 0 && !(len(cellStrings) == 1 && cellStrings[0] == "") {
                // This condition means the line wasn't empty, but after stripping pipes, no "cells" were found.
                // Example: " | " might become [" "]. "||" becomes [""]
                // If actualCellStrings is empty, it implies a line like "|" or an empty line (already skipped).
                // For a line like `|`, cellStrings is `["", ""]`. startIdx=1, endIdx=1. actualCellStrings is `[]`.
                // This means no cells are added.
            }
		}


		for _, cellStr := range actualCellStrings {
			cell, err := parseCell(cellStr) // cellStr is already trimmed by parseCell's own trimming
			if err != nil {
				return table.Table{}, fmt.Errorf("failed to parse cell '%s' in line '%s': %w", cellStr, trimmedLine, err)
			}
			currentRow.Cells = append(currentRow.Cells, cell)
		}

		// Only add row if it has cells. This handles cases like a line of "----" or " | | " parsed into empty content.
		// Or if a line was `|` which results in no cells.
		if len(currentRow.Cells) > 0 {
			t.Rows = append(t.Rows, currentRow)
		}
	}

	return t, nil
}
