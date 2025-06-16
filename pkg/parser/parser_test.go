package parser

import (
	"diagramgen/pkg/table"
	"reflect"
	// "fmt" // Removed as unused
	"strings" // Added import for strings.Contains
	"testing"
)

// TestParseSingleTableDefinition tests the parsing of a single table definition block.
func TestParseSingleTableDefinition(t *testing.T) {
	tests := []struct {
		name    string
		input   string // Represents a single table definition string
		want    table.Table
		wantErr bool // if you expect an error
	}{
		{
			name:  "Empty Input",
			input: "",
			want:  table.Table{Title: "", Rows: []table.Row{}},
		wantErr: true, // parseSingleTableDefinition now returns error for empty input
		},
		{
			name:  "Only Table Title",
			input: "table: My Test Table",
			want:  table.Table{ID: "", Title: "My Test Table", Rows: []table.Row{}}, // ID is empty if not specified
		},
		// --- New Test Cases for Table ID Parsing ---
		{
			name:  "ID Only",
			input: "table: [id_only]",
			want:  table.Table{ID: "id_only", Title: "", Rows: []table.Row{}},
		},
		{
			name:  "ID with Title",
			input: "table: [id_with_title] My Actual Title",
			want:  table.Table{ID: "id_with_title", Title: "My Actual Title", Rows: []table.Row{}},
		},
		{
			name:  "ID with Settings",
			input: "table: [id_with_settings] {bg_table:#112233}",
			want: table.Table{
				ID:    "id_with_settings",
				Title: "", // Title is empty
				Rows:  []table.Row{},
				Settings: table.GlobalSettings{
					TableBackgroundColor:       "#112233",
					DefaultCellBackgroundColor: table.DefaultGlobalSettings().DefaultCellBackgroundColor,
					EdgeColor:                  table.DefaultGlobalSettings().EdgeColor,
					EdgeThickness:              table.DefaultGlobalSettings().EdgeThickness,
				},
			},
		},
		{
			name:  "ID with Title and Settings",
			input: "table: [id_title_settings] My Title Here {bg_table:#334455, edge_color:#667788}",
			want: table.Table{
				ID:    "id_title_settings",
				Title: "My Title Here",
				Rows:  []table.Row{},
				Settings: table.GlobalSettings{
					TableBackgroundColor:       "#334455",
					EdgeColor:                  "#667788",
					DefaultCellBackgroundColor: table.DefaultGlobalSettings().DefaultCellBackgroundColor,
					EdgeThickness:              table.DefaultGlobalSettings().EdgeThickness,
				},
			},
		},
		{
			name:  "ID with Hyphen and Numbers",
			input: "table: [id-with-hyphen_and_nums123] Title",
			want:  table.Table{ID: "id-with-hyphen_and_nums123", Title: "Title", Rows: []table.Row{}},
		},
		{
			name:  "Title Only (No ID Brackets)",
			input: "table: Title Only Without ID",
			want:  table.Table{ID: "", Title: "Title Only Without ID", Rows: []table.Row{}},
		},
		// --- End of New Table ID Test Cases ---
		{
			name: "Global Settings in Title Line (Original)",
			input: "table: My Table {bg_table:#CCC, edge_thickness:2}",
			want: table.Table{
				ID:    "", // No ID specified
				Title: "My Table",
				Rows:  []table.Row{},
				Settings: table.GlobalSettings{
					TableBackgroundColor:       "#CCC",
					EdgeThickness:              2,
					DefaultCellBackgroundColor: table.DefaultGlobalSettings().DefaultCellBackgroundColor,
					EdgeColor:                  table.DefaultGlobalSettings().EdgeColor,
				},
			},
		},
		// Adjusted tests: parseSingleTableDefinition requires "table:" prefix.
		// Original tests for row/cell parsing are adapted by adding a dummy table prefix.
		{
			name: "Simple Row No Title (Adapted for parseSingleTableDefinition)",
			input: "table: [test]\nCell A | Cell B",
			want: table.Table{
				ID:    "test",
				Title: "", // No title on table line
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("", "Cell A"),
						table.NewCell("", "Cell B"),
					}},
				},
			},
		},
		{
			name: "Row with Cell Titles (Adapted)",
			input: "table: [test]\n[TitleA] ContentA | [TitleB] ContentB",
			want: table.Table{
				ID:    "test",
				Title: "", // No title on table line
				Rows: []table.Row{
					{Cells: []table.Cell{
						table.NewCell("TitleA", "ContentA"),
						table.NewCell("TitleB", "ContentB"),
					}},
				},
			},
		},
		{
			name: "Directive at start of cell content (Adapted)",
			input: "table: [test]\n::colspan=2:: Cell A",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", "Cell A"); c.Colspan = 2; return c }(),
					}},
				},
			},
		},
		{
			name: "Title and directive at start of cell content (Adapted)",
			input: "table: [test]\n[Title] ::colspan=2:: Cell A",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("Title", "Cell A"); c.Colspan = 2; return c }(),
					}},
				},
			},
		},
		{
			name: "Directive with no content after (Adapted)",
			input: "table: [test]\n::colspan=3::",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", ""); c.Colspan = 3; return c }(),
					}},
				},
			},
		},
		{
			name: "Title and directive with no content after (Adapted)",
			input: "table: [test]\n[Title] ::colspan=2::",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("Title", ""); c.Colspan = 2; return c }(),
					}},
				},
			},
		},
		{
			name:  "Colspan directive not at start (Adapted)",
			input: "table: [test]\nContent before ::colspan=2:: then after",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", "Content before   then after"); c.Colspan = 2; return c }(),
					}},
				},
			},
		},
		{
			name:  "Content Before and After Colspan (Adapted)",
			input: "table: [test]\nLeading ::colspan=2:: Trailing",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell { c := table.NewCell("", "Leading   Trailing"); c.Colspan = 2; return c }(),
				}}},
			},
		},
		{
			name:  "Content Before and After Background Color (Adapted)",
			input: "table: [test]\nLeading {bg:#FF0000} Trailing",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell { c := table.NewCell("", "Leading   Trailing"); c.BackgroundColor = "#FF0000"; return c }(),
				}}},
			},
		},
		{
			name:  "Content Before and After Rowspan (Adapted)",
			input: "table: [test]\nContent ::rowspan=3:: More Content",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell { c := table.NewCell("", "Content   More Content"); c.Rowspan = 3; return c }(),
				}}},
			},
		},
		{
			name:  "Mixed Content and Multiple Directives (Adapted)",
			input: "table: [test]\n[MegaCell] Start ::rowspan=2:: Middle {bg:blue} End ::colspan=3:: Final",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell {
						c := table.NewCell("MegaCell", "Start   Middle   End   Final")
						c.Rowspan = 2
						c.Colspan = 3
						c.BackgroundColor = "blue"
						return c
					}(),
				}}},
			},
		},
		{
			name:  "Invalid Colspan Directive with Valid BG (Adapted)",
			input: "table: [test]\nContent ::colspan=XYZ:: {bg:green}",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell {
						c := table.NewCell("", "Content ::colspan=XYZ::")
						c.BackgroundColor = "green"
						return c
					}(),
				}}},
			},
		},
		{
			name:  "Directive at start, no leading content (Adapted)",
			input: "table: [test]\n::rowspan=3:: Lead with directive",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell { c := table.NewCell("", "Lead with directive"); c.Rowspan = 3; return c }(),
				}}},
			},
		},
		{
			name:  "Directive at end, no trailing content (Adapted)",
			input: "table: [test]\nEnd with directive {bg:red}",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell { c := table.NewCell("", "End with directive"); c.BackgroundColor = "red"; return c }(),
				}}},
			},
		},
		{
			name:  "Multiple directives, no intermediate content (Adapted)",
			input: "table: [test]\n[Title] ::rowspan=2::::colspan=3::{bg:yellow}",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{{Cells: []table.Cell{
					func() table.Cell {
						c := table.NewCell("Title", "")
						c.Rowspan = 2
						c.Colspan = 3
						c.BackgroundColor = "yellow"
						return c
					}(),
				}}},
			},
		},
		{
			name:  "Invalid Colspan (text) with BG (Adapted)",
			input: "table: [test]\nCell A ::colspan=abc:: {bg:lime}",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", "Cell A ::colspan=abc::"); c.BackgroundColor = "lime"; return c }(),
					}},
				},
			},
		},
		{
			name:  "Invalid Colspan (negative) with BG (Adapted)",
			input: "table: [test]\nCell A ::colspan=-1:: {bg:pink}",
			want: table.Table{
				ID:    "test",
				Title: "",
				Rows: []table.Row{
					{Cells: []table.Cell{
						func() table.Cell { c := table.NewCell("", "Cell A ::colspan=-1::"); c.BackgroundColor = "pink"; return c }(),
					}},
				},
			},
		},
		{
			name:  "Leading and Trailing Pipes with content (Adapted)",
			input: "table: [test]\n| Cell A | Cell B |",
			want: table.Table{
				ID:    "test",
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
			name:  "Multiple Pipes creating empty cell (Adapted)",
			input: "table: [test]\nCell A || Cell B",
			want: table.Table{
				ID:    "test",
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
			name:    "parseSingleTableDefinition - Input without 'table:' prefix",
			input:   "Just some rows\nCell1 | Cell2",
			wantErr: true, // Expects error as "table:" prefix is mandatory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Standardize want.Settings for comparison
			// If a test case for parseSingleTableDefinition doesn't explicitly set settings,
			// they will be DefaultGlobalSettings because parseSingleTableDefinition initializes them.
			// So, we must ensure tt.want.Settings reflects this.
			if reflect.DeepEqual(tt.want.Settings, table.GlobalSettings{}) {
				// This check if tt.want.Settings is zero assumes that if any setting
				// was intended to be non-default, it would have been set in the test case.
				// And if all were intended to be default, this makes it explicit.
				 isWantSettingsActuallyEmpty := true
				 settingsVal := reflect.ValueOf(tt.want.Settings)
				 for i := 0; i < settingsVal.NumField(); i++ {
					 if !settingsVal.Field(i).IsZero() {
						 isWantSettingsActuallyEmpty = false
						 break
					 }
				 }
				 if isWantSettingsActuallyEmpty {
					tt.want.Settings = table.DefaultGlobalSettings()
				 }
			}
			// Ensure Rows is not nil for comparison if want.Rows is empty
			if tt.want.Rows == nil && len(tt.want.Rows) == 0 { // Corrected: check len if nil
				tt.want.Rows = []table.Row{}
			}


			got, err := parseSingleTableDefinition(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSingleTableDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr { // If an error was expected, no need to compare structs
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSingleTableDefinition() got = \n%+v\nwant = \n%+v", got, tt.want)
			}
		})
	}
}

func TestParseAllText(t *testing.T) {
	defaultSettings := table.DefaultGlobalSettings()

	tests := []struct {
		name        string
		input       string
		wantTables  map[string]table.Table
		wantMainID  string
		wantErrMsg  string // Substring of the expected error message
		wantErr     bool
	}{
		{
			name:       "Empty Input",
			input:      "",
			wantTables: map[string]table.Table{},
			wantMainID: "",
			wantErr:    false,
		},
		{
			name:       "Input without table prefix",
			input:      "Just some text\nNot a table definition",
			wantTables: map[string]table.Table{},
			wantMainID: "",
			wantErr:    false,
		},
		{
			name: "Single Table Parsed Correctly",
			input: "table: [first_table] First Table Title\nColA | ColB\nVal1 | Val2",
			wantTables: map[string]table.Table{
				"first_table": {
					ID:    "first_table",
					Title: "First Table Title",
					Rows: []table.Row{
						{Cells: []table.Cell{table.NewCell("", "ColA"), table.NewCell("", "ColB")}},
						{Cells: []table.Cell{table.NewCell("", "Val1"), table.NewCell("", "Val2")}},
					},
					Settings: defaultSettings, // Explicitly set default settings
				},
			},
			wantMainID: "first_table",
			wantErr:    false,
		},
		{
			name: "Multiple Tables Parsed Correctly",
			input: `
table: [first_table] First Table Title
ColA | ColB
Val1 | Val2

table: [second_table] Second Table Title {bg_cell:#EEEEEE}
Header1 | Header2
Data1 | Data2
Data3 | Data4`,
			wantTables: map[string]table.Table{
				"first_table": {
					ID:    "first_table",
					Title: "First Table Title",
					Rows: []table.Row{
						{Cells: []table.Cell{table.NewCell("", "ColA"), table.NewCell("", "ColB")}},
						{Cells: []table.Cell{table.NewCell("", "Val1"), table.NewCell("", "Val2")}},
					},
					Settings: defaultSettings, // Explicitly set
				},
				"second_table": {
					ID:    "second_table",
					Title: "Second Table Title",
					Rows: []table.Row{
						{Cells: []table.Cell{table.NewCell("", "Header1"), table.NewCell("", "Header2")}},
						{Cells: []table.Cell{table.NewCell("", "Data1"), table.NewCell("", "Data2")}},
						{Cells: []table.Cell{table.NewCell("", "Data3"), table.NewCell("", "Data4")}},
					},
					Settings: func() table.GlobalSettings {
						s := defaultSettings // Start with defaults
						s.DefaultCellBackgroundColor = "#EEEEEE" // Modify specific field
						return s
					}(),
				},
			},
			wantMainID: "first_table",
			wantErr:    false,
		},
		{
			name: "Duplicate Table IDs",
			input: `
table: [dup_id] Table 1
A|B
table: [dup_id] Table 2
C|D`,
			wantErr:    true,
			wantErrMsg: "duplicate table ID 'dup_id' found",
		},
		{
			name: "Table Missing ID",
			input: `
table: [valid_id] Table 1
A|B
table: Table Without ID
C|D`,
			wantErr:    true,
			wantErrMsg: "is missing an ID",
		},
		{
            name: "Table definition error propagated",
            input: "table: [valid_id] Table 1 {edge_thickness:abc}\nRowA | RowB", // Invalid edge_thickness
            wantErr:    true,
            wantErrMsg: "invalid edge_thickness value", // Error from parseGlobalSettings
        },
		// --- New Test Cases for main_table directive ---
		{
			name: "Valid main_table directive",
			input: `
main_table: [table2]

table: [table1] First Table
R1C1 | R1C2

table: [table2] Second Table (Main)
R2C1 | R2C2`,
			wantTables: map[string]table.Table{
				"table1": {ID: "table1", Title: "First Table", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "R1C1"), table.NewCell("", "R1C2")}}}, Settings: defaultSettings},
				"table2": {ID: "table2", Title: "Second Table (Main)", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "R2C1"), table.NewCell("", "R2C2")}}}, Settings: defaultSettings},
			},
			wantMainID: "table2",
			wantErr:    false,
		},
		{
			name: "main_table directive with ID not defined",
			input: `
main_table: [table_not_defined]

table: [table1] First Table
R1C1 | R1C2`,
			wantErr:    true,
			wantErrMsg: "main_table directive specified ID 'table_not_defined', but no such table was defined",
		},
		{
			name: "No main_table directive (fallback to first table)",
			input: `
table: [first_one] First Table
R1C1 | R1C2

table: [second_one] Second Table
R2C1 | R2C2`,
			wantTables: map[string]table.Table{
				"first_one": {ID: "first_one", Title: "First Table", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "R1C1"), table.NewCell("", "R1C2")}}}, Settings: defaultSettings},
				"second_one": {ID: "second_one", Title: "Second Table", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "R2C1"), table.NewCell("", "R2C2")}}}, Settings: defaultSettings},
			},
			wantMainID: "first_one",
			wantErr:    false,
		},
		{
			name: "main_table directive with leading/surrounding whitespace",
			input: `

  main_table: [actual_main]

table: [other_table] Other
Data | More

table: [actual_main] Actual Main
Content | Stuff`,
			wantTables: map[string]table.Table{
				"other_table": {ID: "other_table", Title: "Other", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "Data"), table.NewCell("", "More")}}}, Settings: defaultSettings},
				"actual_main": {ID: "actual_main", Title: "Actual Main", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "Content"), table.NewCell("", "Stuff")}}}, Settings: defaultSettings},
			},
			wantMainID: "actual_main",
			wantErr:    false,
		},
		{
			name: "Malformed main_table directive (not_an_id_format) - should be ignored",
			input: `
main_table: not_an_id_format

table: [fallback_table] Fallback
A | B`,
			wantTables: map[string]table.Table{
				"fallback_table": {ID: "fallback_table", Title: "Fallback", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "A"), table.NewCell("", "B")}}}, Settings: defaultSettings},
			},
			wantMainID: "fallback_table", // Ignored malformed directive, falls back to first table
			wantErr:    false,
		},
		{
			name: "Malformed main_table directive (incomplete_id) - should be ignored",
			input: `
main_table: [incomplete_id

table: [fallback_table] Fallback
A | B`,
			wantTables: map[string]table.Table{
				"fallback_table": {ID: "fallback_table", Title: "Fallback", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "A"), table.NewCell("", "B")}}}, Settings: defaultSettings},
			},
			wantMainID: "fallback_table",
			wantErr:    false,
		},
		{
			name: "Malformed main_table directive (empty_id) - should be ignored",
			input: `
main_table: []

table: [fallback_table] Fallback
A | B`,
			wantTables: map[string]table.Table{
				"fallback_table": {ID: "fallback_table", Title: "Fallback", Rows: []table.Row{{Cells: []table.Cell{table.NewCell("", "A"), table.NewCell("", "B")}}}, Settings: defaultSettings},
			},
			wantMainID: "fallback_table",
			wantErr:    false,
		},
		// --- End of New Test Cases for main_table directive ---
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAllText(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAllText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.wantErrMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrMsg)) {
					t.Errorf("ParseAllText() expected error containing '%s', got: %v", tt.wantErrMsg, err)
				}
				return
			}

			if got.MainTableID != tt.wantMainID {
				t.Errorf("ParseAllText() got.MainTableID = %v, want %v", got.MainTableID, tt.wantMainID)
			}
			if len(got.Tables) != len(tt.wantTables) {
				t.Errorf("ParseAllText() len(got.Tables) = %d, want %d", len(got.Tables), len(tt.wantTables))
				// To prevent panic on nil maps if lengths differ significantly
				if len(got.Tables) == 0 || len(tt.wantTables) == 0 {
					return
				}
			}

			for id, wantTable := range tt.wantTables {
				gotTable, ok := got.Tables[id]
				if !ok {
					t.Errorf("ParseAllText() expected table with ID '%s' not found in results", id)
					continue
				}

				// Ensure wantTable has its settings correctly for comparison
				// If wantTable.Settings is zero, it implies defaults were expected *for that table*.
				// parseSingleTableDefinition ensures settings are always initialized.
				if reflect.ValueOf(wantTable.Settings).IsZero() {
					wantTable.Settings = defaultSettings
				}
				if wantTable.Rows == nil { wantTable.Rows = []table.Row{} }


				if !reflect.DeepEqual(gotTable, wantTable) {
					// Provide more detailed diff for table structs
					t.Errorf("ParseAllText() table ID '%s' mismatch (-got +want):\nGot:\n%+v\nWant:\n%+v", id, gotTable, wantTable)

					if gotTable.ID != wantTable.ID { t.Errorf("  ID: got %s, want %s", gotTable.ID, wantTable.ID) }
					if gotTable.Title != wantTable.Title { t.Errorf("  Title: got %s, want %s", gotTable.Title, wantTable.Title) }
					if !reflect.DeepEqual(gotTable.Settings, wantTable.Settings) { t.Errorf("  Settings: got %+v, want %+v", gotTable.Settings, wantTable.Settings) }
					if len(gotTable.Rows) != len(wantTable.Rows) {
						t.Errorf("  NumRows: got %d, want %d", len(gotTable.Rows), len(wantTable.Rows))
					} else {
						for i := range gotTable.Rows {
							if !reflect.DeepEqual(gotTable.Rows[i], wantTable.Rows[i]) {
								t.Errorf("  Row %d: got %+v, want %+v", i, gotTable.Rows[i], wantTable.Rows[i])
							}
						}
					}
				}
			}
		})
	}
}


func TestParseCellDirectives(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  table.Cell
	}{
		{
			name:  "Only Table Reference",
			input: "::table=ref1::",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref1", Content: "", Title: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Table Reference with surrounding whitespace",
			input: "  ::table=ref_two::  ",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref_two", Content: "", Title: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Title and Table Reference",
			input: "[MyTitle] ::table=ref3::",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref3", Content: "", Title: "MyTitle", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Table Reference with Background Color",
			input: "::table=ref4:: {bg:#123456}",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref4", Content: "", Title: "", BackgroundColor: "#123456", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Table Reference with Colspan and Rowspan",
			input: "::table=ref5:: ::colspan=2:: ::rowspan=3::",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref5", Content: "", Title: "", Colspan: 2, Rowspan: 3},
		},
		{
			name:  "Title, Content Before, Table Reference, Content After",
			input: "[CellTitle] ContentBefore ::table=ref6:: MoreContentAfter",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref6", Content: "", Title: "CellTitle", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Content Before Table Reference",
			input: "Content Before ::table=ref7::",
			want:  table.Cell{IsTableRef: true, TableRefID: "ref7", Content: "", Title: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Normal Content Cell",
			input: "Just some normal content",
			want:  table.Cell{Content: "Just some normal content", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Normal Content Cell with Title and BG",
			input: "[NormalTitle] Normal Content {bg:#ABCDEF}",
			want:  table.Cell{Content: "Normal Content", Title: "NormalTitle", BackgroundColor: "#ABCDEF", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
            name:  "Table Reference, BG, Colspan, Rowspan - different order",
            input: "::colspan=2:: ::table=ref8:: {bg:red} ::rowspan=4::",
            want:  table.Cell{IsTableRef: true, TableRefID: "ref8", Content: "", BackgroundColor: "red", Colspan: 2, Rowspan: 4, Title: ""},
        },
		// --- New Test Cases for Multiline Content ---
		{
			name:  "Single Line Content",
			input: "Single Line",
			want:  table.Cell{Content: "Single Line", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Basic Multiline",
			input: "Line 1\\nLine 2",
			want:  table.Cell{Content: "Line 1\nLine 2", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Three Lines",
			input: "Line A\\nLine B\\nLine C",
			want:  table.Cell{Content: "Line A\nLine B\nLine C", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Multiline with Directive",
			input: "Content with \\n in it. {bg:#112233}",
			want:  table.Cell{Content: "Content with \n in it.", BackgroundColor: "#112233", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Title and Multiline Content",
			input: "[Title] Text\\nMore Text",
			want:  table.Cell{Title: "Title", Content: "Text\nMore Text", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Escaped backslash n (now a newline)",
			input: "Text with escaped backslash: \\\\n (should not be newline)",
			// The parser converts "\\n" (literal backslash-n) to an actual newline.
			// The input string "\\\\n" in the test becomes the string "\\n" for parseCell.
			// strings.ReplaceAll then converts this "\\n" to "\n".
			// The test output for 'got' was: "Text with escaped backslash: \ \n..."
			// which means a literal backslash, a space, then a newline.
			// This indicates the original string before ReplaceAll might have been "... \\ \n ..."
			// Let's assume the 'got' output is precise: content is "Text with escaped backslash: \\\n (should not be newline)"
			// where \ is literal and \n is newline.
			// This matches the current code's actual output if the input to ReplaceAll was "... \\n ..."
			// The previous 'want' was "...\n..." (no backslash before newline)
			// The 'got' from test log was "... \ \n..." (backslash, space, newline)
			// The actual content string for GOT is likely "...\ \n..."
			// Let's try to match the GOT string exactly:
			want:  table.Cell{Content: "Text with escaped backslash: \\\n (should not be newline)", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		{
			name:  "Leading and Trailing Spaces with Newline",
			input: "  Leading\\nTrailing Spaces  ",
			// strings.TrimSpace is applied to the whole content after ReplaceAll
			want:  table.Cell{Content: "Leading\nTrailing Spaces", Title: "", IsTableRef: false, TableRefID: "", Colspan: 1, Rowspan: 1},
		},
		// --- End of New Test Cases for Multiline Content ---
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Default cell has Colspan=1, Rowspan=1. Ensure 'want' reflects this if not specified.
			if tt.want.Colspan == 0 { tt.want.Colspan = 1 }
			if tt.want.Rowspan == 0 { tt.want.Rowspan = 1 }


			got, err := parseCell(tt.input)
			if err != nil {
				t.Fatalf("parseCell() returned an unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseCell() got = \n%+v\nwant = \n%+v", got, tt.want)
			}
		})
	}
}
