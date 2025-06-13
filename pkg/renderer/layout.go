package renderer

import (
	"diagramgen/pkg/table"
	"fmt"   // For errors
	"log"   // For logging overlaps or calculation issues
	"math"  // For Max
	"github.com/fogleman/gg" // For gg.Context in measurement
)

// GridCellInfo holds the calculated layout information for a single cell.
type GridCellInfo struct {
	OriginalCell *table.Cell
	X            float64
	Y            float64
	Width        float64
	Height       float64
	GridR        int
	GridC        int
}

// LayoutGrid holds all the computed layout information for a table.
type LayoutGrid struct {
	GridCells      []GridCellInfo
	ColumnWidths   []float64
	RowHeights     []float64
	CanvasWidth    float64
	CanvasHeight   float64
	OccupationMap  [][]*table.Cell
	NumLogicalRows int
	NumLogicalCols int
}

// NewLayoutGrid creates and initializes a LayoutGrid.
func NewLayoutGrid(initialEstimatedRows int, initialEstimatedCols int) *LayoutGrid {
	lg := &LayoutGrid{
		NumLogicalRows: initialEstimatedRows,
		NumLogicalCols: initialEstimatedCols,
		GridCells:      make([]GridCellInfo, 0),
	}

	// Initialize slices only if dimensions are greater than 0
	if initialEstimatedCols > 0 {
		lg.ColumnWidths = make([]float64, initialEstimatedCols)
	} else {
		lg.ColumnWidths = make([]float64, 0) // Explicitly empty if 0 cols
	}
	if initialEstimatedRows > 0 {
		lg.RowHeights = make([]float64, initialEstimatedRows)
	} else {
		lg.RowHeights = make([]float64, 0) // Explicitly empty if 0 rows
	}

	lg.OccupationMap = make([][]*table.Cell, initialEstimatedRows)
	for i := 0; i < initialEstimatedRows; i++ {
		if initialEstimatedCols > 0 {
			lg.OccupationMap[i] = make([]*table.Cell, initialEstimatedCols)
		} else {
			lg.OccupationMap[i] = make([]*table.Cell, 0) // Explicitly empty row if 0 cols
		}
	}
	return lg
}

// ensureCapacity expands the LayoutGrid's internal structures.
func (lg *LayoutGrid) ensureCapacity(targetMaxRow, targetMaxCol int) {
	// Ensure row capacity
	if targetMaxRow >= lg.NumLogicalRows {
		newNumLogicalRows := targetMaxRow + 1
		// Expand RowHeights
		if newNumLogicalRows > cap(lg.RowHeights) {
			newRowHeights := make([]float64, newNumLogicalRows, newNumLogicalRows*2)
			copy(newRowHeights, lg.RowHeights)
			lg.RowHeights = newRowHeights
		} else {
			lg.RowHeights = lg.RowHeights[:newNumLogicalRows]
		}

		// Expand OccupationMap rows
		if newNumLogicalRows > cap(lg.OccupationMap) {
			newOccupationMap := make([][]*table.Cell, newNumLogicalRows, newNumLogicalRows*2)
			copy(newOccupationMap, lg.OccupationMap)
			lg.OccupationMap = newOccupationMap
		} else {
			lg.OccupationMap = lg.OccupationMap[:newNumLogicalRows]
		}

		for i := lg.NumLogicalRows; i < newNumLogicalRows; i++ {
			// New rows are created with the current number of logical columns
			if lg.NumLogicalCols > 0 {
				lg.OccupationMap[i] = make([]*table.Cell, lg.NumLogicalCols)
			} else {
				lg.OccupationMap[i] = make([]*table.Cell, 0) // Empty row if no columns yet
			}
		}
		lg.NumLogicalRows = newNumLogicalRows
	}

	// Ensure column capacity (for all rows)
	if targetMaxCol >= lg.NumLogicalCols {
		newNumLogicalCols := targetMaxCol + 1
		// Expand ColumnWidths
		if newNumLogicalCols > cap(lg.ColumnWidths) {
			newColWidths := make([]float64, newNumLogicalCols, newNumLogicalCols*2)
			copy(newColWidths, lg.ColumnWidths)
			lg.ColumnWidths = newColWidths
		} else {
			lg.ColumnWidths = lg.ColumnWidths[:newNumLogicalCols]
		}

		// Expand OccupationMap columns for all existing rows
		for i := 0; i < lg.NumLogicalRows; i++ {
			// If a row was nil (e.g. new row from row expansion not fully initialized for columns yet)
			if lg.OccupationMap[i] == nil {
				if newNumLogicalCols > 0 {
					lg.OccupationMap[i] = make([]*table.Cell, newNumLogicalCols)
				} else {
					lg.OccupationMap[i] = make([]*table.Cell, 0)
				}
			} else if newNumLogicalCols > cap(lg.OccupationMap[i]) {
				newRow := make([]*table.Cell, newNumLogicalCols, newNumLogicalCols*2)
				copy(newRow, lg.OccupationMap[i])
				lg.OccupationMap[i] = newRow
			} else { // Current capacity is sufficient, just extend the slice length
				currentLen := len(lg.OccupationMap[i])
				lg.OccupationMap[i] = lg.OccupationMap[i][:newNumLogicalCols]
				// Ensure newly exposed elements are nil
				for k := currentLen; k < newNumLogicalCols; k++ {
					lg.OccupationMap[i][k] = nil
				}
			}
		}
		lg.NumLogicalCols = newNumLogicalCols
	}
}

// PopulateOccupationMap processes the input table and maps its cells (respecting spans)
// onto a 2D grid representation (OccupationMap) within a LayoutGrid.
func PopulateOccupationMap(inputTable *table.Table) (*LayoutGrid, error) {
	if inputTable == nil {
		return NewLayoutGrid(0, 0), nil
	}
    if len(inputTable.Rows) == 0 {
        return NewLayoutGrid(0,0), nil
    }

	estRows := len(inputTable.Rows)
	estCols := 0
	for _, r := range inputTable.Rows {
		if len(r.Cells) > estCols {
			estCols = len(r.Cells)
		}
	}
    // estCols can remain 0 if all rows are empty. NewLayoutGrid handles this.
	lg := NewLayoutGrid(estRows, estCols)

	for rIdx, inputRow := range inputTable.Rows {
		gridColCursor := 0 // Tracks the next logical grid column index to try for the current inputRow.

		// Ensure the current row rIdx exists in the occupation map.
		// MaxCol is NumLogicalCols-1. If NumLogicalCols is 0, this will pass -1.
		// ensureCapacity should handle targetMaxCol < 0 by doing nothing for column expansion.
		currentMaxColIdx := lg.NumLogicalCols -1
		if currentMaxColIdx < 0 { currentMaxColIdx = 0 } // Ensure at least 0 for initial capacity check
		lg.ensureCapacity(rIdx, currentMaxColIdx)


		for cInputIdx, _ := range inputRow.Cells { // Changed cellToPlaceData to _
			// cellToPlace is a pointer to the cell in the original table structure.
			cellToPlace := &inputTable.Rows[rIdx].Cells[cInputIdx]

			targetGridR := rIdx
			actualTargetGridC := gridColCursor

            // Ensure capacity up to where we *might* scan or place.
            lg.ensureCapacity(targetGridR, actualTargetGridC)

			// Scan for the first free column in targetGridR, starting from actualTargetGridC.
			// Skips slots already occupied by rowspans from cells in previous rows.
			for {
				if actualTargetGridC >= lg.NumLogicalCols { // If we scan past existing columns
					lg.ensureCapacity(targetGridR, actualTargetGridC) // Expand to this column
					break // This new column is free
				}
				if lg.OccupationMap[targetGridR][actualTargetGridC] == nil {
					break // Found a free slot
				}
				actualTargetGridC++ // Slot occupied, try next one
                // Ensure capacity as we scan right, in case we hit the edge of known columns
                // This is implicitly handled by the check actualTargetGridC >= lg.NumLogicalCols at loop start
			}

			// Ensure grid has capacity for this cell's full span starting at (targetGridR, actualTargetGridC)
			lg.ensureCapacity(targetGridR+cellToPlace.Rowspan-1, actualTargetGridC+cellToPlace.Colspan-1)

			// Mark occupation
			for rOffset := 0; rOffset < cellToPlace.Rowspan; rOffset++ {
				for cOffset := 0; cOffset < cellToPlace.Colspan; cOffset++ {
					mapR := targetGridR + rOffset
					mapC := actualTargetGridC + cOffset

					if lg.OccupationMap[mapR][mapC] != nil && lg.OccupationMap[mapR][mapC] != cellToPlace {
						log.Printf("Warning: Overlap! Cell '%s' (input r%d, c%d) at grid (%d,%d) overwriting cell '%s'.",
						    cellToPlace.Title, rIdx, cInputIdx, mapR, mapC, lg.OccupationMap[mapR][mapC].Title)
					}
					lg.OccupationMap[mapR][mapC] = cellToPlace
				}
			}

			gridColCursor = actualTargetGridC + cellToPlace.Colspan
		}
	}
	return lg, nil
}


// LayoutConstants groups parameters needed for layout calculations.
type LayoutConstants struct {
	FontPath             string
	FontSize             float64
	LineHeightMultiplier float64
	Padding              float64
	MinCellWidth         float64
	MinCellHeight        float64
}

// CalculateColumnWidthsAndRowHeights computes optimal column widths and row heights.
func (lg *LayoutGrid) CalculateColumnWidthsAndRowHeights(constants LayoutConstants) error {
	if lg.NumLogicalCols == 0 || lg.NumLogicalRows == 0 {
		return nil
	}

	tempDc := gg.NewContext(1, 1)
	if err := tempDc.LoadFontFace(constants.FontPath, constants.FontSize); err != nil {
		return fmt.Errorf("failed to load font '%s' for layout calculations: %w", constants.FontPath, err)
	}

	type cellGridPos struct{ r, c int }
	uniqueCellPositions := make(map[*table.Cell]cellGridPos)
	processedForPos := make(map[*table.Cell]bool)

	for r := 0; r < lg.NumLogicalRows; r++ {
		for c := 0; c < lg.NumLogicalCols; c++ {
			cell := lg.OccupationMap[r][c]
			if cell != nil && !processedForPos[cell] {
				 firstR, firstC := -1, -1
				 scanBreak:
				 for rr := 0; rr < lg.NumLogicalRows; rr++ {
					 for cc := 0; cc < lg.NumLogicalCols; cc++ {
						 if lg.OccupationMap[rr][cc] == cell {
							 firstR, firstC = rr, cc
							 break scanBreak
						 }
					 }
				 }
				 if firstR != -1 {
					uniqueCellPositions[cell] = cellGridPos{firstR, firstC}
				 }
				processedForPos[cell] = true
			}
		}
	}

	for i := range lg.ColumnWidths { lg.ColumnWidths[i] = 0.0 }

	for cell, pos := range uniqueCellPositions {
		textIdealW, _, err := calculateCellContentSizeInternal(tempDc, cell, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, 10000.0)
		if err != nil {
			log.Printf("Warning (ideal width calc): %v", err)
			continue
		}
		cellFullIdealW := math.Max(textIdealW+(2*constants.Padding), constants.MinCellWidth)

		if cell.Colspan == 1 {
			if cellFullIdealW > lg.ColumnWidths[pos.c] {
				lg.ColumnWidths[pos.c] = cellFullIdealW
			}
		} else {
			currentSpanWidth := 0.0
			for i := 0; i < cell.Colspan; i++ {
				if pos.c+i < lg.NumLogicalCols {
					currentSpanWidth += lg.ColumnWidths[pos.c+i]
				}
			}
			if cellFullIdealW > currentSpanWidth {
				shortfall := cellFullIdealW - currentSpanWidth
				widthToAddPerCol := shortfall / float64(cell.Colspan)
				for i := 0; i < cell.Colspan; i++ {
					if pos.c+i < lg.NumLogicalCols {
						lg.ColumnWidths[pos.c+i] += widthToAddPerCol
					}
				}
			}
		}
	}

	for i := range lg.RowHeights { lg.RowHeights[i] = 0.0 }

	for cell, pos := range uniqueCellPositions {
		currentCellActualDrawingWidth := 0.0
		for i := 0; i < cell.Colspan; i++ {
			if pos.c+i < lg.NumLogicalCols {
				currentCellActualDrawingWidth += lg.ColumnWidths[pos.c+i]
			}
		}

		_, finalTextH, err := calculateCellContentSizeInternal(tempDc, cell, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, currentCellActualDrawingWidth)
		if err != nil {
			log.Printf("Warning (final height calc): %v", err)
			continue
		}
		cellFullFinalH := math.Max(finalTextH+(2*constants.Padding), constants.MinCellHeight)

		if cell.Rowspan == 1 {
			if cellFullFinalH > lg.RowHeights[pos.r] {
				lg.RowHeights[pos.r] = cellFullFinalH
			}
		} else {
			currentSpanHeight := 0.0
            for i := 0; i < cell.Rowspan; i++ {
                if pos.r+i < lg.NumLogicalRows {
                    currentSpanHeight += lg.RowHeights[pos.r+i]
                }
            }
            if cellFullFinalH > currentSpanHeight {
                shortfall := cellFullFinalH - currentSpanHeight
                heightToAddPerRow := shortfall / float64(cell.Rowspan)
                for i := 0; i < cell.Rowspan; i++ {
                    if pos.r+i < lg.NumLogicalRows {
                        lg.RowHeights[pos.r+i] += heightToAddPerRow
                    }
                }
            }
		}
	}
	return nil
}

// calculateCellContentSizeInternal is a helper for layout calculations.
func calculateCellContentSizeInternal(
	dc *gg.Context,
	cell *table.Cell,
	fontSize float64,
	lineHeightMultiplier float64,
	padding float64,
	availableWidthForTextAndPadding float64,
) (textBlockWidth float64, textBlockHeight float64, err error) {

	currentTotalHeight := 0.0
	actualMaxWidthUsed := 0.0
	lineHeight := fontSize * lineHeightMultiplier

	textAvailableWidth := availableWidthForTextAndPadding - (2 * padding)
	if textAvailableWidth < 0 { textAvailableWidth = 0.0 }

	if cell.Title != "" {
		titleText := "[" + cell.Title + "]"
		titleLines := dc.WordWrap(titleText, textAvailableWidth)
		if len(titleLines) == 0 && titleText != "" { titleLines = []string{""} }
		for _, line := range titleLines {
			w, _ := dc.MeasureString(line)
			if w > actualMaxWidthUsed { actualMaxWidthUsed = w }
		}
		currentTotalHeight += float64(len(titleLines)) * lineHeight
	}

	if cell.Content != "" {
		if cell.Title != "" && currentTotalHeight > 0 { currentTotalHeight += lineHeight * 0.25 }
		contentLines := dc.WordWrap(cell.Content, textAvailableWidth)
		if len(contentLines) == 0 && cell.Content != "" { contentLines = []string{""} }
		for _, line := range contentLines {
			w, _ := dc.MeasureString(line)
			if w > actualMaxWidthUsed { actualMaxWidthUsed = w }
		}
		currentTotalHeight += float64(len(contentLines)) * lineHeight
	}
	return actualMaxWidthUsed, currentTotalHeight, nil
}

// CalculateFinalCellLayouts computes the final X, Y, Width, and Height for each unique cell.
func (lg *LayoutGrid) CalculateFinalCellLayouts(margin float64) {
	lg.GridCells = make([]GridCellInfo, 0)

	if lg.NumLogicalCols == 0 || lg.NumLogicalRows == 0 {
		lg.CanvasWidth = margin * 2; if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 }
		lg.CanvasHeight = margin * 2; if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }
		return
	}

	uniqueCellStartPositions := make(map[*table.Cell]struct{ r, c int })
	for r := 0; r < lg.NumLogicalRows; r++ {
		for c := 0; c < lg.NumLogicalCols; c++ {
			cellPtr := lg.OccupationMap[r][c]
			if cellPtr != nil {
				if _, exists := uniqueCellStartPositions[cellPtr]; !exists {
					uniqueCellStartPositions[cellPtr] = struct{ r, c int }{r, c}
				}
			}
		}
	}

	for cell, startPos := range uniqueCellStartPositions {
		currentX := margin
		for i := 0; i < startPos.c; i++ {
			if i < len(lg.ColumnWidths) { currentX += lg.ColumnWidths[i] }
		}
		currentY := margin
		for i := 0; i < startPos.r; i++ {
			if i < len(lg.RowHeights) { currentY += lg.RowHeights[i] }
		}
		cellDrawingWidth := 0.0
		for i := 0; i < cell.Colspan; i++ {
			colIdx := startPos.c + i
			if colIdx < len(lg.ColumnWidths) { cellDrawingWidth += lg.ColumnWidths[colIdx]
			} else { log.Printf("Warning: Col index %d (for cell '%s') out of bounds.", colIdx, cell.Title) }
		}
		cellDrawingHeight := 0.0
		for i := 0; i < cell.Rowspan; i++ {
			rowIdx := startPos.r + i
			if rowIdx < len(lg.RowHeights) { cellDrawingHeight += lg.RowHeights[rowIdx]
			} else { log.Printf("Warning: Row index %d (for cell '%s') out of bounds.", rowIdx, cell.Title) }
		}

		lg.GridCells = append(lg.GridCells, GridCellInfo{
			OriginalCell: cell, X: currentX, Y: currentY,
			Width: cellDrawingWidth, Height: cellDrawingHeight,
			GridR: startPos.r, GridC: startPos.c,
		})
	}

	totalColWidth := 0.0; for _, w := range lg.ColumnWidths { totalColWidth += w }
	lg.CanvasWidth = totalColWidth + (margin * 2); if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 }
	totalRowHeight := 0.0; for _, h := range lg.RowHeights { totalRowHeight += h }
	lg.CanvasHeight = totalRowHeight + (margin * 2); if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }
}
