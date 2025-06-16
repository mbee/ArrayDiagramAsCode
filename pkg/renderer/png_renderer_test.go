package renderer

import (
	"diagramgen/pkg/parser" // Added for ParseAllText
	"diagramgen/pkg/table"
	"os"
	"path/filepath" // Added for filepath.Join
	"reflect"       // Added for reflect.ValueOf
	"testing"
)

func TestRenderToPNG_FileCreation(t *testing.T) {
	testTable := table.Table{
		Title: "Test PNG Creation",
		Rows: []table.Row{
			{Cells: []table.Cell{table.NewCell("R1C1", "Cell 1"), table.NewCell("R1C2", "Cell 2")}},
		},
		Settings: table.DefaultGlobalSettings(), // Ensure settings are initialized
	}
	outputPath := "test_output_creation.png"
	// Defer removal at the start of the function to ensure cleanup even if test fails early.
	defer func() {
		err := os.Remove(outputPath)
		if err != nil && !os.IsNotExist(err) { // Don't error if file already removed or never created
			t.Logf("Warning: failed to remove test file %s: %v", outputPath, err)
		}
	}()

	// For existing tests, pass an empty map for allTables
	err := RenderToPNG(&testTable, make(map[string]table.Table), outputPath)
	if err != nil {
		t.Fatalf("RenderToPNG failed: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("RenderToPNG did not create the output file '%s'", outputPath)
	}
}

func TestRenderToPNG_EmptyTable(t *testing.T) {
	emptyTable := table.Table{
		Title:    "Empty Table Test", // Give it a title to make image slightly more interesting if needed
		Rows:     []table.Row{},
		Settings: table.DefaultGlobalSettings(),
	}
	outputPath := "test_output_empty.png"
	defer func() {
		err := os.Remove(outputPath)
		if err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove test file %s: %v", outputPath, err)
		}
	}()

	err := RenderToPNG(&emptyTable, make(map[string]table.Table), outputPath)
	if err != nil {
		t.Fatalf("RenderToPNG failed for empty table: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("RenderToPNG did not create the output file '%s' for empty table", outputPath)
	}
	// Optional: Check file size. An empty table should still produce a small PNG (e.g., just margins and table background).
	// info, err := os.Stat(outputPath)
	// if err == nil {
	// 	if info.Size() == 0 { // Or some other threshold
	// 		t.Errorf("RenderToPNG created an empty (0 byte) file for an empty table.")
	// 	}
	// }
}

func TestRenderToPNG_InvalidColorHandling(t *testing.T) {
	tableWithInvalidColor := table.Table{
		Title: "Invalid Color Test",
		Rows: []table.Row{
			{Cells: []table.Cell{
				func() table.Cell {
					c := table.NewCell("C1", "Content")
					c.BackgroundColor = "#INVALIDCOLORSTRING" // Clearly invalid hex
					return c
				}(),
			}},
		},
		Settings: table.DefaultGlobalSettings(),
	}
	// Also test invalid global color
	tableWithInvalidColor.Settings.EdgeColor = "NotAColor"

	outputPath := "test_output_invalid_color.png"
	defer func() {
		err := os.Remove(outputPath)
		if err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove test file %s: %v", outputPath, err)
		}
	}()

	// Current png_renderer logs errors for invalid colors and uses defaults.
	// It should not return an error to RenderToPNG for color parsing issues.
	// The manual parseHexColor implemented returns an error, but RenderToPNG logs it and proceeds.
	err := RenderToPNG(&tableWithInvalidColor, make(map[string]table.Table), outputPath)
	if err != nil {
		// This test assumes that color parsing errors are handled gracefully by the renderer
		// (i.e., logged and defaults used) rather than propagated as fatal errors from RenderToPNG.
		t.Fatalf("RenderToPNG failed: %v. Expected graceful handling of invalid colors.", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("RenderToPNG did not create the output file '%s' when handling invalid color", outputPath)
	}
}

const testInputWithInnerTable = `
table: [inner_sample] Sample Inner Table {bg_table:#FFFFE0}
Detail | Value
Status | OK
Count | 12

table: [main_sample] Main Table with Inner {bg_table:#E0FFFF}
Description | Data | Notes
Item A | Some regular text | Just a note
Item B | ::table=inner_sample:: | Inner table here
Item C | More text | End of section
`

func TestRenderToPNG_WithInnerTable(t *testing.T) {
	// Import parser from the correct module path.
	// This assumes the test is in the 'renderer' package, and 'diagramgen/pkg/parser' is the path.
	// Adjust if your module structure is different.
	// For this test, direct import path is used.
	parserInstance := struct{ ParseAllText func(string) (table.AllTables, error) }{ParseAllText: parser.ParseAllText}


	allTablesData, parseErr := parserInstance.ParseAllText(testInputWithInnerTable)
	if parseErr != nil {
		t.Fatalf("ParseAllText failed: %v", parseErr)
	}

	if allTablesData.MainTableID == "" {
		t.Fatalf("MainTableID is empty after parsing. Input was:\n%s", testInputWithInnerTable)
	}

	mainTable, ok := allTablesData.Tables[allTablesData.MainTableID]
	if !ok {
		t.Fatalf("Main table with ID '%s' not found in parsed tables. MainTableID: %s, Available IDs: %v",
			allTablesData.MainTableID, allTablesData.MainTableID, reflect.ValueOf(allTablesData.Tables).MapKeys())
	}


	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "test_output_with_inner.png")

	renderErr := RenderToPNG(&mainTable, allTablesData.Tables, outputFilePath)
	if renderErr != nil {
		t.Errorf("RenderToPNG failed: %v", renderErr)
	}

	fileInfo, statErr := os.Stat(outputFilePath)
	if os.IsNotExist(statErr) {
		t.Errorf("Output PNG file was not created: %s", outputFilePath)
		return // Stop further checks if file doesn't exist
	} else if statErr != nil {
		t.Errorf("Error checking output file stats: %v", statErr)
		return
	}

	if fileInfo.Size() == 0 {
		t.Errorf("Output PNG file is empty (0 bytes).")
	}
	// For debugging, print the path: t.Logf("Output PNG created at: %s", outputFilePath)
}
