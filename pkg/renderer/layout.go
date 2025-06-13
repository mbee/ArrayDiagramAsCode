package renderer

import (
	"diagramgen/pkg/table"
	"fmt"   // For errors and logging
	"log"   // For logging overlaps or calculation issues
	"math"  // For Max
	"github.com/fogleman/gg" // For gg.Context in measurement
)

// GridCellInfo holds the calculated layout information for a single cell
// that is ready to be drawn on the canvas.
type GridCellInfo struct {
	OriginalCell *table.Cell // Pointer to the original cell data from the input table
	X            float64     // Calculated X position on canvas (top-left corner)
	Y            float64     // Calculated Y position on canvas (top-left corner)
	Width        float64     // Calculated final width for drawing this cell
	Height       float64     // Calculated final height for drawing this cell
	GridR        int         // Top-left logical row index in the grid
	GridC        int         // Top-left logical column index in the grid
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
	if initialEstimatedRows == 0 {
		initialEstimatedCols = 0
	} else if initialEstimatedCols == 0 && initialEstimatedRows > 0 {
		initialEstimatedCols = 1
	}

	lg := &LayoutGrid{
		NumLogicalRows: initialEstimatedRows,
		NumLogicalCols: initialEstimatedCols,
		GridCells:      make([]GridCellInfo, 0),
		ColumnWidths:   make([]float64, initialEstimatedCols),
		RowHeights:     make([]float64, initialEstimatedRows),
	}

	lg.OccupationMap = make([][]*table.Cell, initialEstimatedRows)
	for i := 0; i < initialEstimatedRows; i++ {
		lg.OccupationMap[i] = make([]*table.Cell, initialEstimatedCols)
	}
	return lg
}

// ensureCapacity expands the LayoutGrid's internal structures.
func (lg *LayoutGrid) ensureCapacity(targetMaxRow, targetMaxCol int) {
	if targetMaxRow >= lg.NumLogicalRows {
		newNumLogicalRows := targetMaxRow + 1
		if newNumLogicalRows > cap(lg.RowHeights) {
			newRowHeights := make([]float64, newNumLogicalRows, newNumLogicalRows*2)
			copy(newRowHeights, lg.RowHeights)
			lg.RowHeights = newRowHeights
		} else {
			lg.RowHeights = lg.RowHeights[:newNumLogicalRows]
		}
		if newNumLogicalRows > cap(lg.OccupationMap) {
			newOccupationMap := make([][]*table.Cell, newNumLogicalRows, newNumLogicalRows*2)
			copy(newOccupationMap, lg.OccupationMap)
			lg.OccupationMap = newOccupationMap
		} else {
			lg.OccupationMap = lg.OccupationMap[:newNumLogicalRows]
		}
		for i := lg.NumLogicalRows; i < newNumLogicalRows; i++ {
			lg.OccupationMap[i] = make([]*table.Cell, lg.NumLogicalCols)
		}
		lg.NumLogicalRows = newNumLogicalRows
	}

	if targetMaxCol >= lg.NumLogicalCols {
		newNumLogicalCols := targetMaxCol + 1
		if newNumLogicalCols > cap(lg.ColumnWidths) {
			newColWidths := make([]float64, newNumLogicalCols, newNumLogicalCols*2)
			copy(newColWidths, lg.ColumnWidths)
			lg.ColumnWidths = newColWidths
		} else {
			lg.ColumnWidths = lg.ColumnWidths[:newNumLogicalCols]
		}
		for i := 0; i < lg.NumLogicalRows; i++ {
			currentLen := 0
			if lg.OccupationMap[i] != nil {
				currentLen = len(lg.OccupationMap[i])
			}

			if newNumLogicalCols > cap(lg.OccupationMap[i]) {
				newRow := make([]*table.Cell, newNumLogicalCols, newNumLogicalCols*2)
				if lg.OccupationMap[i] != nil {
					copy(newRow, lg.OccupationMap[i])
				}
				lg.OccupationMap[i] = newRow
			} else {
				if lg.OccupationMap[i] == nil {
					lg.OccupationMap[i] = make([]*table.Cell, newNumLogicalCols)
				} else {
					// Slice to expand. If newNumLogicalCols is larger, this will expose zero values.
					lg.OccupationMap[i] = lg.OccupationMap[i][:newNumLogicalCols]
					// Ensure newly exposed elements are nil (though they should be by default for pointer types)
					for k := currentLen; k < newNumLogicalCols; k++ {
						lg.OccupationMap[i][k] = nil
					}
				}
			}
		}
		lg.NumLogicalCols = newNumLogicalCols
	}
}

// PopulateOccupationMap processes the input table and maps its cells.
func PopulateOccupationMap(inputTable *table.Table) (*LayoutGrid, error) {
	if inputTable == nil {
		return NewLayoutGrid(0, 0), nil
	}
	estRows := len(inputTable.Rows)
	estCols := 0
	if estRows > 0 {
		for _, r := range inputTable.Rows {
			if len(r.Cells) > estCols {
				estCols = len(r.Cells)
			}
		}
	}
	lg := NewLayoutGrid(estRows, estCols)

	for rIdx, inputRow := range inputTable.Rows {
		currentScanC := 0
		for cIdx, cellToPlace := range inputRow.Cells {
			lg.ensureCapacity(rIdx, currentScanC)
			resolvedC := currentScanC
			for {
				if resolvedC >= lg.NumLogicalCols {
					lg.ensureCapacity(rIdx, resolvedC)
					break
				}
				if lg.OccupationMap[rIdx][resolvedC] == nil {
					break
				}
				resolvedC++
			}
			targetMaxRow := rIdx + cellToPlace.Rowspan - 1
			targetMaxCol := resolvedC + cellToPlace.Colspan - 1
			lg.ensureCapacity(targetMaxRow, targetMaxCol)

			for rOffset := 0; rOffset < cellToPlace.Rowspan; rOffset++ {
				for cOffset := 0; cOffset < cellToPlace.Colspan; cOffset++ {
					mapR, mapC := rIdx+rOffset, resolvedC+cOffset
					if lg.OccupationMap[mapR][mapC] != nil {
						log.Printf("Warning: Overlap detected. Cell from input row %d, input col %d (title: '%s') is overwriting existing cell (title: '%s') in grid slot (%d,%d).",
							rIdx, cIdx, cellToPlace.Title, lg.OccupationMap[mapR][mapC].Title, mapR, mapC)
					}
					// Store the address of the cell from the original slice
					lg.OccupationMap[mapR][mapC] = &inputRow.Cells[cIdx]
				}
			}
			currentScanC = resolvedC + cellToPlace.Colspan
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

// CalculateFinalCellLayouts computes the final X, Y, Width, and Height for each unique cell
// on the canvas, based on determined ColumnWidths and RowHeights.
// It populates lg.GridCells and sets lg.CanvasWidth/Height.
func (lg *LayoutGrid) CalculateFinalCellLayouts(margin float64) {
	lg.GridCells = make([]GridCellInfo, 0) // Clear/reset previous layout info

	if lg.NumLogicalCols == 0 || lg.NumLogicalRows == 0 {
		lg.CanvasWidth = margin * 2
		lg.CanvasHeight = margin * 2
		if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 } // Ensure minimum 1px canvas
		if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }
		return
	}

	// Identify unique cells and their top-left starting grid positions.
	uniqueCellStartPositions := make(map[*table.Cell]struct{ r, c int })
	// Iterate OccupationMap to find the first occurrence (top-left) of each cell.
	// This assumes that the first time we encounter a cell pointer during a row-major scan,
	// it's at its top-left defining slot. This holds if PopulateOccupationMap correctly
	// places cells starting from their top-left.
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
		// Calculate X position by summing widths of preceding columns
		currentX := margin
		for i := 0; i < startPos.c; i++ {
			if i < len(lg.ColumnWidths) {
				currentX += lg.ColumnWidths[i]
			}
		}

		// Calculate Y position by summing heights of preceding rows
		currentY := margin
		for i := 0; i < startPos.r; i++ {
			if i < len(lg.RowHeights) {
				currentY += lg.RowHeights[i]
			}
		}

		// Calculate Width for the cell based on spanned columns
		cellDrawingWidth := 0.0
		for i := 0; i < cell.Colspan; i++ {
			colIdx := startPos.c + i
			if colIdx < len(lg.ColumnWidths) {
				cellDrawingWidth += lg.ColumnWidths[colIdx]
			} else {
				log.Printf("Warning: Column index %d (for cell '%s' starting at %d,%d spanning %d cols) out of bounds (%d columns). Using available widths.",
				    colIdx, cell.Title, startPos.r, startPos.c, cell.Colspan, len(lg.ColumnWidths))
			}
		}

		// Calculate Height for the cell based on spanned rows
		cellDrawingHeight := 0.0
		for i := 0; i < cell.Rowspan; i++ {
			rowIdx := startPos.r + i
			if rowIdx < len(lg.RowHeights) {
				cellDrawingHeight += lg.RowHeights[rowIdx]
			} else {
				log.Printf("Warning: Row index %d (for cell '%s' starting at %d,%d spanning %d rows) out of bounds (%d rows). Using available heights.",
				    rowIdx, cell.Title, startPos.r, startPos.c, cell.Rowspan, len(lg.RowHeights))
			}
		}

		gridCell := GridCellInfo{
			OriginalCell: cell,
			X:            currentX,
			Y:            currentY,
			Width:        cellDrawingWidth,
			Height:       cellDrawingHeight,
			GridR:        startPos.r,
			GridC:        startPos.c,
		}
		lg.GridCells = append(lg.GridCells, gridCell)
	}

	// Calculate total canvas width and height
	totalColWidth := 0.0
	for _, w := range lg.ColumnWidths {
		totalColWidth += w
	}
	lg.CanvasWidth = totalColWidth + (margin * 2)

	totalRowHeight := 0.0
	for _, h := range lg.RowHeights {
		totalRowHeight += h
	}
	lg.CanvasHeight = totalRowHeight + (margin * 2)

    if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 } // Ensure minimum 1px canvas
    if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }
}
