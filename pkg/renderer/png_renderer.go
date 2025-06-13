package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	"image/color"
	"log"
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
func RenderToPNG(t *table.Table, outputPath string) error {
	if t == nil {
		return fmt.Errorf("input table is nil")
	}

	// --- 1. Layout Calculation Phase ---
	layoutGrid, err := PopulateOccupationMap(t) // Assumes t is not nil
	if err != nil {
		return fmt.Errorf("failed to populate occupation map: %w", err)
	}

	if layoutGrid.NumLogicalRows == 0 || layoutGrid.NumLogicalCols == 0 {
		dcWidth := int(defaultMargin * 2)
		if dcWidth < 1 { dcWidth = 1 }
		dcHeight := int(defaultMargin * 2)
		if dcHeight < 1 { dcHeight = 1 }

		dc := gg.NewContext(dcWidth, dcHeight)
		tableBG := t.Settings.TableBackgroundColor
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
	layoutConsts := LayoutConstants{
		FontPath:             defaultFontPath,
		FontSize:             defaultFontSize,
		LineHeightMultiplier: defaultLineHeightMultiplier,
		Padding:              defaultPadding,
		MinCellWidth:         defaultMinCellWidth,
		MinCellHeight:        defaultMinCellHeight,
	}

	err = layoutGrid.CalculateColumnWidthsAndRowHeights(layoutConsts)
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
	// Set table background
	tableBgColorHex := t.Settings.TableBackgroundColor
	if tableBgColorHex == "" { tableBgColorHex = "#FFFFFF" } // Default to white if not specified

	if col, errBg := parseHexColor(tableBgColorHex); errBg == nil {
		dc.SetColor(col)
		dc.Clear()
	} else {
		log.Printf("Error parsing TableBackgroundColor '%s': %v. Using white.", t.Settings.TableBackgroundColor, errBg)
		dc.SetColor(color.White); dc.Clear()
	}

	if errFont := dc.LoadFontFace(layoutConsts.FontPath, layoutConsts.FontSize); errFont != nil {
		log.Printf("Error loading font '%s': %v. Text rendering may be affected.", layoutConsts.FontPath, errFont)
	}

	for _, gridCell := range layoutGrid.GridCells {
		cell := gridCell.OriginalCell

		cellBgColorHex := cell.BackgroundColor
		if cellBgColorHex == "" { // If cell has no specific color, use table's default cell BG
			cellBgColorHex = t.Settings.DefaultCellBackgroundColor
		}
		if cellBgColorHex == "" { // If table's default cell BG is also empty, fallback
		    cellBgColorHex = "#FFFFFF" // Ultimate fallback to white
		}

		cellBg, errBgParse := parseHexColor(cellBgColorHex)
		if errBgParse != nil {
			log.Printf("Error parsing cell background color '%s' for cell '%s': %v. Defaulting to white.", cellBgColorHex, cell.Title, errBgParse)
			cellBg = color.White
		}

		dc.SetColor(cellBg)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Fill()

		edgeColorHex := t.Settings.EdgeColor
		if edgeColorHex == "" { edgeColorHex = "#000000" } // Default edge to black

		edgeCol, errEdgeParse := parseHexColor(edgeColorHex)
		if errEdgeParse != nil {
			log.Printf("Error parsing edge color '%s': %v. Defaulting to black.", t.Settings.EdgeColor, errEdgeParse)
			edgeCol = color.Black
		}
		dc.SetColor(edgeCol)
		// Ensure edge thickness is at least 1 if not specified or zero from settings
		edgeThickness := float64(t.Settings.EdgeThickness)
		if edgeThickness <= 0 { edgeThickness = 1.0 }
		dc.SetLineWidth(edgeThickness)
		dc.DrawRoundedRectangle(gridCell.X, gridCell.Y, gridCell.Width, gridCell.Height, defaultCornerRadius)
		dc.Stroke()

		textColor := color.Black
		dc.SetColor(textColor)

		textStartX := gridCell.X + layoutConsts.Padding
		currentTextY := gridCell.Y + layoutConsts.Padding

        textAvailableWidth := gridCell.Width - (2 * layoutConsts.Padding)
        if textAvailableWidth < 0 { textAvailableWidth = 0 }

		lineVisualHeight := layoutConsts.FontSize * layoutConsts.LineHeightMultiplier

		if cell.Title != "" {
			titleText := "[" + cell.Title + "]"
			titleLines := dc.WordWrap(titleText, textAvailableWidth)
			for _, line := range titleLines {
                if currentTextY + lineVisualHeight <= gridCell.Y + gridCell.Height - layoutConsts.Padding + epsilon {
				    dc.DrawString(line, textStartX, currentTextY)
                    currentTextY += lineVisualHeight
                } else { break }
			}
		}

		if cell.Content != "" {
			if cell.Title != "" && (currentTextY > gridCell.Y + layoutConsts.Padding + epsilon) {
				currentTextY += lineVisualHeight * 0.25
			}
			contentLines := dc.WordWrap(cell.Content, textAvailableWidth)
			for _, line := range contentLines {
                if currentTextY + lineVisualHeight <= gridCell.Y + gridCell.Height - layoutConsts.Padding + epsilon {
				    dc.DrawString(line, textStartX, currentTextY)
				    currentTextY += lineVisualHeight
                } else { break }
			}
		}
	}

	return dc.SavePNG(outputPath)
}
