package renderer

import (
	"diagramgen/pkg/table"
	"math"    // For float comparisons
	"strings" // For TestCalculateColumnWidthsAndRowHeights font error check
	"testing"
)

const epsilon_layout_test = 0.1
var defaultFontPath_layout_test = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"

func newTestLayoutCellForLayoutTest(title, content string, colspan, rowspan int) *table.Cell {
	cell := table.NewCell(title, content)
	cell.Colspan = colspan
	cell.Rowspan = rowspan
	return &cell
}

func TestPopulateOccupationMap(t *testing.T) {
	// Test Case 1: Simple 2x2 table (no change in expectation)
	table1 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R0C0", "c1", 1, 1), *newTestLayoutCellForLayoutTest("R0C1", "c2", 1, 1)}},
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R1C0", "c3", 1, 1), *newTestLayoutCellForLayoutTest("R1C1", "c4", 1, 1)}},
	}}
	lg1, _ := PopulateOccupationMap(table1)
	if lg1.NumLogicalRows != 2 { t.Errorf("table1: exp 2 rows, got %d", lg1.NumLogicalRows) }
	if lg1.NumLogicalCols != 2 { t.Errorf("table1: exp 2 cols, got %d", lg1.NumLogicalCols) }
	if lg1.OccupationMap[0][0] != &table1.Rows[0].Cells[0] { t.Error("table1: (0,0)") }
	if lg1.OccupationMap[0][1] != &table1.Rows[0].Cells[1] { t.Error("table1: (0,1)") }
	if lg1.OccupationMap[1][0] != &table1.Rows[1].Cells[0] { t.Error("table1: (1,0)") }
	if lg1.OccupationMap[1][1] != &table1.Rows[1].Cells[1] { t.Error("table1: (1,1)") }

	// Test Case 2: Colspan (no change in expectation for NumLogicalCols from this)
	table2 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R0C0s2", "span2col", 2, 1), *newTestLayoutCellForLayoutTest("R0C2", "after", 1, 1)}},
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R1C0", "r1c0", 1, 1), *newTestLayoutCellForLayoutTest("R1C1", "r1c1", 1, 1), *newTestLayoutCellForLayoutTest("R1C2", "r1c2", 1, 1)}},
	}}
	lg2, _ := PopulateOccupationMap(table2)
	if lg2.NumLogicalRows != 2 { t.Errorf("table2: exp 2 rows, got %d", lg2.NumLogicalRows) }
	if lg2.NumLogicalCols != 3 { t.Errorf("table2: exp 3 cols, got %d", lg2.NumLogicalCols) }
	if lg2.OccupationMap[0][0] != &table2.Rows[0].Cells[0] || lg2.OccupationMap[0][1] != &table2.Rows[0].Cells[0] { t.Error("table2: (0,0)-(0,1) span") }
    if lg2.OccupationMap[0][2] != &table2.Rows[0].Cells[1] { t.Error("table2: (0,2)")}
    if lg2.OccupationMap[1][0] != &table2.Rows[1].Cells[0] || lg2.OccupationMap[1][1] != &table2.Rows[1].Cells[1] || lg2.OccupationMap[1][2] != &table2.Rows[1].Cells[2] {t.Error("table2: row 1 contents")}

	// Test Case 3: Rowspan (EXPECTING "true skip" behavior -> 2 columns)
	table3 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R0C0_rs2", "span2row", 1, 2), *newTestLayoutCellForLayoutTest("R0C1_adj", "adjacent", 1, 1)}},
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("R1C0_skip", "SKIPPED", 1, 1), *newTestLayoutCellForLayoutTest("R1C1_next", "next in row", 1, 1)}},
	}}
	lg3, _ := PopulateOccupationMap(table3)
	if lg3.NumLogicalRows != 2 {t.Errorf("table3: exp 2 rows, got %d", lg3.NumLogicalRows)}
	if lg3.NumLogicalCols != 2 {t.Errorf("table3: exp 2 cols for true skip, got %d. Map: %v", lg3.NumLogicalCols, lg3.OccupationMap)}
	if lg3.OccupationMap[0][0] != &table3.Rows[0].Cells[0] {t.Error("table3: (0,0)")}
	if lg3.OccupationMap[1][0] != &table3.Rows[0].Cells[0] {t.Error("table3: (1,0) from rowspan")}
	if lg3.OccupationMap[0][1] != &table3.Rows[0].Cells[1] {t.Error("table3: (0,1)")}
	if lg3.OccupationMap[1][1] != &table3.Rows[1].Cells[1] {t.Errorf("table3: Expected R1C1_next at (1,1), got %p instead of %p", lg3.OccupationMap[1][1], &table3.Rows[1].Cells[1])}

    // Test Case 4: Rowspan completely blocks (EXPECTING "true skip" -> 2 columns)
    table4 := &table.Table{ Rows: []table.Row{
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("A_rs2", "A", 1, 2), *newTestLayoutCellForLayoutTest("B", "B", 1, 1)}},
		{Cells: []table.Cell{*newTestLayoutCellForLayoutTest("C_skip", "C", 1, 1), *newTestLayoutCellForLayoutTest("D_next", "D", 1, 1)}},
	}}
    lg4, _ := PopulateOccupationMap(table4)
    if lg4.NumLogicalRows != 2 {t.Errorf("table4: exp 2 rows, got %d", lg4.NumLogicalRows)}
    if lg4.NumLogicalCols != 2 {t.Errorf("table4: exp 2 cols for true skip, got %d", lg4.NumLogicalCols)}
    if lg4.OccupationMap[0][0] != &table4.Rows[0].Cells[0] || lg4.OccupationMap[1][0] != &table4.Rows[0].Cells[0] {t.Error("table4: A span")}
    if lg4.OccupationMap[0][1] != &table4.Rows[0].Cells[1] {t.Error("table4: B")}
    if lg4.OccupationMap[1][1] != &table4.Rows[1].Cells[1] {t.Errorf("table4: Expected D_next at (1,1), got %p", lg4.OccupationMap[1][1])}

    // Test Case 5: User example (Hydrogen/Helium) - True Skip Logic (EXPECTING 4 columns)
    tableEx := &table.Table{ Rows: []table.Row{
            {Cells:[]table.Cell{*newTestLayoutCellForLayoutTest("ID1", "1", 1, 1), *newTestLayoutCellForLayoutTest("NameH", "Hydrogen", 1, 2), *newTestLayoutCellForLayoutTest("SymH", "H", 1, 1), *newTestLayoutCellForLayoutTest("OrigH", "Cavendish", 1, 1)}},
            {Cells:[]table.Cell{*newTestLayoutCellForLayoutTest("IDHe", "2", 1, 1), *newTestLayoutCellForLayoutTest("NameHeSK", "HeliumSKIP", 1, 1), *newTestLayoutCellForLayoutTest("SymHePl", "He", 1, 1), *newTestLayoutCellForLayoutTest("OrigHePl", "Janssen", 1, 1)}},
    }}
    lgEx, _ := PopulateOccupationMap(tableEx)
    if lgEx.NumLogicalRows != 2 {t.Errorf("tableEx: Exp 2 rows, got %d", lgEx.NumLogicalRows)}
    if lgEx.NumLogicalCols != 4 {t.Errorf("tableEx: Exp 4 cols for true skip, got %d. Map: %v", lgEx.NumLogicalCols, lgEx.OccupationMap)}
    if lgEx.OccupationMap[0][0] != &tableEx.Rows[0].Cells[0] {t.Error("tableEx: (0,0) h_id")}
    if lgEx.OccupationMap[0][1] != &tableEx.Rows[0].Cells[1] {t.Error("tableEx: (0,1) h_name_rs2")}
    if lgEx.OccupationMap[0][2] != &tableEx.Rows[0].Cells[2] {t.Error("tableEx: (0,2) h_sym")}
    if lgEx.OccupationMap[0][3] != &tableEx.Rows[0].Cells[3] {t.Error("tableEx: (0,3) h_orig")}
    if lgEx.OccupationMap[1][0] != &tableEx.Rows[1].Cells[0] {t.Errorf("tableEx: (1,0) exp he_id")}
    if lgEx.OccupationMap[1][1] != &tableEx.Rows[0].Cells[1] {t.Errorf("tableEx: (1,1) exp h_name_rs2 (rowspan)")}
    if lgEx.OccupationMap[1][2] != &tableEx.Rows[1].Cells[2] {t.Errorf("tableEx: (1,2) exp he_sym_place")}
    if lgEx.OccupationMap[1][3] != &tableEx.Rows[1].Cells[3] {t.Errorf("tableEx: (1,3) exp he_orig_place")}
    if lgEx.NumLogicalCols > 4 { t.Errorf("tableEx: NumLogicalCols became %d, expected 4", lgEx.NumLogicalCols) }
}

func TestCalculateColumnWidthsAndRowHeights(t *testing.T) {
	constants := LayoutConstants{ FontPath: defaultFontPath_layout_test, FontSize: 12.0, LineHeightMultiplier: 1.4, Padding: 5.0, MinCellWidth: 10.0, MinCellHeight: 10.0 }
	c1 := newTestLayoutCellForLayoutTest("C1", "short", 1, 1); c2 := newTestLayoutCellForLayoutTest("C2", "this is a much longer string", 1, 1)
	tableSimple := &table.Table{ Rows: []table.Row{{Cells: []table.Cell{*c1, *c2}}} }
	lgSimple, popErr := PopulateOccupationMap(tableSimple)
    if popErr != nil { t.Fatalf("Simple: PopulateOccupationMap failed: %v", popErr)}
	calcErr := lgSimple.CalculateColumnWidthsAndRowHeights(constants)
	fontLoadFailed := (calcErr != nil && strings.Contains(calcErr.Error(), "failed to load font"))
	if calcErr != nil && !fontLoadFailed { t.Fatalf("Simple: Calculate failed: %v", calcErr) }
    if fontLoadFailed { t.Logf("Simple: Font loading failed ('%s'), text measurement dependent assertions will be less reliable: %v", constants.FontPath, calcErr) }
	if len(lgSimple.ColumnWidths) < 2 {t.Fatalf("Simple: Expected at least 2 col widths, got %d", len(lgSimple.ColumnWidths))}
	if !fontLoadFailed {
		if lgSimple.ColumnWidths[0] < constants.MinCellWidth { t.Errorf("Simple: Col0 width %f < min %f", lgSimple.ColumnWidths[0], constants.MinCellWidth) }
		if lgSimple.ColumnWidths[1] < constants.MinCellWidth { t.Errorf("Simple: Col1 width %f < min %f", lgSimple.ColumnWidths[1], constants.MinCellWidth) }
		if lgSimple.ColumnWidths[1] <= lgSimple.ColumnWidths[0] { t.Errorf("Simple: Expected col1 (%f) > col0 (%f)", lgSimple.ColumnWidths[1], lgSimple.ColumnWidths[0]) }
	} else { t.Log("Simple: Skipping some column width checks due to font loading issue.") }
	if len(lgSimple.RowHeights) < 1 {t.Fatalf("Simple: Expected at least 1 row height, got %d", len(lgSimple.RowHeights))}
	if lgSimple.RowHeights[0] < constants.MinCellHeight {t.Errorf("Simple: Row height %f < min %f", lgSimple.RowHeights[0], constants.MinCellHeight)}
    spanCell := newTestLayoutCellForLayoutTest("Span", "This cell spans two columns", 2, 1)
    tableColspan := &table.Table{ Rows: []table.Row{{Cells: []table.Cell{*spanCell}}} }
    lgColspan, popErrCs := PopulateOccupationMap(tableColspan)
    if popErrCs != nil { t.Fatalf("Colspan: Populate failed: %v", popErrCs) }
    calcErrColspan := lgColspan.CalculateColumnWidthsAndRowHeights(constants)
	fontLoadFailedColspan := (calcErrColspan != nil && strings.Contains(calcErrColspan.Error(), "failed to load font"))
    if calcErrColspan != nil && !fontLoadFailedColspan { t.Fatalf("Colspan: Calculate failed: %v", calcErrColspan) }
    if fontLoadFailedColspan { t.Logf("Colspan: Font loading failed...") }
	if len(lgColspan.ColumnWidths) < 2 { t.Fatalf("Colspan: Expected 2 column widths, got %d", len(lgColspan.ColumnWidths)) }
    totalSpannedWidth := lgColspan.ColumnWidths[0] + lgColspan.ColumnWidths[1]
    if totalSpannedWidth < constants.MinCellWidth-epsilon_layout_test && !fontLoadFailedColspan {
         t.Errorf("Colspan: Total spanned width %f too narrow, min %f", totalSpannedWidth, constants.MinCellWidth)
    }
}

func TestCalculateFinalCellLayouts(t *testing.T) {
	margin := 10.0; c1 := newTestLayoutCellForLayoutTest("C1", "c1", 1, 1); c2_span := newTestLayoutCellForLayoutTest("C2span", "c2span", 2, 1)
	lg := NewLayoutGrid(1, 3);
	lg.NumLogicalRows = 1; lg.NumLogicalCols = 3;
    if len(lg.OccupationMap) < lg.NumLogicalRows { lg.OccupationMap = make([][]*table.Cell, lg.NumLogicalRows) }
    for i := range lg.OccupationMap { if len(lg.OccupationMap[i]) < lg.NumLogicalCols { lg.OccupationMap[i] = make([]*table.Cell, lg.NumLogicalCols) } }
	lg.OccupationMap[0][0] = c1; lg.OccupationMap[0][1] = c2_span; lg.OccupationMap[0][2] = c2_span
	lg.ColumnWidths = []float64{50.0, 60.0, 70.0}; lg.RowHeights = []float64{30.0}
	lg.CalculateFinalCellLayouts(margin)
	expectedCanvasWidth := 50.0 + 60.0 + 70.0 + 2*margin
	if math.Abs(lg.CanvasWidth - expectedCanvasWidth) > epsilon_layout_test { t.Errorf("CanvasWidth: exp %f, got %f", expectedCanvasWidth, lg.CanvasWidth) }
	expectedCanvasHeight := 30.0 + 2*margin
	if math.Abs(lg.CanvasHeight - expectedCanvasHeight) > epsilon_layout_test { t.Errorf("CanvasHeight: exp %f, got %f", expectedCanvasHeight, lg.CanvasHeight) }
	if len(lg.GridCells) != 2 { t.Fatalf("Exp 2 GridCells, got %d", len(lg.GridCells)) }
	foundC1 := false; foundC2span := false
	for _, gridCell := range lg.GridCells {
		if gridCell.OriginalCell == c1 { foundC1 = true
			if math.Abs(gridCell.X - margin) > epsilon_layout_test { t.Errorf("C1.X exp %f, got %f", margin, gridCell.X) }
			if math.Abs(gridCell.Y - margin) > epsilon_layout_test { t.Errorf("C1.Y exp %f, got %f", margin, gridCell.Y) }
			if math.Abs(gridCell.Width - 50.0) > epsilon_layout_test { t.Errorf("C1.Width exp %f, got %f", 50.0, gridCell.Width) }
			if math.Abs(gridCell.Height - 30.0) > epsilon_layout_test { t.Errorf("C1.Height exp %f, got %f", 30.0, gridCell.Height) }
		}
		if gridCell.OriginalCell == c2_span { foundC2span = true
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
		innerTable := table.Table{
			ID: "inner1",
			Rows: []table.Row{
				{Cells: []table.Cell{table.NewCell("R1C1", "Content"), table.NewCell("R1C2", "More Content")}},
				{Cells: []table.Cell{table.NewCell("R2C1", "Even More"), table.NewCell("R2C2", "Final")}},
			},
			Settings: table.DefaultGlobalSettings(),
		}
		allTablesMap := map[string]table.Table{"inner1": innerTable}
		parentCell := table.Cell{
			IsTableRef: true,
			TableRefID: "inner1",
		}

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
		parentCell := table.Cell{IsTableRef: true, TableRefID: "emptyInner"}

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
		parentCell := table.Cell{IsTableRef: true, TableRefID: "non_existent_id"}

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
		parentCell := table.Cell{IsTableRef: true, TableRefID: "anyID"}

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
		parentCell := table.Cell{IsTableRef: true, TableRefID: ""} // Empty TableRefID

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
