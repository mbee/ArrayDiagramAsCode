package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	// "image" // For image.Image - Not directly used, type is inferred.
	"image/color"
	"log"
	"math"    // For math.Min and math.Round
	"runtime" // Added for OS-dependent font path
	"strings" // Needed for parseHexColor if it uses strings.TrimPrefix

	"github.com/fogleman/gg"
)

// Constants for rendering.
// These could be further refined or moved into table settings or LayoutConstants.
const (
	defaultMargin               = 15.0
	defaultFontPath             = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	defaultFontSize             = 12.0
	defaultLineHeightMultiplier = 1.4 // Multiplier for font size for line height
	defaultPadding              = 8.0
	defaultCornerRadius         = 6.0
	defaultMinCellWidth         = 30.0 // Min width for any cell (content+padding)
	defaultMinCellHeight        = 30.0 // Min height for any cell (content+padding)
	epsilon                     = 0.1  // For float comparisons in text fitting
)

// parseHexColor (ensure this function is available - using manual version from previous steps)
func parseHexColor(s string) (color.Color, error) {
	if s == "" {
		return color.Transparent, fmt.Errorf("empty color string is not a valid color for direct parsing")
	}
	s = strings.TrimPrefix(s, "#")
	var r, g, b uint8
	if len(s) == 3 {
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return color.Transparent, fmt.Errorf("error parsing short hex color %s: %w", s, err)
		}
		r *= 17
		g *= 17
		b *= 17
	} else if len(s) == 6 {
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return color.Transparent, fmt.Errorf("error parsing hex color %s: %w", s, err)
		}
	} else {
		return color.Transparent, fmt.Errorf("invalid hex color string format: %s", s)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

// RenderToPNG generates a PNG image of the table using dynamic layout.
// It now accepts the mainTable to render and allTables for potential future use (e.g., rendering nested tables).
func RenderToPNG(mainTable *table.Table, allTables map[string]table.Table, outputPath string) error {
	if mainTable == nil {
		return fmt.Errorf("input mainTable is nil")
	}

	// --- 1. Layout Calculation Phase ---
	// For now, layout calculation is only for the mainTable.
	// Nested table rendering will require deeper changes in layout calculation.
	layoutGrid, err := PopulateOccupationMap(mainTable) // Assumes mainTable is not nil
	if err != nil {
		return fmt.Errorf("failed to populate occupation map: %w", err)
	}

	if layoutGrid.NumLogicalRows == 0 || layoutGrid.NumLogicalCols == 0 {
		dcWidth := int(defaultMargin * 2)
		if dcWidth < 1 { dcWidth = 1 }
		dcHeight := int(defaultMargin * 2)
		if dcHeight < 1 { dcHeight = 1 }

		dc := gg.NewContext(dcWidth, dcHeight)
		tableBG := mainTable.Settings.TableBackgroundColor
		if tableBG == "" { tableBG = "#FFFFFF" } // Default to white if empty

		if col, errBg := parseHexColor(tableBG); errBg == nil {
			dc.SetColor(col)
		} else {
			log.Printf("Error parsing TableBackgroundColor for empty table '%s': %v. Defaulting to white.", tableBG, errBg)
			dc.SetColor(color.White)
		}
		dc.Clear()
		log.Println("RenderToPNG: Table has no logical rows or columns. Saving minimal image.")
		return dc.SavePNG(outputPath)
	}

	// Use default constants for layout calculation.
	// These could be overridden by table.Settings in the future.

	var osSpecificFontPath string
	switch runtime.GOOS {
	case "darwin":
		osSpecificFontPath = "/System/Library/Fonts/Geneva.ttf" // A common macOS font
	case "linux":
		osSpecificFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	default:
		// Fallback for other OSes (e.g., Windows, BSD).
		// Users on these OSes might need to ensure the font exists at this path
		// or modify the code.
		osSpecificFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf" // Defaulting to Linux path
		log.Printf("Warning: OS '%s' not explicitly supported for font path. Defaulting to: %s. Please ensure this font is available.", runtime.GOOS, osSpecificFontPath)
	}

	layoutConsts := LayoutConstants{
		FontPath:             osSpecificFontPath, // Use OS-specific font path
		FontSize:             defaultFontSize,
		LineHeightMultiplier: defaultLineHeightMultiplier,
		Padding:              defaultPadding,
		MinCellWidth:         defaultMinCellWidth,
		MinCellHeight:        defaultMinCellHeight,
	}

	err = layoutGrid.CalculateColumnWidthsAndRowHeights(layoutConsts, allTables)
	if err != nil {
		return fmt.Errorf("failed to calculate column/row sizes: %w", err)
	}

	layoutGrid.CalculateFinalCellLayouts(defaultMargin)

	// --- 2. Canvas Creation ---
	canvasW := int(layoutGrid.CanvasWidth)
	canvasH := int(layoutGrid.CanvasHeight)
	if canvasW <= 0 { canvasW = int(defaultMargin * 2); if canvasW < 1 {canvasW = 1} }
	if canvasH <= 0 { canvasH = int(defaultMargin * 2); if canvasH < 1 {canvasH = 1} }

	dc := gg.NewContext(canvasW, canvasH)

	// --- 3. Drawing Phase ---
	// Set overall canvas background based on mainTable's settings or default white
	canvasBgColorHex := mainTable.Settings.TableBackgroundColor
	if canvasBgColorHex == "" {
		canvasBgColorHex = "#FFFFFF" // Default canvas background to white
	}
	if col, errBg := parseHexColor(canvasBgColorHex); errBg == nil {
		dc.SetColor(col)
		dc.Clear() // Clear the entire canvas with this color
	} else {
		log.Printf("Error parsing TableBackgroundColor '%s' for main canvas: %v. Using white.", mainTable.Settings.TableBackgroundColor, errBg)
		dc.SetColor(color.White)
		dc.Clear()
	}

	// Call the recursive drawing function for the main table
	err = drawTableItself(dc, mainTable, layoutGrid, allTables, layoutConsts)
	if err != nil {
		return fmt.Errorf("error drawing main table: %w", err)
	}

	return dc.SavePNG(outputPath)
}

// drawTableItself handles the rendering of a given table and its cells, including nested tables.
func drawTableItself(dc *gg.Context, tableToDraw *table.Table, lg *LayoutGrid, allTables map[string]table.Table, lConsts LayoutConstants) error {
	// Set background for the current table being drawn, if specified in its settings.
	// This allows nested tables to have their own distinct backgrounds.
	// If not specified, it remains transparent to what's underneath (parent cell or main canvas).
	if tableToDraw.Settings.TableBackgroundColor != "" {
		if col, err := parseHexColor(tableToDraw.Settings.TableBackgroundColor); err == nil {
			// Create a temporary context to draw this table's background only within its bounds
			// This is tricky because dc is for the parent. We need to draw onto dc at tableToDraw's location.
			// For now, this function assumes `dc` is already offset or is the main canvas for the main table.
			// When called recursively, the subDc is sized for the inner table.
			// So, clearing subDc with its own background is correct.
			// This part is more relevant when subDc is passed.
			// If dc is the main canvas, this logic is fine for the first call if tableToDraw is mainTable.
			// If dc is a sub-context, this will fill that sub-context.
			dc.SetColor(col)
			dc.Clear()
		} else {
			log.Printf("Error parsing TableBackgroundColor '%s' for table '%s': %v. Skipping custom BG.", tableToDraw.Settings.TableBackgroundColor, tableToDraw.ID, err)
		}
	}

	if errFont := dc.LoadFontFace(lConsts.FontPath, lConsts.FontSize); errFont != nil {
		log.Printf("Error loading font '%s' for table '%s': %v. Text rendering may be affected.", lConsts.FontPath, tableToDraw.ID, errFont)
		// Continue rendering even if font loading fails, default font might be used or text might be missing.
	}

	for _, gridCell := range lg.GridCells {
		cell := gridCell.OriginalCell

		// Draw cell background
		cellBgColorHex := cell.BackgroundColor
		if cellBgColorHex == "" { // If cell has no specific color, use table's default cell BG
			cellBgColorHex = tableToDraw.Settings.DefaultCellBackgroundColor
		}
		if cellBgColorHex == "" { // If table's default cell BG is also empty, fallback
			cellBgColorHex = "#FFFFFF" // Ultimate fallback to white for cell background
		}
		cellBg, errBgParse := parseHexColor(cellBgColorHex)
		if errBgParse != nil {
			log.Printf("Error parsing cell background color '%s' for cell '%s' (table '%s'): %v. Defaulting to white.", cellBgColorHex, cell.Title, tableToDraw.ID, errBgParse)
			cellBg = color.White
		}
		dc.SetColor(cellBg)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Fill()

		// Draw cell border
		edgeColorHex := tableToDraw.Settings.EdgeColor
		if edgeColorHex == "" {
			edgeColorHex = "#000000"
		} // Default edge to black
		edgeCol, errEdgeParse := parseHexColor(edgeColorHex)
		if errEdgeParse != nil {
			log.Printf("Error parsing edge color '%s' for table '%s': %v. Defaulting to black.", tableToDraw.Settings.EdgeColor, tableToDraw.ID, errEdgeParse)
			edgeCol = color.Black
		}
		dc.SetColor(edgeCol)
		edgeThickness := float64(tableToDraw.Settings.EdgeThickness)
		if edgeThickness <= 0 {
			edgeThickness = 1.0
		}
		dc.SetLineWidth(edgeThickness)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Stroke()

		// --- Start Clipping ---
		dc.Push()
		contentAreaX := gridCell.X + lConsts.Padding
		contentAreaY := gridCell.Y + lConsts.Padding
		contentAreaWidth := gridCell.Width - (2 * lConsts.Padding)
		contentAreaHeight := gridCell.Height - (2 * lConsts.Padding)

		if contentAreaWidth < 0 { contentAreaWidth = 0 }
		if contentAreaHeight < 0 { contentAreaHeight = 0 }

		dc.DrawRectangle(contentAreaX, contentAreaY, contentAreaWidth, contentAreaHeight)
		dc.Clip()
		// --- All subsequent drawing for this cell's content will be clipped ---

		if cell.IsTableRef {
			log.Printf("Cell '%s' in table '%s' is ref to table '%s'. Drawing inner table.", cell.Title, tableToDraw.ID, cell.TableRefID)
			if cell.TableRefID == "" {
				log.Printf("Warning: Cell '%s' in table '%s' is IsTableRef but TableRefID is empty. Skipping rendering.", cell.Title, tableToDraw.ID)
				continue
			}
			refTable, ok := allTables[cell.TableRefID]
			if !ok {
				log.Printf("Warning: Referenced table ID '%s' not found for cell '%s' in table '%s'. Skipping rendering.", cell.TableRefID, cell.Title, tableToDraw.ID)
				continue
			}
			// TODO: Add check for self-reference or cyclical references if parent table ID is available.

			innerLg, mapErr := PopulateOccupationMap(&refTable)
			if mapErr != nil {
				log.Printf("Error populating occupation map for inner table '%s' (cell '%s' in table '%s'): %v. Skipping.", refTable.ID, cell.Title, tableToDraw.ID, mapErr)
				continue
			}
			if innerLg.NumLogicalRows == 0 || innerLg.NumLogicalCols == 0 {
				log.Printf("Info: Inner table '%s' for cell '%s' in table '%s' is empty. Skipping.", refTable.ID, cell.Title, tableToDraw.ID)
				continue
			}

			calcErr := innerLg.CalculateColumnWidthsAndRowHeights(lConsts, allTables)
			if calcErr != nil {
				log.Printf("Error calculating layout for inner table '%s' (cell '%s' in table '%s'): %v. Skipping.", refTable.ID, cell.Title, tableToDraw.ID, calcErr)
				continue
			}
			innerLg.CalculateFinalCellLayouts(0) // 0 margin for sub-tables

			innerDcWidth := int(innerLg.CanvasWidth)
			innerDcHeight := int(innerLg.CanvasHeight)

			if innerDcWidth <= 0 || innerDcHeight <= 0 {
				log.Printf("Warning: Inner table '%s' for cell '%s' in table '%s' has zero or negative dimensions (W:%d, H:%d). Skipping.", refTable.ID, cell.Title, tableToDraw.ID, innerDcWidth, innerDcHeight)
				continue
			}
			subDc := gg.NewContext(innerDcWidth, innerDcHeight)

			// Recursive call to draw the inner table onto the sub-context
			drawErr := drawTableItself(subDc, &refTable, innerLg, allTables, lConsts)
			if drawErr != nil {
				log.Printf("Error drawing inner table '%s' (cell '%s' in table '%s'): %v. Skipping.", refTable.ID, cell.Title, tableToDraw.ID, drawErr)
				continue
			}
			naturalInnerTableImage := subDc.Image()
			naturalInnerWidth := float64(naturalInnerTableImage.Bounds().Dx())
			naturalInnerHeight := float64(naturalInnerTableImage.Bounds().Dy())

			if naturalInnerWidth == 0 || naturalInnerHeight == 0 {
				log.Printf("Info: Inner table '%s' for cell '%s' has zero dimensions after rendering. Skipping drawing.", refTable.ID, cell.Title)
				// dc.Pop() will be called at the end of the cell's drawing logic
				// No drawing of image or border if inner table is effectively empty
			} else {
				parentContentX := gridCell.X + lConsts.Padding
				parentContentY := gridCell.Y + lConsts.Padding
				parentContentWidth := gridCell.Width - (2 * lConsts.Padding)
				parentContentHeight := gridCell.Height - (2 * lConsts.Padding)
				if parentContentWidth < 0 { parentContentWidth = 0 }
				if parentContentHeight < 0 { parentContentHeight = 0 }

				scaledW := naturalInnerWidth
				scaledH := naturalInnerHeight
				imageToDraw := naturalInnerTableImage
				performScale := false

				switch cell.InnerTableScaleMode {
				case "none":
					// No scaling
				case "fit_width":
					if naturalInnerWidth > 0 && parentContentWidth > 0 {
						scaleFactor := parentContentWidth / naturalInnerWidth
						scaledW = parentContentWidth
						scaledH = naturalInnerHeight * scaleFactor
						performScale = true
					}
				case "fit_height":
					if naturalInnerHeight > 0 && parentContentHeight > 0 {
						scaleFactor := parentContentHeight / naturalInnerHeight
						scaledH = parentContentHeight
						scaledW = naturalInnerWidth * scaleFactor
						performScale = true
					}
				case "fit_both":
					if naturalInnerWidth > 0 && naturalInnerHeight > 0 && parentContentWidth > 0 && parentContentHeight > 0 {
						scaleFactorW := parentContentWidth / naturalInnerWidth
						scaleFactorH := parentContentHeight / naturalInnerHeight
						scaleFactor := math.Min(scaleFactorW, scaleFactorH)
						scaledW = naturalInnerWidth * scaleFactor
						scaledH = naturalInnerHeight * scaleFactor
						performScale = true
					}
				case "fill_stretch":
					if parentContentWidth > 0 && parentContentHeight > 0 {
						scaledW = parentContentWidth
						scaledH = parentContentHeight
						if scaledW != naturalInnerWidth || scaledH != naturalInnerHeight {
							performScale = true
						}
					}
				}

				// Ensure scaled dimensions are not zero or negative if parent content area is zero
				if scaledW < 0 { scaledW = 0 }
				if scaledH < 0 { scaledH = 0 }


				if performScale && (math.Abs(scaledW-naturalInnerWidth) > epsilon || math.Abs(scaledH-naturalInnerHeight) > epsilon) &&
					naturalInnerWidth > 0 && naturalInnerHeight > 0 && scaledW > 0 && scaledH > 0 {

					roundedScaledW := int(math.Round(scaledW))
					roundedScaledH := int(math.Round(scaledH))

					if roundedScaledW > 0 && roundedScaledH > 0 {
						scaledDc := gg.NewContext(roundedScaledW, roundedScaledH)
						scaleFactorX := float64(roundedScaledW) / naturalInnerWidth
						scaleFactorY := float64(roundedScaledH) / naturalInnerHeight
						scaledDc.Scale(scaleFactorX, scaleFactorY)
						scaledDc.DrawImage(naturalInnerTableImage, 0, 0)
						imageToDraw = scaledDc.Image()
						// Update scaledW/H to actual dimensions after potential rounding for context creation
						scaledW = float64(imageToDraw.Bounds().Dx())
						scaledH = float64(imageToDraw.Bounds().Dy())
					} else {
						// If rounded scaled dimensions are zero, effectively don't draw
						imageToDraw = nil
					}
				} else if performScale { // if performScale was true but conditions for scaling image weren't fully met (e.g. target is 0)
					// This can happen if parentContentWidth/Height is 0, leading to scaledW/H being 0
					if scaledW <= 0 || scaledH <= 0 {
						imageToDraw = nil // Don't draw if scaled to zero
					}
				}


				offsetX, offsetY := 0.0, 0.0
				switch cell.InnerTableAlignment {
				case "top_center":
					offsetX = (parentContentWidth - scaledW) / 2
				case "top_right":
					offsetX = parentContentWidth - scaledW
				case "middle_left":
					offsetY = (parentContentHeight - scaledH) / 2
				case "center", "middle_center":
					offsetX = (parentContentWidth - scaledW) / 2
					offsetY = (parentContentHeight - scaledH) / 2
				case "middle_right":
					offsetX = parentContentWidth - scaledW
					offsetY = (parentContentHeight - scaledH) / 2
				case "bottom_left":
					offsetY = parentContentHeight - scaledH
				case "bottom_center":
					offsetX = (parentContentWidth - scaledW) / 2
					offsetY = parentContentHeight - scaledH
				case "bottom_right":
					offsetX = parentContentWidth - scaledW
					offsetY = parentContentHeight - scaledH
				// Default is "top_left", so offsetX=0, offsetY=0
				}

				drawX := parentContentX + offsetX
				drawY := parentContentY + offsetY

				if imageToDraw != nil && scaledW > 0 && scaledH > 0 {
					dc.DrawImage(imageToDraw, int(math.Round(drawX)), int(math.Round(drawY)))

					// Draw special border for the inner table, around the scaled and aligned image
					borderColorHex := tableToDraw.Settings.EdgeColor // Use parent table's edge color
					if borderColorHex == "" {
						borderColorHex = "#000000" // Default to black
					}
					parsedBorderColor, err := parseHexColor(borderColorHex)
					if err != nil {
						log.Printf("Error parsing border color '%s' for inner table frame: %v. Defaulting to black.", borderColorHex, err)
						parsedBorderColor = color.Black
					}
					dc.SetColor(parsedBorderColor)
					dc.SetLineWidth(1.0)
					dc.SetDash() // Ensure solid line

					// Border should be around the actual drawn image at its final position and size
					dc.DrawRectangle(math.Round(drawX), math.Round(drawY), math.Round(scaledW), math.Round(scaledH))
					dc.Stroke()
				}
			}
		} else {
			// Original text rendering logic for non-reference cells
			textColor := color.Black // Or from settings
			dc.SetColor(textColor)

			textStartX := gridCell.X + lConsts.Padding
			textAvailableWidth := gridCell.Width - (2 * lConsts.Padding)
			if textAvailableWidth < 0 {
				textAvailableWidth = 0
			}

			contentAreaTopY := gridCell.Y + lConsts.Padding
			contentAreaBottomY := gridCell.Y + gridCell.Height - lConsts.Padding
			lineHeight := lConsts.FontSize * lConsts.LineHeightMultiplier
			firstLineBaselineOffsetY := lConsts.FontSize // Key adjustment for first line

			currentBaselineY := contentAreaTopY + firstLineBaselineOffsetY

			titleProcessed := false
			if cell.Title != "" {
				titleText := "[" + cell.Title + "]"
				titleLines := dc.WordWrap(titleText, textAvailableWidth)
				for _, line := range titleLines {
					if currentBaselineY < contentAreaBottomY+epsilon { // Check if baseline is within content area
						dc.DrawString(line, textStartX, currentBaselineY)
						currentBaselineY += lineHeight
						titleProcessed = true
					} else {
						break // Stop if baseline goes out of bounds
					}
				}
			}

			if cell.Content != "" {
				if titleProcessed && (currentBaselineY <= contentAreaBottomY+epsilon) { // Add gap only if title was drawn and there's space
					currentBaselineY += lineHeight * 0.25
				}
				contentLines := dc.WordWrap(cell.Content, textAvailableWidth)
				for _, line := range contentLines {
					if currentBaselineY < contentAreaBottomY+epsilon { // Check if baseline is within content area
						dc.DrawString(line, textStartX, currentBaselineY)
						currentBaselineY += lineHeight
					} else {
						break
					}
				}
			}
		}
		// --- End Clipping ---
		dc.Pop()
	}
	return nil
}
