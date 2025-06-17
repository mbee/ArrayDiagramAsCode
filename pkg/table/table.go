package table

// GlobalSettings holds default styling for the entire table.
type GlobalSettings struct {
	DefaultCellBackgroundColor string // e.g., "#FFFFFF"
	TableBackgroundColor       string // e.g., "#ECECEC"
	EdgeColor                  string // e.g., "#000000"
	EdgeThickness              int    // e.g., 1
}

// DefaultGlobalSettings provides a default set of global table settings.
func DefaultGlobalSettings() GlobalSettings {
	return GlobalSettings{
		DefaultCellBackgroundColor: "#FFFFFF", // White
		TableBackgroundColor:       "",        // Transparent/None by default, can be overridden by parser
		EdgeColor:                  "#000000", // Black
		EdgeThickness:              1,
	}
}

// Cell represents a single cell in a table
type Cell struct {
	Title           string
	Content         string
	Colspan         int    // For merged cells horizontally
	Rowspan         int    // For merged cells vertically
	BackgroundColor string // Specific background color for this cell, e.g., "#RRGGBB"
	IsTableRef      bool   // Flag to indicate this cell is a table reference
	TableRefID      string // ID of the table to render in this cell

	// New fields for inner table control
	InnerTableAlignment string // e.g., "top_left", "center"
	InnerTableScaleMode string // e.g., "none", "fit_width"

	// New fields for fixed cell dimensions
	FixedWidth  float64 // Specified fixed width in pixels. 0.0 means not set.
	FixedHeight float64 // Specified fixed height in pixels. 0.0 means not set.
}

// NewCell creates a new Cell with default values.
// Title and Content are provided, Colspan and Rowspan default to 1.
// BackgroundColor defaults to empty string, implying global default should be used.
func NewCell(title string, content string) Cell {
	return Cell{
		Title:               title,
		Content:             content,
		Colspan:             1,
		Rowspan:             1,    // Default rowspan is 1
		BackgroundColor:     "",   // Default to no specific background color (use table default)
		IsTableRef:          false,
		TableRefID:          "",
		InnerTableAlignment: "top_left", // Default alignment
		InnerTableScaleMode: "none",     // Default scale mode
		FixedWidth:          0.0,        // Default: not fixed
		FixedHeight:         0.0,        // Default: not fixed
	}
}

// Row represents a row in a table
type Row struct {
	Cells []Cell
}

// Table represents a table, including its data and global settings.
type Table struct {
	ID       string // New field for the table identifier
	Title    string
	Rows     []Row
	Settings GlobalSettings // Holds global settings for the table
}

// You might want a constructor for Table that initializes settings,
// especially if the parser will create Tables.
// For example:
/*
func NewTableWithDefaults() Table {
    return Table {
        Settings: DefaultGlobalSettings(),
        // Rows would be initialized as needed, e.g., Rows: []Row{}
    }
}
*/

// AllTables holds all parsed tables from an input source and identifies the main one.
type AllTables struct {
	Tables      map[string]Table // Stores all parsed tables, keyed by their ID.
	MainTableID string           // ID of the table to be rendered as the primary one.
}
