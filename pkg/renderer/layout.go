package renderer

import (
	"diagramgen/pkg/table"
	"fmt"   // For errors
	"log"   // For logging overlaps or calculation issues
	"math"  // For Max - Not used in this exact PopulateOccupationMap, but often in other layout funcs
	"github.com/fogleman/gg" // For gg.Context in measurement - Not used here, but in other layout funcs
)

// GridCellInfo, LayoutGrid, NewLayoutGrid, ensureCapacity definitions from the prompt
type GridCellInfo struct {OriginalCell *table.Cell; X, Y, Width, Height float64; GridR, GridC int}
type LayoutGrid struct { GridCells []GridCellInfo; ColumnWidths, RowHeights []float64; CanvasWidth, CanvasHeight float64; OccupationMap [][]*table.Cell; NumLogicalRows, NumLogicalCols int }

func NewLayoutGrid(initialEstimatedRows int, initialEstimatedCols int) *LayoutGrid {
	lg := &LayoutGrid{ NumLogicalRows: initialEstimatedRows, NumLogicalCols: initialEstimatedCols, GridCells: make([]GridCellInfo, 0), }
    if initialEstimatedCols > 0 { lg.ColumnWidths = make([]float64, initialEstimatedCols) } else { lg.ColumnWidths = make([]float64, 0) }
    if initialEstimatedRows > 0 { lg.RowHeights = make([]float64, initialEstimatedRows) } else { lg.RowHeights = make([]float64, 0) }
	lg.OccupationMap = make([][]*table.Cell, initialEstimatedRows)
	for i := 0; i < initialEstimatedRows; i++ {
        if initialEstimatedCols > 0 { lg.OccupationMap[i] = make([]*table.Cell, initialEstimatedCols) } else { lg.OccupationMap[i] = make([]*table.Cell, 0) }
	}
	return lg
}

func (lg *LayoutGrid) ensureCapacity(targetMaxRow, targetMaxCol int) {
    // Row expansion logic
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
            if lg.NumLogicalCols > 0 {
                lg.OccupationMap[i] = make([]*table.Cell, lg.NumLogicalCols)
            } else {
                lg.OccupationMap[i] = make([]*table.Cell, 0)
            }
        }
		lg.NumLogicalRows = newNumLogicalRows
	}
    // Column expansion logic
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
        // Expand columns for all existing rows in OccupationMap
		for i := 0; i < lg.NumLogicalRows; i++ {
            currentLen := 0
            if lg.OccupationMap[i] != nil { currentLen = len(lg.OccupationMap[i]) }

            if lg.OccupationMap[i] == nil { // If row was nil (should not happen if row expansion is correct)
                if newNumLogicalCols > 0 { lg.OccupationMap[i] = make([]*table.Cell, newNumLogicalCols)
                } else { lg.OccupationMap[i] = make([]*table.Cell, 0) }
            } else if newNumLogicalCols > cap(lg.OccupationMap[i]) {
                newRow := make([]*table.Cell, newNumLogicalCols, newNumLogicalCols*2)
			    copy(newRow, lg.OccupationMap[i])
			    lg.OccupationMap[i] = newRow
            } else { // Capacity is enough, just extend length
                lg.OccupationMap[i] = lg.OccupationMap[i][:newNumLogicalCols]
                for k := currentLen; k < newNumLogicalCols; k++ { // Ensure new slots are nil
					lg.OccupationMap[i][k] = nil
				}
            }
		}
		lg.NumLogicalCols = newNumLogicalCols
	}
}

// PopulateOccupationMap with "true skip/overwrite" logic.
func PopulateOccupationMap(inputTable *table.Table) (*LayoutGrid, error) {
	if inputTable == nil || len(inputTable.Rows) == 0 {
		return NewLayoutGrid(0, 0), nil
	}

	estRows := len(inputTable.Rows)
	estCols := 0
	for _, r := range inputTable.Rows {
		currentInputRowCols := 0
		for _, cell := range r.Cells { // Iterate over copies
			currentInputRowCols += cell.Colspan
		}
		if currentInputRowCols > estCols {
			estCols = currentInputRowCols
		}
	}
    if estCols == 0 && estRows > 0 { /* No cells, estCols remains 0 */ }

	lg := NewLayoutGrid(estRows, estCols)

	for rIdx, inputRow := range inputTable.Rows {
		gridColPlacementTarget := 0

        initialTargetColForEnsure := estCols -1
        if initialTargetColForEnsure < 0 { initialTargetColForEnsure = 0 }
        if estCols == 0 && len(inputRow.Cells) > 0 {
            lg.ensureCapacity(rIdx, -1)
        } else {
		    lg.ensureCapacity(rIdx, initialTargetColForEnsure)
        }
        // After this, lg.OccupationMap[rIdx] exists and has length lg.NumLogicalCols (which is estCols or 0)

		for cInputIdx, _ := range inputRow.Cells { // Use _ if cellData not needed directly
            cellToPlace := &inputTable.Rows[rIdx].Cells[cInputIdx] // Use pointer to original cell
			targetGridR := rIdx

			isBlockedByRowspan := false
            // Check within current grid column bounds AND table's natural width (estCols)
			if gridColPlacementTarget < estCols && gridColPlacementTarget < lg.NumLogicalCols &&
			   lg.OccupationMap[targetGridR][gridColPlacementTarget] != nil {
				isBlockedByRowspan = true
			}

			if isBlockedByRowspan {
				// log.Printf("Cell '%s' (input r%d, targetC %d) is blocked by rowspan. Skipping.", cellToPlace.Title, rIdx, gridColPlacementTarget)
				gridColPlacementTarget += cellToPlace.Colspan
				continue
			}

            // Check if placing it (starting at gridColPlacementTarget) would exceed estCols.
            // This check is only meaningful if estCols > 0.
			if estCols > 0 && (gridColPlacementTarget + cellToPlace.Colspan > estCols) {
				// log.Printf("Skipping cell '%s' (input r%d): targetC %d + colspan %d > estCols %d.",
				// 	cellToPlace.Title, rIdx, gridColPlacementTarget, cellToPlace.Colspan, estCols)
				gridColPlacementTarget += cellToPlace.Colspan
				continue
			}

            // If we are here, the cell is to be placed at (targetGridR, gridColPlacementTarget)
			lg.ensureCapacity(targetGridR+cellToPlace.Rowspan-1, gridColPlacementTarget+cellToPlace.Colspan-1)

			for rOffset := 0; rOffset < cellToPlace.Rowspan; rOffset++ {
				for cOffset := 0; cOffset < cellToPlace.Colspan; cOffset++ {
					mapR := targetGridR + rOffset
					mapC := gridColPlacementTarget + cOffset

                    if mapR < lg.NumLogicalRows && mapC < lg.NumLogicalCols && lg.OccupationMap[mapR] != nil && mapC < len(lg.OccupationMap[mapR]) {
					    if lg.OccupationMap[mapR][mapC] != nil && lg.OccupationMap[mapR][mapC] != cellToPlace {
						    // log.Printf("Warning: Overlap! Cell '%s' (input r%d, c%d) at grid (%d,%d) overwriting cell '%s'.",
                            // cellToPlace.Title, rIdx, cInputIdx, mapR, mapC, lg.OccupationMap[mapR][mapC].Title)
					    }
					    lg.OccupationMap[mapR][mapC] = cellToPlace
                    } else {
                        // This should not happen if ensureCapacity and estCols logic is correct.
                        // log.Printf("Error: mapR %d or mapC %d is out of bounds. Grid: %dx%d. Cell: %s", mapR, mapC, lg.NumLogicalRows, lg.NumLogicalCols, cellToPlace.Title)
                    }
				}
			}
			gridColPlacementTarget += cellToPlace.Colspan
		}
	}
    // The final NumLogicalCols should be estCols, unless estCols was 0.
    // If estCols was 0, NumLogicalCols is determined by the cells placed.
    // If estCols > 0, ensureCapacity calls might have expanded NumLogicalCols temporarily
    // if a cell placement target was at the edge, but it should not exceed estCols overall
    // due to the skip conditions. We can enforce NumLogicalCols = estCols here if needed.
    if estCols > 0 && lg.NumLogicalCols != estCols {
        // This might happen if the table is narrower than estCols (e.g. last rows are short)
        // Or if ensureCapacity expanded it due to a scan, but skip conditions prevented use.
        // Forcing it to estCols ensures consistency with the definition of table width.
        // log.Printf("Adjusting NumLogicalCols from %d to estCols %d", lg.NumLogicalCols, estCols)
        lg.NumLogicalCols = estCols
        if len(lg.ColumnWidths) > estCols { lg.ColumnWidths = lg.ColumnWidths[:estCols] }
        for i := range lg.OccupationMap {
            if len(lg.OccupationMap[i]) > estCols {
                lg.OccupationMap[i] = lg.OccupationMap[i][:estCols]
            }
        }
    } else if estCols == 0 { // If table had no cells initially, NumLogicalCols is actual max col placed + 1
        // No adjustment needed, NumLogicalCols reflects actual width.
    }


	return lg, nil
}

// --- Other layout functions (LayoutConstants, CalculateColumnWidthsAndRowHeights, etc.) follow ---
// (Assuming they are present from previous steps and are correct)
type LayoutConstants struct {FontPath string; FontSize, LineHeightMultiplier, Padding, MinCellWidth, MinCellHeight float64}
func (lg *LayoutGrid) CalculateColumnWidthsAndRowHeights(constants LayoutConstants, allTables map[string]table.Table) error {
	if lg.NumLogicalCols == 0 || lg.NumLogicalRows == 0 { return nil }
	tempDc := gg.NewContext(1, 1)
	if err := tempDc.LoadFontFace(constants.FontPath, constants.FontSize); err != nil { return fmt.Errorf("failed to load font '%s': %w", constants.FontPath, err) }
	type cellGridPos struct{ r, c int }; uniqueCellPositions := make(map[*table.Cell]cellGridPos); processedForPos := make(map[*table.Cell]bool)
	for r := 0; r < lg.NumLogicalRows; r++ { for c := 0; c < lg.NumLogicalCols; c++ { cell := lg.OccupationMap[r][c]; if cell != nil && !processedForPos[cell] {
		firstR, firstC := -1, -1; scanBreak: for rr := 0; rr < lg.NumLogicalRows; rr++ { for cc := 0; cc < lg.NumLogicalCols; cc++ { if lg.OccupationMap[rr][cc] == cell { firstR, firstC = rr, cc; break scanBreak }}}
		if firstR != -1 { uniqueCellPositions[cell] = cellGridPos{firstR, firstC} }; processedForPos[cell] = true }}}
	for i := range lg.ColumnWidths { lg.ColumnWidths[i] = 0.0 }
	for cell, pos := range uniqueCellPositions {
		textIdealW, _, err := calculateCellContentSizeInternal(tempDc, cell, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, 10000.0, allTables, constants)
		if err != nil {
			log.Printf("Warning (ideal width calc for cell '%s'): %v", cell.Title, err)
			// Fallback to MinCellWidth if content calculation fails, ensuring textIdealW is for content area
			textIdealW = constants.MinCellWidth - (2*constants.Padding)
			if textIdealW < 0 { textIdealW = 0 }
		}

		var cellFullIdealW float64
		if cell.FixedWidth > 0.0 { // FixedWidth is set
			cellFullIdealW = cell.FixedWidth
		} else { // Not fixed, calculate from content
			cellFullIdealW = math.Max(textIdealW + (2*constants.Padding), constants.MinCellWidth)
		}

		if cell.Colspan == 1 { if cellFullIdealW > lg.ColumnWidths[pos.c] { lg.ColumnWidths[pos.c] = cellFullIdealW }
		} else { currentSpanWidth := 0.0; for i := 0; i < cell.Colspan; i++ { if pos.c+i < lg.NumLogicalCols { currentSpanWidth += lg.ColumnWidths[pos.c+i] } }
			if cellFullIdealW > currentSpanWidth { shortfall := cellFullIdealW - currentSpanWidth; widthToAddPerCol := shortfall / float64(cell.Colspan); for i := 0; i < cell.Colspan; i++ { if pos.c+i < lg.NumLogicalCols { lg.ColumnWidths[pos.c+i] += widthToAddPerCol } }}}}
	for i := range lg.RowHeights { lg.RowHeights[i] = 0.0 }
	for cell, pos := range uniqueCellPositions {
		currentCellActualDrawingWidth := 0.0; for i := 0; i < cell.Colspan; i++ { if pos.c+i < lg.NumLogicalCols { currentCellActualDrawingWidth += lg.ColumnWidths[pos.c+i] } }
		_, finalTextH, err := calculateCellContentSizeInternal(tempDc, cell, constants.FontSize, constants.LineHeightMultiplier, constants.Padding, currentCellActualDrawingWidth, allTables, constants)
		if err != nil {
			log.Printf("Warning (final height calc for cell '%s'): %v", cell.Title, err)
			// Fallback to MinCellHeight if content calculation fails, ensuring finalTextH is for content area
			finalTextH = constants.MinCellHeight - (2*constants.Padding)
			if finalTextH < 0 { finalTextH = 0 }
		}

		var cellFullFinalH float64
		if cell.FixedHeight > 0.0 { // FixedHeight is set
			cellFullFinalH = cell.FixedHeight
		} else { // Not fixed, calculate from content
			cellFullFinalH = math.Max(finalTextH + (2*constants.Padding), constants.MinCellHeight)
		}

		if cell.Rowspan == 1 { if cellFullFinalH > lg.RowHeights[pos.r] { lg.RowHeights[pos.r] = cellFullFinalH }
		} else { currentSpanHeight := 0.0; for i := 0; i < cell.Rowspan; i++ { if pos.r+i < lg.NumLogicalRows { currentSpanHeight += lg.RowHeights[pos.r+i] } }
            if cellFullFinalH > currentSpanHeight { shortfall := cellFullFinalH - currentSpanHeight; heightToAddPerRow := shortfall / float64(cell.Rowspan); for i := 0; i < cell.Rowspan; i++ { if pos.r+i < lg.NumLogicalRows { lg.RowHeights[pos.r+i] += heightToAddPerRow } }}}}
	return nil
}
func calculateCellContentSizeInternal(dc *gg.Context, cell *table.Cell, fontSize, lineHeightMultiplier, padding, availableWidthForTextAndPadding float64, allTables map[string]table.Table, layoutConsts LayoutConstants) (textBlockWidth float64, textBlockHeight float64, err error) {
	if cell.IsTableRef {
		minContentWidth := math.Max(0, layoutConsts.MinCellWidth-(2*layoutConsts.Padding))
		minContentHeight := math.Max(0, layoutConsts.MinCellHeight-(2*layoutConsts.Padding))

		if cell.TableRefID == "" {
			log.Printf("Warning: Cell '%s' (Title) is IsTableRef but TableRefID is empty. Using min content size.", cell.Title)
			return minContentWidth, minContentHeight, nil
		}
		if allTables == nil {
			log.Printf("Warning: allTables map is nil while processing table reference for cell '%s' (Title). Using min content size.", cell.Title)
			return minContentWidth, minContentHeight, nil
		}

		refTable, ok := allTables[cell.TableRefID]
		if !ok {
			log.Printf("Warning: Referenced table ID '%s' not found for cell '%s' (Title). Using min content size.", cell.TableRefID, cell.Title)
			return minContentWidth, minContentHeight, nil
		}

		// Basic self-reference check could be added here if parent table ID were available.
		// For now, proceeding without it.

		// Create and calculate layout for the inner table.
		// Note: Using the same layout constants for the inner table.
		// A different set of constants (e.g., smaller font) could be passed if desired.
		innerLayoutGrid, mapErr := PopulateOccupationMap(&refTable)
		if mapErr != nil {
			return 0, 0, fmt.Errorf("error populating occupation map for inner table '%s' (cell '%s'): %w", refTable.ID, cell.Title, mapErr)
		}

		if innerLayoutGrid.NumLogicalRows == 0 || innerLayoutGrid.NumLogicalCols == 0 {
			log.Printf("Info: Inner table '%s' for cell '%s' (Title) is empty. Using zero content size.", refTable.ID, cell.Title)
			return 0, 0, nil // Represents empty content
		}

		calcErr := innerLayoutGrid.CalculateColumnWidthsAndRowHeights(layoutConsts, allTables) // Recursive call
		if calcErr != nil {
			return 0, 0, fmt.Errorf("error calculating layout for inner table '%s' (cell '%s'): %w", refTable.ID, cell.Title, calcErr)
		}

		// Use 0 margin for inner table calculation, as parent cell's padding handles spacing.
		innerLayoutGrid.CalculateFinalCellLayouts(0)

		calculatedWidth := innerLayoutGrid.CanvasWidth
		calculatedHeight := innerLayoutGrid.CanvasHeight
		log.Printf("Info: Calculated inner table '%s' for cell '%s' (Title): width=%.2f, height=%.2f", refTable.ID, cell.Title, calculatedWidth, calculatedHeight)
		return calculatedWidth, calculatedHeight, nil

	} else {
		// Original text measurement logic for non-reference cells
		currentTotalHeight, actualMaxWidthUsed, lineHeight := 0.0, 0.0, fontSize*lineHeightMultiplier
		textAvailableWidth := availableWidthForTextAndPadding - (2 * padding)
		if textAvailableWidth < 0 {
			textAvailableWidth = 0.0
		}

		if cell.Title != "" {
			titleText := "[" + cell.Title + "]"
			titleLines := dc.WordWrap(titleText, textAvailableWidth)
			if len(titleLines) == 0 && titleText != "" {
				// Handle case where WordWrap returns empty for non-empty string (e.g. very narrow width)
				titleLines = []string{""} // Count as one line
			}
			for _, line := range titleLines {
				w, _ := dc.MeasureString(line)
				if w > actualMaxWidthUsed {
					actualMaxWidthUsed = w
				}
			}
			currentTotalHeight += float64(len(titleLines)) * lineHeight
		}

		if cell.Content != "" {
			if cell.Title != "" && currentTotalHeight > 0 {
				currentTotalHeight += lineHeight * 0.25 // Space between title and content
			}
			contentLines := dc.WordWrap(cell.Content, textAvailableWidth)
			if len(contentLines) == 0 && cell.Content != "" {
				contentLines = []string{""} // Count as one line
			}
			for _, line := range contentLines {
				w, _ := dc.MeasureString(line)
				if w > actualMaxWidthUsed {
					actualMaxWidthUsed = w
				}
			}
			currentTotalHeight += float64(len(contentLines)) * lineHeight
		}
		return actualMaxWidthUsed, currentTotalHeight, nil
	}
}
func (lg *LayoutGrid) CalculateFinalCellLayouts(margin float64) {
	lg.GridCells = make([]GridCellInfo, 0) ; if lg.NumLogicalCols == 0 || lg.NumLogicalRows == 0 { lg.CanvasWidth = margin * 2; if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 }; lg.CanvasHeight = margin * 2; if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }; return }
	uniqueCellStartPositions := make(map[*table.Cell]struct{ r, c int }); for r := 0; r < lg.NumLogicalRows; r++ { for c := 0; c < lg.NumLogicalCols; c++ { cellPtr := lg.OccupationMap[r][c]; if cellPtr != nil { if _, exists := uniqueCellStartPositions[cellPtr]; !exists { uniqueCellStartPositions[cellPtr] = struct{ r, c int }{r, c} } } } }
	for cell, startPos := range uniqueCellStartPositions {
		currentX, currentY, cellDrawingWidth, cellDrawingHeight := margin, margin, 0.0, 0.0
		for i := 0; i < startPos.c; i++ { if i < len(lg.ColumnWidths) { currentX += lg.ColumnWidths[i] } }; for i := 0; i < startPos.r; i++ { if i < len(lg.RowHeights) { currentY += lg.RowHeights[i] } }
		for i := 0; i < cell.Colspan; i++ { colIdx := startPos.c + i; if colIdx < len(lg.ColumnWidths) { cellDrawingWidth += lg.ColumnWidths[colIdx] } else { log.Printf("Warning: Col index %d for cell '%s' out of bounds.", colIdx, cell.Title) } }
		for i := 0; i < cell.Rowspan; i++ { rowIdx := startPos.r + i; if rowIdx < len(lg.RowHeights) { cellDrawingHeight += lg.RowHeights[rowIdx] } else { log.Printf("Warning: Row index %d for cell '%s' out of bounds.", rowIdx, cell.Title) } }
		lg.GridCells = append(lg.GridCells, GridCellInfo{ OriginalCell: cell, X: currentX, Y: currentY, Width: cellDrawingWidth, Height: cellDrawingHeight, GridR: startPos.r, GridC: startPos.c, })
	}
	totalColWidth := 0.0; for _, w := range lg.ColumnWidths { totalColWidth += w }; lg.CanvasWidth = totalColWidth + (margin * 2); if lg.CanvasWidth < 1 { lg.CanvasWidth = 1 }
	totalRowHeight := 0.0; for _, h := range lg.RowHeights { totalRowHeight += h }; lg.CanvasHeight = totalRowHeight + (margin * 2); if lg.CanvasHeight < 1 { lg.CanvasHeight = 1 }
}
