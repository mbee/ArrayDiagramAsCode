package parser

import (
	"diagramgen/pkg/table"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseAllText takes a string input that may contain multiple table definitions
// and parses them into an AllTables struct.
func ParseAllText(fullInput string) (table.AllTables, error) {
	allTables := table.AllTables{Tables: make(map[string]table.Table)}
	trimmedFullInput := strings.TrimSpace(fullInput)
	if trimmedFullInput == "" {
		return allTables, nil // Return empty AllTables if input is empty
	}

	lines := strings.Split(trimmedFullInput, "\n")
	var currentTableLines []string
	var tableStartLineNumber int

	for i, line := range lines {
		if strings.HasPrefix(line, "table:") {
			// If currentTableLines has content, then a previous table definition has ended.
			if len(currentTableLines) > 0 {
				tableDef := strings.Join(currentTableLines, "\n")
				parsedTable, err := parseSingleTableDefinition(tableDef)
				if err != nil {
					return table.AllTables{}, fmt.Errorf("error parsing table starting at line %d: %w", tableStartLineNumber, err)
				}
				if parsedTable.ID == "" {
					return table.AllTables{}, fmt.Errorf("parsed table starting at line %d is missing an ID", tableStartLineNumber)
				}
				if _, exists := allTables.Tables[parsedTable.ID]; exists {
					return table.AllTables{}, fmt.Errorf("duplicate table ID '%s' found (originally from table at line %d)", parsedTable.ID, tableStartLineNumber)
				}
				allTables.Tables[parsedTable.ID] = parsedTable
				if allTables.MainTableID == "" {
					allTables.MainTableID = parsedTable.ID
				}
				currentTableLines = nil // Reset for the next table
			}
			tableStartLineNumber = i + 1 // Store line number (1-indexed) for error reporting
			currentTableLines = append(currentTableLines, line)
		} else if len(currentTableLines) > 0 {
			// Only append lines if we are currently inside a table definition
			currentTableLines = append(currentTableLines, line)
		}
	}

	// Process the last table definition if any
	if len(currentTableLines) > 0 {
		tableDef := strings.Join(currentTableLines, "\n")
		parsedTable, err := parseSingleTableDefinition(tableDef)
		if err != nil {
			return table.AllTables{}, fmt.Errorf("error parsing table starting at line %d: %w", tableStartLineNumber, err)
		}
		if parsedTable.ID == "" {
			return table.AllTables{}, fmt.Errorf("parsed table starting at line %d is missing an ID", tableStartLineNumber)
		}
		if _, exists := allTables.Tables[parsedTable.ID]; exists {
			return table.AllTables{}, fmt.Errorf("duplicate table ID '%s' found (originally from table at line %d)", parsedTable.ID, tableStartLineNumber)
		}
		allTables.Tables[parsedTable.ID] = parsedTable
		if allTables.MainTableID == "" {
			allTables.MainTableID = parsedTable.ID
		}
	}

	return allTables, nil
}

// parseSingleTableDefinition takes a string input for a single table definition
// and attempts to parse it into a Table object.
// It also parses global table settings from the title line and
// rowspan and background color from individual cells.
func parseSingleTableDefinition(tableInput string) (table.Table, error) {
	// Initialize table with default settings. These can be overridden by parsed settings.
	t := table.Table{
		Settings: table.DefaultGlobalSettings(),
		Rows:     []table.Row{}, // Ensure Rows is initialized
	}

	trimmedInput := strings.TrimSpace(tableInput)
	if trimmedInput == "" {
		// This case should ideally be handled by the caller (ParseAllText)
		// by not calling parseSingleTableDefinition with empty input.
		// However, returning an empty table or an error are options.
		// For now, let's assume valid non-empty input for a single table.
		return t, fmt.Errorf("empty input provided to parseSingleTableDefinition")
	}

	lines := strings.Split(trimmedInput, "\n")
	// A single table definition must start with "table:"
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "table:") {
		return table.Table{}, fmt.Errorf("table definition must start with 'table:'")
	}

	titleAndSettingsLine := strings.TrimPrefix(lines[0], "table:")
	firstLineIdx := 1 // Start processing rows from the next line

		// Regex to capture optional table ID: e.g., "[my-id] Title {settings}"
		idRegex := regexp.MustCompile(`^\s*\[([\w\-]+)\](.*)`)
		idMatches := idRegex.FindStringSubmatch(titleAndSettingsLine)

		if len(idMatches) == 3 { // 0: full match, 1: ID, 2: rest of the line
			t.ID = idMatches[1]
			titleAndSettingsLine = strings.TrimSpace(idMatches[2])
		}

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
	// THE ERRONEOUS '}' WAS HERE (after the if block, before "Process actual table rows")

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

	// 5. Extract Table Reference (e.g., "::table=ref-id::")
	// This should ideally be the only significant content if present.
	isTableRef := false
	tableRefID := ""
	// Regex: `(.*?)::table=([\w\-]+)::(.*)`
	tableRefRegex := regexp.MustCompile(`(.*?)::table=([\w\-]+)::(.*)`)
	if matches := tableRefRegex.FindStringSubmatch(tempStr); len(matches) == 4 {
		// If a table reference is found, it takes precedence.
		// Content before or after the directive will be trimmed.
		isTableRef = true
		tableRefID = strings.TrimSpace(matches[2])
		// tempStr is reconstructed from parts not including the table directive.
		// If other text was present (matches[1] or matches[3]), it's kept for now,
		// but will be cleared if isTableRef is true.
		tempStr = strings.TrimSpace(matches[1] + " " + matches[3])
	}

	// What's left in tempStr after removing all directives is the actual content.
	content := strings.TrimSpace(tempStr)

	// Create cell using NewCell (which sets defaults) and then override.
	finalCell := table.NewCell(title, content)
	finalCell.Colspan = colspan
	finalCell.Rowspan = rowspan
	finalCell.BackgroundColor = bgColor
	finalCell.IsTableRef = isTableRef
	finalCell.TableRefID = tableRefID

	// If the cell is a table reference, its direct content should be empty.
	if finalCell.IsTableRef {
		finalCell.Content = ""
	}

	return finalCell, nil
}
