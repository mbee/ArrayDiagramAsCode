package renderer

import (
	"diagramgen/pkg/table"
	"fmt" // Import fmt for formatSettingsLine helper
	"strings"
	"testing"
)

// Helper to create cells with specific colspan/rowspan for tests.
func newTestCell(title, content string, cs, rs int) table.Cell {
	cell := table.NewCell(title, content)
	cell.Colspan = cs
	cell.Rowspan = rs
	return cell
}

// Helper to format the settings string consistently for tests
func formatSettingsLine(s table.GlobalSettings) string {
	return fmt.Sprintf("  Settings: EdgeColor: '%s', EdgeThickness: %d, DefaultCellBG: '%s', TableBG: '%s'\n",
		s.EdgeColor, s.EdgeThickness, s.DefaultCellBackgroundColor, s.TableBackgroundColor)
}

func TestRender(t *testing.T) {
	defaultSettings := table.DefaultGlobalSettings()
	defaultSettingsStr := formatSettingsLine(defaultSettings)

	// Example of specific settings for a test, if needed
	// customSettings := table.GlobalSettings{ /* ... */ }
	// customSettingsStr := formatSettingsLine(customSettings)

	tests := []struct {
		name  string
		table table.Table
		want  string // Expected output string
	}{
		{
			name:  "Empty Table",
			table: table.Table{Rows: []table.Row{}, Settings: defaultSettings},
			want:  defaultSettingsStr, // Only settings line for an empty table
		},
		{
			name:  "Empty Table with Nil Rows",
			table: table.Table{Rows: nil, Settings: defaultSettings},
			want:  defaultSettingsStr, // Only settings line
		},
		{
			name:  "Table with Title Only",
			table: table.Table{Title: "Test Title", Rows: []table.Row{}, Settings: defaultSettings},
			want: "Table: Test Title\n" + defaultSettingsStr,
		},
		{
			name: "Simple Table",
			table: table.Table{
				Settings: defaultSettings, // Add this
				Rows: []table.Row{
					{Cells: []table.Cell{
						newTestCell("", "R1C1", 1, 1),
						newTestCell("", "R1C2", 1, 1),
					}},
					{Cells: []table.Cell{
						newTestCell("", "R2C1", 1, 1),
						newTestCell("", "R2C2", 1, 1),
					}},
				},
			},
			want: defaultSettingsStr + // Settings line first
				strings.Join([]string{
					"R1C1 (cs:1) (rs:1) | R1C2 (cs:1) (rs:1)",
					"R2C1 (cs:1) (rs:1) | R2C2 (cs:1) (rs:1)",
					"",
				}, "\n"),
		},
		{
			name: "Table with All Features",
			table: table.Table{
				Title:    "Complex Table",
				Settings: defaultSettings, // Add this
				Rows: []table.Row{
					{Cells: []table.Cell{
						newTestCell("ID", "1", 1, 1),
						newTestCell("Data", "Info", 2, 1),
					}},
					{Cells: []table.Cell{
						newTestCell("", "More", 1, 1),
						newTestCell("Note", "Important", 1, 1),
						newTestCell("", "End", 1, 1),
					}},
				},
			},
			want: "Table: Complex Table\n" + // Title line
				defaultSettingsStr + // Settings line
				strings.Join([]string{
					"[ID] 1 (cs:1) (rs:1) | [Data] Info (cs:2) (rs:1)",
					"More (cs:1) (rs:1) | [Note] Important (cs:1) (rs:1) | End (cs:1) (rs:1)",
					"",
				}, "\n"),
		},
		{
			name: "Table with different rowspans",
			table: table.Table{
				Title:    "Rowspan Test",
				Settings: defaultSettings, // Add this
				Rows: []table.Row{
					{Cells: []table.Cell{
						newTestCell("Cell1", "Spans two rows", 1, 2),
						newTestCell("Cell2", "Normal", 1, 1),
					}},
					{Cells: []table.Cell{
						newTestCell("Cell3", "Next row", 1, 1),
					}},
				},
			},
			want: "Table: Rowspan Test\n" + // Title line
				defaultSettingsStr + // Settings line
				strings.Join([]string{
					"[Cell1] Spans two rows (cs:1) (rs:2) | [Cell2] Normal (cs:1) (rs:1)",
					"[Cell3] Next row (cs:1) (rs:1)",
					"",
				}, "\n"),
		},
		{
			name: "Cell with empty content but with title and colspan",
			table: table.Table{
				Settings: defaultSettings, // Add this
				Rows: []table.Row{
					{Cells: []table.Cell{
						newTestCell("TitleOnly", "", 3, 1),
					}},
				},
			},
			want: defaultSettingsStr + // Settings line
				strings.Join([]string{
					"[TitleOnly]  (cs:3) (rs:1)", // Note the double space
					"",
				}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Crucial: The Render function uses the Settings from tt.table.
			// If tt.table.Settings is zero, it will render with zero values, not defaults.
			// The DefaultGlobalSettings() is applied by the PARSER, not the RENDERER.
			// So, tests must provide settings in tt.table.Settings.
			// The code above now correctly initializes tt.table.Settings for each test case.

			if got := Render(tt.table); got != tt.want {
				t.Errorf("Render() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
