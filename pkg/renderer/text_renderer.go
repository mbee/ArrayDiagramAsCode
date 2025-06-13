package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	// "strconv" // No longer directly used
	"strings"
)

// Render generates a simple text representation of a table, including styling info.
func Render(t table.Table) string {
	var sb strings.Builder

	// Display Table Title
	if t.Title != "" {
		// Changed prefix from "Rendered Table:" to "Table:" to match example
		sb.WriteString(fmt.Sprintf("Table: %s\n", t.Title))
	}

	// Display Global Settings
	// Forcing display for debugging purposes.
	// A more sophisticated check could compare t.Settings against table.DefaultGlobalSettings()
	// to show only non-default values.
	sb.WriteString(fmt.Sprintf("  Settings: EdgeColor: '%s', EdgeThickness: %d, DefaultCellBG: '%s', TableBG: '%s'\n",
		t.Settings.EdgeColor,
		t.Settings.EdgeThickness,
		t.Settings.DefaultCellBackgroundColor,
		t.Settings.TableBackgroundColor,
	))

	for _, row := range t.Rows {
		for i, cell := range row.Cells {
			var cellParts []string

			// Add title if present
			if cell.Title != "" {
				cellParts = append(cellParts, "["+cell.Title+"]")
			}

			// Add content
			// Ensure content is added even if empty, to maintain structure if title is also empty.
			cellParts = append(cellParts, cell.Content)

			// Add attributes: Colspan, Rowspan
			cellParts = append(cellParts, fmt.Sprintf("(cs:%d)", cell.Colspan))
			cellParts = append(cellParts, fmt.Sprintf("(rs:%d)", cell.Rowspan))

			// Add background color if specified for the cell
			if cell.BackgroundColor != "" {
				cellParts = append(cellParts, fmt.Sprintf("{bg:%s}", cell.BackgroundColor))
			}

			// Join parts with space. Filter out empty strings (like empty content if title is also empty)
			// to avoid multiple spaces, unless content itself is a space.
			var finalCellParts []string
			for _, part := range cellParts {
			    if part != "" || (len(finalCellParts) > 0 && finalCellParts[len(finalCellParts)-1] != "") {
			        // Add part if it's not empty.
			        // Or if it IS empty, but the previous part was not empty (to allow one space for empty content).
			        // This logic is tricky. Let's simplify: join and then clean up multiple spaces.
			        finalCellParts = append(finalCellParts, part)
			    }
			}

            // Simpler approach for joining: Let strings.Join handle it.
            // Then replace multiple spaces with single spaces if needed.
            // For now, direct join:
			// joinedString := strings.Join(finalCellParts, " ") // This was unused

			// The original logic for cellParts construction in previous versions was:
			// cellParts = append(cellParts, "["+cell.Title+"]")
			// cellParts = append(cellParts, cell.Content)
			// cellParts = append(cellParts, "(cs:"+strconv.Itoa(cell.Colspan)+")")
			// cellParts = append(cellParts, "(rs:"+strconv.Itoa(cell.Rowspan)+")")
			// if cell.BackgroundColor != "" { cellParts = append(cellParts, "{bg:"+cell.BackgroundColor+"}")}
			// This would result in "Title Content (cs:1) (rs:1) {bg:#fff}"
			// Or if content is empty: "Title  (cs:1) (rs:1)" (double space)
			// The provided code in the prompt was:
			// cellContentBuilder.WriteString("[" + cell.Title + "] ")
			// cellContentBuilder.WriteString(cell.Content)
			// cellContentBuilder.WriteString(fmt.Sprintf(" (cs:%d)", cell.Colspan)) ... etc.
			// This also leads to double spaces if title is present and content is empty.
			// Let's stick to the prompt's direct formatting approach.

			var cellContentBuilder strings.Builder
			hasPreviousPart := false
			if cell.Title != "" {
				cellContentBuilder.WriteString("[" + cell.Title + "]")
				hasPreviousPart = true
			}

			if cell.Content != "" {
				if hasPreviousPart {
					cellContentBuilder.WriteString(" ")
				}
				cellContentBuilder.WriteString(cell.Content)
				hasPreviousPart = true
			} else if cell.Title != "" { // Content is empty but title was there
			    // To avoid "Title(cs:1)" vs "Title (cs:1)"
			    // If there's no content, ensure a space before attributes if title was present.
			    cellContentBuilder.WriteString(" ")
			}


			// Attributes always have a leading space in their format string.
			cellContentBuilder.WriteString(fmt.Sprintf(" (cs:%d)", cell.Colspan))
			cellContentBuilder.WriteString(fmt.Sprintf(" (rs:%d)", cell.Rowspan))

			if cell.BackgroundColor != "" {
				cellContentBuilder.WriteString(fmt.Sprintf(" {bg:%s}", cell.BackgroundColor))
			}

			sb.WriteString(cellContentBuilder.String())

			if i < len(row.Cells)-1 {
				sb.WriteString(" | ")
			}
		}
		sb.WriteString("\n") // Newline after each row
	}

	return sb.String()
}
