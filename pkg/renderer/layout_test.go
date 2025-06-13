package renderer

import (
	"diagramgen/pkg/table"
	"math" // For float comparisons
	"strings" // For strings.Contains
	"testing"
	// "github.com/fogleman/gg" // Not strictly needed unless we create a DC for test setup
)

const testEpsilonFloat = 0.1 // For float comparisons

// defaultFontPath is a package-level variable for tests.
// This path needs to be valid in the test execution environment.
var testDefaultFontPathLayout = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

// Helper to create a simple cell for tests, returning a pointer.
func newTestLayoutCell(title, content string, colspan, rowspan int) *table.Cell {
	cellData := table.NewCell(title, content) // Create the cell data
	cellData.Colspan = colspan
	cellData.Rowspan = rowspan
	return &cellData // Return address of cell data
}

func TestPopulateOccupationMap(t *testing.T) {
	// Test Case 1: Simple 2x2 table, no spans
	table1 := &table.Table{
		Rows: []table.Row{
			{Cells: []table.Cell{
				*newTestLayoutCell("R0C0", "c1", 1, 1), *newTestLayoutCell("R0C1", "c2", 1, 1),
			}},
			{Cells: []table.Cell{
				*newTestLayoutCell("R1C0", "c3", 1, 1), *newTestLayoutCell("R1C1", "c4", 1, 1),
			}},
		},
	}
	lg1, err1 := PopulateOccupationMap(table1)
	if err1 != nil { t.Fatalf("table1: PopulateOccupationMap failed: %v", err1) }
	if lg1.NumLogicalRows != 2 { t.Errorf("table1: expected 2 logical rows, got %d", lg1.NumLogicalRows) }
	if lg1.NumLogicalCols != 2 { t.Errorf("table1: expected 2 logical cols, got %d", lg1.NumLogicalCols) }
	// Compare with pointers to cells within table1 structure
	if lg1.OccupationMap[0][0] != &table1.Rows[0].Cells[0] { t.Errorf("table1: mismatch at (0,0), expected %p got %p", &table1.Rows[0].Cells[0], lg1.OccupationMap[0][0]) }
	if lg1.OccupationMap[0][1] != &table1.Rows[0].Cells[1] { t.Errorf("table1: mismatch at (0,1), expected %p got %p", &table1.Rows[0].Cells[1], lg1.OccupationMap[0][1]) }
	if lg1.OccupationMap[1][0] != &table1.Rows[1].Cells[0] { t.Errorf("table1: mismatch at (1,0), expected %p got %p", &table1.Rows[1].Cells[0], lg1.OccupationMap[1][0]) }
	if lg1.OccupationMap[1][1] != &table1.Rows[1].Cells[1] { t.Errorf("table1: mismatch at (1,1), expected %p got %p", &table1.Rows[1].Cells[1], lg1.OccupationMap[1][1]) }

	// Test Case 2: Colspan
	table2 := &table.Table{
		Rows: []table.Row{
			{Cells: []table.Cell{
				*newTestLayoutCell("R0C0_span2", "span2col", 2, 1), *newTestLayoutCell("R0C2_after", "afterspan", 1, 1),
			}},
			{Cells: []table.Cell{
				*newTestLayoutCell("R1C0_val", "r1c0", 1, 1), *newTestLayoutCell("R1C1_val", "r1c1", 1, 1), *newTestLayoutCell("R1C2_val", "r1c2", 1, 1),
			}},
		},
	}
	lg2, err2 := PopulateOccupationMap(table2)
	if err2 != nil { t.Fatalf("table2: PopulateOccupationMap failed: %v", err2) }
	if lg2.NumLogicalRows != 2 { t.Errorf("table2: expected 2 logical rows, got %d", lg2.NumLogicalRows) }
	if lg2.NumLogicalCols != 3 { t.Errorf("table2: expected 3 logical cols, got %d", lg2.NumLogicalCols) }
	if lg2.OccupationMap[0][0] != &table2.Rows[0].Cells[0] { t.Errorf("table2: mismatch at (0,0)") }
	if lg2.OccupationMap[0][1] != &table2.Rows[0].Cells[0] { t.Errorf("table2: mismatch at (0,1) (span)") }
	if lg2.OccupationMap[0][2] != &table2.Rows[0].Cells[1] { t.Errorf("table2: mismatch at (0,2)") }
	if lg2.OccupationMap[1][0] != &table2.Rows[1].Cells[0] { t.Errorf("table2: mismatch at (1,0)") }
	if lg2.OccupationMap[1][1] != &table2.Rows[1].Cells[1] { t.Errorf("table2: mismatch at (1,1)") }
	if lg2.OccupationMap[1][2] != &table2.Rows[1].Cells[2] { t.Errorf("table2: mismatch at (1,2)") }

	// Test Case 3: Rowspan
	table3 := &table.Table{
		Rows: []table.Row{
			{Cells: []table.Cell{
				*newTestLayoutCell("R0C0_rowspan2", "span2row", 1, 2), *newTestLayoutCell("R0C1_옆", "옆", 1, 1),
			}},
			{Cells: []table.Cell{ // This cell will be placed at grid (1,1)
				*newTestLayoutCell("R1C1_아래옆", "아래옆", 1, 1),
			}},
		},
	}
	lg3, err3 := PopulateOccupationMap(table3)
	if err3 != nil { t.Fatalf("table3: PopulateOccupationMap failed: %v", err3) }
	if lg3.NumLogicalRows != 2 { t.Errorf("table3: expected 2 logical rows, got %d", lg3.NumLogicalRows) }
	if lg3.NumLogicalCols != 2 { t.Errorf("table3: expected 2 logical cols, got %d", lg3.NumLogicalCols) }
	if lg3.OccupationMap[0][0] != &table3.Rows[0].Cells[0] { t.Errorf("table3: mismatch at (0,0)") }
	if lg3.OccupationMap[1][0] != &table3.Rows[0].Cells[0] { t.Errorf("table3: mismatch at (1,0) (span)") }
	if lg3.OccupationMap[0][1] != &table3.Rows[0].Cells[1] { t.Errorf("table3: mismatch at (0,1)") }
	if lg3.OccupationMap[1][1] != &table3.Rows[1].Cells[0] { t.Errorf("table3: mismatch at (1,1)") }
}

func TestCalculateColumnWidthsAndRowHeights(t *testing.T) {
	constants := LayoutConstants{
		FontPath:             testDefaultFontPathLayout,
		FontSize:             12.0,
		LineHeightMultiplier: 1.4,
		Padding:              5.0,
		MinCellWidth:         10.0,
		MinCellHeight:        10.0,
	}

	c1 := newTestLayoutCell("C1", "short", 1, 1)
	c2 := newTestLayoutCell("C2", "this is a much longer string", 1, 1)
	tableSimple := &table.Table{Rows: []table.Row{{Cells: []table.Cell{*c1, *c2}}}}
	lgSimple, popErr := PopulateOccupationMap(tableSimple)
	if popErr != nil { t.Fatalf("Simple: PopulateOccupationMap failed: %v", popErr) }

	calcErr := lgSimple.CalculateColumnWidthsAndRowHeights(constants)
	fontLoadFailed := (calcErr != nil && strings.Contains(calcErr.Error(), "failed to load font"))
	if calcErr != nil && !fontLoadFailed {
		t.Fatalf("Simple: CalculateColumnWidthsAndRowHeights failed unexpectedly: %v", calcErr)
	}
	if fontLoadFailed {
		t.Logf("Simple: Font loading failed ('%s'), text measurement dependent assertions will be less reliable: %v", constants.FontPath, calcErr)
	}

	if len(lgSimple.ColumnWidths) < 2 {t.Fatalf("Simple: Expected at least 2 col widths, got %d", len(lgSimple.ColumnWidths))}
	if lgSimple.ColumnWidths[0] < constants.MinCellWidth && !fontLoadFailed {
		t.Errorf("Simple: Col0 width %f less than min %f", lgSimple.ColumnWidths[0], constants.MinCellWidth)
	}
	if lgSimple.ColumnWidths[1] < constants.MinCellWidth && !fontLoadFailed {
		t.Errorf("Simple: Col1 width %f less than min %f", lgSimple.ColumnWidths[1], constants.MinCellWidth)
	}

	if !fontLoadFailed {
		if lgSimple.ColumnWidths[1] <= lgSimple.ColumnWidths[0] {
			t.Errorf("Simple: Expected col1 (%f) to be wider than col0 (%f) due to content", lgSimple.ColumnWidths[1], lgSimple.ColumnWidths[0])
		}
	} else {
		t.Log("Simple: Skipping relative column width check due to font loading issue.")
	}

	if len(lgSimple.RowHeights) < 1 {t.Fatalf("Simple: Expected at least 1 row height, got %d", len(lgSimple.RowHeights))}
	if lgSimple.RowHeights[0] < constants.MinCellHeight {
		t.Errorf("Simple: Row height %f less than min %f", lgSimple.RowHeights[0], constants.MinCellHeight)
	}

	spanCell := newTestLayoutCell("Span", "This cell spans two columns", 2, 1)
	tableColspan := &table.Table{Rows: []table.Row{{Cells: []table.Cell{*spanCell}}}}
	lgColspan, popErrCs := PopulateOccupationMap(tableColspan)
	if popErrCs != nil { t.Fatalf("Colspan: Populate failed: %v", popErrCs) }

	calcErrColspan := lgColspan.CalculateColumnWidthsAndRowHeights(constants)
	fontLoadFailedColspan := (calcErrColspan != nil && strings.Contains(calcErrColspan.Error(), "failed to load font"))
	if calcErrColspan != nil && !fontLoadFailedColspan {
		t.Fatalf("Colspan: Calculate failed: %v", calcErrColspan)
	}
	if fontLoadFailedColspan {
		t.Logf("Colspan: Font loading failed, text measurement dependent assertions will be less reliable: %v", calcErrColspan)
	}

	if len(lgColspan.ColumnWidths) < 2 { t.Fatalf("Colspan: Expected 2 column widths, got %d", len(lgColspan.ColumnWidths)) }
	totalSpannedWidth := lgColspan.ColumnWidths[0] + lgColspan.ColumnWidths[1]
	if totalSpannedWidth < constants.MinCellWidth-testEpsilonFloat && !fontLoadFailedColspan {
		t.Errorf("Colspan: Total spanned width %f too narrow, min %f", totalSpannedWidth, constants.MinCellWidth)
	}
}

func TestCalculateFinalCellLayouts(t *testing.T) {
	margin := 10.0
	c1 := newTestLayoutCell("C1", "c1", 1, 1)
	c2_span := newTestLayoutCell("C2span", "c2span", 2, 1)

	lg := NewLayoutGrid(1, 3)
	lg.NumLogicalRows = 1 // Explicitly set after NewLayoutGrid if it defaults differently
	lg.NumLogicalCols = 3

	// Manually populate OccupationMap for this specific test
	// Ensure OccupationMap rows are initialized
	if len(lg.OccupationMap) < 1 { lg.OccupationMap = make([][]*table.Cell, 1)}
	if len(lg.OccupationMap[0]) < 3 { lg.OccupationMap[0] = make([]*table.Cell, 3)}

	lg.OccupationMap[0][0] = c1
	lg.OccupationMap[0][1] = c2_span
	lg.OccupationMap[0][2] = c2_span

	lg.ColumnWidths = []float64{50.0, 60.0, 70.0}
	lg.RowHeights = []float64{30.0}
	// lg.NumLogicalRows and lg.NumLogicalCols should be consistent with these slices.
	// NewLayoutGrid(1,3) should handle this, but we override slices for test.

	lg.CalculateFinalCellLayouts(margin)

	expectedCanvasWidth := 50.0 + 60.0 + 70.0 + 2*margin
	if math.Abs(lg.CanvasWidth-expectedCanvasWidth) > testEpsilonFloat {
		t.Errorf("CanvasWidth: expected %f, got %f", expectedCanvasWidth, lg.CanvasWidth)
	}
	// For RowHeights, if lg was created with NewLayoutGrid(1,3), RowHeights has len 1.
	// If it was NewLayoutGrid(2,3) as in prompt, but only RowHeights[0] is used, then sum should be RowHeights[0].
	// Assuming RowHeights = []float64{30.0} means totalRowHeight = 30.0.
	expectedCanvasHeight := 30.0 + 2*margin
	if math.Abs(lg.CanvasHeight-expectedCanvasHeight) > testEpsilonFloat {
		t.Errorf("CanvasHeight: expected %f, got %f", expectedCanvasHeight, lg.CanvasHeight)
	}

	if len(lg.GridCells) != 2 { t.Fatalf("Expected 2 GridCells, got %d", len(lg.GridCells)) }

	foundC1, foundC2span := false, false
	for _, gridCell := range lg.GridCells {
		if gridCell.OriginalCell == c1 {
			foundC1 = true
			if math.Abs(gridCell.X-margin) > testEpsilonFloat { t.Errorf("C1.X: expected %f, got %f", margin, gridCell.X) }
			if math.Abs(gridCell.Y-margin) > testEpsilonFloat { t.Errorf("C1.Y: expected %f, got %f", margin, gridCell.Y) }
			if math.Abs(gridCell.Width-50.0) > testEpsilonFloat { t.Errorf("C1.Width: expected %f, got %f", 50.0, gridCell.Width) }
			if math.Abs(gridCell.Height-30.0) > testEpsilonFloat { t.Errorf("C1.Height: expected %f, got %f", 30.0, gridCell.Height) }
		} else if gridCell.OriginalCell == c2_span {
			foundC2span = true
			expectedX_c2 := margin + 50.0
			if math.Abs(gridCell.X-expectedX_c2) > testEpsilonFloat { t.Errorf("C2span.X: expected %f, got %f", expectedX_c2, gridCell.X) }
			if math.Abs(gridCell.Y-margin) > testEpsilonFloat { t.Errorf("C2span.Y: expected %f, got %f", margin, gridCell.Y) }
			expectedWidth_c2 := 60.0 + 70.0
			if math.Abs(gridCell.Width-expectedWidth_c2) > testEpsilonFloat { t.Errorf("C2span.Width: expected %f, got %f", expectedWidth_c2, gridCell.Width) }
			if math.Abs(gridCell.Height-30.0) > testEpsilonFloat { t.Errorf("C2span.Height: expected %f, got %f", 30.0, gridCell.Height) }
		}
	}
	if !foundC1 { t.Error("Cell C1 not found in GridCells") }
	if !foundC2span { t.Error("Cell C2span not found in GridCells") }
}
