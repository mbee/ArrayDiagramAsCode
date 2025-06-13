package parser

import (
	"diagramgen/pkg/table"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Parse takes a string input and attempts to parse it into a Table object.
// It now also parses global table settings from the title line and
// rowspan and background color from individual cells.
func Parse(input string) (table.Table, error) {
	// Initialize table with default settings. These can be overridden by parsed settings.
	t := table.Table{
		Settings: table.DefaultGlobalSettings(),
		Rows:     []table.Row{}, // Ensure Rows is initialized
	}

	trimmedInput := strings.TrimSpace(input)
	if trimmedInput == "" {
		return t, nil // Return empty table with defaults if input is empty
	}

	lines := strings.Split(trimmedInput, "\n")
	firstLineIdx := 0

	// Parse table title and global settings from the first line if present
	if strings.HasPrefix(lines[0], "table:") {
		titleAndSettingsLine := strings.TrimPrefix(lines[0], "table:")
		firstLineIdx = 1

		// Regex to separate title from settings block (e.g., "My Title {setting:value}")
		// Title part is (.*?), settings part is (.*) within {}.
		settingsRegex := regexp.MustCompile(`^(.*?)\s*\{(.*)\}\s*$`)
		settingsMatch := settingsRegex.FindStringSubmatch(strings.TrimSpace(titleAndSettingsLine))

		if len(settingsMatch) == 3 { // 0: full match, 1: title part, 2: settings string
			t.Title = strings.TrimSpace(settingsMatch[1])
			settingsStr := settingsMatch[2]
			if err := parseGlobalSettings(settingsStr, &t.Settings); err != nil {
				return table.Table{}, fmt.Errorf("failed to parse global settings '%s': %w", settingsStr, err)
			}
		} else {
			// No settings block found, the whole string is the title
			t.Title = strings.TrimSpace(titleAndSettingsLine)
		}

		if len(lines) <= 1 { // Only title/settings line, no rows
			return t, nil
		}
	}

	// Process actual table rows
	for i := firstLineIdx; i < len(lines); i++ {
		line := lines[i]
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "" {
			continue // Skip empty lines
		}

		var currentRow table.Row
		cellStrings := strings.Split(trimmedLine, "|")

		startIdx := 0
		endIdx := len(cellStrings)
		if len(cellStrings) > 0 && cellStrings[0] == "" && strings.HasPrefix(trimmedLine, "|") {
			startIdx = 1
		}
		if len(cellStrings) > 0 && cellStrings[len(cellStrings)-1] == "" && strings.HasSuffix(trimmedLine, "|") {
			// Ensure endIdx doesn't go below startIdx for single empty elements like from "|"
			if endIdx > startIdx {
				endIdx = len(cellStrings) - 1
			}
		}

		actualCellStrings := cellStrings[startIdx:endIdx]

		if len(actualCellStrings) == 0 && trimmedLine != "" {
		    if strings.ReplaceAll(trimmedLine, "|", "") == "" {
		        for j := 0; j < len(cellStrings)-1 ; j++ {
		             cell, _ := parseCell("")
		             currentRow.Cells = append(currentRow.Cells, cell)
		        }
		    }
		} else {
			for _, cellStr := range actualCellStrings {
				cell, err := parseCell(cellStr)
				if err != nil {
					return table.Table{}, fmt.Errorf("failed to parse cell '%s' in line '%s': %w", cellStr, trimmedLine, err)
				}
				currentRow.Cells = append(currentRow.Cells, cell)
			}
		}


		if len(currentRow.Cells) > 0 {
			t.Rows = append(t.Rows, currentRow)
		}
	}

	return t, nil
}

// parseGlobalSettings parses key-value pairs for global table settings.
func parseGlobalSettings(settingsStr string, settings *table.GlobalSettings) error {
	pairs := strings.Split(settingsStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "bg_table":
			settings.TableBackgroundColor = value
		case "bg_cell":
			settings.DefaultCellBackgroundColor = value
		case "edge_color":
			settings.EdgeColor = value
		case "edge_thickness":
			thickness, err := strconv.Atoi(value)
			if err != nil {
				return fmt.Errorf("invalid edge_thickness value '%s': %w", value, err)
			}
			if thickness < 0 {
				return fmt.Errorf("edge_thickness must be non-negative, got %d", thickness)
			}
			settings.EdgeThickness = thickness
		}
	}
	return nil
}

// parseCell refines the parsing of individual cell strings to more flexibly extract
// title, rowspan, colspan, background color, and content.
// Directives can be mixed with content.
func parseCell(cellInput string) (table.Cell, error) {
	remainingStr := strings.TrimSpace(cellInput)

	title := ""
	rowspan := 1 // Default
	colspan := 1 // Default
	bgColor := ""  // Default (empty means use table default)

	// 1. Extract Title (must be at the very beginning if present)
	// Regex: `^\[([^\]]*)\]` - matches "[Title]" at the start.
	// `([^\]]*)` captures content inside brackets.
	// We also trim space around the title and the rest of the string.
	titleRegex := regexp.MustCompile(`^\s*\[([^\]]*)\](.*)`)
	titleMatches := titleRegex.FindStringSubmatch(remainingStr)
	if len(titleMatches) == 3 { // 0: full match, 1: title content, 2: rest of string
		title = strings.TrimSpace(titleMatches[1])
		remainingStr = strings.TrimSpace(titleMatches[2])
	}

	// tempStr will be modified as directives are extracted.
	tempStr := remainingStr

	// 2. Extract Rowspan (e.g., "Some Content ::rowspan=2:: More Content")
	// Regex: `(.*?)::rowspan=(\d+)::(.*)`
	// (.*?) non-greedy before directive, (\d+) for number, (.*) for after.
	rowspanRegex := regexp.MustCompile(`(.*?)::rowspan=(\d+)::(.*)`)
	if matches := rowspanRegex.FindStringSubmatch(tempStr); len(matches) == 4 {
		parsedVal, err := strconv.Atoi(matches[2])
		if err == nil && parsedVal >= 1 {
			rowspan = parsedVal
			// Reconstruct string without the directive part, joining with a space
			tempStr = strings.TrimSpace(matches[1] + " " + matches[3])
		}
	}

	// 3. Extract Colspan (similarly to rowspan)
	colspanRegex := regexp.MustCompile(`(.*?)::colspan=(\d+)::(.*)`)
	if matches := colspanRegex.FindStringSubmatch(tempStr); len(matches) == 4 {
		parsedVal, err := strconv.Atoi(matches[2])
		if err == nil && parsedVal >= 1 {
			colspan = parsedVal
			tempStr = strings.TrimSpace(matches[1] + " " + matches[3])
		}
	}

	// 4. Extract Background Color (e.g., "Some Content {bg:#RRGGBB} More Content")
	// Regex: `(.*?)\{bg:([#\w\d]+)\}(.*)`
	// ([#\w\d]+) captures the color code.
	bgColorRegex := regexp.MustCompile(`(.*?)\{bg:([#\w\d]+)\}(.*)`)
	if matches := bgColorRegex.FindStringSubmatch(tempStr); len(matches) == 4 {
		bgColor = strings.TrimSpace(matches[2]) // The color code itself
		tempStr = strings.TrimSpace(matches[1] + " " + matches[3])
	}

	// What's left in tempStr after removing all directives is the actual content.
	content := strings.TrimSpace(tempStr)

	// Create cell using NewCell (which sets defaults) and then override.
	finalCell := table.NewCell(title, content)
	finalCell.Colspan = colspan
	finalCell.Rowspan = rowspan
	finalCell.BackgroundColor = bgColor

	return finalCell, nil
}
