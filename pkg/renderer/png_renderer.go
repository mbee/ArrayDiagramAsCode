package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	"image/color"
	"log"
	"math"    // For math.Min and math.Round
	"runtime" // Added for OS-dependent font path
	"strings" // Needed for parseHexColor if it uses strings.TrimPrefix
	// "image" // No longer directly needed after subDc.Image() is image.Image
	"github.com/fogleman/gg"
)

// Constants for rendering.
const (
	defaultMargin               = 15.0
	defaultFontPath             = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	defaultFontSize             = 12.0
	defaultLineHeightMultiplier = 1.4
	defaultPadding              = 8.0
	defaultCornerRadius         = 6.0
	defaultMinCellWidth         = 30.0
	defaultMinCellHeight        = 30.0
	epsilon                     = 0.1
)

func parseHexColor(s string) (color.Color, error) {
	if s == "" { return color.Transparent, fmt.Errorf("empty color string") }
	s = strings.TrimPrefix(s, "#")
	var r, g, b uint8
	if len(s) == 3 {
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b); if err != nil { return color.Transparent, fmt.Errorf("short hex: %w", err) }
		r *= 17; g *= 17; b *= 17
	} else if len(s) == 6 {
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b); if err != nil { return color.Transparent, fmt.Errorf("hex: %w", err) }
	} else { return color.Transparent, fmt.Errorf("invalid hex format: %s", s) }
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}

func RenderToPNG(mainTable *table.Table, allTables map[string]table.Table, outputPath string) error {
	if mainTable == nil { return fmt.Errorf("input mainTable is nil") }
	layoutGrid, err := PopulateOccupationMap(mainTable)
	if err != nil { return fmt.Errorf("populate occupation map: %w", err) }

	if layoutGrid.NumLogicalRows == 0 || layoutGrid.NumLogicalCols == 0 {
		dcWidth := int(defaultMargin * 2); if dcWidth < 1 { dcWidth = 1 }
		dcHeight := int(defaultMargin * 2); if dcHeight < 1 { dcHeight = 1 }
		dc := gg.NewContext(dcWidth, dcHeight)
		tableBG := mainTable.Settings.TableBackgroundColor; if tableBG == "" { tableBG = "#FFFFFF" }
		if col, errBg := parseHexColor(tableBG); errBg == nil { dc.SetColor(col) } else { dc.SetColor(color.White) }
		dc.Clear(); log.Println("RenderToPNG: Empty table. Saving minimal image."); return dc.SavePNG(outputPath)
	}

	var osSpecificFontPath string
	switch runtime.GOOS {
	case "darwin": osSpecificFontPath = "/System/Library/Fonts/Geneva.ttf"
	case "linux": osSpecificFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	default: osSpecificFontPath = defaultFontPath; log.Printf("Warning: OS '%s' unsupported, defaulting to %s", runtime.GOOS, osSpecificFontPath)
	}

	layoutConsts := LayoutConstants{
		FontPath: osSpecificFontPath, FontSize: defaultFontSize, LineHeightMultiplier: defaultLineHeightMultiplier,
		Padding: defaultPadding, MinCellWidth: defaultMinCellWidth, MinCellHeight: defaultMinCellHeight,
	}
	if err = layoutGrid.CalculateColumnWidthsAndRowHeights(layoutConsts, allTables); err != nil { return fmt.Errorf("calc sizes: %w", err) }
	layoutGrid.CalculateFinalCellLayouts(defaultMargin)

	canvasW := int(layoutGrid.CanvasWidth); if canvasW <= 0 { canvasW = 1 }
	canvasH := int(layoutGrid.CanvasHeight); if canvasH <= 0 { canvasH = 1 }
	dc := gg.NewContext(canvasW, canvasH)

	canvasBgColorHex := mainTable.Settings.TableBackgroundColor; if canvasBgColorHex == "" { canvasBgColorHex = "#FFFFFF" }
	if col, errBg := parseHexColor(canvasBgColorHex); errBg == nil { dc.SetColor(col) } else { dc.SetColor(color.White) }
	dc.Clear()

	if err = drawTableItself(dc, mainTable, layoutGrid, allTables, layoutConsts); err != nil { return fmt.Errorf("draw main table: %w", err) }
	return dc.SavePNG(outputPath)
}

func drawTableItself(dc *gg.Context, tableToDraw *table.Table, lg *LayoutGrid, allTables map[string]table.Table, lConsts LayoutConstants) error {
	log.Printf("drawTableItself START: Drawing table ID '%s' on dc (size %dx%d)", tableToDraw.ID, dc.Width(), dc.Height())
	if tableToDraw.Settings.TableBackgroundColor != "" {
		if col, err := parseHexColor(tableToDraw.Settings.TableBackgroundColor); err == nil { dc.SetColor(col); dc.Clear()
		} else { log.Printf("Error parsing table BG color '%s' for table '%s': %v", tableToDraw.Settings.TableBackgroundColor, tableToDraw.ID, err) }
	}
	// Note: Font is loaded onto the main dc by RenderToPNG. For subDc in recursion, it's loaded there.
	// If this drawTableItself is called for the main table, font is loaded by the caller.
	// If it's for a sub-table, its subDc needs font loading. (This is handled in the IsTableRef block for subDc)

	for _, gridCell := range lg.GridCells {
		cell := gridCell.OriginalCell
		log.Printf("CELL [%d,%d]: Start. Title:'%s', Content:'%.20s', IsRef:%t. Geom: X:%.1f, Y:%.1f, W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, cell.Title, strings.ReplaceAll(cell.Content, "\n", "\\n"), cell.IsTableRef, gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height)

		cellBgColorHex := cell.BackgroundColor; if cellBgColorHex == "" { cellBgColorHex = tableToDraw.Settings.DefaultCellBackgroundColor }
		if cellBgColorHex == "" { cellBgColorHex = "#FFFFFF" }
		cellBg, _ := parseHexColor(cellBgColorHex); dc.SetColor(cellBg)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius); dc.Fill()

		edgeColorHex := tableToDraw.Settings.EdgeColor; if edgeColorHex == "" { edgeColorHex = "#000000" }
		edgeCol, _ := parseHexColor(edgeColorHex); dc.SetColor(edgeCol)
		edgeThickness := float64(tableToDraw.Settings.EdgeThickness); if edgeThickness <= 0 { edgeThickness = 1.0 }
		dc.SetLineWidth(edgeThickness)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius); dc.Stroke()

		// --- New strategy: Draw content to a temporary context, then draw that context's image ---
		contentAreaX_on_main_dc := gridCell.X + lConsts.Padding
		contentAreaY_on_main_dc := gridCell.Y + lConsts.Padding
		contentAreaW := gridCell.Width - (2 * lConsts.Padding)
		contentAreaH := gridCell.Height - (2 * lConsts.Padding)

		if contentAreaW <= 0 || contentAreaH <= 0 {
			log.Printf("CELL [%d,%d]: Content area is zero or negative (W:%.1f, H:%.1f). Skipping content drawing.", gridCell.GridR, gridCell.GridC, contentAreaW, contentAreaH)
			continue
		}
		roundedContentW, roundedContentH := int(math.Round(contentAreaW)), int(math.Round(contentAreaH))
		if roundedContentW <= 0 || roundedContentH <= 0 {
			log.Printf("CELL [%d,%d]: Rounded content area is zero or negative (W:%d, H:%d). Skipping content drawing.", gridCell.GridR, gridCell.GridC, roundedContentW, roundedContentH)
			continue
		}

		contentDc := gg.NewContext(roundedContentW, roundedContentH)
		if errFont := contentDc.LoadFontFace(lConsts.FontPath, lConsts.FontSize); errFont != nil {
			log.Printf("CELL [%d,%d]: Error loading font for contentDc: %v", gridCell.GridR, gridCell.GridC, errFont)
			// Continue, default font might be used or text might be missing.
		}

		// All drawing coordinates from here are relative to contentDc (origin 0,0)

		if cell.IsTableRef {
			log.Printf("CELL [%d,%d]: IsTableRef TRUE. RefID: '%s'", gridCell.GridR, gridCell.GridC, cell.TableRefID)
			if cell.TableRefID == "" { log.Printf("CELL [%d,%d]: Warning: TableRefID is empty. Skipping.", gridCell.GridR, gridCell.GridC); continue }
			refTable, ok := allTables[cell.TableRefID]
			if !ok { log.Printf("CELL [%d,%d]: Warning: Referenced table ID '%s' not found. Skipping.", gridCell.GridR, gridCell.GridC, cell.TableRefID); continue }

			// Parent content area for alignment/scaling is contentDc's size
			parentEffContentW, parentEffContentH := float64(roundedContentW), float64(roundedContentH)
			log.Printf("CELL [%d,%d]: InnerTable: ID '%s'. ParentEffectiveContentArea W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, refTable.ID, parentEffContentW, parentEffContentH)

			innerLg, mapErr := PopulateOccupationMap(&refTable)
			if mapErr != nil { log.Printf("CELL [%d,%d]: Error populating inner map for '%s': %v. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, mapErr); continue }
			if innerLg.NumLogicalRows == 0 || innerLg.NumLogicalCols == 0 { log.Printf("CELL [%d,%d]: Info: Inner table '%s' is empty. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID); continue }

			calcErr := innerLg.CalculateColumnWidthsAndRowHeights(lConsts, allTables)
			if calcErr != nil { log.Printf("CELL [%d,%d]: Error calculating inner layout for '%s': %v. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, calcErr); continue }
			innerLg.CalculateFinalCellLayouts(0)

			innerDcWidth := int(innerLg.CanvasWidth); innerDcHeight := int(innerLg.CanvasHeight)
			log.Printf("CELL [%d,%d]: InnerTable: Natural canvas size W:%d, H:%d for subDc.", gridCell.GridR, gridCell.GridC, innerDcWidth, innerDcHeight)
			if innerDcWidth <= 0 || innerDcHeight <= 0 { log.Printf("CELL [%d,%d]: Warning: Inner table '%s' zero/neg dims (W:%d, H:%d). Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, innerDcWidth, innerDcHeight); continue }

			subDc := gg.NewContext(innerDcWidth, innerDcHeight) // This is for the inner table's natural size
			drawErr := drawTableItself(subDc, &refTable, innerLg, allTables, lConsts)
			if drawErr != nil { log.Printf("CELL [%d,%d]: Error drawing inner table '%s': %v. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, drawErr); continue }

			naturalInnerTableImage := subDc.Image()
			naturalInnerWidth := float64(naturalInnerTableImage.Bounds().Dx())
			naturalInnerHeight := float64(naturalInnerTableImage.Bounds().Dy())

			if naturalInnerWidth > 0 && naturalInnerHeight > 0 {
				scaledW, scaledH := naturalInnerWidth, naturalInnerHeight
				imageToDraw := naturalInnerTableImage
				performScale := false

				switch cell.InnerTableScaleMode {
				case "fit_width":
					if naturalInnerWidth > 0 && parentEffContentW > 0 {
						scaleFactor := parentEffContentW / naturalInnerWidth
						scaledW = parentEffContentW; scaledH = naturalInnerHeight * scaleFactor; performScale = true
					}
				case "fit_height":
					if naturalInnerHeight > 0 && parentEffContentH > 0 {
						scaleFactor := parentEffContentH / naturalInnerHeight
						scaledH = parentEffContentH; scaledW = naturalInnerWidth * scaleFactor; performScale = true
					}
				case "fit_both":
					if naturalInnerWidth > 0 && naturalInnerHeight > 0 && parentEffContentW > 0 && parentEffContentH > 0 {
						scaleFactorW := parentEffContentW / naturalInnerWidth
						scaleFactorH := parentEffContentH / naturalInnerHeight
						scaleFactor := math.Min(scaleFactorW, scaleFactorH)
						if scaleFactor > 0 { scaledW = naturalInnerWidth * scaleFactor; scaledH = naturalInnerHeight * scaleFactor; performScale = true }
					}
				case "fill_stretch":
					if parentEffContentW > 0 && parentEffContentH > 0 {
						scaledW = parentEffContentW; scaledH = parentEffContentH
						if math.Abs(scaledW-naturalInnerWidth) > epsilon || math.Abs(scaledH-naturalInnerHeight) > epsilon { performScale = true }
					}
				} // Default "none" means scaledW/H remain naturalInnerWidth/Height

				if scaledW <= 0 || scaledH <= 0 { imageToDraw = nil }

				if performScale && imageToDraw != nil && (math.Abs(scaledW-naturalInnerWidth) > epsilon || math.Abs(scaledH-naturalInnerHeight) > epsilon) {
					rSw, rSh := int(math.Round(scaledW)), int(math.Round(scaledH))
					if rSw > 0 && rSh > 0 {
						scaledSubDc := gg.NewContext(rSw, rSh)
						scaledSubDc.Scale(float64(rSw)/naturalInnerWidth, float64(rSh)/naturalInnerHeight)
						scaledSubDc.DrawImage(naturalInnerTableImage, 0, 0)
						imageToDraw = scaledSubDc.Image()
						scaledW, scaledH = float64(imageToDraw.Bounds().Dx()), float64(imageToDraw.Bounds().Dy())
					} else { imageToDraw = nil }
				}
				log.Printf("CELL [%d,%d]: InnerTable: ScaledDims W:%.1f, H:%.1f. AlignMode:'%s', ScaleMode:'%s'", gridCell.GridR, gridCell.GridC, scaledW, scaledH, cell.InnerTableAlignment, cell.InnerTableScaleMode)

				offsetX, offsetY := 0.0, 0.0
				if imageToDraw != nil && scaledW > 0 && scaledH > 0 {
					switch cell.InnerTableAlignment {
					case "top_center": offsetX = (parentEffContentW - scaledW) / 2
					case "top_right": offsetX = parentEffContentW - scaledW
					case "middle_left": offsetY = (parentEffContentH - scaledH) / 2
					case "center", "middle_center": offsetX = (parentEffContentW - scaledW) / 2; offsetY = (parentEffContentH - scaledH) / 2
					case "middle_right": offsetX = parentEffContentW - scaledW; offsetY = (parentEffContentH - scaledH) / 2
					case "bottom_left": offsetY = parentEffContentH - scaledH
					case "bottom_center": offsetX = (parentEffContentW - scaledW) / 2; offsetY = parentEffContentH - scaledH
					case "bottom_right": offsetX = parentEffContentW - scaledW; offsetY = parentEffContentH - scaledH
					}
				}

				if imageToDraw != nil && scaledW > 0 && scaledH > 0 {
					log.Printf("CELL [%d,%d]: InnerTable: Drawing image on contentDc at X:%.1f, Y:%.1f", gridCell.GridR, gridCell.GridC, math.Round(offsetX), math.Round(offsetY))
					contentDc.DrawImage(imageToDraw, int(math.Round(offsetX)), int(math.Round(offsetY)))

					log.Printf("CELL [%d,%d]: InnerTable: Drawing special border on contentDc at X:%.1f, Y:%.1f, W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, math.Round(offsetX), math.Round(offsetY), math.Round(scaledW), math.Round(scaledH))
					borderColHex := tableToDraw.Settings.EdgeColor; if borderColHex == "" { borderColHex = "#000000" }
					parsedBorderCol, errBr := parseHexColor(borderColHex); if errBr != nil { parsedBorderCol = color.Black }
					contentDc.SetColor(parsedBorderCol); contentDc.SetLineWidth(1.0); contentDc.SetDash()
					contentDc.DrawRectangle(math.Round(offsetX), math.Round(offsetY), math.Round(scaledW), math.Round(scaledH)); contentDc.Stroke()
				}
			} else if naturalInnerWidth == 0 || naturalInnerHeight == 0 {
                 log.Printf("CELL [%d,%d]: Info: Inner table '%s' for cell '%s' has zero natural dimensions. Nothing to draw.", gridCell.GridR, gridCell.GridC, refTable.ID, cell.Title)
            }
		} else { // Not IsTableRef - draw text content
			contentDc.SetColor(color.Black) // Assuming text is black
			textStartX := 0.0 // Relative to contentDc
			textAvailableWidth := float64(roundedContentW)

			contentAreaTopY_on_contentDc := 0.0
			contentAreaBottomY_on_contentDc := float64(roundedContentH)
			lineHeight := lConsts.FontSize * lConsts.LineHeightMultiplier
			firstLineBaselineOffsetY := lConsts.FontSize
			currentBaselineY := contentAreaTopY_on_contentDc + firstLineBaselineOffsetY
			titleProcessed := false

			if cell.Title != "" {
				log.Printf("CELL [%d,%d]: Text: Drawing title. Initial baselineY: %.1f", gridCell.GridR, gridCell.GridC, currentBaselineY)
				titleLines := contentDc.WordWrap("["+cell.Title+"]", textAvailableWidth)
				for _, line := range titleLines {
					if currentBaselineY < contentAreaBottomY_on_contentDc+epsilon {
						contentDc.DrawString(line, textStartX, currentBaselineY); currentBaselineY += lineHeight; titleProcessed = true
					} else { break }
				}
			}
			if cell.Content != "" {
				if titleProcessed && (currentBaselineY <= contentAreaBottomY_on_contentDc+epsilon) { currentBaselineY += lineHeight * 0.25 }
				log.Printf("CELL [%d,%d]: Text: Drawing content. Initial baselineY: %.1f", gridCell.GridR, gridCell.GridC, currentBaselineY)
				contentLines := contentDc.WordWrap(cell.Content, textAvailableWidth)
				for _, line := range contentLines {
					if currentBaselineY < contentAreaBottomY_on_contentDc+epsilon {
						contentDc.DrawString(line, textStartX, currentBaselineY); currentBaselineY += lineHeight
					} else { break }
				}
			}
		}
		// Draw the contentDc (with all its drawings) onto the main dc
		dc.DrawImage(contentDc.Image(), int(math.Round(contentAreaX_on_main_dc)), int(math.Round(contentAreaY_on_main_dc)))
	}
	return nil
}
