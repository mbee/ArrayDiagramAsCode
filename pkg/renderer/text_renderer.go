package renderer

import (
	"diagramgen/pkg/table"
	"strconv"
	"strings"
)

// Render generates a simple text representation of a table.
// This representation is primarily for debugging and basic output,
// not for complex visual alignment.
func Render(t table.Table) string {
	var sb strings.Builder

	// 1. Append table title if it exists
	if t.Title != "" {
		// Using a slightly different prefix to distinguish from parser's "table:"
		sb.WriteString("Rendered Table: " + t.Title + "\n")
	}

	// 2. Iterate over each Row
	for _, row := range t.Rows {
		// 3. For each Cell in the Row
		for i, cell := range row.Cells {
			var cellParts []string

			// Add title if present
			if cell.Title != "" {
				cellParts = append(cellParts, "["+cell.Title+"]")
			}

			// Add content
			cellParts = append(cellParts, cell.Content)

			// Add colspan, always display for clarity, even if 1
			cellParts = append(cellParts, "(cs:"+strconv.Itoa(cell.Colspan)+")")

			// Add rowspan, always display for clarity, even if 1 (as it's part of Cell struct)
			cellParts = append(cellParts, "(rs:"+strconv.Itoa(cell.Rowspan)+")")


			sb.WriteString(strings.Join(cellParts, " "))

			// Append separator if not the last cell
			if i < len(row.Cells)-1 {
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n") // Newline after each row
	}

	return sb.String()
}
