package table

// Cell represents a single cell in a table
type Cell struct {
	Title   string
	Content string
	Colspan int // For merged cells
	Rowspan int // For merged cells (future)
}

// NewCell creates a new Cell with default Colspan and Rowspan values.
func NewCell(title, content string) Cell {
	return Cell{
		Title:   title,
		Content: content,
		Colspan: 1,
		Rowspan: 1, // Default Rowspan to 1 as well
	}
}

// Row represents a row in a table
type Row struct {
	Cells []Cell
}

// Table represents a table
type Table struct {
	Title string
	Rows  []Row
}
