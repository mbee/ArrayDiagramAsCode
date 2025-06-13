package renderer

import (
	"diagramgen/pkg/table"
	"fmt"
	"image/color" // For parsing hex colors
	"log"         // For basic logging
	"strings"     // For strings.TrimPrefix

	"github.com/fogleman/gg"
)

const (
	defaultPadding      = 10.0
	defaultMargin       = 20.0
	defaultFontPath     = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf" // Common path
	defaultFontSize     = 12.0
	defaultLineHeight   = 1.6 // Multiplier for font size to get line spacing
	defaultCornerRadius = 8.0
)

// RenderToPNG generates a PNG image of the table.
func RenderToPNG(t table.Table, outputPath string) error {
	// --- Stage 1: Pre-computation (Layout) ---
	// For now, let's use fixed cell sizes for simplicity.
	// This will be replaced by a proper layout engine later.
	numRows := len(t.Rows)
	numCols := 0
	if numRows > 0 && len(t.Rows[0].Cells) > 0 {
		// Basic assumption: table is rectangular based on the first row.
		// A proper layout would sum colspans to find max visual columns.
		tempNumCols := 0
		for _, cell := range t.Rows[0].Cells {
			tempNumCols += cell.Colspan // Sum colspans for a more accurate width estimate
		}
		numCols = tempNumCols
	}

	if numRows == 0 || numCols == 0 {
		// Create a small empty image if table is empty or has no columns
		dc := gg.NewContext(int(defaultMargin*2), int(defaultMargin*2))
		if t.Settings.TableBackgroundColor != "" {
			if col, err := parseHexColor(t.Settings.TableBackgroundColor); err == nil {
				dc.SetColor(col)
				dc.Clear()
			} else {
				dc.SetRGB(1, 1, 1) // Default to white
				dc.Clear()
			}
		} else {
			dc.SetRGB(1, 1, 1) // Default white background
			dc.Clear()
		}
		log.Println("Table has 0 rows or 0 effective columns, saving empty image.")
		return dc.SavePNG(outputPath)
	}

	// Fixed cell size for initial implementation
	// These are for a single span cell (colspan=1, rowspan=1)
	fixedCellWidth := 150.0
	fixedCellHeight := 50.0

	// Estimate table dimensions based on fixed sizes and number of rows/cols
	// This doesn't account for actual content fitting yet.
	estimatedTableWidth := float64(numCols) * fixedCellWidth
	estimatedTableHeight := float64(numRows) * fixedCellHeight // Max height based on rows

	canvasWidth := estimatedTableWidth + 2*defaultMargin
	canvasHeight := estimatedTableHeight + 2*defaultMargin

	dc := gg.NewContext(int(canvasWidth), int(canvasHeight))

	// Set table background
	tableBgColHex := t.Settings.TableBackgroundColor
	if tableBgColHex != "" {
		if col, err := parseHexColor(tableBgColHex); err == nil {
			dc.SetColor(col)
		} else {
			log.Printf("Error parsing TableBackgroundColor '%s': %v. Defaulting to transparent/white.", tableBgColHex, err)
			dc.SetColor(color.Transparent) // Or white: color.RGBA{R: 255, G: 255, B: 255, A: 255}
		}
	} else {
		// If no table background is set, default to fully transparent so if drawn on something else, it shows through.
		// Or, if a visible default is preferred (e.g. white for standalone images):
		dc.SetColor(color.White) // Defaulting to white if not specified
	}
	dc.Clear() // Clear with the chosen background color

	// Load font
	if err := dc.LoadFontFace(defaultFontPath, defaultFontSize); err != nil {
		log.Printf("Error loading font '%s': %v. Text might not render correctly.", defaultFontPath, err)
		// gg might use a system default. If not, text drawing calls will likely fail or be no-ops.
	}

	// --- Stage 2: Drawing ---
	// This simple loop doesn't handle true colspan/rowspan layout yet.
	// It draws each cell based on its (r,c) index and its own colspan/rowspan for sizing,
	// but doesn't skip cells that would be covered by a previous span.
	for r, row := range t.Rows {
		currentX := defaultMargin
		for _, cell := range row.Cells {
			cellX := currentX
			cellY := defaultMargin + float64(r)*fixedCellHeight // Simplified Y position

			// Adjust cell width/height based on colspan/rowspan
			currentCellWidth := float64(cell.Colspan) * fixedCellWidth
			currentCellHeight := float64(cell.Rowspan) * fixedCellHeight // This is a simplification

			// Determine cell background color
			cellBgColorHex := cell.BackgroundColor
			if cellBgColorHex == "" {
				cellBgColorHex = t.Settings.DefaultCellBackgroundColor
			}

			var cellBg color.Color = color.White // Default to white if parsing fails or string is empty initially
			if cellBgColorHex != "" { // Attempt parse only if color string is not empty
				var err error
				cellBg, err = parseHexColor(cellBgColorHex)
				if err != nil {
					log.Printf("Error parsing cell background color '%s': %v. Defaulting to white.", cellBgColorHex, err)
					cellBg = color.White // Ensure it's white on error
				}
			}

			// Draw cell background
			dc.SetColor(cellBg)
			dc.DrawRoundedRectangle(cellX, cellY, currentCellWidth, currentCellHeight, defaultCornerRadius)
			dc.Fill()

			// Draw cell border
			edgeCol, err := parseHexColor(t.Settings.EdgeColor)
			if err != nil {
				log.Printf("Error parsing edge color '%s': %v. Defaulting to black.", t.Settings.EdgeColor, err)
				edgeCol = color.Black
			}
			dc.SetColor(edgeCol)
			dc.SetLineWidth(float64(t.Settings.EdgeThickness))
			dc.DrawRoundedRectangle(cellX, cellY, currentCellWidth, currentCellHeight, defaultCornerRadius)
			dc.Stroke()

			// Prepare for text rendering
			textColor := color.Black // Default text color, could be made configurable
			dc.SetColor(textColor)

			textRenderX := cellX + defaultPadding
			textRenderY := cellY + defaultPadding + defaultFontSize // Baseline for first line of text

			// Render Title
			if cell.Title != "" {
				// TODO: Could use a different font style/size for title
				titleLines := dc.WordWrap("["+cell.Title+"]", currentCellWidth-2*defaultPadding)
				for _, line := range titleLines {
					if textRenderY < cellY+currentCellHeight-defaultPadding { // Check fit
						dc.DrawStringAnchored(line, textRenderX, textRenderY, 0, 0) // Anchor 0,0 is top-left
						textRenderY += defaultFontSize * defaultLineHeight
					} else {
						break
					}
				}
			}

			// Render Content
			if cell.Content != "" {
				contentLines := dc.WordWrap(cell.Content, currentCellWidth-2*defaultPadding)
				for _, line := range contentLines {
					if textRenderY < cellY+currentCellHeight-defaultPadding { // Check fit
						dc.DrawStringAnchored(line, textRenderX, textRenderY, 0, 0)
						textRenderY += defaultFontSize * defaultLineHeight
					} else {
						break
					}
				}
			}
			currentX += currentCellWidth // Move X for the next cell in the row
		}
	}

	return dc.SavePNG(outputPath)
}

// parseHexColor converts a hex color string (e.g., "#RRGGBB" or "#RGB") to a color.Color.
// parseHexColor converts a hex color string (e.g., "#RRGGBB" or "#RGB") to a color.Color.
func parseHexColor(s string) (color.Color, error) {
	if s == "" {
		return color.Transparent, fmt.Errorf("empty color string is not a valid color")
	}

	s = strings.TrimPrefix(s, "#")
	var r, g, b uint8

	if len(s) == 3 { // #RGB format
		_, err := fmt.Sscanf(s, "%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return color.Transparent, fmt.Errorf("error parsing short hex color %s: %w", s, err)
		}
		// Scale a single hex digit (0-F) to two (00-FF)
		r *= 17 // 0xF * 17 = 0xFF
		g *= 17
		b *= 17
	} else if len(s) == 6 { // #RRGGBB format
		_, err := fmt.Sscanf(s, "%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return color.Transparent, fmt.Errorf("error parsing hex color %s: %w", s, err)
		}
	} else {
		return color.Transparent, fmt.Errorf("invalid hex color string format: %s", s)
	}
	return color.RGBA{R: r, G: g, B: b, A: 255}, nil
}
