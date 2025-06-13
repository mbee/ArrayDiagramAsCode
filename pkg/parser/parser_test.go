package parser

import (
	"diagramgen/pkg/table"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    table.Table
		wantErr bool // if you expect an error
	}{
		{
			name:  "Empty Input",
			input: "",
			// Parse returns an empty table (Title="", Rows=nil) for empty string input.
			// Initialize Rows to empty slice for DeepEqual consistency if parser does.
			// Current parser initializes t.Rows = []table.Row{}
			want: table.Table{Title: "", Rows: []table.Row{}},
		},
		{
			name:  "Only Table Title",
			input: "table: My Test Table",
			// Parser initializes t.Rows = []table.Row{}
			want: table.Table{Title: "My Test Table", Rows: []table.Row{}},
		},
		{
			name:  "Only Table Title with trailing newline",
			input: "table: My Test Table\n",
			want:  table.Table{Title: "My Test Table", Rows: []table.Row{}},
		},
		{
			name: "Simple Row No Title",
			input: "Cell A | Cell B | Cell C",
			want: table.Table{
				Title: "", // Explicitly set default
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "Cell A"),
						table.NewCell("", "Cell B"),
						table.NewCell("", "Cell C"),
					}},
				},
			},
		},
		{
			name: "Row with Cell Titles",
			input: "[TitleA] ContentA | [TitleB] ContentB",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("TitleA", "ContentA"),
						table.NewCell("TitleB", "ContentB"),
					}},
				},
			},
		},
		{
			name: "Row with Colspan",
			input: "::colspan=2:: Cell A | Cell B", // Moved colspan to start of cell content
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", "Cell A"); c.Colspan = 2; return c }(),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
			name: "Full Table Example",
			input: `table: Full Test
[ID] 1 | [Name] ::colspan=2:: Test Name
Value A | Value B | Value C`, // Colspan for Name cell
			want: table.Table{
				Title: "Full Test",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("ID", "1"),
						func() table.Cell { c := table.NewCell("Name", "Test Name"); c.Colspan = 2; return c }(),
					}},
					{Cells: []table.Cell{
						table.NewCell("", "Value A"),
						table.NewCell("", "Value B"),
						table.NewCell("", "Value C"),
					}},
				},
			},
		},
		{
			name: "Table with Empty Lines",
			input: `table: Test Empty Lines

Row 1 Cell 1 | Row 1 Cell 2

Row 2 Cell 1
`,
			want: table.Table{
				Title: "Test Empty Lines",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "Row 1 Cell 1"),
						table.NewCell("", "Row 1 Cell 2"),
					}},
					{Cells: []table.Cell{
						table.NewCell("", "Row 2 Cell 1"),
					}},
				},
			},
		},
		{
			name: "Colspan with no content after",
			input: "::colspan=3::",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", ""); c.Colspan = 3; return c }(),
					}},
				},
			},
		},
		{
			name: "Title and Colspan with no content after",
			input: "[Title] ::colspan=2::",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("Title", ""); c.Colspan = 2; return c }(),
					}},
				},
			},
		},
		{
			name: "Invalid Colspan (text) treated as content",
			input: "Cell A ::colspan=abc:: | Cell B",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						// Regex ^::colspan=(\d+):: won't match "::colspan=abc::" as a whole if it's not at the start.
						// If "Cell A ::colspan=abc::" is the cell string, title is empty.
						// Content is "Cell A ::colspan=abc::". Regex doesn't match. Colspan remains 1.
						table.NewCell("", "Cell A ::colspan=abc::"),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
			name: "Invalid Colspan (negative) treated as content",
			input: "Cell A ::colspan=-1:: | Cell B",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						// Similar to above, "::colspan=-1::" is not matched by `\d+` for the number part.
						table.NewCell("", "Cell A ::colspan=-1::"),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
            name: "Colspan directive not at start of content",
            input: "Content before ::colspan=2:: then after",
            want: table.Table{
				Title: "",
                Rows: []table.Row{
                    {Cells: []table.Cell{
                        table.NewCell("", "Content before ::colspan=2:: then after"),
                    }},
                },
            },
        },
		{
			name:  "Leading and Trailing Pipes with content",
			input: "| Cell A | Cell B |",
			// Parser logic:
			// trimmedLine = "| Cell A | Cell B |"
			// cellStrings = ["", " Cell A ", " Cell B ", ""]
			// startIdx = 1, endIdx = 3
			// actualCellStrings = [" Cell A ", " Cell B "]
			// cell1: parseCell(" Cell A ") -> Title:"", Content:"Cell A", Colspan:1
			// cell2: parseCell(" Cell B ") -> Title:"", Content:"Cell B", Colspan:1
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "Cell A"),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
			name:  "Leading and Trailing Pipes with empty content",
			input: "| | |",
			// trimmedLine = "| | |"
			// cellStrings = ["", " ", " ", ""]
			// startIdx = 1, endIdx = 3
			// actualCellStrings = [" ", " "]
			// cell1: parseCell(" ") -> Title:"", Content:"", Colspan:1
			// cell2: parseCell(" ") -> Title:"", Content:"", Colspan:1
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", ""),
						table.NewCell("", ""),
					}},
				},
			},
		},
		{
			name:  "Multiple Pipes creating empty cell",
			input: "Cell A || Cell B",
			// trimmedLine = "Cell A || Cell B"
			// cellStrings = ["Cell A ", "", " Cell B"]
			// startIdx = 0, endIdx = 3
			// actualCellStrings = ["Cell A ", "", " Cell B"]
			// cell1: parseCell("Cell A ") -> Content "Cell A"
			// cell2: parseCell("") -> Content ""
			// cell3: parseCell(" Cell B") -> Content "Cell B"
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "Cell A"),
						table.NewCell("", ""),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
			name:  "Single cell, no pipes",
			input: "JustOneCell",
			want: table.Table{
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "JustOneCell"),
					}},
				},
			},
		},
		{
            name:  "Line with only pipes |||",
            input: "|||",
            // trimmedLine = "|||"
            // cellStrings = ["", "", "", ""]
            // startIdx = 1, endIdx = 3
            // actualCellStrings = ["", ""]
            // cell1: parseCell("") -> Content ""
            // cell2: parseCell("") -> Content ""
            want: table.Table{
				Title: "",
                Rows: []table.Row{
                    {Cells: []table.Cell{
                        table.NewCell("", ""),
                        table.NewCell("", ""),
                    }},
                },
            },
        },
        {
            name:  "Line with only one pipe |",
            input: "|",
            // trimmedLine = "|"
            // cellStrings = ["", ""]
            // startIdx = 1, endIdx = 1
            // actualCellStrings = []
            // No cells added, so no row is added.
            want: table.Table{Title: "", Rows: []table.Row{}},
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// For DeepEqual, ensure Rows is not nil if want.Rows is an empty slice
			if got.Rows == nil && len(tt.want.Rows) == 0 {
				got.Rows = []table.Row{}
			}


			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %+v, want %+v", got, tt.want)
				// For more detailed diff, could print them as JSON or iterate
                if len(got.Rows) != len(tt.want.Rows) {
                    t.Errorf("Parse() got %d rows, want %d rows", len(got.Rows), len(tt.want.Rows))
                } else {
                    for i := range got.Rows {
                        if len(got.Rows[i].Cells) != len(tt.want.Rows[i].Cells) {
                             t.Errorf("Parse() row %d got %d cells, want %d cells", i, len(got.Rows[i].Cells), len(tt.want.Rows[i].Cells))
                        } else {
                            for j := range got.Rows[i].Cells {
                                if !reflect.DeepEqual(got.Rows[i].Cells[j], tt.want.Rows[i].Cells[j]) {
                                    t.Errorf("Parse() row %d cell %d got = %+v, want %+v", i, j, got.Rows[i].Cells[j], tt.want.Rows[i].Cells[j])
                                }
                            }
                        }
                    }
                }
			}
		})
	}
}
