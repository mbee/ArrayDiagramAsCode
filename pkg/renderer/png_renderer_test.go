package renderer

import (
	"diagramgen/pkg/parser"
	"diagramgen/pkg/table"
	"os"
	"path/filepath"
	// "reflect" // No longer needed after changes to FullExampleFile test
	"testing"
)

func TestRenderToPNG_FileCreation(t *testing.T) {
	testTable := table.Table{
		Title: "Test PNG Creation",
		Rows: []table.Row{
			{Cells: []table.Cell{table.NewCell("R1C1", "Cell 1"), table.NewCell("R1C2", "Cell 2")}},
		},
		Settings: table.DefaultGlobalSettings(),
	}
	outputPath := "test_output_creation.png"
	defer func() {
		err := os.Remove(outputPath)
		if err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove test file %s: %v", outputPath, err)
		}
	}()

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
		Title:    "Empty Table Test",
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
}

func TestRenderToPNG_InvalidColorHandling(t *testing.T) {
	tableWithInvalidColor := table.Table{
		Title: "Invalid Color Test",
		Rows: []table.Row{
			{Cells: []table.Cell{
				func() table.Cell {
					c := table.NewCell("C1", "Content")
					c.BackgroundColor = "#INVALIDCOLORSTRING"
					return c
				}(),
			}},
		},
		Settings: table.DefaultGlobalSettings(),
	}
	tableWithInvalidColor.Settings.EdgeColor = "NotAColor"

	outputPath := "test_output_invalid_color.png"
	defer func() {
		err := os.Remove(outputPath)
		if err != nil && !os.IsNotExist(err) {
			t.Logf("Warning: failed to remove test file %s: %v", outputPath, err)
		}
	}()

	err := RenderToPNG(&tableWithInvalidColor, make(map[string]table.Table), outputPath)
	if err != nil {
		t.Fatalf("RenderToPNG failed: %v. Expected graceful handling of invalid colors.", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("RenderToPNG did not create the output file '%s' when handling invalid color", outputPath)
	}
}

// TestRenderToPNG_FullExampleFile reads example.txt and renders it.
func TestRenderToPNG_FullExampleFile(t *testing.T) {
	content, err := os.ReadFile("../../example.txt") // Assumes test run from pkg/renderer directory
	if err != nil {
		t.Fatalf("Failed to read example.txt: %v", err)
	}
	inputStr := string(content)

	allTablesData, parseErr := parser.ParseAllText(inputStr)
	if parseErr != nil {
		t.Fatalf("ParseAllText failed for example.txt: %v\nInput:\n%s", parseErr, inputStr)
	}

	if allTablesData.MainTableID == "" {
		t.Fatalf("MainTableID is empty after parsing example.txt. Input was:\n%s", inputStr)
	}

	mainTable, ok := allTablesData.Tables[allTablesData.MainTableID]
	if !ok {
		// Log available keys for easier debugging if main table ID is wrong
		availableIDs := make([]string, 0, len(allTablesData.Tables))
		for k := range allTablesData.Tables {
			availableIDs = append(availableIDs, k)
		}
		t.Fatalf("Main table with ID '%s' (from MainTableID) not found in parsed tables from example.txt. Available IDs: %v",
			allTablesData.MainTableID, availableIDs)
	}

	tempDir := t.TempDir()
	outputFilePath := filepath.Join(tempDir, "test_example_output.png")

	tablesToRender := allTablesData.Tables
	if tablesToRender == nil { // Should not happen if parsing succeeded and MainTableID is set
		tablesToRender = make(map[string]table.Table)
	}

	renderErr := RenderToPNG(&mainTable, tablesToRender, outputFilePath)
	if renderErr != nil {
		t.Errorf("RenderToPNG failed for example.txt: %v", renderErr)
	}

	fileInfo, statErr := os.Stat(outputFilePath)
	if os.IsNotExist(statErr) {
		t.Errorf("Output PNG file was not created: %s", outputFilePath)
		return
	} else if statErr != nil {
		t.Errorf("Error checking output file stats: %v", statErr)
		return
	}

	if fileInfo.Size() == 0 {
		t.Errorf("Output PNG file is empty (0 bytes).")
	}
	// For debugging, print the path: t.Logf("Output PNG created at: %s", outputFilePath)
}
