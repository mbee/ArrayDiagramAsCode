package renderer

import (
	"diagramgen/pkg/table"
	"math"    // For float comparisons
	"strings" // For TestCalculateColumnWidthsAndRowHeights font error check
	"testing"

	"github.com/fogleman/gg" // Added for NewContext
)

const epsilon_layout_test = 0.1
var defaultFontPath_layout_test = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

// Helper to create a new cell with default alignment/scale for testing layout.
func newLayoutTestCell(title, content string, colspan, rowspan int) table.Cell {
	c := table.NewCell(title, content)
	c.Colspan = colspan
	c.Rowspan = rowspan
	// Explicitly set defaults that are now part of NewCell
	c.InnerTableAlignment = "top_left"
	c.InnerTableScaleMode = "none"
	c.FixedWidth = 0.0
	c.FixedHeight = 0.0
	return c
}


func TestPopulateOccupationMap(t *testing.T) {
	// Test Case 1: Simple 2x2 table (no change in expectation)
	table1 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{newLayoutTestCell("R0C0", "c1", 1, 1), newLayoutTestCell("R0C1", "c2", 1, 1)}},
		{Cells: []table.Cell{newLayoutTestCell("R1C0", "c3", 1, 1), newLayoutTestCell("R1C1", "c4", 1, 1)}},
	}}
	lg1, _ := PopulateOccupationMap(table1)
	if lg1.NumLogicalRows != 2 { t.Errorf("table1: exp 2 rows, got %d", lg1.NumLogicalRows) }
	if lg1.NumLogicalCols != 2 { t.Errorf("table1: exp 2 cols, got %d", lg1.NumLogicalCols) }
	if lg1.OccupationMap[0][0].Content != "c1" { t.Error("table1: (0,0)") } // Compare content for simplicity
	if lg1.OccupationMap[0][1].Content != "c2" { t.Error("table1: (0,1)") }
	if lg1.OccupationMap[1][0].Content != "c3" { t.Error("table1: (1,0)") }
	if lg1.OccupationMap[1][1].Content != "c4" { t.Error("table1: (1,1)") }

	// Test Case 2: Colspan (no change in expectation for NumLogicalCols from this)
	table2Cell0 := newLayoutTestCell("R0C0s2", "span2col", 2, 1)
	table2Cell1 := newLayoutTestCell("R0C2", "after", 1, 1)
	table2 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{table2Cell0, table2Cell1}},
		{Cells: []table.Cell{newLayoutTestCell("R1C0", "r1c0", 1, 1), newLayoutTestCell("R1C1", "r1c1", 1, 1), newLayoutTestCell("R1C2", "r1c2", 1, 1)}},
	}}
	lg2, _ := PopulateOccupationMap(table2)
	if lg2.NumLogicalRows != 2 { t.Errorf("table2: exp 2 rows, got %d", lg2.NumLogicalRows) }
	if lg2.NumLogicalCols != 3 { t.Errorf("table2: exp 3 cols, got %d", lg2.NumLogicalCols) }
	if lg2.OccupationMap[0][0] != &table2.Rows[0].Cells[0] || lg2.OccupationMap[0][1] != &table2.Rows[0].Cells[0] { t.Error("table2: (0,0)-(0,1) span") }
    if lg2.OccupationMap[0][2] != &table2.Rows[0].Cells[1] { t.Error("table2: (0,2)")}

	// Test Case 3: Rowspan
	table3Cell0 := newLayoutTestCell("R0C0_rs2", "span2row", 1, 2)
	table3Cell1 := newLayoutTestCell("R0C1_adj", "adjacent", 1, 1)
	table3Cell2 := newLayoutTestCell("R1C0_skip", "SKIPPED", 1, 1) // This cell should be skipped by placement logic
	table3Cell3 := newLayoutTestCell("R1C1_next", "next in row", 1, 1)
	table3 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{table3Cell0, table3Cell1}},
		{Cells: []table.Cell{table3Cell2, table3Cell3}},
	}}
	lg3, _ := PopulateOccupationMap(table3)
	if lg3.NumLogicalRows != 2 {t.Errorf("table3: exp 2 rows, got %d", lg3.NumLogicalRows)}
	if lg3.NumLogicalCols != 2 {t.Errorf("table3: exp 2 cols for true skip, got %d. Map: %v", lg3.NumLogicalCols, lg3.OccupationMap)}
	if lg3.OccupationMap[0][0] != &table3.Rows[0].Cells[0] {t.Error("table3: (0,0)")}
	if lg3.OccupationMap[1][0] != &table3.Rows[0].Cells[0] {t.Error("table3: (1,0) from rowspan")}
	if lg3.OccupationMap[0][1] != &table3.Rows[0].Cells[1] {t.Error("table3: (0,1)")}
	if lg3.OccupationMap[1][1] != &table3.Rows[1].Cells[1] {t.Errorf("table3: Expected R1C1_next at (1,1)")}
}


func TestCalculateColumnWidthsAndRowHeights(t *testing.T) {
	constants := LayoutConstants{ FontPath: defaultFontPath_layout_test, FontSize: 12.0, LineHeightMultiplier: 1.4, Padding: 5.0, MinCellWidth: 10.0, MinCellHeight: 10.0 }
	c1 := newLayoutTestCell("C1", "short", 1, 1);
	c2 := newLayoutTestCell("C2", "this is a much longer string", 1, 1)
	tableSimple := &table.Table{ Rows: []table.Row{{Cells: []table.Cell{c1, c2}}} }
	lgSimple, popErr := PopulateOccupationMap(tableSimple)
    if popErr != nil { t.Fatalf("Simple: PopulateOccupationMap failed: %v", popErr)}
	calcErr := lgSimple.CalculateColumnWidthsAndRowHeights(constants, make(map[string]table.Table))
	fontLoadFailed := (calcErr != nil && strings.Contains(calcErr.Error(), "failed to load font"))
	if calcErr != nil && !fontLoadFailed { t.Fatalf("Simple: Calculate failed: %v", calcErr) }
    if fontLoadFailed { t.Logf("Simple: Font loading failed ('%s'), text measurement dependent assertions will be less reliable: %v", constants.FontPath, calcErr) }

	if len(lgSimple.ColumnWidths) < 2 {t.Fatalf("Simple: Expected at least 2 col widths, got %d", len(lgSimple.ColumnWidths))}
	if !fontLoadFailed {
		dcCalc := gg.NewContext(1,1)
		dcCalc.LoadFontFace(constants.FontPath, constants.FontSize)
		widthC1, _, _ := calculateCellContentSizeInternal(dcCalc, &c1, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, 10000.0, nil, constants)
		expectedWidthC1 := math.Max(widthC1 + (2*constants.Padding), constants.MinCellWidth)
		if !floatEquals(lgSimple.ColumnWidths[0], expectedWidthC1, epsilon_layout_test) {t.Errorf("Simple: Col0 width %f != expected %f", lgSimple.ColumnWidths[0], expectedWidthC1) }

		widthC2, _, _ := calculateCellContentSizeInternal(dcCalc, &c2, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, 10000.0, nil, constants)
		expectedWidthC2 := math.Max(widthC2+(2*constants.Padding), constants.MinCellWidth)
		if !floatEquals(lgSimple.ColumnWidths[1], expectedWidthC2, epsilon_layout_test) { t.Errorf("Simple: Col1 width %f != expected %f", lgSimple.ColumnWidths[1], expectedWidthC2) }

		if lgSimple.ColumnWidths[1] <= lgSimple.ColumnWidths[0] { t.Errorf("Simple: Expected col1 (%f) > col0 (%f)", lgSimple.ColumnWidths[1], lgSimple.ColumnWidths[0]) }
	} else { t.Log("Simple: Skipping some column width checks due to font loading issue.") }

	if len(lgSimple.RowHeights) < 1 {t.Fatalf("Simple: Expected at least 1 row height, got %d", len(lgSimple.RowHeights))}

	if !fontLoadFailed {
		dcCalc := gg.NewContext(1,1)
		dcCalc.LoadFontFace(constants.FontPath, constants.FontSize)
		_, heightC1, _ := calculateCellContentSizeInternal(dcCalc, &c1, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, lgSimple.ColumnWidths[0], nil, constants)
		_, heightC2, _ := calculateCellContentSizeInternal(dcCalc, &c2, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, lgSimple.ColumnWidths[1], nil, constants)
		expectedRowHeight := math.Max(math.Max(heightC1+(2*constants.Padding), heightC2+(2*constants.Padding)), constants.MinCellHeight)
		if !floatEquals(lgSimple.RowHeights[0], expectedRowHeight, epsilon_layout_test) {t.Errorf("Simple: Row height %f != expected %f", lgSimple.RowHeights[0], expectedRowHeight)}
	}


    spanCell := newLayoutTestCell("Span", "This cell spans two columns", 2, 1)
    tableColspan := &table.Table{ Rows: []table.Row{{Cells: []table.Cell{spanCell}}} }
    lgColspan, popErrCs := PopulateOccupationMap(tableColspan)
    if popErrCs != nil { t.Fatalf("Colspan: Populate failed: %v", popErrCs) }
    calcErrColspan := lgColspan.CalculateColumnWidthsAndRowHeights(constants, make(map[string]table.Table))
	fontLoadFailedColspan := (calcErrColspan != nil && strings.Contains(calcErrColspan.Error(), "failed to load font"))
    if calcErrColspan != nil && !fontLoadFailedColspan { t.Fatalf("Colspan: Calculate failed: %v", calcErrColspan) }
    if fontLoadFailedColspan { t.Logf("Colspan: Font loading failed...") }
	if len(lgColspan.ColumnWidths) < 2 { t.Fatalf("Colspan: Expected 2 column widths, got %d", len(lgColspan.ColumnWidths)) }

	if !fontLoadFailedColspan {
		dcCalc := gg.NewContext(1,1)
		dcCalc.LoadFontFace(constants.FontPath, constants.FontSize)
		spanCellWidth, _, _ := calculateCellContentSizeInternal(dcCalc, &spanCell, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, 10000.0, nil, constants)
		expectedTotalSpannedWidth := math.Max(spanCellWidth+(2*constants.Padding), constants.MinCellWidth*2)
		totalSpannedWidth := lgColspan.ColumnWidths[0] + lgColspan.ColumnWidths[1]
		if math.Abs(totalSpannedWidth - expectedTotalSpannedWidth) > epsilon_layout_test*float64(spanCell.Colspan) {
			 t.Errorf("Colspan: Total spanned width %f too different from expected %f", totalSpannedWidth, expectedTotalSpannedWidth)
		}
	}
}

func TestCalculateColumnWidthsAndRowHeights_WithFixedSizes(t *testing.T) {
	testLayoutConsts := LayoutConstants{
		FontPath:             defaultFontPath_layout_test,
		FontSize:             12.0,
		LineHeightMultiplier: 1.4,
		Padding:              5.0,
		MinCellWidth:         50.0, // Ensure MinCellWidth is clearly different from fixed values
		MinCellHeight:        30.0,
	}
	dc := gg.NewContext(1,1) // For text measurement if needed by non-fixed cells
	if err := dc.LoadFontFace(testLayoutConsts.FontPath, testLayoutConsts.FontSize); err != nil {
		t.Logf("Warning: Font load failed in TestCalculateColumnWidthsAndRowHeights_WithFixedSizes - %v", err)
	}

	t.Run("FixedWidthDictatesColumnWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "text", 1, 1)
		cellB := newLayoutTestCell("B", "fixed", 1, 1)
		cellB.FixedWidth = 150.0

		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA, cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)

		contentWidthA, _, _ := calculateCellContentSizeInternal(dc, &cellA, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, 10000, nil, testLayoutConsts)
		expectedWidthA := math.Max(contentWidthA + (2*testLayoutConsts.Padding), testLayoutConsts.MinCellWidth)

		if !floatEquals(lg.ColumnWidths[0], expectedWidthA, epsilon_layout_test) {
			t.Errorf("Expected Col0 width %.2f, got %.2f", expectedWidthA, lg.ColumnWidths[0])
		}
		if !floatEquals(lg.ColumnWidths[1], 150.0, epsilon_layout_test) {
			t.Errorf("Expected Col1 (CellB FixedWidth) width 150.0, got %.2f", lg.ColumnWidths[1])
		}
	})

	t.Run("FixedHeightDictatesRowHeight", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "text", 1, 1) // Row 0
		cellC := newLayoutTestCell("C", "fixed height", 1, 1) // Row 1
		cellC.FixedHeight = 80.0

		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellC}}}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)

		_, contentHeightA, _ := calculateCellContentSizeInternal(dc, &cellA, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, lg.ColumnWidths[0], nil, testLayoutConsts)
		expectedHeightA := math.Max(contentHeightA + (2*testLayoutConsts.Padding), testLayoutConsts.MinCellHeight)

		if !floatEquals(lg.RowHeights[0], expectedHeightA, epsilon_layout_test) {
			t.Errorf("Expected Row0 height %.2f, got %.2f", expectedHeightA, lg.RowHeights[0])
		}
		if !floatEquals(lg.RowHeights[1], 80.0, epsilon_layout_test) {
			t.Errorf("Expected Row1 (CellC FixedHeight) height 80.0, got %.2f", lg.RowHeights[1])
		}
	})

	t.Run("MaxFixedWidthInColumnWinsOverContent", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "fixed", 1, 1); cellA.FixedWidth = 200.0
		cellB := newLayoutTestCell("B", "short", 1, 1)
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)
		if !floatEquals(lg.ColumnWidths[0], 200.0, epsilon_layout_test) {
			t.Errorf("Expected Col0 width 200.0 (from CellA FixedWidth), got %.2f", lg.ColumnWidths[0])
		}
	})

	t.Run("ContentCanMakeColWiderThanOtherCellFixedWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "fixed", 1, 1); cellA.FixedWidth = 100.0
		cellB := newLayoutTestCell("B", "very very long text that is wider than 100px", 1, 1)
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)

		contentWidthB, _, _ := calculateCellContentSizeInternal(dc, &cellB, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, 10000, nil, testLayoutConsts)
		expectedWidthB := math.Max(contentWidthB + (2*testLayoutConsts.Padding), testLayoutConsts.MinCellWidth)

		if !floatEquals(lg.ColumnWidths[0], expectedWidthB, epsilon_layout_test) {
			t.Errorf("Expected Col0 width %.2f (from CellB content), got %.2f", expectedWidthB, lg.ColumnWidths[0])
		}
	})

	t.Run("FixedWidthWithColspan", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "span text", 2, 1); cellA.FixedWidth = 200.0
		cellB := newLayoutTestCell("B", "end", 1, 1)
		cellC := newLayoutTestCell("C", "c0", 1, 1)
		cellD := newLayoutTestCell("D", "c1", 1, 1)
		cellE := newLayoutTestCell("E", "c2", 1, 1)
		tbl := &table.Table{Rows: []table.Row{
			{Cells: []table.Cell{cellA, cellB}},
			{Cells: []table.Cell{cellC, cellD, cellE}},
		}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)

		if len(lg.ColumnWidths) < 2 { t.Fatalf("Expected at least 2 columns for colspan test, got %d", len(lg.ColumnWidths))}
		sumOfSpannedColumns := lg.ColumnWidths[0] + lg.ColumnWidths[1]
		if sumOfSpannedColumns < 200.0 - epsilon_layout_test { // Sum should be at least the fixed width
			t.Errorf("Expected Col0+Col1 width to be approx 200.0, got %.2f", sumOfSpannedColumns)
		}
	})

	t.Run("FixedWidthLessThanMinCellWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "tiny", 1, 1); cellA.FixedWidth = 20.0
		// testLayoutConsts.MinCellWidth is 50.0
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}}}
		lg, _ := PopulateOccupationMap(tbl)
		_ = lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)
		if !floatEquals(lg.ColumnWidths[0], 20.0, epsilon_layout_test) {
			t.Errorf("Expected Col0 width 20.0 (FixedWidt should override MinCellWidth), got %.2f", lg.ColumnWidths[0])
		}
	})
}


func TestCalculateFinalCellLayouts(t *testing.T) {
	margin := 10.0;
	c1 := newLayoutTestCell("C1", "c1", 1, 1);
	c2_span := newLayoutTestCell("C2span", "c2span", 2, 1)

	lg := NewLayoutGrid(1, 3);
	lg.NumLogicalRows = 1; lg.NumLogicalCols = 3;
    if len(lg.OccupationMap) < lg.NumLogicalRows { lg.OccupationMap = make([][]*table.Cell, lg.NumLogicalRows) }
    for i := range lg.OccupationMap {
		if lg.OccupationMap[i] == nil { lg.OccupationMap[i] = make([]*table.Cell, lg.NumLogicalCols) }
	}

	// Manually assign cells to occupation map for this test
	lg.OccupationMap[0][0] = &c1
	lg.OccupationMap[0][1] = &c2_span
	lg.OccupationMap[0][2] = &c2_span

	lg.ColumnWidths = []float64{50.0, 60.0, 70.0}; lg.RowHeights = []float64{30.0}
	lg.CalculateFinalCellLayouts(margin)
	expectedCanvasWidth := 50.0 + 60.0 + 70.0 + 2*margin
	if math.Abs(lg.CanvasWidth - expectedCanvasWidth) > epsilon_layout_test { t.Errorf("CanvasWidth: exp %f, got %f", expectedCanvasWidth, lg.CanvasWidth) }
	expectedCanvasHeight := 30.0 + 2*margin
	if math.Abs(lg.CanvasHeight - expectedCanvasHeight) > epsilon_layout_test { t.Errorf("CanvasHeight: exp %f, got %f", expectedCanvasHeight, lg.CanvasHeight) }
	if len(lg.GridCells) != 2 { t.Fatalf("Exp 2 GridCells, got %d", len(lg.GridCells)) }
	foundC1 := false; foundC2span := false
	for _, gridCell := range lg.GridCells {
		if gridCell.OriginalCell.Content == c1.Content { foundC1 = true // Compare by content as pointers will differ
			if math.Abs(gridCell.X - margin) > epsilon_layout_test { t.Errorf("C1.X exp %f, got %f", margin, gridCell.X) }
			if math.Abs(gridCell.Y - margin) > epsilon_layout_test { t.Errorf("C1.Y exp %f, got %f", margin, gridCell.Y) }
			if math.Abs(gridCell.Width - 50.0) > epsilon_layout_test { t.Errorf("C1.Width exp %f, got %f", 50.0, gridCell.Width) }
			if math.Abs(gridCell.Height - 30.0) > epsilon_layout_test { t.Errorf("C1.Height exp %f, got %f", 30.0, gridCell.Height) }
		}
		if gridCell.OriginalCell.Content == c2_span.Content { foundC2span = true
			if math.Abs(gridCell.X - (margin + 50.0)) > epsilon_layout_test { t.Errorf("C2span.X exp %f, got %f", margin+50.0, gridCell.X) }
			if math.Abs(gridCell.Y - margin) > epsilon_layout_test { t.Errorf("C2span.Y exp %f, got %f", margin, gridCell.Y) }
			if math.Abs(gridCell.Width - (60.0 + 70.0)) > epsilon_layout_test { t.Errorf("C2span.Width exp %f, got %f", 60.0+70.0, gridCell.Width) }
			if math.Abs(gridCell.Height - 30.0) > epsilon_layout_test { t.Errorf("C2span.Height exp %f, got %f", 30.0, gridCell.Height) }
		}
	}
	if !foundC1 { t.Error("C1 not found") }; if !foundC2span { t.Error("C2span not found") }
}

// floatEquals compares two float64 values with a given epsilon.
func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestCalculateCellContentSizeInternal_TableRef(t *testing.T) {
	// Define LayoutConstants for tests - using values similar to png_renderer defaults
	testLayoutConsts := LayoutConstants{
		FontPath:             defaultFontPath_layout_test, // Defined in this test file
		FontSize:             12.0,
		LineHeightMultiplier: 1.4,
		Padding:              8.0,
		MinCellWidth:         30.0,
		MinCellHeight:        30.0,
	}

	// Dummy gg.Context for text measurement (if needed by inner tables)
	dc := gg.NewContext(1, 1)
	if err := dc.LoadFontFace(testLayoutConsts.FontPath, testLayoutConsts.FontSize); err != nil {
		t.Logf("Warning: could not load font at '%s': %v. Tests relying on precise text measurement might be affected.", testLayoutConsts.FontPath, err)
	}

	// --- Test Case 1: Valid Table Reference ---
	t.Run("ValidTableReference", func(t *testing.T) {
		innerTableCellContent := newLayoutTestCell("R1C1", "Content", 1,1)
		innerTableCellMoreContent := newLayoutTestCell("R1C2", "More Content",1,1)
		innerTableCellEvenMore := newLayoutTestCell("R2C1", "Even More",1,1)
		innerTableCellFinal := newLayoutTestCell("R2C2", "Final",1,1)

		innerTable := table.Table{
			ID: "inner1",
			Rows: []table.Row{
				{Cells: []table.Cell{innerTableCellContent, innerTableCellMoreContent}},
				{Cells: []table.Cell{innerTableCellEvenMore, innerTableCellFinal}},
			},
			Settings: table.DefaultGlobalSettings(),
		}
		allTablesMap := map[string]table.Table{"inner1": innerTable}
		parentCell := newLayoutTestCell("", "", 1,1)
		parentCell.IsTableRef = true
		parentCell.TableRefID = "inner1"


		// Calculate expected dimensions for the inner table
		expectedInnerLg, mapErr := PopulateOccupationMap(&innerTable)
		if mapErr != nil {
			t.Fatalf("Error populating map for expected inner table: %v", mapErr)
		}
		// Note: Pass allTablesMap here in case innerTable itself had nested tables.
		calcErr := expectedInnerLg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if calcErr != nil && !strings.Contains(calcErr.Error(), "failed to load font") { // Ignore font load error for this part
			t.Fatalf("Error calculating column/row sizes for expected inner table: %v", calcErr)
		}
		expectedInnerLg.CalculateFinalCellLayouts(0) // 0 margin for inner table's own canvas
		expectedWidth := expectedInnerLg.CanvasWidth
		expectedHeight := expectedInnerLg.CanvasHeight

		actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &parentCell,
			testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding,
			2000.0, // large availableWidth
			allTablesMap, testLayoutConsts)

		if err != nil {
			t.Errorf("calculateCellContentSizeInternal returned an unexpected error: %v", err)
		}
		if !floatEquals(actualWidth, expectedWidth, epsilon_layout_test) {
			t.Errorf("Expected width %.2f, got %.2f", expectedWidth, actualWidth)
		}
		if !floatEquals(actualHeight, expectedHeight, epsilon_layout_test) {
			t.Errorf("Expected height %.2f, got %.2f", expectedHeight, actualHeight)
		}
	})

	// --- Test Case 2: Empty Inner Table ---
	t.Run("EmptyInnerTable", func(t *testing.T) {
		emptyInnerTable := table.Table{ID: "emptyInner", Settings: table.DefaultGlobalSettings()} // No rows/cols
		allTablesMap := map[string]table.Table{"emptyInner": emptyInnerTable}
		parentCell := newLayoutTestCell("","",1,1)
		parentCell.IsTableRef = true
		parentCell.TableRefID = "emptyInner"


		actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &parentCell,
			testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding,
			2000.0, allTablesMap, testLayoutConsts)

		if err != nil {
			t.Errorf("Expected no error for empty inner table, got: %v", err)
		}
		if !floatEquals(actualWidth, 0.0, epsilon_layout_test) {
			t.Errorf("Expected width 0.0 for empty inner table, got %.2f", actualWidth)
		}
		if !floatEquals(actualHeight, 0.0, epsilon_layout_test) {
			t.Errorf("Expected height 0.0 for empty inner table, got %.2f", actualHeight)
		}
	})

	// --- Test Case 3: Non-Existent TableRefID ---
	t.Run("NonExistentTableRefID", func(t *testing.T) {
		allTablesMap := map[string]table.Table{} // Empty map
		parentCell := newLayoutTestCell("","",1,1)
		parentCell.IsTableRef = true
		parentCell.TableRefID = "non_existent_id"


		expectedFallbackWidth := math.Max(0, testLayoutConsts.MinCellWidth-(2*testLayoutConsts.Padding))
		expectedFallbackHeight := math.Max(0, testLayoutConsts.MinCellHeight-(2*testLayoutConsts.Padding))

		actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &parentCell,
			testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding,
			2000.0, allTablesMap, testLayoutConsts)

		if err != nil {
			t.Errorf("Expected no error for non-existent TableRefID (should use fallback), got: %v", err)
		}
		if !floatEquals(actualWidth, expectedFallbackWidth, epsilon_layout_test) {
			t.Errorf("Expected fallback width %.2f for non-existent ID, got %.2f", expectedFallbackWidth, actualWidth)
		}
		if !floatEquals(actualHeight, expectedFallbackHeight, epsilon_layout_test) {
			t.Errorf("Expected fallback height %.2f for non-existent ID, got %.2f", expectedFallbackHeight, actualHeight)
		}
	})

	// --- Test Case 4: Nil allTablesMap ---
	t.Run("NilAllTablesMap", func(t *testing.T) {
		parentCell := newLayoutTestCell("","",1,1)
		parentCell.IsTableRef = true
		parentCell.TableRefID = "anyID"


		expectedFallbackWidth := math.Max(0, testLayoutConsts.MinCellWidth-(2*testLayoutConsts.Padding))
		expectedFallbackHeight := math.Max(0, testLayoutConsts.MinCellHeight-(2*testLayoutConsts.Padding))

		actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &parentCell,
			testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding,
			2000.0, nil, testLayoutConsts) // Pass nil for allTablesMap

		if err != nil {
			t.Errorf("Expected no error for nil allTablesMap (should use fallback), got: %v", err)
		}
		if !floatEquals(actualWidth, expectedFallbackWidth, epsilon_layout_test) {
			t.Errorf("Expected fallback width %.2f for nil allTablesMap, got %.2f", expectedFallbackWidth, actualWidth)
		}
		if !floatEquals(actualHeight, expectedFallbackHeight, epsilon_layout_test) {
			t.Errorf("Expected fallback height %.2f for nil allTablesMap, got %.2f", expectedFallbackHeight, actualHeight)
		}
	})

	// --- Test Case 5: Empty TableRefID ---
	t.Run("EmptyTableRefID", func(t *testing.T) {
		allTablesMap := map[string]table.Table{}
		parentCell := newLayoutTestCell("","",1,1)
		parentCell.IsTableRef = true
		parentCell.TableRefID = "" // Empty TableRefID


		expectedFallbackWidth := math.Max(0, testLayoutConsts.MinCellWidth-(2*testLayoutConsts.Padding))
		expectedFallbackHeight := math.Max(0, testLayoutConsts.MinCellHeight-(2*testLayoutConsts.Padding))

		actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &parentCell,
			testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding,
			2000.0, allTablesMap, testLayoutConsts)

		if err != nil {
			t.Errorf("Expected no error for empty TableRefID (should use fallback), got: %v", err)
		}
		if !floatEquals(actualWidth, expectedFallbackWidth, epsilon_layout_test) {
			t.Errorf("Expected fallback width %.2f for empty TableRefID, got %.2f", expectedFallbackWidth, actualWidth)
		}
		if !floatEquals(actualHeight, expectedFallbackHeight, epsilon_layout_test) {
			t.Errorf("Expected fallback height %.2f for empty TableRefID, got %.2f", expectedFallbackHeight, actualHeight)
		}
	})
}

func TestCalculateCellContentSizeInternal_Multiline(t *testing.T) {
	testLayoutConsts := LayoutConstants{
		FontPath:             defaultFontPath_layout_test,
		FontSize:             10.0, // Using 10.0 for easier calculation with lineHeightMultiplier
		LineHeightMultiplier: 1.5,  // Results in lineHeight of 15.0
		Padding:              5.0,
		MinCellWidth:         20.0,
		MinCellHeight:        20.0,
	}
	lineHeight := testLayoutConsts.FontSize * testLayoutConsts.LineHeightMultiplier // 15.0

	dc := gg.NewContext(1, 1) // Dummy context for text measurement
	if err := dc.LoadFontFace(testLayoutConsts.FontPath, testLayoutConsts.FontSize); err != nil {
		t.Fatalf("Failed to load font %s: %v. This test requires a valid font.", testLayoutConsts.FontPath, err)
	}

	tests := []struct {
		name                      string
		cellContent               string
		availableWidthForContent  float64 // This is availableWidthForTextAndPadding in func signature
		expectedTextBlockWidth    float64 // Approximate, depends on font
		expectedTextBlockHeight   float64
		expectFontLoadErrorToWarn bool // If true, width assertions might be skipped if font fails
	}{
		{
			name:                    "Single Line",
			cellContent:             "Hello World",
			availableWidthForContent: 200.0,
			// Width of "Hello World" at 10pt. Manually get this via a quick gg run or estimate.
			// For DejaVuSans at 10pt, "Hello World" is approx 60-70px. Let's say 65 for test.
			expectedTextBlockWidth:  func() float64 { w, _ := dc.MeasureString("Hello World"); return w }(),
			expectedTextBlockHeight: lineHeight, // 1 line * 15.0
		},
		{
			name:                    "Two Lines Simple",
			cellContent:             "First line\nSecond line",
			availableWidthForContent: 200.0,
			// Width of "Second line" is likely wider.
			expectedTextBlockWidth:  func() float64 { w, _ := dc.MeasureString("Second line"); return w }(),
			expectedTextBlockHeight: 2 * lineHeight, // 2 lines * 15.0
		},
		{
			name:                    "Three Lines Blank Middle",
			cellContent:             "Line 1\n\nLine 3", // gg.WordWrap likely treats \n\n as one separator -> ["Line 1", "Line 3"]
			availableWidthForContent: 200.0,
			expectedTextBlockWidth:  func() float64 { w, _ := dc.MeasureString("Line 1"); return w }(),
			expectedTextBlockHeight: 2 * lineHeight, // Expect 2 lines if gg.WordWrap collapses \n\n
		},
		{
			name:                    "Word Wrap with Newlines",
			cellContent:             "Short line\nThis is a much longer line that will be wrapped",
			availableWidthForContent: 100.0,
			expectedTextBlockWidth:  0, // Placeholder, will be calculated by test logic below
			expectedTextBlockHeight: 0, // Placeholder, will be calculated by test logic below
		},
		{
			name:                    "Content requiring no wrap but has newline",
			cellContent:             "Line one\nLine two also short",
			availableWidthForContent: 500.0, // Ample width
			expectedTextBlockWidth:  func() float64 { w, _ := dc.MeasureString("Line two also short"); return w }(),
			expectedTextBlockHeight: 2 * lineHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// dc.LoadFontFace is already called once.
			// For each test, create a new cell
			cell := table.NewCell("", tt.cellContent) // Title is empty for these content tests

			// Recalculate expected width for wrapped case more accurately inside test run
			currentExpectedWidth := tt.expectedTextBlockWidth
			currentExpectedHeight := tt.expectedTextBlockHeight

			if tt.name == "Word Wrap with Newlines" {
				// Manually calculate wrapping for this specific case to make test more robust
				textAvailableWidth := tt.availableWidthForContent - (2 * testLayoutConsts.Padding)
				lines := strings.Split(tt.cellContent, "\n")
				totalLines := 0
				maxWidth := 0.0
				for _, part := range lines {
					wrappedLines := dc.WordWrap(part, textAvailableWidth)
					if len(wrappedLines) == 0 && part != "" { wrappedLines = []string{""} } // gg.WordWrap behavior
					totalLines += len(wrappedLines)
					for _, wl := range wrappedLines {
						w, _ := dc.MeasureString(wl)
						if w > maxWidth {
							maxWidth = w
						}
					}
				}
				currentExpectedHeight = float64(totalLines) * lineHeight
				// If any line was wrapped, the width taken would be the max of wrapped lines.
				// If no line was wrapped, it's the max of actual line widths.
				currentExpectedWidth = maxWidth // Use the actual max width of lines after potential wrapping
			}
			// For tests not named "Word Wrap with Newlines", use pre-set expected width.
			if tt.name != "Word Wrap with Newlines" {
			    currentExpectedWidth = tt.expectedTextBlockWidth
			}


			actualWidth, actualHeight, err := calculateCellContentSizeInternal(dc, &cell,
				testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier,
				testLayoutConsts.Padding, tt.availableWidthForContent,
				nil, testLayoutConsts) // allTables is nil, not testing table refs here

			if err != nil {
				t.Errorf("calculateCellContentSizeInternal returned error: %v", err)
			}

			if !floatEquals(actualWidth, currentExpectedWidth, epsilon_layout_test*2) { // Wider epsilon for wrapped width
				t.Errorf("For content '%s':\nExpected textBlockWidth %.2f, got %.2f", tt.cellContent, currentExpectedWidth, actualWidth)
			}
			if !floatEquals(actualHeight, currentExpectedHeight, epsilon_layout_test) {
				t.Errorf("For content '%s':\nExpected textBlockHeight %.2f, got %.2f", tt.cellContent, currentExpectedHeight, actualHeight)
			}
		})
	}
}
// TestCalculateColumnWidthsAndRowHeights_WithFixedSizes tests how fixed cell dimensions
// influence overall column and row size calculations.
func TestCalculateColumnWidthsAndRowHeights_WithFixedSizes(t *testing.T) {
	testLayoutConsts := LayoutConstants{
		FontPath:             defaultFontPath_layout_test,
		FontSize:             12.0,
		LineHeightMultiplier: 1.4,
		Padding:              5.0,  // Standard padding
		MinCellWidth:         50.0, // Min width for non-fixed cells
		MinCellHeight:        30.0, // Min height for non-fixed cells
	}
	dc := gg.NewContext(1, 1) // For text measurement if needed by non-fixed cells
	if err := dc.LoadFontFace(testLayoutConsts.FontPath, testLayoutConsts.FontSize); err != nil {
		t.Logf("Warning: Font load failed in TestCalculateColumnWidthsAndRowHeights_WithFixedSizes - %v. Some text-dependent assertions might be less precise.", err)
	}
	allTablesMap := make(map[string]table.Table) // Empty, as not testing inner table refs here

	t.Run("FixedWidthDictatesColumnWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "text", 1, 1) // Normal cell
		cellB := newLayoutTestCell("B", "fixed", 1, 1)
		cellB.FixedWidth = 150.0

		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA, cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){
			t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)
		}

		contentWidthA, _, _ := calculateCellContentSizeInternal(dc, &cellA, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, 10000, nil, testLayoutConsts)
		expectedWidthA := math.Max(contentWidthA+(2*testLayoutConsts.Padding), testLayoutConsts.MinCellWidth)

		if len(lg.ColumnWidths) < 2 { t.Fatalf("Expected at least 2 column widths, got %d", len(lg.ColumnWidths)) }
		if !floatEquals(lg.ColumnWidths[0], expectedWidthA, epsilon_layout_test) {
			t.Errorf("Expected Col0 width %.2f, got %.2f", expectedWidthA, lg.ColumnWidths[0])
		}
		if !floatEquals(lg.ColumnWidths[1], 150.0, epsilon_layout_test) {
			t.Errorf("Expected Col1 (CellB FixedWidth) width 150.0, got %.2f", lg.ColumnWidths[1])
		}
	})

	t.Run("FixedHeightDictatesRowHeight", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "text", 1, 1)
		cellC := newLayoutTestCell("C", "fixed height", 1, 1)
		cellC.FixedHeight = 80.0

		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellC}}}}
		lg, _ := PopulateOccupationMap(tbl)
		// Need column widths to be calculated first for accurate height calculation of cellA
		tempColWidth := math.Max(testLayoutConsts.MinCellWidth, dc.MeasureString("text") + 2*testLayoutConsts.Padding)
		lg.ColumnWidths[0] = tempColWidth

		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){
			t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)
		}

		_, contentHeightA, _ := calculateCellContentSizeInternal(dc, &cellA, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, lg.ColumnWidths[0], nil, testLayoutConsts)
		expectedHeightA := math.Max(contentHeightA + (2*testLayoutConsts.Padding), testLayoutConsts.MinCellHeight)

		if len(lg.RowHeights) < 2 { t.Fatalf("Expected at least 2 row heights, got %d", len(lg.RowHeights)) }
		if !floatEquals(lg.RowHeights[0], expectedHeightA, epsilon_layout_test) {
			t.Errorf("Expected Row0 height %.2f, got %.2f", expectedHeightA, lg.RowHeights[0])
		}
		if !floatEquals(lg.RowHeights[1], 80.0, epsilon_layout_test) {
			t.Errorf("Expected Row1 (CellC FixedHeight) height 80.0, got %.2f", lg.RowHeights[1])
		}
	})

	t.Run("MaxFixedWidthInColumnWinsOverContent", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "fixed", 1, 1); cellA.FixedWidth = 200.0
		cellB := newLayoutTestCell("B", "short", 1, 1)
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){ t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)}

		if len(lg.ColumnWidths) < 1 { t.Fatalf("Expected at least 1 column width, got %d", len(lg.ColumnWidths)) }
		if !floatEquals(lg.ColumnWidths[0], 200.0, epsilon_layout_test) {
			t.Errorf("Expected Col0 width 200.0 (from CellA FixedWidth), got %.2f", lg.ColumnWidths[0])
		}
	})

	t.Run("ContentCanMakeColWiderThanOtherCellFixedWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "fixed", 1, 1); cellA.FixedWidth = 100.0
		cellBContent := "very very long text that is wider than 100px"
		cellB := newLayoutTestCell("B", cellBContent, 1, 1)
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}, {Cells: []table.Cell{cellB}}}}
		lg, _ := PopulateOccupationMap(tbl)
		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){ t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)}

		contentWidthB, _, _ := calculateCellContentSizeInternal(dc, &cellB, testLayoutConsts.FontSize, testLayoutConsts.LineHeightMultiplier, testLayoutConsts.Padding, 10000, nil, testLayoutConsts)
		expectedWidthB := math.Max(contentWidthB + (2*testLayoutConsts.Padding), testLayoutConsts.MinCellWidth)
		// The column width should be the max of cellA's fixedWidth (100) and cellB's content-derived width.
		finalExpectedColWidth := math.Max(cellA.FixedWidth, expectedWidthB)

		if len(lg.ColumnWidths) < 1 { t.Fatalf("Expected at least 1 column width, got %d", len(lg.ColumnWidths)) }
		if !floatEquals(lg.ColumnWidths[0], finalExpectedColWidth, epsilon_layout_test) {
			t.Errorf("Expected Col0 width %.2f, got %.2f", finalExpectedColWidth, lg.ColumnWidths[0])
		}
	})

	t.Run("FixedWidthWithColspan", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "span text", 2, 1); cellA.FixedWidth = 200.0
		cellB := newLayoutTestCell("B", "end", 1, 1)
		cellC := newLayoutTestCell("C", "c0", 1, 1)
		cellD := newLayoutTestCell("D", "c1", 1, 1)
		cellE := newLayoutTestCell("E", "c2", 1, 1)
		tbl := &table.Table{Rows: []table.Row{
			{Cells: []table.Cell{cellA, cellB}},
			{Cells: []table.Cell{cellC, cellD, cellE}},
		}}
		lg, _ := PopulateOccupationMap(tbl)
		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, allTablesMap)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){ t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)}


		if len(lg.ColumnWidths) < 3 { t.Fatalf("Expected at least 3 columns for colspan test, got %d", len(lg.ColumnWidths))}
		sumOfSpannedColumns := lg.ColumnWidths[0] + lg.ColumnWidths[1]
		// The sum must be AT LEAST FixedWidth. It can be more if other cells in those columns need more space.
		if sumOfSpannedColumns < (200.0 - epsilon_layout_test) {
			t.Errorf("Expected Col0+Col1 width to be at least 200.0, got %.2f", sumOfSpannedColumns)
		}
	})

	t.Run("FixedWidthLessThanMinCellWidth", func(t *testing.T) {
		cellA := newLayoutTestCell("A", "tiny", 1, 1); cellA.FixedWidth = 20.0
		// testLayoutConsts.MinCellWidth is 50.0
		tbl := &table.Table{Rows: []table.Row{{Cells: []table.Cell{cellA}}}}
		lg, _ := PopulateOccupationMap(tbl)
		err := lg.CalculateColumnWidthsAndRowHeights(testLayoutConsts, nil)
		if err != nil && !strings.Contains(err.Error(), "failed to load font"){ t.Fatalf("CalculateColumnWidthsAndRowHeights failed: %v", err)}

		if len(lg.ColumnWidths) < 1 { t.Fatalf("Expected at least 1 column width, got %d", len(lg.ColumnWidths)) }
		if !floatEquals(lg.ColumnWidths[0], 20.0, epsilon_layout_test) {
			t.Errorf("Expected Col0 width 20.0 (FixedWidth should override MinCellWidth), got %.2f", lg.ColumnWidths[0])
		}
	})
}
