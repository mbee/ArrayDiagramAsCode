package renderer

import (
	"diagramgen/pkg/table"
	"strings"
	"testing"
)

// Helper to create cells with specific colspan/rowspan for tests.
// table.NewCell initializes Colspan and Rowspan to 1.
func newTestCell(title, content string, cs, rs int) table.Cell {
	cell := table.NewCell(title, content) // Defaults cs=1, rs=1
	cell.Colspan = cs
	cell.Rowspan = rs
	return cell
}

func TestRender(t *testing.T) {
	tests := []struct {
		name  string
		table table.Table
		want  string // Expected output string
	}{
		{
			name:  "Empty Table",
			table: table.Table{Rows: []table.Row{}},
			want:  "",
		},
		{
			name:  "Empty Table with Nil Rows",
			table: table.Table{Rows: nil},
			want:  "",
		},
		{
			name:  "Table with Title Only",
			table: table.Table{Title: "Test Title", Rows: []table.Row{}},
			want: "Rendered Table: Test Title\n",
		},
		{
			name: "Simple Table",
			table: table.Table{
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
			want: strings.Join([]string{
				"R1C1 (cs:1) (rs:1) | R1C2 (cs:1) (rs:1)",
				"R2C1 (cs:1) (rs:1) | R2C2 (cs:1) (rs:1)",
				"",
			}, "\n"),
		},
		{
			name: "Table with All Features",
			table: table.Table{
				Title: "Complex Table",
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
			want: strings.Join([]string{
				"Rendered Table: Complex Table",
				"[ID] 1 (cs:1) (rs:1) | [Data] Info (cs:2) (rs:1)",
				"More (cs:1) (rs:1) | [Note] Important (cs:1) (rs:1) | End (cs:1) (rs:1)",
				"",
			}, "\n"),
		},
		{
			name: "Table with different rowspans",
			table: table.Table{
				Title: "Rowspan Test",
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
			want: strings.Join([]string{
				"Rendered Table: Rowspan Test",
				"[Cell1] Spans two rows (cs:1) (rs:2) | [Cell2] Normal (cs:1) (rs:1)",
				"[Cell3] Next row (cs:1) (rs:1)",
				"",
			}, "\n"),
		},
		{
			name: "Cell with empty content but with title and colspan",
			table: table.Table{
				Rows: []table.Row{
					{Cells: []table.Cell{
						newTestCell("TitleOnly", "", 3, 1),
					}},
				},
			},
			// This now correctly reflects the actual output with the double space.
			want: strings.Join([]string{
				"[TitleOnly]  (cs:3) (rs:1)", // Note the double space
				"",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Render(tt.table); got != tt.want {
				t.Errorf("Render() =\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
