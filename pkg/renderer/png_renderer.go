package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	"image/color"
	"log"
	"math"    // For math.Min and math.Round
	"runtime" // Added for OS-dependent font path
	"strings" // Needed for parseHexColor if it uses strings.TrimPrefix

	"github.com/fogleman/gg"
)

// Constants for rendering.
const (
	defaultMargin               = 15.0
	defaultFontPath             = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf" // Used by OS-specific logic if needed
	defaultFontSize             = 12.0
	defaultLineHeightMultiplier = 1.4
	defaultPadding              = 8.0
	defaultCornerRadius         = 6.0
	defaultMinCellWidth         = 30.0
	defaultMinCellHeight        = 30.0
	epsilon                     = 0.1
)

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
		r *= 17; g *= 17; b *= 17
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

func RenderToPNG(mainTable *table.Table, allTables map[string]table.Table, outputPath string) error {
	if mainTable == nil {
		return fmt.Errorf("input mainTable is nil")
	}

	layoutGrid, err := PopulateOccupationMap(mainTable)
	if err != nil {
		return fmt.Errorf("failed to populate occupation map: %w", err)
	}

	if layoutGrid.NumLogicalRows == 0 || layoutGrid.NumLogicalCols == 0 {
		dcWidth := int(defaultMargin * 2); if dcWidth < 1 { dcWidth = 1 }
		dcHeight := int(defaultMargin * 2); if dcHeight < 1 { dcHeight = 1 }
		dc := gg.NewContext(dcWidth, dcHeight)
		tableBG := mainTable.Settings.TableBackgroundColor
		if tableBG == "" { tableBG = "#FFFFFF" }
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

	var osSpecificFontPath string
	switch runtime.GOOS {
	case "darwin":
		osSpecificFontPath = "/System/Library/Fonts/Geneva.ttf"
	case "linux":
		osSpecificFontPath = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
	default:
		osSpecificFontPath = defaultFontPath
		log.Printf("Warning: OS '%s' not explicitly supported for font path. Defaulting to: %s. Please ensure this font is available.", runtime.GOOS, osSpecificFontPath)
	}

	layoutConsts := LayoutConstants{
		FontPath:             osSpecificFontPath,
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

	canvasW := int(layoutGrid.CanvasWidth); if canvasW <=0 { canvasW =1 }
	canvasH := int(layoutGrid.CanvasHeight); if canvasH <=0 { canvasH =1 }
	dc := gg.NewContext(canvasW, canvasH)

	canvasBgColorHex := mainTable.Settings.TableBackgroundColor
	if canvasBgColorHex == "" { canvasBgColorHex = "#FFFFFF" }
	if col, errBg := parseHexColor(canvasBgColorHex); errBg == nil {
		dc.SetColor(col); dc.Clear()
	} else {
		log.Printf("Error parsing TableBackgroundColor '%s' for main canvas: %v. Using white.", mainTable.Settings.TableBackgroundColor, errBg)
		dc.SetColor(color.White); dc.Clear()
	}

	err = drawTableItself(dc, mainTable, layoutGrid, allTables, layoutConsts)
	if err != nil {
		return fmt.Errorf("error drawing main table: %w", err)
	}
	return dc.SavePNG(outputPath)
}

func drawTableItself(dc *gg.Context, tableToDraw *table.Table, lg *LayoutGrid, allTables map[string]table.Table, lConsts LayoutConstants) error {
	log.Printf("drawTableItself START: Drawing table ID '%s' on dc (size %dx%d)", tableToDraw.ID, dc.Width(), dc.Height())
	if tableToDraw.Settings.TableBackgroundColor != "" {
		if col, err := parseHexColor(tableToDraw.Settings.TableBackgroundColor); err == nil {
			dc.SetColor(col)
			dc.Clear()
		} else {
			log.Printf("Error parsing TableBackgroundColor '%s' for table '%s': %v. Skipping custom BG.", tableToDraw.Settings.TableBackgroundColor, tableToDraw.ID, err)
		}
	}

	if errFont := dc.LoadFontFace(lConsts.FontPath, lConsts.FontSize); errFont != nil {
		log.Printf("Error loading font '%s' for table '%s': %v. Text rendering may be affected.", lConsts.FontPath, tableToDraw.ID, errFont)
	}

	for _, gridCell := range lg.GridCells {
		cell := gridCell.OriginalCell
		log.Printf("CELL [%d,%d]: Start. Title:'%s', Content:'%.20s', IsRef:%t. Geom: X:%.1f, Y:%.1f, W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, cell.Title, strings.ReplaceAll(cell.Content, "\n", "\\n"), cell.IsTableRef, gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height)

		cellBgColorHex := cell.BackgroundColor
		if cellBgColorHex == "" { cellBgColorHex = tableToDraw.Settings.DefaultCellBackgroundColor }
		if cellBgColorHex == "" { cellBgColorHex = "#FFFFFF" }
		cellBg, errBgParse := parseHexColor(cellBgColorHex)
		if errBgParse != nil {
			log.Printf("Error parsing cell background color '%s': %v. Defaulting to white.", cellBgColorHex, errBgParse)
			cellBg = color.White
		}
		dc.SetColor(cellBg)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Fill()

		edgeColorHex := tableToDraw.Settings.EdgeColor
		if edgeColorHex == "" { edgeColorHex = "#000000" }
		edgeCol, errEdgeParse := parseHexColor(edgeColorHex)
		if errEdgeParse != nil {
			log.Printf("Error parsing edge color '%s': %v. Defaulting to black.", tableToDraw.Settings.EdgeColor, errEdgeParse)
			edgeCol = color.Black
		}
		dc.SetColor(edgeCol)
		edgeThickness := float64(tableToDraw.Settings.EdgeThickness); if edgeThickness <= 0 { edgeThickness = 1.0 }
		dc.SetLineWidth(edgeThickness)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Stroke()

		log.Printf("CELL [%d,%d]: Pre-Push/Clip.", gridCell.GridR, gridCell.GridC)
		// dc.Push()
		// defer dc.Pop() // Pop is also commented out at the end of the loop for this debug
		log.Printf("CELL [%d,%d]: Post-Push, defer dc.Pop() active. (NOTE: dc.Push/Pop/Clip TEMPORARILY DISABLED FOR DEBUGGING)", gridCell.GridR, gridCell.GridC)

		contentAreaX := gridCell.X + lConsts.Padding // Still needed for positioning content
		contentAreaY := gridCell.Y + lConsts.Padding // Still needed for positioning content
		contentAreaWidth := gridCell.Width - (2 * lConsts.Padding) // Still useful for alignment logic
		contentAreaHeight := gridCell.Height - (2 * lConsts.Padding) // Still useful for alignment logic
		if contentAreaWidth < 0 { contentAreaWidth = 0 }
		if contentAreaHeight < 0 { contentAreaHeight = 0 }

		// log.Printf("CELL [%d,%d]: Applying Clip: X:%.1f, Y:%.1f, W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, contentAreaX, contentAreaY, contentAreaWidth, contentAreaHeight)
		// dc.DrawRectangle(contentAreaX, contentAreaY, contentAreaWidth, contentAreaHeight)
		// dc.Clip()

		if cell.IsTableRef {
			log.Printf("CELL [%d,%d]: IsTableRef TRUE. RefID: '%s'", gridCell.GridR, gridCell.GridC, cell.TableRefID)
			if cell.TableRefID == "" { log.Printf("CELL [%d,%d]: Warning: TableRefID is empty. Skipping.", gridCell.GridR, gridCell.GridC); continue }

			refTable, ok := allTables[cell.TableRefID]
			if !ok { log.Printf("CELL [%d,%d]: Warning: Referenced table ID '%s' not found. Skipping.", gridCell.GridR, gridCell.GridC, cell.TableRefID); continue }
			log.Printf("CELL [%d,%d]: InnerTable: ID '%s'. ParentContentArea W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, refTable.ID, contentAreaWidth, contentAreaHeight)

			innerLg, mapErr := PopulateOccupationMap(&refTable)
			if mapErr != nil { log.Printf("CELL [%d,%d]: Error populating inner map for '%s': %v. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, mapErr); continue }
			if innerLg.NumLogicalRows == 0 || innerLg.NumLogicalCols == 0 { log.Printf("CELL [%d,%d]: Info: Inner table '%s' is empty. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID); continue }

			calcErr := innerLg.CalculateColumnWidthsAndRowHeights(lConsts, allTables)
			if calcErr != nil { log.Printf("CELL [%d,%d]: Error calculating inner layout for '%s': %v. Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, calcErr); continue }
			innerLg.CalculateFinalCellLayouts(0)

			innerDcWidth := int(innerLg.CanvasWidth); innerDcHeight := int(innerLg.CanvasHeight)
			log.Printf("CELL [%d,%d]: InnerTable: Natural canvas size W:%d, H:%d for subDc.", gridCell.GridR, gridCell.GridC, innerDcWidth, innerDcHeight)
			if innerDcWidth <= 0 || innerDcHeight <= 0 { log.Printf("CELL [%d,%d]: Warning: Inner table '%s' zero/neg dims (W:%d, H:%d). Skipping.", gridCell.GridR, gridCell.GridC, refTable.ID, innerDcWidth, innerDcHeight); continue }

			subDc := gg.NewContext(innerDcWidth, innerDcHeight)
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
					if naturalInnerWidth > 0 && contentAreaWidth > 0 {
						scaleFactor := contentAreaWidth / naturalInnerWidth
						scaledW = contentAreaWidth; scaledH = naturalInnerHeight * scaleFactor
						performScale = true
					}
				case "fit_height":
					if naturalInnerHeight > 0 && contentAreaHeight > 0 {
						scaleFactor := contentAreaHeight / naturalInnerHeight
						scaledH = contentAreaHeight; scaledW = naturalInnerWidth * scaleFactor
						performScale = true
					}
				case "fit_both":
					if naturalInnerWidth > 0 && naturalInnerHeight > 0 && contentAreaWidth > 0 && contentAreaHeight > 0 {
						scaleFactorW := contentAreaWidth / naturalInnerWidth
						scaleFactorH := contentAreaHeight / naturalInnerHeight
						scaleFactor := math.Min(scaleFactorW, scaleFactorH)
						if scaleFactor > 0 {
							scaledW = naturalInnerWidth * scaleFactor; scaledH = naturalInnerHeight * scaleFactor
							performScale = true
						}
					}
				case "fill_stretch":
					if contentAreaWidth > 0 && contentAreaHeight > 0 {
						scaledW = contentAreaWidth; scaledH = contentAreaHeight
						if math.Abs(scaledW-naturalInnerWidth) > epsilon || math.Abs(scaledH-naturalInnerHeight) > epsilon { performScale = true }
					}
				}

				if scaledW <= 0 || scaledH <= 0 { imageToDraw = nil }

				if performScale && imageToDraw != nil && (math.Abs(scaledW-naturalInnerWidth) > epsilon || math.Abs(scaledH-naturalInnerHeight) > epsilon) {
					roundedScaledW, roundedScaledH := int(math.Round(scaledW)), int(math.Round(scaledH))
					if roundedScaledW > 0 && roundedScaledH > 0 {
						scaledSubDc := gg.NewContext(roundedScaledW, roundedScaledH)
						scaledSubDc.Scale(float64(roundedScaledW)/naturalInnerWidth, float64(roundedScaledH)/naturalInnerHeight)
						scaledSubDc.DrawImage(naturalInnerTableImage, 0, 0)
						imageToDraw = scaledSubDc.Image()
						scaledW, scaledH = float64(imageToDraw.Bounds().Dx()), float64(imageToDraw.Bounds().Dy())
					} else { imageToDraw = nil }
				}
				log.Printf("CELL [%d,%d]: InnerTable: ScaledDims W:%.1f, H:%.1f. AlignMode:'%s', ScaleMode:'%s'", gridCell.GridR, gridCell.GridC, scaledW, scaledH, cell.InnerTableAlignment, cell.InnerTableScaleMode)

				offsetX, offsetY := 0.0, 0.0
				if imageToDraw != nil && scaledW > 0 && scaledH > 0 {
					switch cell.InnerTableAlignment {
					case "top_center": offsetX = (contentAreaWidth - scaledW) / 2
					case "top_right": offsetX = contentAreaWidth - scaledW
					case "middle_left": offsetY = (contentAreaHeight - scaledH) / 2
					case "center", "middle_center": offsetX = (contentAreaWidth - scaledW) / 2; offsetY = (contentAreaHeight - scaledH) / 2
					case "middle_right": offsetX = contentAreaWidth - scaledW; offsetY = (contentAreaHeight - scaledH) / 2
					case "bottom_left": offsetY = contentAreaHeight - scaledH
					case "bottom_center": offsetX = (contentAreaWidth - scaledW) / 2; offsetY = contentAreaHeight - scaledH
					case "bottom_right": offsetX = contentAreaWidth - scaledW; offsetY = contentAreaHeight - scaledH
					}
				}

				drawX, drawY := contentAreaX+offsetX, contentAreaY+offsetY
				if imageToDraw != nil && scaledW > 0 && scaledH > 0 {
					log.Printf("CELL [%d,%d]: InnerTable: Drawing image at X:%.1f, Y:%.1f", gridCell.GridR, gridCell.GridC, math.Round(drawX), math.Round(drawY))
					dc.DrawImage(imageToDraw, int(math.Round(drawX)), int(math.Round(drawY)))

					log.Printf("CELL [%d,%d]: InnerTable: Drawing special border at X:%.1f, Y:%.1f, W:%.1f, H:%.1f", gridCell.GridR, gridCell.GridC, math.Round(drawX), math.Round(drawY), math.Round(scaledW), math.Round(scaledH))
					borderColHex := tableToDraw.Settings.EdgeColor; if borderColHex == "" { borderColHex = "#000000" }
					parsedBorderCol, errBr := parseHexColor(borderColHex)
					if errBr != nil { parsedBorderCol = color.Black }
					dc.SetColor(parsedBorderCol); dc.SetLineWidth(1.0); dc.SetDash()
					dc.DrawRectangle(math.Round(drawX), math.Round(drawY), math.Round(scaledW), math.Round(scaledH)); dc.Stroke()
				}
			} else if naturalInnerWidth == 0 || naturalInnerHeight == 0 {
                 log.Printf("CELL [%d,%d]: Info: Inner table '%s' for cell '%s' has zero natural dimensions. Nothing to draw.", gridCell.GridR, gridCell.GridC, refTable.ID, cell.Title)
            }
		} else {
			textColor := color.Black; dc.SetColor(textColor)
			textStartX := contentAreaX
			textAvailableWidth := contentAreaWidth
			lineHeight := lConsts.FontSize * lConsts.LineHeightMultiplier
			firstLineBaselineOffsetY := lConsts.FontSize
			currentBaselineY := contentAreaY + firstLineBaselineOffsetY
			titleProcessed := false
			if cell.Title != "" {
				log.Printf("CELL [%d,%d]: Text: Drawing title. Initial baselineY: %.1f", gridCell.GridR, gridCell.GridC, currentBaselineY)
				titleLines := dc.WordWrap("["+cell.Title+"]", textAvailableWidth)
				for _, line := range titleLines {
					if currentBaselineY < contentAreaY+contentAreaHeight+epsilon {
						dc.DrawString(line, textStartX, currentBaselineY); currentBaselineY += lineHeight; titleProcessed = true
					} else { break }
				}
			}
			if cell.Content != "" {
				if titleProcessed && (currentBaselineY <= contentAreaY+contentAreaHeight+epsilon) { currentBaselineY += lineHeight * 0.25 }
				log.Printf("CELL [%d,%d]: Text: Drawing content. Initial baselineY: %.1f", gridCell.GridR, gridCell.GridC, currentBaselineY)
				contentLines := dc.WordWrap(cell.Content, textAvailableWidth)
				for _, line := range contentLines {
					if currentBaselineY < contentAreaY+contentAreaHeight+epsilon {
						dc.DrawString(line, textStartX, currentBaselineY); currentBaselineY += lineHeight
					} else { break }
				}
			}
		}
	} // Explicit dc.Pop() was already removed, defer is now commented.
	return nil
}
